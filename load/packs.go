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

type packs []Pack

var PacksWatchInterval time.Duration

func init() { ufs.WalkReadDirFunc = ufs.Dir }

func (me *Ctx) WithKnownPacks(do func([]Pack)) {
	me.maybeInitPanic(false)
	me.state.Lock()
	do(me.packs.all)
	me.state.Unlock()
	return
}

func (me *Ctx) WithKnownPacksWhere(where func(*Pack) bool, do func([]*Pack)) {
	me.maybeInitPanic(false)
	me.state.Lock()
	doall, packs := where == nil, make([]*Pack, 0, len(me.packs.all))
	for i := range me.packs.all {
		if pack := &me.packs.all[i]; doall || where(pack) {
			packs = append(packs, pack)
		}
	}
	do(packs)
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

func (me *Ctx) ReloadModifiedPacksUnlessAlreadyWatching() (numFileSystemModsNoticedAndActedUpon int) {
	me.maybeInitPanic(false)
	if me.state.fileModsWatch.doManually != nil {
		numFileSystemModsNoticedAndActedUpon = me.state.fileModsWatch.doManually()
	}
	return
}

func (me *Ctx) initPacks() {
	var handledir func(string, map[string]int)
	handledir = func(dirfullpath string, modpackdirs map[string]int) {
		if idx := me.packs.all.indexDirPath(dirfullpath); idx >= 0 { // dir was previously known as a pack
			modpackdirs[dirfullpath] = cap(me.packs.all[idx].srcFiles)
		} else if dirfullpath == me.Dirs.Session {
			modpackdirs[dirfullpath] = 1
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

	var watchdircur []string
	if !me.Dirs.curAlreadyInPacksDirs {
		watchdircur = []string{me.Dirs.Session}
	}
	modswatcher := ufs.ModificationsWatcher(PacksWatchInterval/2, me.Dirs.Packs, watchdircur, atmo.SrcFileExt, func(mods map[string]os.FileInfo, starttime int64) {
		var filemodwatchduration int64
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

			if len(me.packs.all) == 0 && !me.Dirs.curAlreadyInPacksDirs {
				modpackdirs[me.Dirs.Session] = 1
			}
			if filemodwatchduration = time.Now().UnixNano() - starttime; len(modpackdirs) > 0 {
				shouldrefresh := make(map[string]bool, len(modpackdirs))
				// handle new-or-modified packs
				for packdirpath, numfilesguess := range modpackdirs {
					if me.packs.all.indexDirPath(packdirpath) < 0 {
						if numfilesguess < 4 {
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
							if packdirpath == me.Dirs.Session {
								packimppath = "."
							} else {
								panic("the impossible, debug+fix stat")
							}
						}
						me.packs.all = append(me.packs.all, Pack{DirPath: packdirpath, ImpPath: packimppath,
							srcFiles: make(atmolang.AstFiles, 0, numfilesguess)})
					}
					shouldrefresh[packdirpath] = true
				}
				// remove packs that have vanished from the file-system
				var numremoved int
				for i := 0; i < len(me.packs.all); i++ {
					if cur := &me.packs.all[i]; !ufs.IsDir(cur.DirPath) {
						delete(shouldrefresh, cur.DirPath)
						me.packs.all.removeAt(i)
						i, numremoved = i-1, numremoved+1
					}
				}
				// ensure no duplicate imp-paths
				for i := len(me.packs.all) - 1; i >= 0; i-- {
					cur := &me.packs.all[i]
					if idx := me.packs.all.indexImpPath(cur.ImpPath); idx != i {
						delete(shouldrefresh, cur.DirPath)
						delete(shouldrefresh, me.packs.all[idx].DirPath)
						me.msg(true, "duplicate import path `"+cur.ImpPath+"`", "in "+cur.PacksDirPath(), "and "+me.packs.all[idx].PacksDirPath(), "─── both will not load until fixed")
						if idx > i {
							me.packs.all.removeAt(idx)
							me.packs.all.removeAt(i)
						} else {
							me.packs.all.removeAt(i)
							me.packs.all.removeAt(idx)
						}
						i--
					}
				}
				// for stable listings etc.
				sort.Sort(me.packs.all)
				// timing until now, before reloads
				nowtime := time.Now().UnixNano()
				starttime, filemodwatchduration = nowtime, nowtime-starttime
				// per-file refresher
				for packdirpath := range shouldrefresh {
					me.packRefresh(me.packs.all.indexDirPath(packdirpath))
				}
				if me.state.fileModsWatch.emitMsgs {
					me.msg(true, "Modifications in "+ustr.Plu(len(modpackdirs), "pack")+" led to dropping "+ustr.Plu(numremoved, "pack"), "and then (re)loading "+ustr.Plu(len(shouldrefresh), "pack")+", which took "+time.Duration(time.Now().UnixNano()-starttime).String()+".")
				}
			}
			me.state.Unlock()
		}
		const modswatchdurationcritical = int64(23 * time.Millisecond)
		if filemodwatchduration > modswatchdurationcritical {
			me.msg(false, "[DBG] note to dev, mods-watch took "+time.Duration(filemodwatchduration).String())
		}
	})
	if modswatchcancel := ustd.DoNowAndThenEvery(PacksWatchInterval, me.OngoingPacksWatch.ShouldNow, func() { _ = modswatcher() }); modswatchcancel != nil {
		me.state.fileModsWatch.runningAutomaticallyPeriodically, me.state.cleanUps =
			true, append(me.state.cleanUps, modswatchcancel)
	} else {
		me.state.fileModsWatch.emitMsgs, me.state.fileModsWatch.doManually = true, modswatcher
	}
}

func (me packs) Len() int          { return len(me) }
func (me packs) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }
func (me packs) Less(i int, j int) bool {
	pi, pj := &me[i], &me[j]
	if pi.DirPath != pj.DirPath {
		if piau, pjau := (pi.ImpPath == atmo.NameAutoPack), (pj.ImpPath == atmo.NameAutoPack); piau || pjau {
			return piau || !pjau
		}
	}
	return pi.DirPath < pj.DirPath
}

func (me *packs) removeAt(idx int) {
	this := *me
	for i := idx; i < len(this)-1; i++ {
		this[i] = this[i+1]
	}
	this = this[:len(this)-1]
	*me = this
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

func PacksDirPathFrom(packDirPath string, packImpPath string) string {
	return filepath.Clean(packDirPath[:len(packDirPath)-len(packImpPath)])
}
