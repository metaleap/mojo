package atmosess

import (
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

	topLevel atmolang_irfun.AstTopDefs
	srcFiles atmolang.AstFiles
	state    struct {
		defsGone []string
		defsNew  []string
	}
	lookups struct {
		tlDefsByID     map[string]*atmolang_irfun.AstDefTop
		tlDefIDsByName map[string][]string
	}
	errs struct {
		dirAccessDuringRefresh error
	}
}

// KitEnsureLoaded forces (re)loading the `kit` only if it never was.
// (Primarily for interactive load-on-demand scenarios like REPLs or editor language servers.))
func (me *Ctx) KitEnsureLoaded(kit *Kit) {
	if !kit.WasEverToBeLoaded {
		me.kitForceReload(kit)
		me.renewAndRevalidateAffectedIRsIfAnyKitsReloaded()
	}
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
			me.kitForceReload(kit)
		}
	}
	me.renewAndRevalidateAffectedIRsIfAnyKitsReloaded()
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

func (me *Ctx) kitRefreshFilesAndReloadIfWasLoaded(idx int) {
	this := me.Kits.all[idx]
	var diritems []os.FileInfo
	if diritems, this.errs.dirAccessDuringRefresh = ufs.Dir(this.DirPath); this.errs.dirAccessDuringRefresh != nil {
		this.srcFiles, this.topLevel = nil, nil
		return
	}

	// any deleted files get forgotten now
	for i := 0; i < len(this.srcFiles); i++ {
		if this.srcFiles[i].SrcFilePath != "" && !ufs.IsFile(this.srcFiles[i].SrcFilePath) {
			this.srcFiles.RemoveAt(i)
			i--
		}
	}

	// any new files get added
	for _, file := range diritems {
		if (!file.IsDir()) && ustr.Suff(file.Name(), atmo.SrcFileExt) {
			if fp := filepath.Join(this.DirPath, file.Name()); this.srcFiles.Index(fp) < 0 {
				this.srcFiles = append(this.srcFiles, atmolang.AstFile{SrcFilePath: fp})
			}
		}
	}
	if atmo.SortMaybe(this.srcFiles); this.WasEverToBeLoaded {
		me.kitForceReload(this)
	}
}

func (me *Ctx) kitForceReload(kit *Kit) {
	kit.WasEverToBeLoaded = true
	var fresherrs []error

	for i := range kit.srcFiles {
		sf := &kit.srcFiles[i]
		fresherrs = append(fresherrs, sf.LexAndParseFile(true, false)...)
	}
	{
		od, nd, fe := kit.topLevel.ReInitFrom(kit.srcFiles)
		kit.state.defsGone, kit.state.defsNew, fresherrs = od, nd, append(fresherrs, fe...)
		if len(od) > 0 || len(nd) > 0 || len(fe) > 0 {
			me.state.someKitsReloaded = true
		}
	}
	kit.lookups.tlDefIDsByName, kit.lookups.tlDefsByID = make(map[string][]string, len(kit.topLevel)), make(map[string]*atmolang_irfun.AstDefTop, len(kit.topLevel))
	for i := range kit.topLevel {
		tldef := &kit.topLevel[i]
		kit.lookups.tlDefsByID[tldef.ID] = tldef
		kit.lookups.tlDefIDsByName[tldef.Name.Val] = append(kit.lookups.tlDefIDsByName[tldef.Name.Val], tldef.ID)
	}

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

func (me *Kit) Defs(name string, resolveNakedAliases bool) (defs []*atmolang_irfun.AstDefTop) {
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
