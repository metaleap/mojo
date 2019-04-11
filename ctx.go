package atem

import (
	"fmt"
)

type Ctx struct {
	Dir     string
	WorkDir string
}

func (this *Ctx) DirPath() string { return this.Dir }

func (this *Ctx) ReadEvalPrint(in string) (out fmt.Stringer, err error) {
	err = fmt.Errorf("to-do: evaluation of %q", in)
	return
}
