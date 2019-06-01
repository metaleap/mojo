package atmosess

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/go-leap/fs"
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
	"github.com/metaleap/atmo/lang/irfun"
)

// Kit is a pile of atmo source files residing in the same directory and
// being interpreted or compiled all together as a unit.
type Kit struct {
	DirPath           string
	ImpPath           string
	WasEverToBeLoaded bool
	Imports           []string

	topLevelDefs atmolang_irfun.AstTopDefs
	defsFacts    map[string]*defNameFacts
	SrcFiles     atmolang.AstFiles
	state        struct {
		defsGoneIdsNames map[string]string
		defsBornIdsNames map[string]string
	}
	lookups struct {
		tlDefsByID      map[string]*atmolang_irfun.AstDefTop
		tlDefIDsByName  map[string][]string
		namesInScopeOwn atmolang_irfun.AnnNamesInScope
		namesInScopeExt atmolang_irfun.AnnNamesInScope
		namesInScopeAll atmolang_irfun.AnnNamesInScope
	}
	Errs struct {
		Stage0DirAccessDuringRefresh error
		Stage0BadImports             []error
		Stage1BadNames               atmo.Errors
	}
}

func (me *Ctx) kitEnsureLoaded(kit *Kit) {
	me.kitRefreshFilesAndMaybeReload(kit, !me.state.fileModsWatch.collectingFileModsAutomaticallyPeriodically, !kit.WasEverToBeLoaded)
}

func (me *Ctx) KitEnsureLoaded(kit *Kit) {
	me.kitEnsureLoaded(kit)
	me.reprocessAffectedDefsIfAnyKitsReloaded()
}

func (me *Ctx) KitsEnsureLoaded(plusSessDirFauxKits bool, kitImpPaths ...string) {
	me.maybeInitPanic(false)
	if plusSessDirFauxKits {
		for _, dirsess := range me.Dirs.fauxKits {
			if idx := me.Kits.All.IndexDirPath(dirsess); idx >= 0 {
				kitImpPaths = append(kitImpPaths, me.Kits.All[idx].ImpPath)
			}
		}
	}
	if len(kitImpPaths) > 0 {
		for _, kip := range kitImpPaths {
			if kit := me.Kits.All.ByImpPath(kip); kit != nil {
				me.kitRefreshFilesAndMaybeReload(kit, !me.state.fileModsWatch.collectingFileModsAutomaticallyPeriodically, true)
			}
		}
		me.reprocessAffectedDefsIfAnyKitsReloaded()
	}
}

func (me *Ctx) KitDefFacts(kit *Kit, def *atmolang_irfun.AstDefTop) ValFacts {
	return ValFacts{valFacts: me.substantiateKitTopLevelDefFacts(kit, def.Id, false).valFacts}
}

func (me *Ctx) KitByDirPath(dirPath string, tryToAddToFauxKits bool) (kit *Kit) {
	if kit = me.Kits.All.ByDirPath(dirPath); kit == nil && tryToAddToFauxKits {
		if ok, _ := me.FauxKitsAdd(dirPath); ok {
			kit = me.Kits.All.ByDirPath(dirPath)
		}
	}
	return
}

func (me *Ctx) KitByImpPath(impPath string) *Kit {
	me.maybeInitPanic(false)
	idx := me.Kits.All.IndexImpPath(impPath)
	if idx < 0 && (impPath == "" || impPath == "." || impPath == "·") {
		if fauxkitdirs := me.Dirs.fauxKits; len(fauxkitdirs) > 0 {
			idx = me.Kits.All.IndexDirPath(fauxkitdirs[0])
		}
	}
	if idx >= 0 {
		return me.Kits.All[idx]
	}
	return nil
}

