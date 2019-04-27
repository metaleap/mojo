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

type kits []Kit

var KitsWatchInterval time.Duration

func init() { ufs.WalkReadDirFunc = ufs.Dir }

func (me *Ctx) WithKnownKits(do func([]Kit)) {
	me.maybeInitPanic(false)
	me.state.Lock()
	do(me.kits.all)
	me.state.Unlock()
	return
}

func (me *Ctx) WithKnownKitsWhere(where func(*Kit) bool, do func([]*Kit)) {
	me.maybeInitPanic(false)
	me.state.Lock()
	doall, kits := (where == nil), make([]*Kit, 0, len(me.kits.all))
	for i := range me.kits.all {
		if kit := &me.kits.all[i]; doall || where(kit) {
			kits = append(kits, kit)
		}
	}
	do(kits)
	me.state.Unlock()
	return
}

func (me *Ctx) KnownKitImpPaths() (kitImpPaths []string) {
	me.maybeInitPanic(false)
	me.state.Lock()
	kitImpPaths = make([]string, len(me.kits.all))
	for i := range me.kits.all {
		kitImpPaths[i] = me.kits.all[i].ImpPath
	}
	me.state.Unlock()
	return
}

func (me *Ctx) ReloadModifiedKitsUnlessAlreadyWatching() (numFileSystemModsNoticedAndActedUpon int) {
	me.maybeInitPanic(false)
	if me.state.fileModsWatch.doManually == nil {
		numFileSystemModsNoticedAndActedUpon = -1
	} else {
		numFileSystemModsNoticedAndActedUpon = me.state.fileModsWatch.doManually()
	}
	return
}

