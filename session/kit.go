package atmosess

import (
	"os"
	"path/filepath"

	"github.com/go-leap/fs"
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
	"github.com/metaleap/atmo/lang"
)

// Kit is a pile of atmo source files residing in the same directory and
// being interpreted or compiled all together as a unit.
type Kit struct {
	DirPath           string
	ImpPath           string
	WasEverToBeLoaded bool

	imports      []string
	topLevelDefs atmoil.IrTopDefs
	SrcFiles     atmolang.AstFiles
	state        struct {
		defsGoneIdsNames map[string]string
		defsBornIdsNames map[string]string
	}
	lookups struct {
		tlDefsByID      map[string]*atmoil.IrDefTop
		tlDefIDsByName  map[string][]string
		namesInScopeOwn atmoil.AnnNamesInScope
		namesInScopeExt atmoil.AnnNamesInScope
		namesInScopeAll atmoil.AnnNamesInScope
	}
	Errs struct {
		Stage0DirAccessDuringRefresh *atmo.Error
		Stage0BadImports             atmo.Errors
	}
}

func (me *Ctx) KitEnsureLoaded(kit *Kit) (freshErrs atmo.Errors) {
	return me.kitEnsureLoaded(kit, true)
}

func (me *Ctx) kitEnsureLoaded(kit *Kit, reprocessEverythingInSessionAsNeededImmediatelyAfterwardsAndThenNotifySubscribers bool) (freshErrs atmo.Errors) {
	stage0errs := me.kitRefreshFilesAndMaybeReload(kit, !kit.WasEverToBeLoaded)
	freshErrs.Add(stage0errs...)
	if reprocessEverythingInSessionAsNeededImmediatelyAfterwardsAndThenNotifySubscribers {
		stage1andbeyonderrs := me.reprocessAffectedDefsIfAnyKitsReloaded()
		freshErrs.Add(stage1andbeyonderrs...)
		me.onSomeOrAllKitsPartiallyOrFullyRefreshed(stage0errs, stage1andbeyonderrs)
	}
	return
}

func (me *Ctx) KitsEnsureLoaded(plusSessDirFauxKits bool, kitImpPaths ...string) {
	if plusSessDirFauxKits {
		for _, dirsess := range me.Dirs.fauxKits {
			if idx := me.Kits.All.IndexDirPath(dirsess); idx >= 0 {
				kitImpPaths = append(kitImpPaths, me.Kits.All[idx].ImpPath)
			}
		}
	}

	var fresherrs atmo.Errors
	if len(kitImpPaths) > 0 {
		for _, kip := range kitImpPaths {
			if kit := me.Kits.All.ByImpPath(kip); kit != nil {
				fresherrs.Add(me.kitRefreshFilesAndMaybeReload(kit, !kit.WasEverToBeLoaded)...)
			}
		}
	}
	me.onSomeOrAllKitsPartiallyOrFullyRefreshed(fresherrs, me.reprocessAffectedDefsIfAnyKitsReloaded())
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
	idx := me.Kits.All.IndexImpPath(impPath)
	if idx < 0 && (impPath == "" || impPath == "." || impPath == "·") {
		if fauxkitdirs := me.Dirs.fauxKits; len(fauxkitdirs) > 0 {
			idx = me.Kits.All.IndexDirPath(fauxkitdirs[0])
		}
		if idx < 0 {
			if curdirpath, err := os.Getwd(); err == nil {
				idx = me.Kits.All.IndexDirPath(curdirpath)
			}
		}
	}
	if idx >= 0 {
		return me.Kits.All[idx]
	}
	return nil
}

func (me *Ctx) kitGatherAllUnparsedGlobalsNames(kit *Kit, unparsedGlobalsNames map[string]int) {
	kitimports := kit.Imports()
	kits := make(Kits, 1, 1+len(kitimports))
	kits[0] = kit
	for _, imppath := range kitimports {
		if kitimp := me.Kits.All.ByImpPath(imppath); kitimp != nil {
			kits = append(kits, kitimp)
		}
	}
	for _, kit = range kits {
		for _, srcfile := range kit.SrcFiles {
			for i := range srcfile.TopLevel {
				if tld := &srcfile.TopLevel[i].Ast.Def; tld.NameIfErr != "" {
					unparsedGlobalsNames[tld.NameIfErr] = 1 + unparsedGlobalsNames[tld.NameIfErr]
				}
			}
		}
	}
	return
}

