package atmosess

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/go-leap/fs"
	"github.com/go-leap/str"
	"github.com/go-leap/sys"
	. "github.com/metaleap/atmo"
	. "github.com/metaleap/atmo/ast"
)

// CtxDefaultCacheDirPath returns the default used by `Ctx.Init` if `Ctx.Dirs.Cache`
// was left empty. It returns a platform-specific dir path such as `~/.cache/atmo`,
// `~/.config/atmo` etc. or in the worst case the current user's home directory.
func CtxDefaultCacheDirPath() string {
	return filepath.Join(usys.UserDataDirPath(true), "atmo")
}

// Init validates the `Ctx.Dirs` fields currently set, then builds up its
// `Kits` reflective of the structures found in the various `me.Dirs.Kits`
// search paths and from now on in sync with live modifications to those.
func (me *Ctx) Init(clearCacheDir bool, sessionFauxKitDir string) (kitImpPathIfFauxKitDirActualKit string, err *Error) {
	me.Kits.All = make(Kits, 0, 32)
	cachedir := me.Dirs.CacheData
	if cachedir == "" {
		cachedir = CtxDefaultCacheDirPath()
	}
	if !ufs.IsDir(cachedir) {
		err = ErrFrom(ErrCatSess, ErrSessInit_IoCacheDirCreationFailure, cachedir, ufs.EnsureDir(cachedir))
	} else if clearCacheDir {
		err = ErrFrom(ErrCatSess, ErrSessInit_IoCacheDirDeletionFailure, cachedir, ufs.Del(cachedir))
	}
	if kitsdirs := me.Dirs.KitsStashes; err == nil {
		kitsdirsenv := ustr.Split(os.Getenv(EnvVarKitsDirs), string(os.PathListSeparator))
		kitsdirdefault := filepath.Join(usys.UserHomeDirPath(), ".atmo")
		for i := range kitsdirsenv {
			kitsdirsenv[i] = filepath.Clean(kitsdirsenv[i])
		}
		for i := range kitsdirs {
			kitsdirs[i] = filepath.Clean(kitsdirs[i])
		}
		kitsdirsorig := kitsdirs
		if tmp := kitsdirdefault + string(filepath.Separator); ustr.Index(kitsdirs, func(s string) bool {
			return s == kitsdirdefault || ustr.Pref(s, tmp)
		}) < 0 {
			kitsdirs = append(kitsdirs, kitsdirdefault)
		}
		kitsdirs = ustr.Merge(kitsdirsenv, kitsdirs, func(ldp string) bool {
			if ldp != "" && !ufs.IsDir(ldp) {
				if ldp != kitsdirdefault {
					me.bgMsg(true, "kitstash dir "+ldp+" not found")
				}
				return true
			}
			return ldp == ""
		})
		for i := range kitsdirs {
			for j := range kitsdirs {
				if iinj, jini := ustr.Pref(kitsdirs[i], kitsdirs[j]), ustr.Pref(kitsdirs[j], kitsdirs[i]); i != j && (iinj || jini) {
					err = ErrSess(ErrSessInit_KitsDirsConflict, "", "conflicting kitstash dirs, because one contains the other: `"+kitsdirs[i]+"` vs. `"+kitsdirs[j]+"`")
					break
				}
			}
			if err != nil {
				break
			}
		}
		if err == nil && len(kitsdirs) == 0 {
			if kitsdirstried := append(kitsdirsenv, kitsdirsorig...); len(kitsdirstried) == 0 {
				err = ErrSess(ErrSessInit_KitsDirsNotSpecified, "", "no kitstash dirs were specified, neither via env-var "+EnvVarKitsDirs+" nor via command-line flags")
			} else {
				err = ErrSess(ErrSessInit_KitsDirsNotFound, "", "none of the specified kitstash dirs were found:\n    "+ustr.Join(kitsdirstried, "\n    "))
			}
		}
		if err == nil {
			var autokitexists bool
			for _, kd := range kitsdirs {
				if autokitexists = ufs.DoesDirHaveFilesWithSuffix(filepath.Join(kd, NameAutoKit), SrcFileExt); autokitexists {
					break
				}
			}
			if !autokitexists {
				err = ErrSess(ErrSessInit_KitsDirAutoNotFound, "", "Standard auto-imported kit `"+NameAutoKit+"` not found in any of:\n    "+ustr.Join(kitsdirs, "\n    "))
			}
		}
		if err == nil {
			if me.Dirs.CacheData, me.Dirs.KitsStashes = cachedir, kitsdirs; len(sessionFauxKitDir) != 0 {
				_, kip, e := me.fauxKitsAddDir(sessionFauxKitDir, true)
				kitImpPathIfFauxKitDirActualKit, err = kip, ErrFrom(ErrCatSess, ErrSessInit_IoFauxKitDirFailure, sessionFauxKitDir, e)
			}
		}
		if err == nil {
			me.initKits()
		}
	}
	return
}

