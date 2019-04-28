package atmosess

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-leap/fs"
	"github.com/go-leap/str"
	"github.com/go-leap/sys"
	"github.com/metaleap/atmo"
)

type CtxBgMsg struct {
	Issue bool
	Time  time.Time
	Lines []string
}

type Ctx struct {
	ClearCacheDir bool
	Dirs          struct {
		Session               string
		Cache                 string
		Kits                  []string
		sessAlreadyInKitsDirs bool
	}
	OngoingKitsWatch struct {
		ShouldNow func() bool
	}

	kits struct {
		all Kits
	}
	state struct {
		sync.Mutex
		initCalled    bool
		cleanUps      []func()
		bgMsgs        []CtxBgMsg
		fileModsWatch struct {
			runningAutomaticallyPeriodically bool
			doManually                       func() int
			emitMsgs                         bool
		}
	}
}

func CtxDefaultCacheDirPath() string {
	return filepath.Join(usys.UserDataDirPath(true), "atmo")
}

func (me *Ctx) maybeInitPanic(initingNow bool) {
	if me.state.initCalled == initingNow {
		panic("atmo.Ctx.Init must be called exactly once only")
	}
}

func (me *Ctx) Init() (err error) {
	me.maybeInitPanic(true)
	dirsession := me.Dirs.Session
	if me.state.initCalled, me.kits.all = true, make(Kits, 0, 32); dirsession == "" || dirsession == "." {
		dirsession, err = os.Getwd()
	} else if dirsession[0] == '~' {
		if len(dirsession) == 1 {
			dirsession = usys.UserHomeDirPath()
		} else if dirsession[1] == filepath.Separator {
			dirsession = filepath.Join(usys.UserHomeDirPath(), dirsession[2:])
		}
	}
	if err == nil {
		dirsession, err = filepath.Abs(dirsession)
	}
	if err == nil && !ufs.IsDir(dirsession) {
		err = &os.PathError{Path: dirsession, Op: "directory", Err: os.ErrNotExist}
	}
	if cachedir := me.Dirs.Cache; err == nil {
		if cachedir == "" {
			cachedir = CtxDefaultCacheDirPath()
		}
		if !ufs.IsDir(cachedir) {
			err = ufs.EnsureDir(cachedir)
		} else if me.ClearCacheDir {
			err = ufs.Del(cachedir)
		}
		if kitsdirs := me.Dirs.Kits; err == nil {
			kitsdirsenv := ustr.Split(os.Getenv(atmo.EnvVarKitsDirs), string(os.PathListSeparator))
			for i := range kitsdirsenv {
				kitsdirsenv[i] = filepath.Clean(kitsdirsenv[i])
			}
			for i := range kitsdirs {
				kitsdirs[i] = filepath.Clean(kitsdirs[i])
			}
			kitsdirsorig := kitsdirs
			kitsdirs = ustr.Merge(kitsdirsenv, kitsdirs, func(ldp string) bool {
				if ldp != "" && !ufs.IsDir(ldp) {
					me.bgMsg(true, true, "kits-dir "+ldp+" not found")
					return true
				}
				return ldp == ""
			})
			for i := range kitsdirs {
				for j := range kitsdirs {
					if iinj, jini := ustr.Pref(kitsdirs[i], kitsdirs[j]), ustr.Pref(kitsdirs[j], kitsdirs[i]); i != j && (iinj || jini) {
						err = errors.New("conflicting kits-dirs: " + kitsdirs[i] + " vs. " + kitsdirs[j])
						break
					}
				}
				if err != nil {
					break
				}
			}
			if err == nil && len(kitsdirs) == 0 {
				if kitsdirstried := append(kitsdirsenv, kitsdirsorig...); len(kitsdirstried) == 0 {
					err = errors.New("no kits-dirs were specified, neither via env-var " + atmo.EnvVarKitsDirs + " nor via command-line flags")
				} else {
					err = errors.New("none of the specified kits-dirs were found:\n    " + ustr.Join(kitsdirstried, "\n    "))
				}
			}
			if err == nil {
				me.Dirs.sessAlreadyInKitsDirs = false
				for _, kitsdirpath := range kitsdirs {
					if me.Dirs.sessAlreadyInKitsDirs = ustr.Pref(dirsession, kitsdirpath+string(os.PathSeparator)); me.Dirs.sessAlreadyInKitsDirs {
						break
					}
				}
				me.Dirs.Session, me.Dirs.Cache, me.Dirs.Kits = dirsession, cachedir, kitsdirs
				me.initKits()
			}
		}
	}
	if err != nil {
		me.state.initCalled = false
	}
	return
}

func (me *Ctx) ReadEvalPrint(in string) (out fmt.Stringer, err error) {
	me.maybeInitPanic(false)
	err = fmt.Errorf("to-do: evaluation of %q", in)
	return
}

func (me *Ctx) Dispose() {
	me.maybeInitPanic(false)
	for _, cleanup := range me.state.cleanUps {
		if cleanup != nil {
			cleanup()
		}
	}
	me.state.cleanUps = nil
}

func (me *Ctx) bgMsg(alreadyLocked bool, issue bool, lines ...string) {
	msg := CtxBgMsg{Issue: issue, Time: time.Now(), Lines: lines}
	if !alreadyLocked {
		me.state.Lock()
	}
	me.state.bgMsgs = append(me.state.bgMsgs, msg)
	if !alreadyLocked {
		me.state.Unlock()
	}
}

func (me *Ctx) BackgroundMessages(clear bool) (msgs []CtxBgMsg) {
	me.maybeInitPanic(false)
	me.state.Lock()
	if msgs = me.state.bgMsgs; clear {
		me.state.bgMsgs = nil
	}
	me.state.Unlock()
	return
}
