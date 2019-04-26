package atmoload

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
	ImpPath string
	DirPath string

	topLevel          atmocorefn.AstDefs
	srcFiles          atmolang.AstFiles
	wasEverToBeLoaded bool
	errs              struct {
		refresh error
	}
}

func (me *Ctx) WithKit(impPath string, ensureLoaded bool, do func(*Kit)) {
	me.maybeInitPanic(false)
	me.state.Lock()
	if idx := me.kits.all.indexImpPath(impPath); idx < 0 {
		do(nil)
	} else {
		if ensureLoaded && !me.kits.all[idx].wasEverToBeLoaded {
			me.kitReload(idx)
		}
		do(&me.kits.all[idx])
	}
	me.state.Unlock()
	return
}

func (me *Ctx) kitRefresh(idx int) {
	this := &me.kits.all[idx]
	var diritems []os.FileInfo
	if diritems, this.errs.refresh = ufs.Dir(this.DirPath); this.errs.refresh != nil {
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

	if this.wasEverToBeLoaded {
		me.kitReload(idx)
	}
}

func (me *Ctx) kitReload(idx int) {
	this := &me.kits.all[idx]
	this.wasEverToBeLoaded = true
	for i := range this.srcFiles {
		sf := &this.srcFiles[i]
		sf.LexAndParseFile(true, false)
		if errs := sf.Errors(); len(errs) > 0 {
			for _, e := range errs {
				me.msg(true, e.Error())
			}
		}
	}
	this.topLevel.Reload(this.srcFiles)
}

func (me *Kit) Errs() (errs []error) {
	if me.errs.refresh != nil {
		errs = append(errs, me.errs.refresh)
	}
	for i := range me.srcFiles {
		for _, e := range me.srcFiles[i].Errors() {
			errs = append(errs, e)
		}
	}
	for i := range me.topLevel {
		for e := range me.topLevel[i].Errs {
			errs = append(errs, &me.topLevel[i].Errs[e])
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

func (me *Kit) Defs(name string) (defs []*atmocorefn.AstDef) {
	wantall := (name == "")
	for i := range me.topLevel {
		if def := &me.topLevel[i]; wantall || def.Name.String() == name {
			defs = append(defs, def)
		}
	}
	return
}
