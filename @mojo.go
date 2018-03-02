package mo

import (
	"fmt"
	"path/filepath"
)

type ICtx interface {
	DirPath() string
	ReadEvalPrint(string) (fmt.Stringer, error)
}

func New(dirPath string) (mojoCtx ICtx, err error) {
	if dirPath, err = filepath.Abs(dirPath); err == nil {
		mojoCtx = &ctx{Dir: dirPath}
	}
	return
}
