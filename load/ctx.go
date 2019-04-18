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

var (
	dirPathAutoPack string
)

type CtxMsg struct {
	Time time.Time
	Text string
}

type Ctx struct {
	ClearCacheDir bool
	Dirs          struct {
		Cur   string
		Cache string
		Packs []string
	}
	AutoPacksWatch struct {
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

func (me *Ctx) maybeInitPanic(initingNow bool) {
	if me.state.initCalled == initingNow {
		panic("atmo.Ctx.Init must be called exactly once only")
	}
}

func (me *Ctx) Init(dirCur string) (err error) {
	me.maybeInitPanic(true)
	if me.state.initCalled, me.packs.all = true, nil; dirCur == "" || dirCur == "." {
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
						return
					}
				}
				if dirPathAutoPack == "" {
					if dp := filepath.Join(packsdirs[i], atmo.NameAutoPack); ufs.IsDir(dp) {
						dirPathAutoPack = dp
					}
				}
			}
			if len(packsdirs) == 0 {
				err = errors.New("none of the specified packs dirs were found:\n    " + ustr.Join(append(packsdirsenv, packsdirsorig...), "\n    "))
			} else if dirPathAutoPack == "" {
				err = errors.New("`" + atmo.NameAutoPack + "` pack not found in any of these paths:\n    " + ustr.Join(packsdirs, "\n    "))
			} else {
				me.Dirs.Cur, me.Dirs.Cache, me.Dirs.Packs = dirCur, cachedir, packsdirs
				me.initPacks()
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
