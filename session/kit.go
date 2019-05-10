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

	topLevel    atmolang_irfun.AstTopDefs
	defsReduced map[string]*defReduced
	srcFiles    atmolang.AstFiles
	state       struct {
		defsGoneIDsNames map[string]string
		defsNew          []string
	}
	lookups struct {
		allNames        []string
		tlDefsByID      map[string]*atmolang_irfun.AstDefTop
		tlDefIDsByName  map[string][]string
		namesInScopeOwn map[string][]atmolang_irfun.IAstNode
		namesInScopeExt map[string][]atmolang_irfun.IAstNode
	}
	errs struct {
		dirAccessDuringRefresh error
		badImports             []error
	}
}

func (me *Ctx) kitEnsureLoaded(kit *Kit, redoIRs bool) {
	me.kitRefreshFilesAndMaybeReload(kit, !me.state.fileModsWatch.runningAutomaticallyPeriodically, !kit.WasEverToBeLoaded)
	if redoIRs {
		me.reReduceAffectedIRsIfAnyKitsReloaded()
	}
}

// KitEnsureLoaded forces (re)loading the `kit` only if it never was.
// (Primarily for interactive load-on-demand scenarios like REPLs or editor language servers.))
func (me *Ctx) KitEnsureLoaded(kit *Kit) {
	me.kitEnsureLoaded(kit, true)
}

func (me *Ctx) KitsEnsureLoaded(plusSessDirFauxKits bool, kitImpPaths ...string) {
	me.maybeInitPanic(false)
	me.state.Lock()
	if plusSessDirFauxKits {
		for _, dirsess := range me.Dirs.sess {
			if idx := me.Kits.all.indexDirPath(dirsess); idx >= 0 {
				kitImpPaths = append(kitImpPaths, me.Kits.all[idx].ImpPath)
			}
		}
	}
	for _, kip := range kitImpPaths {
		if kit := me.Kits.all.ByImpPath(kip); kit != nil {
			me.kitRefreshFilesAndMaybeReload(kit, !me.state.fileModsWatch.runningAutomaticallyPeriodically, true)
		}
	}
	me.reReduceAffectedIRsIfAnyKitsReloaded()
	me.state.Unlock()
}

func (me *Ctx) KitIsSessionDirFauxKit(kit *Kit) bool {
	return ustr.In(kit.DirPath, me.Dirs.sess...)
}

// WithKit runs `do` with the specified `Kit` if it exists, else with `nil`.
// The `Kit` must not be written to. While `do` runs, the `Kit` is blocked
// for updates triggered by file modifications etc.
func (me *Ctx) WithKit(impPath string, do func(*Kit)) {
	me.maybeInitPanic(false)
	me.state.Lock()
	idx := me.Kits.all.indexImpPath(impPath)
	if idx < 0 && (impPath == "" || impPath == "." || impPath == "~") && len(me.Dirs.sess) == 1 {
		idx = me.Kits.all.indexDirPath(me.Dirs.sess[0])
	}
	if idx < 0 {
		do(nil)
	} else {
		do(me.Kits.all[idx])
	}
	me.state.Unlock()
	return
}

