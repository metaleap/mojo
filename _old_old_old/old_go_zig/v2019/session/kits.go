package atmosess

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-leap/fs"
	"github.com/go-leap/str"
	. "github.com/metaleap/atmo/old/v2019"
	. "github.com/metaleap/atmo/old/v2019/ast"
	. "github.com/metaleap/atmo/old/v2019/il"
)

func init() { ufs.ReadDirFunc = ufs.Dir }

// KnownKitImpPaths returns all the import-paths of all currently known `Kit`s.
func (me *Ctx) KnownKitImpPaths() (kitImpPaths []string) {
	kitImpPaths = make([]string, len(me.Kits.All))
	for i := range me.Kits.All {
		kitImpPaths[i] = me.Kits.All[i].ImpPath
	}
	return
}

func (me *Ctx) initKits() {
	me.state.fileModsWatch.collectFileModsForNextCatchup = ufs.ModificationsWatcher(SrcFileExt, me.fileModsDirOk, 0,
		func(mods map[string]os.FileInfo, starttime int64, wasfirstrun bool) {
			if len(mods) != 0 {
				me.state.fileModsWatch.latest = append(me.state.fileModsWatch.latest, mods)
			}
		},
	)
	me.CatchUpOnFileMods()
}

func (me *Ctx) fileModsHandle(kitsDirs []string, fauxKitDirs []string, latest []map[string]os.FileInfo, forceFor *Kit) {
	var alllatestfilemods map[string]os.FileInfo
	if len(latest) != 0 {
		alllatestfilemods = make(map[string]os.FileInfo, len(latest[0]))
		for i := range latest {
			for k, v := range latest[i] {
				alllatestfilemods[k] = v
			}
		}
	}

	modkitdirs := map[string]int{}
	if forceFor != nil {
		modkitdirs[forceFor.DirPath] = 1
	}
	for fullpath, fileinfo := range alllatestfilemods {
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

	if len(modkitdirs) != 0 || forceFor != nil {
		shouldrefresh := make(StringKeys, 1+len(modkitdirs))
		if forceFor != nil {
			shouldrefresh[forceFor.DirPath] = Є
		}

		// handle new-or-modified kits
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
							kitimppath = kitdirpath
							break
						}
					}
				}
				me.Kits.All = append(me.Kits.All, &Kit{DirPath: kitdirpath, ImpPath: kitimppath,
					SrcFiles: make(AstFiles, 0, numfilesguess)})
			}
			shouldrefresh[kitdirpath] = Є
			for _, dk := range me.Kits.All.DependantKitsOf(me.Kits.All.ByDirPath(kitdirpath).ImpPath, true) {
				shouldrefresh[dk.DirPath] = Є
			}
		}

		// remove kits that have vanished from the file-system
		delkits := StringKeys{}
		for i := 0; i < len(me.Kits.All); i++ {
			if kit := me.Kits.All[i]; (!ustr.In(kit.DirPath, fauxKitDirs...)) && (!ufs.DoesDirHaveFilesWithSuffix(kit.DirPath, SrcFileExt)) {
				delkits[kit.ImpPath] = Є
				delete(shouldrefresh, kit.DirPath)
				me.Kits.All.removeAt(i)
				i--
			}
		}
		for delkitimppath := range delkits {
			for _, k := range me.Kits.All.DependantKitsOf(delkitimppath, true) {
				if _, wasdel := delkits[k.ImpPath]; !wasdel {
					shouldrefresh[k.DirPath] = Є
				}
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
		SortMaybe(me.Kits.All)

		// per-kit refreshers
		var freshstage1errs Errors
		for _, kitdirpath := range shouldrefresh.Sorted(func(kdp1 string, kdp2 string) bool {
			k1, k2 := me.Kits.All.ByDirPath(kdp1), me.Kits.All.ByDirPath(kdp2)
			return k1.ImpPath == NameAutoKit || k2.DoesImport(k1.ImpPath) || !k1.DoesImport(k2.ImpPath)
		}) {
			if kit := me.Kits.All.ByDirPath(kitdirpath); kit != nil {
				freshstage1errs.Add(me.kitRefreshFilesAndMaybeReload(kit, false, false)...)
			} else {
				panic(kitdirpath)
			}
		}
		freshstage2errs := me.reprocessAffectedDefsIfAnyKitsReloaded()

		// reprocess maybe
		me.onSomeOrAllKitsPartiallyOrFullyRefreshed(freshstage1errs, freshstage2errs)
	}
	return
}

func (me *Ctx) fileModsDirOk(kitsDirs []string, fauxKitDirs []string, dirFullPath string, dirName string) bool {
	return ustr.In(dirFullPath, fauxKitDirs...) ||
		ustr.In(dirFullPath, kitsDirs...) ||
		IsValidKitDirName(dirName)
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
		} else if (!isdir) && (!added) && ustr.Suff(fileinfo.Name(), SrcFileExt) && !ustr.In(dirFullPath, kitsDirs...) {
			added, modKitDirs[dirFullPath] = true, modKitDirs[dirFullPath]+1
		}
	}
}

