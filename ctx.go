package mo

import (
	"fmt"
)

type ctx struct {
	Dir string
}

func (this *ctx) DirPath() string { return this.Dir }

func (this *ctx) ReadEvalPrint(in string) (out fmt.Stringer, err error) {
	err = fmt.Errorf("to-do: evaluation of %q", in)
	return
}