func (me *Ctx) kitRefreshFilesAndMaybeReload(kit *Kit, forceFilesCheck bool, forceReload bool) {
	var fresherrs []error
	if forceFilesCheck {
		var diritems []os.FileInfo
		if diritems, kit.errs.dirAccessDuringRefresh = ufs.Dir(kit.DirPath); kit.errs.dirAccessDuringRefresh != nil {
			kit.srcFiles, kit.topLevel, fresherrs = nil, nil, append(fresherrs, kit.errs.dirAccessDuringRefresh)
			goto end
		}

		// any deleted files get forgotten now
		for i := 0; i < len(kit.srcFiles); i++ {
			if kit.srcFiles[i].SrcFilePath != "" && !ufs.IsFile(kit.srcFiles[i].SrcFilePath) {
				kit.srcFiles.RemoveAt(i)
				i--
			}
		}

		// any new files get added
		for _, file := range diritems {
			if (!file.IsDir()) && ustr.Suff(file.Name(), atmo.SrcFileExt) {
				if fp := filepath.Join(kit.DirPath, file.Name()); kit.srcFiles.Index(fp) < 0 {
					kit.srcFiles = append(kit.srcFiles, &atmolang.AstFile{SrcFilePath: fp})
				}
			}
		}
		atmo.SortMaybe(kit.srcFiles)
	}
	if kit.WasEverToBeLoaded || forceReload {
		kit.WasEverToBeLoaded, kit.errs.badImports, kit.lookups.tlDefIDsByName, kit.lookups.tlDefsByID, kit.lookups.allNames =
			true, nil, nil, nil, nil

		for _, sf := range kit.srcFiles {
			fresherrs = append(fresherrs, sf.LexAndParseFile(true, false)...)
		}

		for _, imp := range kit.Imports {
			if kimp := me.Kits.all.ByImpPath(imp); kimp == nil {
				kit.errs.badImports = append(kit.errs.badImports, errors.New("import not found: `"+imp+"`"))
			} else {
				me.kitEnsureLoaded(kimp, true)
			}
		}
		if len(kit.errs.badImports) > 0 {
			fresherrs = append(fresherrs, kit.errs.badImports...)
		}

		{
			od, nd, fe := kit.topLevel.ReInitFrom(kit.srcFiles)
			kit.state.defsGoneIDsNames, kit.state.defsNew, fresherrs = od, nd, append(fresherrs, fe...)
			if len(od) > 0 || len(nd) > 0 || len(fe) > 0 {
				me.state.someKitsReloaded = true
			}
		}
		kit.lookups.allNames, kit.lookups.tlDefIDsByName, kit.lookups.tlDefsByID =
			make([]string, 0, len(kit.topLevel)), make(map[string][]string, len(kit.topLevel)), make(map[string]*atmolang_irfun.AstDefTop, len(kit.topLevel))
		for _, tldef := range kit.topLevel {
			kit.lookups.tlDefsByID[tldef.ID] = tldef
			if n, ok := kit.lookups.tlDefIDsByName[tldef.Name.Val]; !ok {
				kit.lookups.tlDefIDsByName[tldef.Name.Val], kit.lookups.allNames =
					[]string{tldef.ID}, append(kit.lookups.allNames, tldef.Name.Val)
			} else {
				kit.lookups.tlDefIDsByName[tldef.Name.Val] = append(n, tldef.ID)
			}
		}

	}
end:
	me.onErrs(fresherrs, nil)
}

// Errors collects whatever issues exist in any of the `Kit`'s source files
// (file-system errors, lexing/parsing errors, semantic errors etc).
func (me *Kit) Errors() (errs []error) {
	if me.errs.dirAccessDuringRefresh != nil {
		errs = append(errs, me.errs.dirAccessDuringRefresh)
	}
	for i := range me.srcFiles {
		for _, e := range me.srcFiles[i].Errors() {
			errs = append(errs, e)
		}
	}
	for i := range me.topLevel {
		for e := range me.topLevel[i].Errors {
			errs = append(errs, &me.topLevel[i].Errors[e])
		}
	}
	for _, defred := range me.defsReduced {
		for _, rc := range defred.Cases {
			if rc.Err != nil {
				errs = append(errs, rc.Err)
			}
		}
	}
	return
}

func (me *Kit) kitsDirPath() string {
	return kitsDirPathFrom(me.DirPath, me.ImpPath)
}

// SrcFiles returns all source files belonging to the `Kit`.
// The slice or its contents must not be written to.
func (me *Kit) SrcFiles() atmolang.AstFiles {
	return me.srcFiles
}

// HasDefs returns whether any of the `Kit`'s source files define `name`.
func (me *Kit) HasDefs(name string) bool {
	return len(me.lookups.tlDefIDsByName[name]) > 0
}

func (me *Kit) Defs(name string, resolveNakedAliases bool) (defs atmolang_irfun.AstTopDefs) {
	for len(name) > 0 && name[0] == '_' {
		name = name[1:]
	}
start:
	if len(name) > 0 {
		for _, id := range me.lookups.tlDefIDsByName[name] {
			if def := me.lookups.tlDefsByID[id]; def != nil {
				if def.Orig.IsNakedAliasTo != "" {
					name = def.Orig.IsNakedAliasTo
					goto start
				}
				defs = append(defs, def)
			}
		}
	}
	return
}
