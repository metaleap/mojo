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
	errs     struct {
		dirAccessDuringRefresh error
	}
}

// KitEnsureLoaded forces (re)loading the `kit` only if it never was.
// (Primarily for interactive load-on-demand scenarios like REPLs or editor language servers.))
func (me *Ctx) KitEnsureLoaded(kit *Kit) {
	if !kit.WasEverToBeLoaded {
		me.kitForceReload(kit)
	}
}

// WithKit runs `do` with the specified `Kit` if it exists, else with `nil`.
// The `Kit` must not be written to. While `do` runs, the `Kit` is blocked
// for updates triggered by file modifications etc.
func (me *Ctx) WithKit(impPath string, do func(*Kit)) {
	me.maybeInitPanic(false)
	me.state.Lock()
	if idx := me.Kits.all.indexImpPath(impPath); idx < 0 {
		do(nil)
	} else {
		do(&me.Kits.all[idx])
	}
	me.state.Unlock()
	return
}

func (me *Ctx) kitRefreshFilesAndReloadIfWasLoaded(idx int) {
	this := &me.Kits.all[idx]
	var diritems []os.FileInfo
	if diritems, this.errs.dirAccessDuringRefresh = ufs.Dir(this.DirPath); this.errs.dirAccessDuringRefresh != nil {
		this.srcFiles, this.topLevel = nil, nil
		return
	}

	// any deleted files get forgotten now
	for i := 0; i < len(this.srcFiles); i++ {
		if !ufs.IsFile(this.srcFiles[i].SrcFilePath) {
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

	fresherrs = append(fresherrs, kit.topLevel.ReInitFrom(kit.srcFiles)...)
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
	for i := range me.srcFiles {
		if me.srcFiles[i].HasDefs(name) {
			return true
		}
	}
	return false
}

func (me *Kit) Defs(name string, resolveNakedAliases bool) (defs []*atmolang_irfun.AstDefTop) {
	if name[0] == '_' {
		name = name[1:]
	}
start:
	for i := range me.topLevel {
		if def := &me.topLevel[i]; def.Orig.Name.Val == name {
			if def.Orig.IsNakedAliasTo != "" {
				name = def.Orig.IsNakedAliasTo
				goto start
			}
			defs = append(defs, def)
		}
	}
	return
}