func (me *Ctx) initKits() {
	var handledir func(string, map[string]int)
	handledir = func(dirfullpath string, modkitdirs map[string]int) {
		if idx := me.kits.all.indexDirPath(dirfullpath); idx >= 0 { // dir was previously known as a kit
			modkitdirs[dirfullpath] = cap(me.kits.all[idx].srcFiles)
		} else if dirfullpath == me.Dirs.Session {
			modkitdirs[dirfullpath] = 1
		}
		for i := range me.kits.all {
			if ustr.Pref(me.kits.all[i].DirPath, dirfullpath+string(os.PathSeparator)) {
				modkitdirs[me.kits.all[i].DirPath] = cap(me.kits.all[i].srcFiles)
			}
		}
		dircontents, _ := ufs.Dir(dirfullpath)
		var added bool
		for _, file := range dircontents {
			if file.IsDir() {
				handledir(filepath.Join(dirfullpath, file.Name()), modkitdirs)
			} else if (!added) && ustr.Suff(file.Name(), atmo.SrcFileExt) {
				added, modkitdirs[dirfullpath] = true, modkitdirs[dirfullpath]+1
			}
		}
	}

	var watchdircur []string
	if !me.Dirs.curAlreadyInKitsDirs {
		watchdircur = []string{me.Dirs.Session}
	}
	modswatcher := ufs.ModificationsWatcher(KitsWatchInterval/2, me.Dirs.Kits, watchdircur, atmo.SrcFileExt, func(mods map[string]os.FileInfo, starttime int64) {
		var filemodwatchduration int64
		if len(mods) > 0 {
			me.state.Lock()
			modkitdirs := map[string]int{}
			for fullpath, fileinfo := range mods {
				if fileinfo.IsDir() {
					handledir(fullpath, modkitdirs)
				} else {
					dp := filepath.Dir(fullpath)
					modkitdirs[dp] = modkitdirs[dp] + 1
				}
			}

			if len(me.kits.all) == 0 && !me.Dirs.curAlreadyInKitsDirs {
				modkitdirs[me.Dirs.Session] = 1
			}
			if filemodwatchduration = time.Now().UnixNano() - starttime; len(modkitdirs) > 0 {
				shouldrefresh := make(map[string]bool, len(modkitdirs))
				// handle new-or-modified kits
				for kitdirpath, numfilesguess := range modkitdirs {
					if me.kits.all.indexDirPath(kitdirpath) < 0 {
						if numfilesguess < 4 {
							numfilesguess = 4
						}
						var kitimppath string
						for _, ldp := range me.Dirs.Kits {
							if ustr.Pref(kitdirpath, ldp+string(os.PathSeparator)) {
								if kitimppath = filepath.Clean(kitdirpath[len(ldp)+1:]); os.PathSeparator != '/' {
									kitimppath = ustr.Replace(kitimppath, string(os.PathSeparator), "/")
								}
								break
							}
						}
						if kitimppath == "" {
							if kitdirpath == me.Dirs.Session {
								kitimppath = "."
							} else {
								panic("the impossible, debug+fix stat")
							}
						}
						me.kits.all = append(me.kits.all, Kit{DirPath: kitdirpath, ImpPath: kitimppath,
							srcFiles: make(atmolang.AstFiles, 0, numfilesguess)})
					}
					shouldrefresh[kitdirpath] = true
				}
				// remove kits that have vanished from the file-system
				var numremoved int
				for i := 0; i < len(me.kits.all); i++ {
					if cur := &me.kits.all[i]; !ufs.IsDir(cur.DirPath) {
						delete(shouldrefresh, cur.DirPath)
						me.kits.all.removeAt(i)
						i, numremoved = i-1, numremoved+1
					}
				}
				// ensure no duplicate imp-paths
				for i := len(me.kits.all) - 1; i >= 0; i-- {
					cur := &me.kits.all[i]
					if idx := me.kits.all.indexImpPath(cur.ImpPath); idx != i {
						delete(shouldrefresh, cur.DirPath)
						delete(shouldrefresh, me.kits.all[idx].DirPath)
						me.msg(true, "duplicate import path `"+cur.ImpPath+"`", "in "+cur.KitsDirPath(), "and "+me.kits.all[idx].KitsDirPath(), "─── both will not load until fixed")
						if idx > i {
							me.kits.all.removeAt(idx)
							me.kits.all.removeAt(i)
						} else {
							me.kits.all.removeAt(i)
							me.kits.all.removeAt(idx)
						}
						i--
					}
				}
				// for stable listings etc.
				sort.Sort(me.kits.all)
				// timing until now, before reloads
				nowtime := time.Now().UnixNano()
				starttime, filemodwatchduration = nowtime, nowtime-starttime
				// per-file refresher
				for kitdirpath := range shouldrefresh {
					me.kitRefresh(me.kits.all.indexDirPath(kitdirpath))
				}
				if me.state.fileModsWatch.emitMsgs {
					me.msg(true, "Modifications in "+ustr.Plu(len(modkitdirs), "kit")+" led to dropping "+ustr.Plu(numremoved, "kit"), "and then (re)loading "+ustr.Plu(len(shouldrefresh), "kit")+", which took "+time.Duration(time.Now().UnixNano()-starttime).String()+".")
				}
			}
			me.state.Unlock()
		}
		const modswatchdurationcritical = int64(23 * time.Millisecond)
		if filemodwatchduration > modswatchdurationcritical {
			me.msg(false, "[DBG] note to dev, mods-watch took "+time.Duration(filemodwatchduration).String())
		}
	})
	if modswatchstart, modswatchcancel := ustd.DoNowAndThenEvery(KitsWatchInterval, me.OngoingKitsWatch.ShouldNow, func() { _ = modswatcher() }); modswatchstart != nil {
		me.state.fileModsWatch.runningAutomaticallyPeriodically, me.state.cleanUps =
			true, append(me.state.cleanUps, modswatchcancel)
		go modswatchstart()
	} else {
		me.state.fileModsWatch.emitMsgs, me.state.fileModsWatch.doManually = true, modswatcher
	}
}

func (me kits) Len() int          { return len(me) }
func (me kits) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }
func (me kits) Less(i int, j int) bool {
	pi, pj := &me[i], &me[j]
	if pi.DirPath != pj.DirPath {
		if piau, pjau := (pi.ImpPath == atmo.NameAutoKit), (pj.ImpPath == atmo.NameAutoKit); piau || pjau {
			return piau || !pjau
		}
	}
	return pi.DirPath < pj.DirPath
}

func (me *kits) removeAt(idx int) {
	this := *me
	for i := idx; i < len(this)-1; i++ {
		this[i] = this[i+1]
	}
	this = this[:len(this)-1]
	*me = this
}

func (me kits) indexDirPath(dirPath string) int {
	for i := range me {
		if me[i].DirPath == dirPath {
			return i
		}
	}
	return -1
}

func (me kits) indexImpPath(impPath string) int {
	for i := range me {
		if me[i].ImpPath == impPath {
			return i
		}
	}
	return -1
}

func KitsDirPathFrom(kitDirPath string, kitImpPath string) string {
	return filepath.Clean(kitDirPath[:len(kitDirPath)-len(kitImpPath)])
}
