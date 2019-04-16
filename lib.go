package atem

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-leap/fs"
	"github.com/go-leap/str"
	"github.com/metaleap/atem/lang"
)

var LibWatchInterval = 1 * time.Second

func init() { ufs.WalkReadDirFunc = ufs.Dir }

type Libs []Lib

type Lib struct {
	LibPath string
	DirPath string
	Errs    struct {
		Refresh error
	}
	SrcFiles atemlang.AstFiles
}

func (me *Ctx) KnownLibs() (known Libs) {
	me.libs.Lock()
	known = me.libs.Known
	me.libs.Unlock()
	return
}

func (me *Ctx) initLibs() {
	var handledir func(string, map[string]bool)

	me.cleanUps = append(me.cleanUps,
		ufs.WatchModTimesEvery(LibWatchInterval, me.Dirs.Libs, SrcFileExt, func(mods map[string]os.FileInfo) {
			modlibdirs := map[string]bool{}
			for fullpath, fileinfo := range mods {
				if !fileinfo.IsDir() {
					modlibdirs[filepath.Dir(fullpath)] = true
				} else {
					handledir(fullpath, modlibdirs)
				}
			}

			me.libs.Lock()
			if len(modlibdirs) > 0 {
				for libdirpath := range modlibdirs {
					idx, ok := me.libs.lookups.dirPaths[libdirpath]
					if !ok {
						var libpath string
						for _, ldp := range me.Dirs.Libs {
							if ustr.Pref(libdirpath, ustr.TrimR(ldp, "/\\")+string(os.PathSeparator)) {
								libpath = ustr.TrimLR(ustr.ReplB(libdirpath[len(ldp):], '\\', '/'), "/")
							}
						}
						idx = len(me.libs.Known)
						me.libs.lookups.dirPaths[libdirpath], me.libs.lookups.libPaths[libpath] = idx, idx
						me.libs.Known = append(me.libs.Known, Lib{DirPath: libdirpath, LibPath: libpath,
							SrcFiles: make(atemlang.AstFiles, 0, 8)})
					}
					if ufs.IsDir(libdirpath) {
						me.libRefresh(idx)
					}
				}
			}

			var gonelibs bool
			for i := 0; i < len(me.libs.Known); i++ {
				if lib := &me.libs.Known[i]; !ufs.IsDir(lib.DirPath) {
					me.libs.Known = append(me.libs.Known[:i], me.libs.Known[i+1:]...)
					gonelibs, i = true, i-1
				}
			}
			if gonelibs {
				me.libs.lookups.dirPaths, me.libs.lookups.libPaths = make(map[string]int, len(me.libs.Known)), make(map[string]int, len(me.libs.Known))
				for i := range me.libs.Known {
					me.libs.lookups.dirPaths[me.libs.Known[i].DirPath], me.libs.lookups.libPaths[me.libs.Known[i].LibPath] = i, i
				}
			}
			me.libs.Unlock()
		}))

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
}

func (me *Ctx) libRefresh(idx int) {
	lib := &me.libs.Known[idx]

	// any deleted files get forgotten now
	for i := 0; i < len(lib.SrcFiles); i++ {
		if lib.SrcFiles[i].SrcFilePath != "" && !ufs.IsFile(lib.SrcFiles[i].SrcFilePath) {
			lib.SrcFiles.RemoveAt(i)
			i--
		}
	}

	// any new files get added
	var files []os.FileInfo
	if files, lib.Errs.Refresh = ufs.Files(lib.DirPath, SrcFileExt); lib.Errs.Refresh == nil {
		for _, file := range files {
			if file != nil {
				if fp := filepath.Join(lib.DirPath, file.Name()); !lib.SrcFiles.Contains(fp) {
					lib.SrcFiles = append(lib.SrcFiles, atemlang.AstFile{SrcFilePath: fp})
				}
			}
		}
	}
}

func (me *Lib) Error() (errMsg string) {
	if me.Errs.Refresh != nil {
		errMsg = me.Errs.Refresh.Error()
	}
	return
}
