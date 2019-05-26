package atmosess

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-leap/fs"
	"github.com/go-leap/std"
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

// KitsWatchInterval is the default file-watching interval that is picked up
// by each `Ctx.Init` call and then used throughout the `Ctx`'s life time.
var KitsWatchInterval time.Duration

type Kits []*Kit

func init() { ufs.WalkReadDirFunc = ufs.Dir }

// WithKnownKits runs `do` with all currently-known (loaded or not) `Kit`s
// passed to it. The `Kits` slice or its contents must not be written to.
func (me *Ctx) WithKnownKits(do func(Kits)) {
	me.maybeInitPanic(false)
	do(me.Kits.all)
	return
}

// WithKnownKitsWhere works like `WithKnownKits` but with pre-filtering via `where`.
func (me *Ctx) WithKnownKitsWhere(where func(*Kit) bool, do func(Kits)) {
	me.maybeInitPanic(false)
	doall, kits := (where == nil), make(Kits, 0, len(me.Kits.all))
	for i := range me.Kits.all {
		if kit := me.Kits.all[i]; doall || where(kit) {
			kits = append(kits, kit)
		}
	}
	do(kits)
	return
}

// KnownKitImpPaths returns all the import-paths of all currently known `Kit`s.
func (me *Ctx) KnownKitImpPaths() (kitImpPaths []string) {
	me.maybeInitPanic(false)
	kitImpPaths = make([]string, len(me.Kits.all))
	for i := range me.Kits.all {
		kitImpPaths[i] = me.Kits.all[i].ImpPath
	}
	return
}

func (me *Ctx) initKits() {
	checkforfilemodsnow := ufs.ModificationsWatcher(atmo.SrcFileExt, me.fileModsDirOk, 0,
		func(mods map[string]os.FileInfo, starttime int64, wasfirstrun bool) {
			// isconcurrent := me.state.fileModsWatch.runningAutomaticallyPeriodically && (!wasfirstrun)
			if len(mods) > 0 {
				me.state.fileModsWatch.latestMutex.Lock()
				me.state.fileModsWatch.latest = append(me.state.fileModsWatch.latest, mods)
				me.state.fileModsWatch.latestMutex.Unlock()
			}
		},
	)
	if modswatchstart, modswatchcancel := ustd.DoNowAndThenEvery(KitsWatchInterval,
		me.Kits.RecurringBackgroundWatch.ShouldNow,
		func() {
			me.Dirs.fauxKitsMutex.Lock()
			fauxkitdirpaths := me.Dirs.fauxKits
			me.Dirs.fauxKitsMutex.Unlock()
			_ = checkforfilemodsnow(me.Dirs.Kits, fauxkitdirpaths)
		},
	); modswatchstart != nil {
		me.state.fileModsWatch.runningAutomaticallyPeriodically, me.state.cleanUps =
			true, append(me.state.cleanUps, modswatchcancel)
		go modswatchstart()
	} else {
		me.state.fileModsWatch.doManually = checkforfilemodsnow
	}
}

func (me *Ctx) fileModsDirOk(kitsDirs []string, fauxKitDirs []string, dirFullPath string, dirName string) bool {
	return ustr.In(dirFullPath, fauxKitDirs...) || ustr.In(dirFullPath, kitsDirs...) ||
		((!ustr.IsLen1And(dirName, '_', '*', '.', ' ')) && dirName != "·" && (!ustr.HasAnyOf(dirName, ' ')))
}

func (me *Ctx) fileModsHandleDir(kitsDirs []string, fauxKitDirs []string, dirFullPath string, modKitDirs map[string]int) {
	isdirsess := ustr.In(dirFullPath, fauxKitDirs...)
	if idx := me.Kits.all.indexDirPath(dirFullPath); idx >= 0 {
		// dir was previously known as a kit
		modKitDirs[dirFullPath] = cap(me.Kits.all[idx].srcFiles)
	} else if isdirsess {
		// cur sess dir is a (real or faux) "kit"
		modKitDirs[dirFullPath] = 1
	}
	if !isdirsess {
		for i := range me.Kits.all {
			if ustr.Pref(me.Kits.all[i].DirPath, dirFullPath+string(os.PathSeparator)) {
				modKitDirs[me.Kits.all[i].DirPath] = cap(me.Kits.all[i].srcFiles)
			}
		}
	}
	dircontents, _ := ufs.Dir(dirFullPath)
	var added bool
	for _, fileinfo := range dircontents {
		if isdir, fp := fileinfo.IsDir(), filepath.Join(dirFullPath, fileinfo.Name()); isdir && isdirsess {
			// continue next one
		} else if isdir && me.fileModsDirOk(kitsDirs, fauxKitDirs, fp, fileinfo.Name()) {
			me.fileModsHandleDir(kitsDirs, fauxKitDirs, fp, modKitDirs)
		} else if (!isdir) && (!added) && ustr.Suff(fileinfo.Name(), atmo.SrcFileExt) && !ustr.In(dirFullPath, kitsDirs...) {
			added, modKitDirs[dirFullPath] = true, modKitDirs[dirFullPath]+1
		}
	}
}

