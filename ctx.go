package atem

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-leap/fs"
	"github.com/go-leap/sys"
)

type Ctx struct {
	initCalled bool

	ClearCacheDir bool
	Dirs          struct {
		Cur     string
		Cache   string
		StdLibs string
	}
}

func (me *Ctx) maybeInitPanic(initingNow bool) {
	if me.initCalled && initingNow {
		panic("atem.Ctx.Init was called more than once: Ctx is not for reuse")
	} else if (!me.initCalled) && !initingNow {
		panic("atem.Ctx.Init wasn't called prior to Ctx use")
	}
}

func (me *Ctx) Init(dirCur string) (err error) {
	me.maybeInitPanic(true)
	if me.initCalled = true; dirCur == "" || dirCur == "." {
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
		if stdlibsdir := me.Dirs.StdLibs; err == nil {
			if stdlibsdir == "" {
				stdlibsdir = os.Getenv("ATEM_PATH_STDLIBS")
			}

			me.Dirs.Cur, me.Dirs.Cache, me.Dirs.StdLibs = dirCur, cachedir, stdlibsdir
		}
	}
	return
}

func (me *Ctx) ReadEvalPrint(in string) (out fmt.Stringer, err error) {
	me.maybeInitPanic(false)
	err = fmt.Errorf("to-do: evaluation of %q", in)
	return
}