func (me *Ctx) kitRefreshFilesAndMaybeReload(kit *Kit, reloadForceInsteadOfAuto bool) (freshErrs atmo.Errors) {
	var srcfileschanged bool
	var err error

	{ // step 1: files refresh
		var diritems []os.FileInfo
		if diritems, err = ufs.Dir(kit.DirPath); err != nil {
			kit.SrcFiles, kit.topLevelDefs, kit.Errs.Stage0DirAccessDuringRefresh =
				nil, nil, freshErrs.AddSess(ErrSessKits_IoReadDirFailure, kit.DirPath, err.Error())
			return
		}

		// any deleted files get forgotten now
		for i := 0; i < len(kit.SrcFiles); i++ {
			if kit.SrcFiles[i].SrcFilePath != "" && !ufs.IsFile(kit.SrcFiles[i].SrcFilePath) {
				kit.SrcFiles.RemoveAt(i)
				i, srcfileschanged = i-1, true
			}
		}

		// any new files get added
		for _, file := range diritems {
			if (!file.IsDir()) && ustr.Suff(file.Name(), atmo.SrcFileExt) {
				if fp := filepath.Join(kit.DirPath, file.Name()); kit.SrcFiles.Index(fp) < 0 {
					srcfileschanged, kit.SrcFiles = true, append(kit.SrcFiles, &atmolang.AstFile{SrcFilePath: fp})
				}
			}
		}
		if srcfileschanged {
			atmo.SortMaybe(kit.SrcFiles)
		}
	}

	{ // step 2: maybe (re)load
		if was := kit.WasEverToBeLoaded; was || reloadForceInsteadOfAuto {
			kit.WasEverToBeLoaded, kit.Errs.Stage0BadImports =
				true, nil

			allunchanged := !srcfileschanged
			for _, sf := range kit.SrcFiles {
				var nochanges bool
				freshErrs.Add(sf.LexAndParseFile(true, false, &nochanges)...)
				allunchanged = allunchanged && nochanges
			}

			for _, imp := range kit.Imports() {
				if kimp := me.Kits.All.ByImpPath(imp); kimp == nil {
					kit.Errs.Stage0BadImports.AddSess(ErrSessKits_ImportNotFound, "", "import not found: `"+imp+"`")
				} else {
					freshErrs.Add(me.kitEnsureLoaded(kimp, false)...)
				}
			}

			if len(kit.Errs.Stage0BadImports) > 0 {
				freshErrs.Add(kit.Errs.Stage0BadImports...)
			}

			if was && allunchanged && len(freshErrs) == 0 && !reloadForceInsteadOfAuto {
				return
			}
			{
				defsgone, defsborn, fresherrs := kit.topLevelDefs.ReInitFrom(kit.SrcFiles)
				if len(kit.state.defsGoneIdsNames) == 0 {
					kit.state.defsGoneIdsNames = defsgone
				} else if len(defsgone) > 0 {
					panic("TO-BE-INVESTIGATED (GONES)")
				}
				if len(kit.state.defsBornIdsNames) == 0 {
					kit.state.defsBornIdsNames = defsborn
				} else if len(defsborn) > 0 {
					panic("TO-BE-INVESTIGATED (BORNS)")
				}
				if len(kit.state.defsGoneIdsNames) > 0 || len(kit.state.defsBornIdsNames) > 0 || len(fresherrs) > 0 {
					me.Kits.reprocessingNeeded = true
				}
				freshErrs.Add(fresherrs...)
			}
			kit.lookups.tlDefIDsByName, kit.lookups.tlDefsByID = make(map[string][]string, len(kit.topLevelDefs)), make(map[string]*atmoil.IrDefTop, len(kit.topLevelDefs))
			for _, tldef := range kit.topLevelDefs {
				kit.lookups.tlDefsByID[tldef.Id], kit.lookups.tlDefIDsByName[tldef.Name.Val] =
					tldef, append(kit.lookups.tlDefIDsByName[tldef.Name.Val], tldef.Id)
			}
		}
	}
	return
}

func (me *Kit) ensureErrTldPosOffsets() {
	for _, srcfile := range me.SrcFiles {
		for i := range srcfile.TopLevel {
			tlc := &srcfile.TopLevel[i]
			tlc.Errs().UpdatePosOffsets(tlc)
		}
	}
	for _, tld := range me.topLevelDefs {
		tld.Errs.Stage0Init.UpdatePosOffsets(tld.OrigTopChunk)
		tld.Errs.Stage1BadNames.UpdatePosOffsets(tld.OrigTopChunk)
		tld.Errs.Stage2Preduce.UpdatePosOffsets(tld.OrigTopChunk)
	}
}

