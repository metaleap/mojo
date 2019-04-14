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

	Dirs struct {
		Cur     string
		Cache   string
		StdLibs string
	}
}

func (me *Ctx) Init(dirCur string) (err error) {
	if me.initCalled {
		panic("atem.Ctx.Init was called more than once: Ctx is not for reuse")
	}
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
		}
		if stdlibsdir := me.Dirs.StdLibs; err == nil {
			if stdlibsdir == "" {
				if stdlibsdir = os.Getenv("ATEM_PATH_STDLIBS"); stdlibsdir == "" {
					stdlibsdir = "/home/_/c/atem/stdlibs"
				}
			}

			me.Dirs.Cur, me.Dirs.Cache, me.Dirs.StdLibs = dirCur, cachedir, stdlibsdir
		}
	}
	return
}

func (me *Ctx) ReadEvalPrint(in string) (out fmt.Stringer, err error) {
	if !me.initCalled {
		panic("atem.Ctx.Init wasn't called")
	}
	err = fmt.Errorf("to-do: evaluation of %q", in)
	return
}
