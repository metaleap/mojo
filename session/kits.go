package atmosess

import (
	"os"
	"path/filepath"

	"github.com/go-leap/fs"
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
	"github.com/metaleap/atmo/lang"
)

type Kits []*Kit

func init() { ufs.ReadDirFunc = ufs.Dir }

// KnownKitImpPaths returns all the import-paths of all currently known `Kit`s.
func (me *Ctx) KnownKitImpPaths() (kitImpPaths []string) {
	me.maybeInitPanic(false)
	kitImpPaths = make([]string, len(me.Kits.All))
	for i := range me.Kits.All {
		kitImpPaths[i] = me.Kits.All[i].ImpPath
	}
	return
}

func (me *Ctx) initKits() {
	me.state.fileModsWatch.collectFileModsForNextCatchup = ufs.ModificationsWatcher(atmo.SrcFileExt, me.fileModsDirOk, 0,
		func(mods map[string]os.FileInfo, starttime int64, wasfirstrun bool) {
			if len(mods) > 0 {
				me.state.fileModsWatch.latest = append(me.state.fileModsWatch.latest, mods)
			}
		},
	)
	me.CatchUpOnFileMods()
}

func (me *Ctx) fileModsHandle(kitsDirs []string, fauxKitDirs []string, latest []map[string]os.FileInfo) {
	latestfilemods := make(map[string]os.FileInfo, len(latest[0]))
	for i := range latest {
		for k, v := range latest[i] {
			latestfilemods[k] = v
		}
	}

	modkitdirs := map[string]int{}
	for fullpath, fileinfo := range latestfilemods {
		if fileinfo.IsDir() {
			me.fileModsHandleDir(kitsDirs, fauxKitDirs, fullpath, modkitdirs)
		} else if dp := filepath.Dir(fullpath); !ustr.In(dp, kitsDirs...) {
			modkitdirs[dp] = modkitdirs[dp] + 1
		}
	}
	if len(me.Kits.All) == 0 {
		for _, dirsess := range fauxKitDirs {
			modkitdirs[dirsess] = 1
		}
	}

	if len(modkitdirs) > 0 {
		shouldrefresh := make(atmo.StringKeys, len(modkitdirs))
		// handle new-or-modified kits
		// TODO: mark all existing&new direct&indirect dependants (as per Kit.Imports) for full-refresh
		for kitdirpath, numfilesguess := range modkitdirs {
			if me.Kits.All.IndexDirPath(kitdirpath) < 0 {
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
				me.Kits.All = append(me.Kits.All, &Kit{DirPath: kitdirpath, ImpPath: kitimppath, Imports: kitimps,
					SrcFiles: make(atmolang.AstFiles, 0, numfilesguess)})
			}
			shouldrefresh[kitdirpath] = atmo.Є
		}
		// remove kits that have vanished from the file-system
		// TODO: mark all existing&new direct&indirect dependants (as per Kit.Imports) for full-refresh
		var numremoved int
		for i := 0; i < len(me.Kits.All); i++ {
			if kit := me.Kits.All[i]; (!ustr.In(kit.DirPath, fauxKitDirs...)) && (!ufs.DoesDirHaveFilesWithSuffix(kit.DirPath, atmo.SrcFileExt)) {
				delete(shouldrefresh, kit.DirPath)
				me.Kits.All.removeAt(i)
				i, numremoved = i-1, numremoved+1
			}
		}
		// ensure no duplicate imp-paths
		for i := len(me.Kits.All) - 1; i >= 0; i-- {
			kit := me.Kits.All[i]
			if idx := me.Kits.All.IndexImpPath(kit.ImpPath); idx != i {
				delete(shouldrefresh, kit.DirPath)
				delete(shouldrefresh, me.Kits.All[idx].DirPath)
				me.bgMsg(true, "duplicate import path `"+kit.ImpPath+"`", "in "+kit.kitsDirPath(), "and "+me.Kits.All[idx].kitsDirPath(), "─── both will not load until fixed")
				if idx > i {
					me.Kits.All.removeAt(idx)
					me.Kits.All.removeAt(i)
				} else {
					me.Kits.All.removeAt(i)
					me.Kits.All.removeAt(idx)
				}
				i--
			}
		}
		// for stable listings etc.
		atmo.SortMaybe(me.Kits.All)
		// per-file refresher
		for kitdirpath := range shouldrefresh {
			if idx := me.Kits.All.IndexDirPath(kitdirpath); idx >= 0 {
				me.kitRefreshFilesAndMaybeReload(me.Kits.All[idx], false)
			} else {
				panic(kitdirpath)
			}
		}
		me.reprocessAffectedDefsIfAnyKitsReloaded()
	}
}

func (me *Ctx) fileModsDirOk(kitsDirs []string, fauxKitDirs []string, dirFullPath string, dirName string) bool {
	return ustr.In(dirFullPath, fauxKitDirs...) || ustr.In(dirFullPath, kitsDirs...) ||
		((!ustr.IsLen1And(dirName, '_', '*', '.', ' ')) && dirName != "·" && (!ustr.HasAnyOf(dirName, ' ')))
}

