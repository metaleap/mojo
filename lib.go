package atem

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-leap/fs"
	"github.com/go-leap/str"
)

func init() { ufs.WalkReadDirFunc = ufs.Dir }

type Libs []Lib

type Lib struct {
	LibPath string
	DirPath string
}

func (me *Ctx) initLibs() {
	var handledir func(string, map[string]bool)
	handledir = func(dirfullpath string, modlibdirs map[string]bool) {
		if _, ok := me.libs.lookups.dirPaths[dirfullpath]; ok {
			// dir was previously known as a lib
			modlibdirs[dirfullpath] = true
		}
		dircontents, _ := ufs.Dir(dirfullpath)
		var added bool
		for _, file := range dircontents {
			if file.IsDir() {
				handledir(filepath.Join(dirfullpath, file.Name()), modlibdirs)
			} else if (!added) && ustr.Suff(file.Name(), SrcFileExt) {
				added, modlibdirs[dirfullpath] = true, true
			}
		}
	}

	me.cleanUps = append(me.cleanUps,
		ufs.WatchModTimesEvery(3*time.Second, me.Dirs.Libs, SrcFileExt, func(mods map[string]os.FileInfo) {
			modlibdirs := map[string]bool{}
			for fullpath, fileinfo := range mods {
				if !fileinfo.IsDir() {
					modlibdirs[filepath.Dir(fullpath)] = true
				} else {
					handledir(fullpath, modlibdirs)
				}
			}
			if len(modlibdirs) > 0 {

			}
		}))
}
