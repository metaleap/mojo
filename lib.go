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

type Lib struct {
	LibPath string
	DirPath string
	Errors  struct {
		Refresh error
	}
	SrcFiles atemlang.AstFiles
}

func (me *Ctx) KnownLibs() (known Libs) {
	me.maybeInitPanic(false)
	me.libs.Lock()
	known = me.libs.Known
	me.libs.Unlock()
	return
}

func (me *Ctx) KnownLibPaths() (libPaths []string) {
	me.maybeInitPanic(false)
	me.libs.Lock()
	already := make(map[string]bool, len(me.libs.Known))
	libPaths = make([]string, 0, len(me.libs.Known))
	for i := range me.libs.Known {
		if libpath := me.libs.Known[i].LibPath; !already[libpath] {
			already[libpath], libPaths = true, append(libPaths, libpath)
		}
	}
	me.libs.Unlock()
	return
}

func (me *Ctx) Lib(libPath string) (lib *Lib) {
	me.maybeInitPanic(false)
	me.libs.Lock()
	if idx := me.libs.Known.indexLibPath(libPath); idx >= 0 {
		lib = &me.libs.Known[idx]
	}
	me.libs.Unlock()
	return
}

func (me *Ctx) LibReachable(lib *Lib) (reachable bool) {
	me.maybeInitPanic(false)
	me.libs.Lock()
	if idx := me.libs.Known.indexLibPath(lib.LibPath); idx >= 0 {
		reachable = (lib == &me.libs.Known[idx])
	}
	me.libs.Unlock()
	return
}

func (me *Ctx) initLibs() {
	var handledir func(string, map[string]int)

	me.cleanUps = append(me.cleanUps,
		ufs.WatchModTimesEvery(LibWatchInterval, me.Dirs.Libs, SrcFileExt, func(mods map[string]os.FileInfo) {
			modlibdirs := map[string]int{}
			for fullpath, fileinfo := range mods {
				if fileinfo.IsDir() {
					handledir(fullpath, modlibdirs)
				} else {
					dp := filepath.Dir(fullpath)
					modlibdirs[dp] = modlibdirs[dp] + 1
				}
			}

			if len(modlibdirs) > 0 {
				me.libs.Lock()
				// remove libs that have vanished from the file-system
				for i := 0; i < len(me.libs.Known); i++ {
					if !ufs.IsDir(me.libs.Known[i].DirPath) {
						me.libs.Known = append(me.libs.Known[:i], me.libs.Known[i+1:]...)
						i--
					}
				}
				// add any new ones, refresh any potentially-modified ones
				for libdirpath, numfilesguess := range modlibdirs {
					if ufs.IsDir(libdirpath) {
						idx := me.libs.Known.indexDirPath(libdirpath)
						if idx < 0 {
							if idx = len(me.libs.Known); numfilesguess < 4 {
								numfilesguess = 4
							}
							var libpath string
							for _, ldp := range me.Dirs.Libs {
								if ustr.Pref(libdirpath, ustr.TrimR(ldp, "/\\")+string(os.PathSeparator)) {
									libpath = ustr.TrimLR(ustr.ReplB(libdirpath[len(ldp):], '\\', '/'), "/")
									break
								}
							}
							me.libs.Known = append(me.libs.Known, Lib{DirPath: libdirpath, LibPath: libpath,
								SrcFiles: make(atemlang.AstFiles, 0, numfilesguess)})
						}
						me.libRefresh(idx)
					}
				}
				me.libs.Unlock()
			}
		}))

	handledir = func(dirfullpath string, modlibdirs map[string]int) {
		if idx := me.libs.Known.indexDirPath(dirfullpath); idx >= 0 { // dir was previously known as a lib
			modlibdirs[dirfullpath] = cap(me.libs.Known[idx].SrcFiles)
		}
		dircontents, _ := ufs.Dir(dirfullpath)
		var added bool
		for _, file := range dircontents {
			if file.IsDir() {
				handledir(filepath.Join(dirfullpath, file.Name()), modlibdirs)
			} else if (!added) && ustr.Suff(file.Name(), SrcFileExt) {
				added, modlibdirs[dirfullpath] = true, modlibdirs[dirfullpath]+1
			}
		}
	}
}

func (me *Ctx) libRefresh(idx int) {
	lib := &me.libs.Known[idx]
	diritems, e := ufs.Dir(lib.DirPath)
	if e != nil {
		lib.SrcFiles = lib.SrcFiles[0:0]
		lib.Errors.Refresh = e
		return
	}
	// any deleted files get forgotten now
	for i := 0; i < len(lib.SrcFiles); i++ {
		if lib.SrcFiles[i].SrcFilePath != "" && !ufs.IsFile(lib.SrcFiles[i].SrcFilePath) {
			lib.SrcFiles.RemoveAt(i)
			i--
		}
	}

	// any new files get added
	for _, file := range diritems {
		if (!file.IsDir()) && ustr.Suff(file.Name(), SrcFileExt) {
			if fp := filepath.Join(lib.DirPath, file.Name()); !lib.SrcFiles.Contains(fp) {
				lib.SrcFiles = append(lib.SrcFiles, atemlang.AstFile{SrcFilePath: fp})
			}
		}
	}

	for i := range lib.SrcFiles {
		lib.SrcFiles[i].LexAndParseFile(true, false)
	}
}

func (me *Lib) Errs() (errs []error) {
	if me.Errors.Refresh != nil {
		errs = append(errs, me.Errors.Refresh)
	}
	for i := range me.SrcFiles {
		for _, e := range me.SrcFiles[i].Errs() {
			errs = append(errs, e)
		}
	}
	return
}

func (me *Lib) Err() error {
	if errs := me.Errs(); len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (me *Lib) Error() (errMsg string) {
	if e := me.Err(); e != nil {
		errMsg = e.Error()
	}
	return
}

func (me *Lib) IsEverLib() bool { return me.LibPath == "ever" }

type Libs []Lib

func (me Libs) indexDirPath(dirPath string) int {
	for i := range me {
		if me[i].DirPath == dirPath {
			return i
		}
	}
	return -1
}

func (me Libs) indexLibPath(libPath string) int {
	for i := range me {
		if me[i].LibPath == libPath {
			return i
		}
	}
	return -1
}