func (me *Ctx) fileModsHandle(kitsDirs []string, fauxKitDirs []string, latest []map[string]os.FileInfo) {
	for _, latestfilemods := range latest {
		modkitdirs := map[string]int{}
		for fullpath, fileinfo := range latestfilemods {
			if fileinfo.IsDir() {
				me.fileModsHandleDir(kitsDirs, fauxKitDirs, fullpath, modkitdirs)
			} else if dp := filepath.Dir(fullpath); !ustr.In(dp, kitsDirs...) {
				modkitdirs[dp] = modkitdirs[dp] + 1
			}
		}
		if len(me.Kits.all) == 0 {
			for _, dirsess := range fauxKitDirs {
				modkitdirs[dirsess] = 1
			}
		}
		if len(modkitdirs) > 0 {
			shouldrefresh := make(atmo.StringsUnorderedButUnique, len(modkitdirs))
			// handle new-or-modified kits
			// TODO: mark all existing&new direct&indirect dependants (as per Kit.Imports) for full-refresh
			for kitdirpath, numfilesguess := range modkitdirs {
				if me.Kits.all.indexDirPath(kitdirpath) < 0 {
					if numfilesguess < 2 {
						numfilesguess = 2
					}
					var kitimppath string
					for _, ldp := range kitsDirs {
						if ustr.Pref(kitdirpath, ldp+string(os.PathSeparator)) {
							if kitimppath = filepath.Clean(kitdirpath[len(ldp)+1:]); os.PathSeparator != '/' {
								kitimppath = ustr.Replace(kitimppath, string(os.PathSeparator), "/")
							}
							break
						}
					}
					if kitimppath == "" {
						for _, dirsess := range fauxKitDirs {
							if dirsess == kitdirpath {
								kitimppath = ustr.Replace(kitdirpath, "/", "·")
								break
							}
						}
					}
					kitimps := []string{atmo.NameAutoKit}
					if kitimppath == atmo.NameAutoKit {
						kitimps = nil
					}
					me.Kits.all = append(me.Kits.all, &Kit{DirPath: kitdirpath, ImpPath: kitimppath, Imports: kitimps,
						srcFiles: make(atmolang.AstFiles, 0, numfilesguess), defsFacts: make(map[string]*defNameFacts, numfilesguess*8)})
				}
				shouldrefresh[kitdirpath] = atmo.Exists
			}
			// remove kits that have vanished from the file-system
			// TODO: mark all existing&new direct&indirect dependants (as per Kit.Imports) for full-refresh
			var numremoved int
			for i := 0; i < len(me.Kits.all); i++ {
				if kit := me.Kits.all[i]; !ustr.In(kit.DirPath, fauxKitDirs...) &&
					((!ufs.IsDir(kit.DirPath)) || !ufs.HasFilesWithSuffix(kit.DirPath, atmo.SrcFileExt)) {
					delete(shouldrefresh, kit.DirPath)
					me.Kits.all.removeAt(i)
					i, numremoved = i-1, numremoved+1
				}
			}
			// ensure no duplicate imp-paths
			for i := len(me.Kits.all) - 1; i >= 0; i-- {
				kit := me.Kits.all[i]
				if idx := me.Kits.all.indexImpPath(kit.ImpPath); idx != i {
					delete(shouldrefresh, kit.DirPath)
					delete(shouldrefresh, me.Kits.all[idx].DirPath)
					me.bgMsg(true, "duplicate import path `"+kit.ImpPath+"`", "in "+kit.kitsDirPath(), "and "+me.Kits.all[idx].kitsDirPath(), "─── both will not load until fixed")
					if idx > i {
						me.Kits.all.removeAt(idx)
						me.Kits.all.removeAt(i)
					} else {
						me.Kits.all.removeAt(i)
						me.Kits.all.removeAt(idx)
					}
					i--
				}
			}
			// for stable listings etc.
			atmo.SortMaybe(me.Kits.all)
			// per-file refresher
			for kitdirpath := range shouldrefresh {
				if idx := me.Kits.all.indexDirPath(kitdirpath); idx >= 0 {
					me.kitRefreshFilesAndMaybeReload(me.Kits.all[idx], true, false)
				} else {
					panic(kitdirpath)
				}
			}
			me.reprocessAffectedIRsIfAnyKitsReloaded()
			if me.state.fileModsWatch.emitMsgsIfManual {
				me.bgMsg(false, "Modifications in "+ustr.Plu(len(modkitdirs), "kit")+" led to dropping "+ustr.Plu(numremoved, "kit"), "and then (re)loading "+ustr.Plu(len(shouldrefresh), "kit")+".")
			}
		}
	}
}

