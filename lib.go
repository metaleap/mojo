package atem

import (
	"os"
	"time"

	"github.com/go-leap/fs"
)

type Libs []Lib

type Lib struct {
	LibPath string
	DirPath string
}

func (me *Ctx) initLibs() {
	me.cleanUps = append(me.cleanUps,
		ufs.WatchModTimesEvery(3*time.Second, me.Dirs.Libs, ".at", func(mods map[string]os.FileInfo) {
			for fullpath := range mods {
				println("MOD", fullpath)
			}
		}))
}
