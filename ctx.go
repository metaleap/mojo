package mo

type ctx struct {
	Dir string
}

func (me *ctx) DirPath() string { return me.Dir }
