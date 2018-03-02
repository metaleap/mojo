package mo

import (
	"fmt"
)

type ctx struct {
	Dir string
}

func (me *ctx) DirPath() string { return me.Dir }

func (me *ctx) ReadEvalPrint(in string) (out fmt.Stringer, err error) {
	err = fmt.Errorf("to-do: evaluation of %q", in)
	return
}