func (me *Ctx) fileModsHandleDir(kitsDirs []string, fauxKitDirs []string, dirFullPath string, modKitDirs map[string]int) {
	isdirsess := ustr.In(dirFullPath, fauxKitDirs...)
	if idx := me.Kits.All.IndexDirPath(dirFullPath); idx >= 0 {
		// dir was previously known as a kit
		modKitDirs[dirFullPath] = cap(me.Kits.All[idx].SrcFiles)
	} else if isdirsess {
		// cur sess dir is a (real or faux) "kit"
		modKitDirs[dirFullPath] = 1
	}
	if !isdirsess {
		for i := range me.Kits.All {
			if ustr.Pref(me.Kits.All[i].DirPath, dirFullPath+string(os.PathSeparator)) {
				modKitDirs[me.Kits.All[i].DirPath] = cap(me.Kits.All[i].SrcFiles)
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

func (me *Ctx) KitsCollectReferences(forceLoadAllKnownKits bool, name string) map[*atmoil.AstDefTop][]atmoil.IAstExpr {
	if name == "" {
		return nil
	}
	if forceLoadAllKnownKits {
		me.KitsEnsureLoaded(true, me.KnownKitImpPaths()...)
	}
	return me.Kits.All.collectReferences(name)
}

func (me *Ctx) KitsCollectDependants(forceLoadAllKnownKits bool, defNames atmo.StringKeys, indirects bool) (dependantsDefIds map[string]*Kit) {
	if forceLoadAllKnownKits {
		me.KitsEnsureLoaded(true, me.KnownKitImpPaths()...)
	}
	dependantsDefIds = make(map[string]*Kit)
	var dones atmo.StringKeys
	if indirects {
		dones = make(atmo.StringKeys, len(defNames))
	}
	me.Kits.All.collectDependants(defNames, dependantsDefIds, dones)
	return
}

func (me Kits) SrcFilePaths() (srcFilePaths []string) {
	var count int
	for _, kit := range me {
		count += len(kit.SrcFiles)
	}
	srcFilePaths = make([]string, 0, count)

	for _, kit := range me {
		for _, srcfile := range kit.SrcFiles {
			srcFilePaths = append(srcFilePaths, srcfile.SrcFilePath)
		}
	}
	return
}

func (me *Ctx) KitsReloadModifiedsUnlessAlreadyWatching() {
	me.maybeInitPanic(false)
	me.CatchUpOnFileMods()
}

func (me *Ctx) reprocessAffectedDefsIfAnyKitsReloaded() {
	if me.Kits.reprocessingNeeded {
		me.Kits.reprocessingNeeded = false

		namesofchange, defidsborn, _, fresherrs := me.kitsRepopulateAstNamesInScope()
		defidsdependantsofnamesofchange := make(map[string]*Kit)
		me.Kits.All.collectDependants(namesofchange, defidsdependantsofnamesofchange, make(atmo.StringKeys, len(namesofchange)))

		fresherrs.Add(me.refreshFactsForTopLevelDefs(defidsborn, defidsdependantsofnamesofchange))

		me.Kits.All.ensureErrTldPosOffsets()
		me.onErrs(fresherrs, nil)
		if me.Kits.OnSomeReprocessed != nil {
			me.Kits.OnSomeReprocessed()
		}
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

func (me Kits) IndexDirPath(dirPath string) int {
	for i := range me {
		if me[i].DirPath == dirPath {
			return i
		}
	}
	return -1
}

func (me Kits) IndexImpPath(impPath string) int {
	if impPath != "" {
		for i := range me {
			if me[i].ImpPath == impPath {
				return i
			}
		}
	}
	return -1
}

func (me Kits) ByDirPath(kitDirPath string) *Kit {
	if idx := me.IndexDirPath(kitDirPath); idx >= 0 {
		return me[idx]
	}
	return nil
}

// ByImpPath finds the `Kit` in `Kits` with the given import-path.
func (me Kits) ByImpPath(kitImpPath string) *Kit {
	if idx := me.IndexImpPath(kitImpPath); idx >= 0 {
		return me[idx]
	}
	return nil
}

func (me Kits) Where(check func(*Kit) bool) (kits Kits) {
	if check == nil {
		return me
	}
	kits = make(Kits, 0, len(me))
	for _, kit := range me {
		if check(kit) {
			kits = append(kits, kit)
		}
	}
	return
}

func (me Kits) collectReferences(name string) (refs map[*atmoil.AstDefTop][]atmoil.IAstExpr) {
	for _, kit := range me {
		for _, tld := range kit.topLevelDefs {
			if nodes := tld.RefsTo(name); len(nodes) > 0 {
				if refs == nil {
					refs = make(map[*atmoil.AstDefTop][]atmoil.IAstExpr)
				}
				refs[tld] = nodes
			}
		}
	}
	return
}

func (me Kits) collectDependants(defNames atmo.StringKeys, dependantsDefIds map[string]*Kit, doneAlready atmo.StringKeys) {
	if len(defNames) == 0 {
		return
	}
	indirects := (doneAlready != nil)
	var morenames atmo.StringKeys
	if indirects {
		morenames = make(atmo.StringKeys, 4)
	}

	for defname := range defNames {
		if indirects {
			doneAlready[defname] = atmo.Є
		}
		for _, kit := range me {
			for _, tld := range kit.topLevelDefs {
				if tld.RefersTo(defname) {
					if dependantsDefIds[tld.Id] = kit; indirects {
						if _, doneearlier := doneAlready[tld.Name.Val]; !doneearlier {
							morenames[tld.Name.Val] = atmo.Є
						}
					}
				}
			}
		}
	}
	if indirects {
		me.collectDependants(morenames, dependantsDefIds, doneAlready)
	}
}

func (me Kits) ensureErrTldPosOffsets() {
	for _, kit := range me {
		kit.ensureErrTldPosOffsets()
	}
}
