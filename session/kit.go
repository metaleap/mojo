package atmosess

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/go-leap/fs"
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
	"github.com/metaleap/atmo/lang/corefn"
)

type Kit struct {
	ImpPath           string
	DirPath           string
	WasEverToBeLoaded bool
	DefNames          []string

	topLevel atmocorefn.AstDefs
	srcFiles atmolang.AstFiles
	errs     struct {
		dirAccessDuringRefresh error
	}
}

func (me *Ctx) KitEnsureLoaded(kit *Kit) {
	if !kit.WasEverToBeLoaded {
		me.kitForceReload(kit)
	}
}

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
	sort.Sort(this.srcFiles)

	if this.WasEverToBeLoaded {
		me.kitForceReload(this)
	}
}

func (me *Ctx) kitForceReload(this *Kit) {
	this.WasEverToBeLoaded = true
	for i := range this.srcFiles {
		sf := &this.srcFiles[i]
		sf.LexAndParseFile(true, false)
	}

	this.DefNames = this.topLevel.Reload(this.srcFiles)
	me.onErrs(this.Errors(), nil)
}

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

func (me *Kit) KitsDirPath() string {
	return KitsDirPathFrom(me.DirPath, me.ImpPath)
}

func (me *Kit) SrcFiles() atmolang.AstFiles {
	return me.srcFiles
}

func (me *Kit) HasDefs(name string) bool {
	for i := range me.srcFiles {
		if me.srcFiles[i].HasDefs(name) {
			return true
		}
	}
	return false
}

func (me *Kit) Defs(name string) (defs []*atmocorefn.AstDef) {
	for i := range me.topLevel {
		if def := &me.topLevel[i]; def.Orig.Name.Val == name {
			defs = append(defs, def)
		}
	}
	return
}
