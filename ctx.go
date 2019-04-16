package atem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-leap/fs"
	"github.com/go-leap/str"
	"github.com/go-leap/sys"
)

type Ctx struct {
	initCalled bool

	ClearCacheDir bool
	Dirs          struct {
		Cur   string
		Cache string
		Libs  []string
	}
	libs struct {
		sync.Mutex
		Known   Libs
		lookups struct {
			dirPaths map[string]int
			libPaths map[string]int
		}
	}

	cleanUps []func()
}

func (me *Ctx) maybeInitPanic(initingNow bool) {
	if me.initCalled == initingNow {
		panic("atem.Ctx.Init must be called exactly once only")
	}
}

func (me *Ctx) Init(dirCur string) (err error) {
	me.maybeInitPanic(true)
	if me.initCalled, me.libs.Known = true, nil; dirCur == "" || dirCur == "." {
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
			cachedir = filepath.Join(usys.UserDataDirPath(true), "atem")
		}
		if !ufs.IsDir(cachedir) {
			err = ufs.EnsureDir(cachedir)
		} else if me.ClearCacheDir {
			err = ufs.Del(cachedir)
		}
		if libsdirs := me.Dirs.Libs; err == nil {
			libsdirs = ustr.Merge(ustr.Split(os.Getenv(EnvVarLibDirs), string(os.PathListSeparator)), libsdirs, true)
			for i := range libsdirs {
				for j := range libsdirs {
					if i != j && (ustr.Pref(libsdirs[i], libsdirs[j]) || ustr.Pref(libsdirs[j], libsdirs[i])) {
						err = errors.New("conflicting libs-dirs: " + libsdirs[i] + " vs. " + libsdirs[j])
						return
					}
				}
			}

			me.Dirs.Cur, me.Dirs.Cache, me.Dirs.Libs = dirCur, cachedir, libsdirs
			me.initLibs()
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
	for _, cleanup := range me.cleanUps {
		if cleanup != nil {
			cleanup()
		}
	}
}
