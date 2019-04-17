package atmo

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
)

type CtxMsg struct {
	Time time.Time
	Text string
}

type Ctx struct {
	sync.Mutex
	ClearCacheDir bool
	Dirs          struct {
		Cur   string
		Cache string
		Libs  []string
	}
	LibsWatch struct {
		Should func() bool
	}

	libs struct {
		all libs
	}
	state struct {
		initCalled         bool
		cleanUps           []func()
		msgs               []CtxMsg
		modsWatcher        func()
		modsWatcherRunning bool
	}
}

func (me *Ctx) maybeInitPanic(initingNow bool) {
	if me.state.initCalled == initingNow {
		panic("atmo.Ctx.Init must be called exactly once only")
	}
}

func (me *Ctx) Init(dirCur string) (err error) {
	me.maybeInitPanic(true)
	if me.state.initCalled, me.libs.all = true, nil; dirCur == "" || dirCur == "." {
		dirCur, err = os.Getwd()
	} else if dirCur[0] == '~' {
		if len(dirCur) > 1 && dirCur[1] == filepath.Separator {
			dirCur = filepath.Join(usys.UserHomeDirPath(), dirCur[2:])
		} else {
			dirCur = usys.UserHomeDirPath()
		}
	}
	if err == nil {
		dirCur, err = filepath.Abs(dirCur)
	}
	if err == nil && !ufs.IsDir(dirCur) {
		err = &os.PathError{Path: dirCur, Op: "directory", Err: os.ErrNotExist}
	}
	if cachedir := me.Dirs.Cache; err == nil {
		if cachedir == "" {
			cachedir = filepath.Join(usys.UserDataDirPath(true), "atmo")
		}
		if !ufs.IsDir(cachedir) {
			err = ufs.EnsureDir(cachedir)
		} else if me.ClearCacheDir {
			err = ufs.Del(cachedir)
		}
		if libsdirs := me.Dirs.Libs; err == nil {
			libsdirsenv := ustr.Split(os.Getenv(EnvVarLibDirs), string(os.PathListSeparator))
			for i := range libsdirsenv {
				libsdirsenv[i] = filepath.Clean(libsdirsenv[i])
			}
			for i := range libsdirs {
				libsdirs[i] = filepath.Clean(libsdirs[i])
			}
			libsdirs = ustr.Merge(libsdirsenv, libsdirs, func(ldp string) bool {
				if ldp != "" && !ufs.IsDir(ldp) {
					me.msg(true, "libs-dir "+ldp+" not found")
					return true
				}
				return ldp == ""
			})
			for i := range libsdirs {
				for j := range libsdirs {
					if iinj, jini := ustr.Pref(libsdirs[i], libsdirs[j]), ustr.Pref(libsdirs[j], libsdirs[i]); i != j && (iinj || jini) {
						err = errors.New("conflicting libs-dirs: " + libsdirs[i] + " vs. " + libsdirs[j])
						return
					}
				}
				if dirPathAutoLib == "" {
					if dp := filepath.Join(libsdirs[i], NameAutoLib); ufs.IsDir(dp) {
						dirPathAutoLib = dp
					}
				}
			}
			if dirPathAutoLib == "" {
				err = errors.New("`" + NameAutoLib + "` lib not found in any of these paths: " + ustr.Join(libsdirs, "  ──  "))
			} else {
				me.Dirs.Cur, me.Dirs.Cache, me.Dirs.Libs = dirCur, cachedir, libsdirs
				me.initLibs()
			}
		}
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
}

func (me *Ctx) msg(alreadyLocked bool, text string) {
	msg := CtxMsg{Time: time.Now(), Text: text}
	if !alreadyLocked {
		me.Lock()
	}
	me.state.msgs = append(me.state.msgs, msg)
	if !alreadyLocked {
		me.Unlock()
	}
}

func (me *Ctx) Messages(clear bool) (msgs []CtxMsg) {
	me.Lock()
	if msgs = me.state.msgs; clear {
		me.state.msgs = nil
	}
	me.Unlock()
	return
}
