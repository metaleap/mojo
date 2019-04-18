package main

import (
	"os"

	"github.com/go-leap/dev/lex"
	"github.com/metaleap/atmo/lang"
)

func mainFmt() {
	var astfile atmolang.AstFile

	if len(os.Args) > 2 {
		astfile.SrcFilePath = os.Args[2]
	}
	udevlex.RestrictedWhitespaceRewriter = func(char rune) int {
		if char == '\t' || char == '\v' {
			return 4
		}
		return 1
	}

	astfile.LexAndParseFile(false, true)
	if errs := astfile.Errs(); len(errs) > 0 {
		for _, e := range errs {
			println(e.Error())
		}
	} else {
		_, _ = os.Stdout.Write(astfile.Print(&atmolang.PrintFormatterMinimal{}))
	}
}
