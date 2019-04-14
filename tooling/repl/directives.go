package atemrepl

func (me *Repl) DQuit(string) {
	me.quit = true
}

func (me *Repl) DWelcome(string) {
	me.IO.writeLns(
		"", "— repl directives begin with `:`,\n  all other inputs are eval'd",
		"", "— a line ending in "+me.IO.MultiLineSuffix+" either begins\n  or ends a multi-line input",
		"", "— for line-editing, consider using\n  `rlwrap` or some equivalent",
		"",
	)
}
