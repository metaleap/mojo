package atemrepl

type directive struct {
	Desc string
	Run  func(string)
}

type directives []directive

func (me *directives) ensure(desc string, run func(string)) {
	if found := me.By(desc[0]); found != nil {
		found.Desc, found.Run = desc, run
	} else {
		*me = append(*me, directive{Desc: desc, Run: run})
	}
}

func (me directives) By(letter byte) *directive {
	for i := range me {
		if me[i].Desc[0] == letter {
			return &me[i]
		}
	}
	return nil
}

func (me *Repl) DQuit(string) {
	me.run.quit = true
}

func (me *Repl) DWelcomeMsg(string) {
	me.IO.writeLns(
		"", "— repl directives begin with `:`,\n  any other inputs are eval'd",
		"", "— a line ending in "+me.IO.MultiLineSuffix+" begins\n  or ends a multi-line input",
		"", "— for proper line-editing, run via\n  `rlwrap` or `rlfe` or similar.",
		"",
	)
}