func (me *Ctx) KitsCollectReferences(forceLoadAllKnownKits bool, name string) map[*IrDef][]IIrExpr {
	if name == "" {
		return nil
	}
	if forceLoadAllKnownKits {
		me.KitsEnsureLoaded(true, me.KnownKitImpPaths()...)
	}
	return me.Kits.All.collectReferences(name)
}

func (me *Ctx) KitsCollectAcquaintances(forceLoadAllKnownKits bool, defNames StringKeys, indirects bool) (acquaintancesDefs map[*IrDef]*Kit) {
	if forceLoadAllKnownKits {
		me.KitsEnsureLoaded(true, me.KnownKitImpPaths()...)
	}
	acquaintancesDefs = make(map[*IrDef]*Kit)
	var dones StringKeys
	if indirects {
		dones = make(StringKeys, len(defNames))
	}
	me.Kits.All.collectAcquaintances(defNames, acquaintancesDefs, dones)
	return
}

func (me Kits) SrcFilePaths() (srcFilePaths []string) {
	var count, i int
	for _, kit := range me {
		count += len(kit.SrcFiles)
	}
	srcFilePaths = make([]string, count)

	for _, kit := range me {
		for _, srcfile := range kit.SrcFiles {
			i, srcFilePaths[i] = i+1, srcfile.SrcFilePath
		}
	}
	return
}

func (me *Ctx) reprocessAffectedDefsIfAnyKitsReloaded() (freshErrs Errors) {
	if me.Kits.reprocessingNeeded {
		me.Kits.reprocessingNeeded = false

		timestarted := time.Now().UnixNano()
		namesofchange, _, _, ferrs := me.kitsRepopulateNamesInScope()
		freshErrs = ferrs
		defidsofacquaintancesofnamesofchange := make(map[*IrDef]*Kit)
		me.Kits.All.collectAcquaintances(namesofchange, defidsofacquaintancesofnamesofchange, make(StringKeys, len(namesofchange)))
		freshErrs.Add(me.rePreduceTopLevelDefs(defidsofacquaintancesofnamesofchange)...)
		timedone := time.Now().UnixNano()
		me.bgMsg(false, time.Duration(timedone-timestarted).String()+" for "+ustr.Join(namesofchange.Sorted(nil), " + "))
	}
	return
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
		if pi.ImpPath == NameAutoKit {
			return true
		}
		if pj.ImpPath == NameAutoKit {
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

func (me Kits) collectReferences(name string) (refs map[*IrDef][]IIrExpr) {
	for _, kit := range me {
		for _, tld := range kit.topLevelDefs {
			if nodes := tld.RefsTo(name); len(nodes) != 0 {
				if refs == nil {
					refs = make(map[*IrDef][]IIrExpr)
				}
				refs[tld] = nodes
			}
		}
	}
	return
}

func (me Kits) collectAcquaintances(defNames StringKeys, acquaintancesDefs map[*IrDef]*Kit, doneAlready StringKeys) {
	if len(defNames) == 0 {
		return
	}
	indirects := (doneAlready != nil)
	var morenames StringKeys
	if indirects {
		morenames = make(StringKeys, 2)
	}

	for defname := range defNames {
		if indirects {
			doneAlready[defname] = Є
		}
		for _, kit := range me {
			for _, tld := range kit.topLevelDefs {
				if tld.RefersToOrDefines(defname) {
					if acquaintancesDefs[tld] = kit; indirects {
						if _, doneearlier := doneAlready[tld.Ident.Name]; !doneearlier {
							morenames[tld.Ident.Name] = Є
						}
					}
				}
			}
		}
	}
	if indirects {
		me.collectAcquaintances(morenames, acquaintancesDefs, doneAlready)
	}
}

func (me Kits) ensureErrTldPosOffsets() {
	for _, kit := range me {
		kit.ensureErrTldPosOffsets()
	}
}

func (me Kits) DependantKitsOf(kitImpPath string, indirects bool) (dependantKitImpPaths map[string]*Kit) {
	dependantKitImpPaths = make(map[string]*Kit, 8)
	me.collectDependantKitsOf(kitImpPath, indirects, dependantKitImpPaths, make(StringKeys, 8))
	if len(dependantKitImpPaths) == 0 {
		dependantKitImpPaths = nil
	}
	return
}

func (me Kits) ImportersOf(kitImpPath string) Kits {
	return me.Where(func(k *Kit) bool { return k.DoesImport(kitImpPath) })
}

func (me Kits) collectDependantKitsOf(kitImpPath string, indirects bool, collectDependantKitImpPathsInto map[string]*Kit, alreadyTraversedInto StringKeys) {
	if _, donealready := alreadyTraversedInto[kitImpPath]; !donealready {
		alreadyTraversedInto[kitImpPath] = Є
		for _, kit := range me {
			if kit.DoesImport(kitImpPath) {
				collectDependantKitImpPathsInto[kit.ImpPath] = kit
				if indirects {
					me.collectDependantKitsOf(kit.ImpPath, true, collectDependantKitImpPathsInto, alreadyTraversedInto)
				}
			}
		}
	}
}