func (me *Ctx) FauxKitsAdd(dirPath string) (is bool, err error) {
	was := ustr.In(dirPath, me.Dirs.fauxKits...)
	if is = was; !is {
		is, _, err = me.fauxKitsAddDir(dirPath, false)
	}
	if is && !was {
		me.CatchUpOnFileMods()
	}
	return
}

func (me *Ctx) fauxKitsAddDir(dirPath string, forceAcceptEvenIfNoSrcFiles bool) (dirHasSrcFiles bool, existingKitImpPath string, err error) {
	if dirPath == "" || dirPath == "." {
		dirPath, err = os.Getwd()
	} else if dirPath[0] == '~' {
		if len(dirPath) == 1 {
			dirPath = usys.UserHomeDirPath()
		} else if dirPath[1] == filepath.Separator {
			dirPath = filepath.Join(usys.UserHomeDirPath(), dirPath[2:])
		}
	}
	if err == nil {
		dirPath, err = filepath.Abs(dirPath)
	}
	if err == nil && !ufs.IsDir(dirPath) {
		err = &os.PathError{Path: dirPath, Op: "directory", Err: os.ErrNotExist}
	}
	if err == nil {
		if dirHasSrcFiles = ufs.DoesDirHaveFilesWithSuffix(dirPath, SrcFileExt); dirHasSrcFiles || forceAcceptEvenIfNoSrcFiles {
			var in bool
			for _, kitsdirpath := range me.Dirs.KitsStashes {
				pref := kitsdirpath + string(os.PathSeparator)
				if in = ustr.Pref(dirPath, pref); in {
					existingKitImpPath = dirPath[len(pref):]
					break
				}
			}
			if !in {
				in = ustr.In(dirPath, me.Dirs.fauxKits...)
			}
			if !in {
				me.Dirs.fauxKits = append(me.Dirs.fauxKits, dirPath)
			}
		}
	}
	return
}

func (me *Ctx) FauxKitsHas(dirPath string) bool {
	return ustr.In(dirPath, me.Dirs.fauxKits...)
}

func (me *Ctx) bgMsg(issue bool, lines ...string) {
	msg := ctxBgMsg{Issue: issue, Time: time.Now(), Lines: lines}
	me.state.bgMsgs = append(me.state.bgMsgs, msg)
	if me.On.NewBackgroundMessages != nil {
		me.On.NewBackgroundMessages(me)
	}
}

func (me *Ctx) BackgroundMessages(clear bool) (msgs []ctxBgMsg) {
	if msgs = me.state.bgMsgs; clear {
		me.state.bgMsgs = nil
	}
	return
}

func (me *Ctx) BackgroundMessagesCount() (count int) {
	count = len(me.state.bgMsgs)
	return
}

func (me *Ctx) onSomeOrAllKitsPartiallyOrFullyRefreshed(freshStage1Errs Errors, freshStage2AndBeyondErrs Errors) {
	me.Kits.All.ensureErrTldPosOffsets()
	hadfresherrs := len(freshStage1Errs) != 0 || len(freshStage2AndBeyondErrs) != 0
	if hadfresherrs {
		if me.Options.BgMsgs.IncludeLiveKitsErrs {
			for _, e := range freshStage1Errs {
				if pos := e.Pos(); pos == nil || (pos.FilePath != "" && pos.FilePath != me.Options.Scratchpad.FauxFileNameForErrorMessages) {
					me.bgMsg(true, e.Error())
				}
			}
			SortMaybe(freshStage2AndBeyondErrs)
			for _, e := range freshStage2AndBeyondErrs {
				if pos := e.Pos(); pos == nil || (pos.FilePath != "" && pos.FilePath != me.Options.Scratchpad.FauxFileNameForErrorMessages) {
					me.bgMsg(true, e.Error())
				}
			}
		}
	}
	if me.On.SomeKitsRefreshed != nil {
		me.On.SomeKitsRefreshed(me, hadfresherrs)
	}
}

