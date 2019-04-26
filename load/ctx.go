package atmoload

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

type CtxMsg struct {
	Time time.Time
	Text string
}

type Ctx struct {
	ClearCacheDir bool
	Dirs          struct {
		Session               string
		Cache                 string
		Packs                 []string
		curAlreadyInPacksDirs bool
	}
	OngoingPacksWatch struct {
		ShouldNow func() bool
	}

	packs struct {
		all packs
	}
	state struct {
		sync.Mutex
		initCalled         bool
		cleanUps           []func()
		msgs               []CtxMsg
		modsWatcher        func()
		modsWatcherRunning bool
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
	if me.state.initCalled, me.packs.all = true, make(packs, 0, 32); dirsession == "" || dirsession == "." {
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
		if packsdirs := me.Dirs.Packs; err == nil {
			packsdirsenv := ustr.Split(os.Getenv(atmo.EnvVarPacksDirs), string(os.PathListSeparator))
			for i := range packsdirsenv {
				packsdirsenv[i] = filepath.Clean(packsdirsenv[i])
			}
			for i := range packsdirs {
				packsdirs[i] = filepath.Clean(packsdirs[i])
			}
			packsdirsorig := packsdirs
			packsdirs = ustr.Merge(packsdirsenv, packsdirs, func(ldp string) bool {
				if ldp != "" && !ufs.IsDir(ldp) {
					me.msg(true, "packs dir "+ldp+" not found")
					return true
				}
				return ldp == ""
			})
			for i := range packsdirs {
				for j := range packsdirs {
					if iinj, jini := ustr.Pref(packsdirs[i], packsdirs[j]), ustr.Pref(packsdirs[j], packsdirs[i]); i != j && (iinj || jini) {
						err = errors.New("conflicting packs dirs: " + packsdirs[i] + " vs. " + packsdirs[j])
						break
					}
				}
				if err != nil {
					break
				}
			}
			if err == nil && len(packsdirs) == 0 {
				err = errors.New("none of the specified packs dirs were found:\n    " + ustr.Join(append(packsdirsenv, packsdirsorig...), "\n    "))
			}
			if err == nil {
				me.Dirs.curAlreadyInPacksDirs = false
				for _, packsdirpath := range packsdirs {
					if me.Dirs.curAlreadyInPacksDirs = ustr.Pref(dirsession, packsdirpath+string(os.PathSeparator)); me.Dirs.curAlreadyInPacksDirs {
						break
					}
				}
				me.Dirs.Session, me.Dirs.Cache, me.Dirs.Packs = dirsession, cachedir, packsdirs
				me.initPacks()
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

func (me *Ctx) msg(alreadyLocked bool, text string) {
	msg := CtxMsg{Time: time.Now(), Text: text}
	if !alreadyLocked {
		me.state.Lock()
	}
	me.state.msgs = append(me.state.msgs, msg)
	if !alreadyLocked {
		me.state.Unlock()
	}
}

func (me *Ctx) Messages(clear bool) (msgs []CtxMsg) {
	me.maybeInitPanic(false)
	me.state.Lock()
	if msgs = me.state.msgs; clear {
		me.state.msgs = nil
	}
	me.state.Unlock()
	return
}
