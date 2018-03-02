package mo

import (
	"path/filepath"
)

type ICtx interface {
	DirPath() string
}

func New(dirPath string) (mojoCtx ICtx, err error) {
	if dirPath, err = filepath.Abs(dirPath); err == nil {
		mojoCtx = &ctx{Dir: dirPath}
	}
	return
}