func (me *Ctx) CatchUpOnFileMods(ensureFilesMarkedAsChanged ...*AstFile) {
	if me.Options.FileModsCatchup.BurstLimit > 0 {
		now := time.Now()
		if (!me.state.fileModsWatch.lastCatchup.IsZero()) &&
			now.Sub(me.state.fileModsWatch.lastCatchup) < me.Options.FileModsCatchup.BurstLimit {
			return
		}
		me.state.fileModsWatch.lastCatchup = now
	}
	me.catchUpOnFileMods(nil, ensureFilesMarkedAsChanged...)
}

func (me *Ctx) catchUpOnFileMods(forceFor *Kit, ensureFilesMarkedAsChanged ...*AstFile) {
	me.state.fileModsWatch.collectFileModsForNextCatchup(me.Dirs.KitsStashes, me.Dirs.fauxKits)

	var latest []map[string]os.FileInfo
	latest, me.state.fileModsWatch.latest = me.state.fileModsWatch.latest, nil

	if len(ensureFilesMarkedAsChanged) != 0 {
		extra := make(map[string]os.FileInfo, len(ensureFilesMarkedAsChanged))
		for _, srcfile := range ensureFilesMarkedAsChanged {
			if srcfile.SrcFilePath != "" {
				var have bool
				for _, modset := range latest {
					if _, have = modset[srcfile.SrcFilePath]; have {
						break
					}
				}
				if !have {
					if fileinfo, _ := os.Stat(srcfile.SrcFilePath); fileinfo != nil {
						extra[srcfile.SrcFilePath] = fileinfo
					}
				}
			}
		}
		if len(extra) != 0 {
			latest = append(latest, extra)
		}
	}

	if len(latest) != 0 || forceFor != nil {
		me.fileModsHandle(me.Dirs.KitsStashes, me.Dirs.fauxKits, latest, forceFor)
	}
}

// Locked is never used by `atmosess` itself but a convenience helper for
// outside callers that run parallel code-paths and thus need to serialize
// concurrent accesses to their `Ctx`. Wrap any and all of your `Ctx` uses in
// a `func` passed to `Locked` and concurrent accesses will queue up. Caution:
// calling `Locked` again from  inside such a wrapper `func` will deadlock.
func (me *Ctx) Locked(do func()) {
	me.state.notUsedInternallyButAvailableForOutsideCallersConvenience.Lock()
	defer me.state.notUsedInternallyButAvailableForOutsideCallersConvenience.Unlock()
	do()
}

func (me *Ctx) WithInMemFileMod(srcFilePath string, altSrc string, do func()) (recoveredPanic interface{}) {
	var one map[string]string
	if altSrc != "" {
		one = map[string]string{srcFilePath: altSrc}
	}
	return me.WithInMemFileMods(one, do)
}

func (me *Ctx) WithInMemFileMods(srcFilePathsAndAltSrcs map[string]string, do func()) (recoveredPanic interface{}) {
	if len(srcFilePathsAndAltSrcs) != 0 {
		srcfiles := make(AstFiles, 0, len(srcFilePathsAndAltSrcs))
		restoreFinally := func() {
			if recoveredPanic = recover(); recoveredPanic != nil {
				debug.PrintStack()
			}
			for _, srcfile := range srcfiles {
				srcfile.Options.TmpAltSrc = nil
			}
			me.CatchUpOnFileMods(srcfiles...)
		}
		defer restoreFinally()

		for srcfilepath, altsrc := range srcFilePathsAndAltSrcs {
			if kit := me.KitByDirPath(filepath.Dir(srcfilepath), false); kit != nil {
				if srcfile := kit.SrcFiles.ByFilePath(srcfilepath); srcfile != nil {
					srcfiles, srcfile.Options.TmpAltSrc = append(srcfiles, srcfile), []byte(altsrc)
				}
			}
		}
		me.CatchUpOnFileMods(srcfiles...)
	}
	do()
	return
}
