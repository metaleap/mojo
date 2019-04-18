package atmoload

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/go-leap/fs"
	"github.com/go-leap/std"
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

var PacksWatchInterval time.Duration

func init() { ufs.WalkReadDirFunc = ufs.Dir }

type Pack struct {
	ImpPath  string
	DirPath  string
	srcFiles atmolang.AstFiles

	errs struct {
		reload error
	}
}

func (me *Ctx) WithKnownPacks(do func([]Pack)) {
	me.maybeInitPanic(false)
	me.state.Lock()
	do(me.packs.all)
	me.state.Unlock()
	return
}

func (me *Ctx) KnownPackImpPaths() (packImpPaths []string) {
	me.maybeInitPanic(false)
	me.state.Lock()
	packImpPaths = make([]string, len(me.packs.all))
	for i := range me.packs.all {
		packImpPaths[i] = me.packs.all[i].ImpPath
	}
	me.state.Unlock()
	return
}

func (me *Ctx) WithPack(impPath string, do func(*Pack)) {
	me.maybeInitPanic(false)
	me.state.Lock()
	if idx := me.packs.all.indexImpPath(impPath); idx >= 0 {
		do(&me.packs.all[idx])
	} else {
		do(nil)
	}
	me.state.Unlock()
	return
}

func (me *Ctx) ReloadModifiedPacksUnlessAlreadyWatching() {
	me.maybeInitPanic(false)
	if me.state.modsWatcher != nil {
		me.state.modsWatcher()
	}
}

func (me *Ctx) initPacks() {
	var handledir func(string, map[string]int)
	handledir = func(dirfullpath string, modpackdirs map[string]int) {
		if idx := me.packs.all.indexDirPath(dirfullpath); idx >= 0 { // dir was previously known as a this
			modpackdirs[dirfullpath] = cap(me.packs.all[idx].srcFiles)
		}
		for i := range me.packs.all {
			if ustr.Pref(me.packs.all[i].DirPath, dirfullpath+string(os.PathSeparator)) {
				modpackdirs[me.packs.all[i].DirPath] = cap(me.packs.all[i].srcFiles)
			}
		}
		dircontents, _ := ufs.Dir(dirfullpath)
		var added bool
		for _, file := range dircontents {
			if file.IsDir() {
				handledir(filepath.Join(dirfullpath, file.Name()), modpackdirs)
			} else if (!added) && ustr.Suff(file.Name(), atmo.SrcFileExt) {
				added, modpackdirs[dirfullpath] = true, modpackdirs[dirfullpath]+1
			}
		}
	}

	const modswatchdurationcritical = int64(3 * time.Millisecond)
	modswatcher := ufs.ModificationsWatcher(PacksWatchInterval/2, me.Dirs.Packs, atmo.SrcFileExt, func(mods map[string]os.FileInfo, starttime int64) {
		if len(mods) > 0 {
			me.state.Lock()
			modpackdirs := map[string]int{}
			for fullpath, fileinfo := range mods {
				if fileinfo.IsDir() {
					handledir(fullpath, modpackdirs)
				} else {
					dp := filepath.Dir(fullpath)
					modpackdirs[dp] = modpackdirs[dp] + 1
				}
			}

			if len(modpackdirs) > 0 {
				// remove packs that have vanished from the file-system
				for i := 0; i < len(me.packs.all); i++ {
					if me.packs.all[i].DirPath != dirPathAutoPack && !ufs.IsDir(me.packs.all[i].DirPath) {
						me.packs.all = append(me.packs.all[:i], me.packs.all[i+1:]...)
						i--
					}
				}
				// add any new ones, reload any potentially-modified ones
				for packdirpath, numfilesguess := range modpackdirs {
					if isdropped := false; ufs.IsDir(packdirpath) || packdirpath == dirPathAutoPack {
						idx := me.packs.all.indexDirPath(packdirpath)
						if idx < 0 {
							if idx = len(me.packs.all); numfilesguess < 4 {
								numfilesguess = 4
							}
							var packimppath string
							for _, ldp := range me.Dirs.Packs {
								if ustr.Pref(packdirpath, ldp+string(os.PathSeparator)) {
									if packimppath = filepath.Clean(packdirpath[len(ldp)+1:]); os.PathSeparator != '/' {
										packimppath = ustr.Replace(packimppath, string(os.PathSeparator), "/")
									}
									break
								}
							}
							if packimppath == "" {
								panic("should never happen, debug immediately")
							}
							for i := range me.packs.all {
								if me.packs.all[i].ImpPath == packimppath {
									isdropped = true
									me.msg(true, "duplicate import path `"+packimppath+"`:\n    ignoring the one in "+packdirpath+"\n    and using the one in "+me.packs.all[i].DirPath)
									break
								}
							}
							if !isdropped {
								me.packs.all = append(me.packs.all, Pack{DirPath: packdirpath, ImpPath: packimppath,
									srcFiles: make(atmolang.AstFiles, 0, numfilesguess)})
							}
						}
						if !isdropped {
							me.packReload(idx)
						}
					}
				}
				sort.Sort(me.packs.all)
			}
			me.state.Unlock()
		}
		if duration := time.Now().UnixNano() - starttime; duration > modswatchdurationcritical {
			me.msg(false, "[DBG] note to self, mods-watch took "+time.Duration(duration).String())
		}
	})
	if modswatchcancel := ustd.DoNowAndThenEvery(PacksWatchInterval, me.AutoPacksWatch.ShouldNow, modswatcher); modswatchcancel != nil {
		me.state.modsWatcherRunning, me.state.cleanUps =
			true, append(me.state.cleanUps, modswatchcancel)
	} else {
		me.state.modsWatcher = modswatcher
	}
}

func (me *Ctx) packReload(idx int) {
	this := &me.packs.all[idx]

	var diritems []os.FileInfo
	if diritems, this.errs.reload = ufs.Dir(this.DirPath); this.errs.reload != nil {
		this.srcFiles = nil
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

	for i := range this.srcFiles {
		this.srcFiles[i].LexAndParseFile(true, false)
		if errs := this.srcFiles[i].Errs(); len(errs) > 0 {
			for _, e := range errs {
				me.msg(true, e.Error())
			}
		}
	}
}

func (me *Pack) Errs() (errs []error) {
	if me.errs.reload != nil {
		errs = append(errs, me.errs.reload)
	}
	for i := range me.srcFiles {
		for _, e := range me.srcFiles[i].Errs() {
			errs = append(errs, e)
		}
	}
	return
}

func (me *Pack) SrcFiles() atmolang.AstFiles {
	return me.srcFiles
}

type packs []Pack

func (me packs) Len() int          { return len(me) }
func (me packs) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }
func (me packs) Less(i int, j int) bool {
	li, lj := &me[i], &me[j]
	if li.DirPath != lj.DirPath {
		if liev, ljev := li.DirPath == dirPathAutoPack, lj.DirPath == dirPathAutoPack; liev || ljev {
			return liev || !ljev
		}
	}
	return li.DirPath < lj.DirPath
}

func (me packs) indexDirPath(dirPath string) int {
	for i := range me {
		if me[i].DirPath == dirPath {
			return i
		}
	}
	return -1
}

func (me packs) indexImpPath(impPath string) int {
	for i := range me {
		if me[i].ImpPath == impPath {
			return i
		}
	}
	return -1
}
