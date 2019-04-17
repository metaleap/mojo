package atmo

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/go-leap/fs"
	"github.com/go-leap/std"
	"github.com/go-leap/str"
	"github.com/metaleap/atmo/lang"
)

var LibsWatchInterval = 1 * time.Second

func init() { ufs.WalkReadDirFunc = ufs.Dir }

type Lib struct {
	LibPath string
	DirPath string
	Errors  struct {
		Reload error
	}
	SrcFiles atmolang.AstFiles
}

func (me *Ctx) KnownLibs() (known []Lib) {
	me.maybeInitPanic(false)
	me.Lock()
	known = me.libs.all
	me.Unlock()
	return
}

func (me *Ctx) KnownLibPaths() (libPaths []string) {
	me.maybeInitPanic(false)
	me.Lock()
	already := make(map[string]bool, len(me.libs.all))
	libPaths = make([]string, 0, len(me.libs.all))
	for i := range me.libs.all {
		if libpath := me.libs.all[i].LibPath; !already[libpath] {
			already[libpath], libPaths = true, append(libPaths, libpath)
		}
	}
	me.Unlock()
	return
}

func (me *Ctx) Lib(libPath string) (lib *Lib) {
	me.maybeInitPanic(false)
	me.Lock()
	if idx := me.libs.all.indexLibPath(libPath); idx >= 0 {
		lib = &me.libs.all[idx]
	}
	me.Unlock()
	return
}

func (me *Ctx) LibEver() (lib *Lib) {
	me.maybeInitPanic(false)
	me.Lock()
	lib = &me.libs.all[me.libs.all.indexDirPath(dirPathAutoLib)]
	me.Unlock()
	return
}

func (me *Ctx) LibReachable(lib *Lib) (reachable bool) {
	me.maybeInitPanic(false)
	me.Lock()
	if idx := me.libs.all.indexLibPath(lib.LibPath); idx >= 0 {
		reachable = (lib == &me.libs.all[idx])
	}
	me.Unlock()
	return
}

func (me *Ctx) ReloadModifiedLibsUnlessAlreadyWatching() {
	if me.state.modsWatcher != nil {
		me.state.modsWatcher()
	}
}

func (me *Ctx) initLibs() {
	var handledir func(string, map[string]int)
	handledir = func(dirfullpath string, modlibdirs map[string]int) {
		if idx := me.libs.all.indexDirPath(dirfullpath); idx >= 0 { // dir was previously known as a lib
			modlibdirs[dirfullpath] = cap(me.libs.all[idx].SrcFiles)
		}
		for i := range me.libs.all {
			if ustr.Pref(me.libs.all[i].DirPath, dirfullpath+string(os.PathSeparator)) {
				modlibdirs[me.libs.all[i].DirPath] = cap(me.libs.all[i].SrcFiles)
			}
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

	const modswatchdurationcritical = int64(3 * time.Millisecond)
	modswatcher := ufs.ModificationsWatcher(LibsWatchInterval/2, me.Dirs.Libs, SrcFileExt, func(mods map[string]os.FileInfo, starttime int64) {
		if len(mods) > 0 {
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
				me.Lock()
				// remove libs that have vanished from the file-system
				for i := 0; i < len(me.libs.all); i++ {
					if me.libs.all[i].DirPath != dirPathAutoLib && !ufs.IsDir(me.libs.all[i].DirPath) {
						me.libs.all = append(me.libs.all[:i], me.libs.all[i+1:]...)
						i--
					}
				}
				// add any new ones, reload any potentially-modified ones
				for libdirpath, numfilesguess := range modlibdirs {
					if ufs.IsDir(libdirpath) || libdirpath == dirPathAutoLib {
						idx := me.libs.all.indexDirPath(libdirpath)
						if idx < 0 {
							if idx = len(me.libs.all); numfilesguess < 4 {
								numfilesguess = 4
							}
							var libpath string
							for _, ldp := range me.Dirs.Libs {
								if ustr.Pref(libdirpath, ldp+string(os.PathSeparator)) {
									if libpath = filepath.Clean(libdirpath[len(ldp)+1:]); os.PathSeparator != '/' {
										libpath = ustr.Replace(libpath, string(os.PathSeparator), "/")
									}
									break
								}
							}
							me.libs.all = append(me.libs.all, Lib{DirPath: libdirpath, LibPath: libpath,
								SrcFiles: make(atmolang.AstFiles, 0, numfilesguess)})
						}
						me.libReload(idx)
					}
				}
				sort.Sort(me.libs.all)
				me.Unlock()
			}
		}
		if duration := time.Now().UnixNano() - starttime; duration > modswatchdurationcritical {
			me.msg(false, "[DBG] note to self, mods-watch took "+time.Duration(duration).String())
		}

	})
	if modswatchcancel := ustd.DoNowAndThenEvery(LibsWatchInterval, me.LibsWatch.Should, modswatcher); modswatchcancel != nil {
		me.state.modsWatcherRunning, me.state.cleanUps =
			true, append(me.state.cleanUps, modswatchcancel)
	} else {
		me.state.modsWatcher = modswatcher
	}
}
func (me *Ctx) libReload(idx int) {
	this := &me.libs.all[idx]

	var diritems []os.FileInfo
	if diritems, this.Errors.Reload = ufs.Dir(this.DirPath); this.Errors.Reload != nil {
		this.SrcFiles = nil
		return
	}

	// any deleted files get forgotten now
	for i := 0; i < len(this.SrcFiles); i++ {
		if !ufs.IsFile(this.SrcFiles[i].SrcFilePath) {
			this.SrcFiles.RemoveAt(i)
			i--
		}
	}

	// any new files get added
	for _, file := range diritems {
		if (!file.IsDir()) && ustr.Suff(file.Name(), SrcFileExt) {
			if fp := filepath.Join(this.DirPath, file.Name()); !this.SrcFiles.Contains(fp) {
				this.SrcFiles = append(this.SrcFiles, atmolang.AstFile{SrcFilePath: fp})
			}
		}
	}

	for i := range this.SrcFiles {
		this.SrcFiles[i].LexAndParseFile(true, false)
		if errs := this.SrcFiles[i].Errs(); len(errs) > 0 {
			for _, e := range errs {
				me.msg(true, e.Error())
			}
		}
	}
}
func (me *Lib) Errs() (errs []error) {
	if me.Errors.Reload != nil {
		errs = append(errs, me.Errors.Reload)
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

func (me *Lib) IsEverLib() bool { return me.DirPath == dirPathAutoLib }

type libs []Lib

func (me libs) Len() int          { return len(me) }
func (me libs) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }
func (me libs) Less(i int, j int) bool {
	li, lj := &me[i], &me[j]
	if li.DirPath != lj.DirPath {
		if liev, ljev := li.DirPath == dirPathAutoLib, lj.DirPath == dirPathAutoLib; liev || ljev {
			return liev || !ljev
		}
	}
	return li.DirPath < lj.DirPath
}

func (me libs) indexDirPath(dirPath string) int {
	for i := range me {
		if me[i].DirPath == dirPath {
			return i
		}
	}
	return -1
}

func (me libs) indexLibPath(libPath string) int {
	for i := range me {
		if me[i].LibPath == libPath {
			return i
		}
	}
	return -1
}
