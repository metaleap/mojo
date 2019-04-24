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

type Pack struct {
	ImpPath string
	DirPath string

	topLevel          atmocorefn.AstDefs
	srcFiles          atmolang.AstFiles
	wasEverToBeLoaded bool
	errs              struct {
		refresh error
	}
}

func (me *Ctx) WithPack(impPath string, ensureLoaded bool, do func(*Pack)) {
	me.maybeInitPanic(false)
	me.state.Lock()
	if idx := me.packs.all.indexImpPath(impPath); idx < 0 {
		do(nil)
	} else {
		if ensureLoaded && !me.packs.all[idx].wasEverToBeLoaded {
			me.packReload(idx)
		}
		do(&me.packs.all[idx])
	}
	me.state.Unlock()
	return
}

func (me *Ctx) packRefresh(idx int) {
	this := &me.packs.all[idx]
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
		me.packReload(idx)
	}
}

func (me *Ctx) packReload(idx int) {
	this := &me.packs.all[idx]
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

func (me *Pack) Errs() (errs []error) {
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

func (me *Pack) PacksDirPath() string {
	return PacksDirPathFrom(me.DirPath, me.ImpPath)
}

func (me *Pack) SrcFiles() atmolang.AstFiles {
	return me.srcFiles
}