func (me *Ctx) kitRefreshFilesAndMaybeReload(kit *Kit, forceFilesCheck bool, forceReload bool) {
	var fresherrs []error
	if forceFilesCheck {
		var diritems []os.FileInfo
		if diritems, kit.Errs.Stage0DirAccessDuringRefresh = ufs.Dir(kit.DirPath); kit.Errs.Stage0DirAccessDuringRefresh != nil {
			kit.SrcFiles, kit.topLevelDefs, fresherrs = nil, nil, append(fresherrs, kit.Errs.Stage0DirAccessDuringRefresh)
			goto end
		}

		// any deleted files get forgotten now
		for i := 0; i < len(kit.SrcFiles); i++ {
			if kit.SrcFiles[i].SrcFilePath != "" && !ufs.IsFile(kit.SrcFiles[i].SrcFilePath) {
				kit.SrcFiles.RemoveAt(i)
				i--
			}
		}

		// any new files get added
		for _, file := range diritems {
			if (!file.IsDir()) && ustr.Suff(file.Name(), atmo.SrcFileExt) {
				if fp := filepath.Join(kit.DirPath, file.Name()); kit.SrcFiles.Index(fp) < 0 {
					kit.SrcFiles = append(kit.SrcFiles, &atmolang.AstFile{SrcFilePath: fp})
				}
			}
		}
		atmo.SortMaybe(kit.SrcFiles)
	}
	if kit.WasEverToBeLoaded || forceReload || me.Kits.AlwaysEnsureLoadedAsSoonAsDiscovered {
		kit.WasEverToBeLoaded, kit.Errs.Stage0BadImports, kit.lookups.tlDefIDsByName, kit.lookups.tlDefsByID =
			true, nil, nil, nil

		for _, sf := range kit.SrcFiles {
			fresherrs = append(fresherrs, sf.LexAndParseFile(true, false)...)
		}

		for _, imp := range kit.Imports {
			if kimp := me.Kits.All.ByImpPath(imp); kimp == nil {
				kit.Errs.Stage0BadImports = append(kit.Errs.Stage0BadImports, errors.New("import not found: `"+imp+"`"))
			} else {
				me.kitEnsureLoaded(kimp)
			}
		}
		if len(kit.Errs.Stage0BadImports) > 0 {
			fresherrs = append(fresherrs, kit.Errs.Stage0BadImports...)
		}

		{
			od, nd, fe := kit.topLevelDefs.ReInitFrom(kit.SrcFiles)
			kit.state.defsGoneIdsNames, kit.state.defsBornIdsNames, fresherrs = od, nd, append(fresherrs, fe...)
			if len(od) > 0 || len(nd) > 0 || len(fe) > 0 {
				me.state.kitsReprocessing.needed = true
			}
		}
		kit.lookups.tlDefIDsByName, kit.lookups.tlDefsByID = make(map[string][]string, len(kit.topLevelDefs)), make(map[string]*atmolang_irfun.AstDefTop, len(kit.topLevelDefs))
		for _, tldef := range kit.topLevelDefs {
			kit.lookups.tlDefsByID[tldef.Id] = tldef
			if n, ok := kit.lookups.tlDefIDsByName[tldef.Name.Val]; !ok {
				kit.lookups.tlDefIDsByName[tldef.Name.Val] = []string{tldef.Id}
			} else {
				kit.lookups.tlDefIDsByName[tldef.Name.Val] = append(n, tldef.Id)
			}
		}
	}
end:
	me.onErrs(nil, fresherrs)
}

// Errors collects whatever issues exist in any of the `Kit`'s source files
// (file-system errors, lexing/parsing errors, semantic errors etc).
func (me *Kit) Errors() (errs []error) {
	if me.Errs.Stage0DirAccessDuringRefresh != nil {
		errs = append(errs, me.Errs.Stage0DirAccessDuringRefresh)
	}
	errs = append(errs, me.Errs.Stage0BadImports...)
	for i := range me.SrcFiles {
		for _, e := range me.SrcFiles[i].Errors() {
			errs = append(errs, e)
		}
	}
	for i := range me.topLevelDefs {
		errs = append(errs, me.topLevelDefs[i].Errs.Errors()...)
	}
	errs = append(errs, me.Errs.Stage1BadNames.Errors()...)
	for _, dins := range me.defsFacts {
		for _, dol := range dins.overloads {
			dolerrs := dol.Errs()
			for _, e := range dolerrs {
				if !e.IsRef() {
					errs = append(errs, e)
				}
			}
		}
	}
	return
}

func (me *Kit) kitsDirPath() string {
	return kitsDirPathFrom(me.DirPath, me.ImpPath)
}

// HasDefs returns whether any of the `Kit`'s source files define `name`.
func (me *Kit) HasDefs(name string) bool {
	return len(me.lookups.tlDefIDsByName[name]) > 0
}

func (me *Kit) Defs(name string) (defs atmolang_irfun.AstTopDefs) {
	for len(name) > 0 && name[0] == '_' {
		name = name[1:]
	}
	if len(name) > 0 {
		for _, id := range me.lookups.tlDefIDsByName[name] {
			if def := me.lookups.tlDefsByID[id]; def != nil {
				defs = append(defs, def)
			}
		}
	}
	return
}

func (me *Kit) AstNodeAt(srcFilePath string, pos0ByteOffset int) (topLevelChunk *atmolang.AstFileTopLevelChunk, theNodeAndItsAncestors []atmolang.IAstNode) {
	if srcfile := me.SrcFiles.ByFilePath(srcFilePath); srcfile != nil {
		if topLevelChunk = srcfile.TopLevelChunkAt(pos0ByteOffset); topLevelChunk != nil {
			theNodeAndItsAncestors = topLevelChunk.At(pos0ByteOffset)
		}
	}
	return
}

func (me *Kit) AstNodeIrFunFor(defId string, origNode atmolang.IAstNode) (theNodeAndItsAncestors []atmolang_irfun.IAstNode) {
	if astdeftop := me.lookups.tlDefsByID[defId]; astdeftop != nil {
		theNodeAndItsAncestors = astdeftop.Find(origNode)
	}
	return
}
