package atemrepl

func (me *Repl) DQuit(string) {
	me.quit = true
}

func (me *Repl) DWelcome(string) {
	me.IO.writeLns(
		"", "— repl directives begin with `:`,\n  any other inputs are eval'd",
		"", "— a line ending in "+me.IO.MultiLineSuffix+" begins\n  or ends a multi-line input",
		"", "— for proper line-editing, run via\n  `rlwrap` or some equivalent",
		"",
	)
}