// Errors collects whatever issues exist in any of the `Kit`'s source files
// (file-system errors, lexing/parsing errors, semantic errors etc).
func (me *Kit) Errors(maybeErrsToSrcs map[*atmo.Error][]byte) (errs atmo.Errors) {
	if me.Errs.Stage0DirAccessDuringRefresh != nil {
		errs.Add(me.Errs.Stage0DirAccessDuringRefresh)
	}
	errs.Add(me.Errs.Stage0BadImports...)
	for i := range me.SrcFiles {
		for _, e := range me.SrcFiles[i].Errors() {
			if errs = append(errs, e); maybeErrsToSrcs != nil {
				maybeErrsToSrcs[e] = me.SrcFiles[i].LastLoad.Src
			}
		}
	}
	for i := range me.topLevelDefs {
		deferrs := append(append(me.topLevelDefs[i].Errs.Stage0Init, me.topLevelDefs[i].Errs.Stage1BadNames...), me.topLevelDefs[i].Errs.Stage2Preduce...)
		if maybeErrsToSrcs != nil {
			for _, e := range deferrs {
				maybeErrsToSrcs[e] = me.topLevelDefs[i].OrigTopChunk.SrcFile.LastLoad.Src
			}
		}
		errs.Add(deferrs...)
	}
	return
}

func (me *Kit) DoesImport(kitImpPath string) bool {
	return kitImpPath != me.ImpPath && ustr.In(kitImpPath, me.Imports()...)
}

func (me *Kit) Imports() []string {
	if me.imports == nil {
		me.imports = make([]string, 0, 1)
		if me.ImpPath != atmo.NameAutoKit {
			me.imports = append(me.imports, atmo.NameAutoKit)
		}
	}
	return me.imports
}

func (me *Kit) kitsDirPath() string {
	return kitsDirPathFrom(me.DirPath, me.ImpPath)
}

// HasDefs returns whether any of the `Kit`'s source files define `name`.
func (me *Kit) HasDefs(name string) bool {
	return len(me.lookups.tlDefIDsByName[name]) > 0
}

func (me *Kit) Defs(name string, includeUnparsedOnes bool) (defs atmoil.IrTopDefs) {
	for len(name) > 0 && name[0] == '_' {
		name = name[1:]
	}
	if len(name) > 0 {
		for _, id := range me.lookups.tlDefIDsByName[name] {
			if def := me.lookups.tlDefsByID[id]; def != nil {
				defs = append(defs, def)
			}
		}
		if includeUnparsedOnes {
			for _, tld := range me.topLevelDefs {
				if tld.OrigTopChunk.Ast.Def.NameIfErr == name {
					defs = append(defs, tld)
				}
			}
		}
	}
	return
}

func (me *Kit) AstNodeAt(srcFilePath string, pos0ByteOffset int) (topLevelChunk *atmolang.SrcTopChunk, theNodeAndItsAncestors []atmolang.IAstNode) {
	if srcfile := me.SrcFiles.ByFilePath(srcFilePath); srcfile != nil {
		if topLevelChunk = srcfile.TopLevelChunkAt(pos0ByteOffset); topLevelChunk != nil {
			theNodeAndItsAncestors = topLevelChunk.At(pos0ByteOffset)
		}
	}
	return
}

func (me *Kit) IrNodeOfAstNode(defId string, origNode atmolang.IAstNode) (astDefTop *atmoil.IrDefTop, theNodeAndItsAncestors []atmoil.INode) {
	if astDefTop = me.lookups.tlDefsByID[defId]; astDefTop != nil {
		theNodeAndItsAncestors = astDefTop.FindByOrig(origNode)
	}
	return
}

func (me *Kit) SelectNodes(tldOk func(*atmoil.IrDefTop) bool, nodeOk func([]atmoil.INode, atmoil.INode, []atmoil.INode) (ismatch bool, dontdescend bool, tlddone bool, alldone bool)) (matches map[atmoil.INode]*atmoil.IrDefTop) {
	var alldone bool
	matches = map[atmoil.INode]*atmoil.IrDefTop{}
	for _, tld := range me.topLevelDefs {
		if tldOk(tld) {
			tld.Walk(func(curnodeancestors []atmoil.INode, curnode atmoil.INode, curnodedescendantsthatwillbetraversedifreturningtrue ...atmoil.INode) (traverse bool) {
				ismatch, dontdescend, donetld, doneall := nodeOk(curnodeancestors, curnode, curnodedescendantsthatwillbetraversedifreturningtrue)
				if doneall {
					alldone = true
				}
				if ismatch {
					matches[curnode] = tld
				}
				return !(dontdescend || donetld || doneall)
			})
			if alldone {
				break
			}
		}
	}
	return
}

func IsValidKitDirName(dirName string) bool {
	return (!ustr.IsLen1And(dirName, '_', '*', '.', ' ')) &&
		dirName != "·" &&
		ustr.BeginsUpper(dirName) &&
		(!ustr.HasAnyOf(dirName, ' '))
}