// KitsReloadModifiedsUnlessAlreadyWatching returns -1 if file-watching is
// enabled, otherwise it scans all currently-known kits-dirs for modifications
// and refreshes the `Ctx`'s internal represenation of `Kits` if any were noted.
func (me *Ctx) KitsReloadModifiedsUnlessAlreadyWatching() (numFileSystemModsNoticedAndActedUpon int) {
	me.maybeInitPanic(false)
	if me.state.fileModsWatch.doManually == nil {
		numFileSystemModsNoticedAndActedUpon = -1
	} else {
		numFileSystemModsNoticedAndActedUpon = me.state.fileModsWatch.doManually(me.Dirs.Kits, me.Dirs.fauxKits)
	}
	return
}

func (me *Ctx) reprocessAffectedIRsIfAnyKitsReloaded() {
	if me.state.kitsReprocessing.needed {
		me.state.kitsReprocessing.needed = false
		me.kitsRepopulateIdentNamesInScope()
		me.onErrs(me.substantiateKitsDefsFactsAsNeeded(), nil)
	}
}

func kitsDirPathFrom(kitDirPath string, kitImpPath string) string {
	return filepath.Clean(kitDirPath[:len(kitDirPath)-len(kitImpPath)])
}

// Len implements Go's standard `sort.Interface`.
func (me Kits) Len() int { return len(me) }

// Swap implements Go's standard `sort.Interface`.
func (me Kits) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }

// Less implements Go's standard `sort.Interface`.
func (me Kits) Less(i int, j int) bool {
	pi, pj := me[i], me[j]
	if pi.DirPath != pj.DirPath {
		if pi.ImpPath == atmo.NameAutoKit {
			return true
		}
		if pj.ImpPath == atmo.NameAutoKit {
			return false
		}
	}
	return pi.DirPath < pj.DirPath
}

func (me *Kits) removeAt(idx int) {
	this := *me
	for i := idx; i < len(this)-1; i++ {
		this[i] = this[i+1]
	}
	this = this[:len(this)-1]
	*me = this
}

func (me Kits) indexDirPath(dirPath string) int {
	for i := range me {
		if me[i].DirPath == dirPath {
			return i
		}
	}
	return -1
}

func (me Kits) indexImpPath(impPath string) int {
	if impPath != "" {
		for i := range me {
			if me[i].ImpPath == impPath {
				return i
			}
		}
	}
	return -1
}

func (me Kits) byDirPath(kitDirPath string) *Kit {
	if idx := me.indexDirPath(kitDirPath); idx >= 0 {
		return me[idx]
	}
	return nil
}

// ByImpPath finds the `Kit` in `Kits` with the given import-path.
func (me Kits) ByImpPath(kitImpPath string) *Kit {
	if idx := me.indexImpPath(kitImpPath); idx >= 0 {
		return me[idx]
	}
	return nil
}

func (me Kits) Where(check func(*Kit) bool) (kits Kits) {
	for _, kit := range me {
		if check(kit) {
			kits = append(kits, kit)
		}
	}
	return
}

func (me Kits) collectReferencers(defNames atmo.StringsUnorderedButUnique, into map[string]*Kit, indirects bool) {
	if len(defNames) == 0 {
		return
	}
	var morenames atmo.StringsUnorderedButUnique
	if indirects {
		morenames = make(atmo.StringsUnorderedButUnique, 4)
	}
	for _, kit := range me {
		for _, tld := range kit.topLevelDefs {
			for defname := range defNames {
				if tld.RefersTo(defname) {
					if into[tld.Id] = kit; indirects {
						morenames[tld.Name.Val] = atmo.Exists
					}
				}
			}
		}
	}
	if indirects {
		me.collectReferencers(morenames, into, true)
	}
}
