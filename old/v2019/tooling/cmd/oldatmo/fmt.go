package main

import (
	"os"

	"github.com/go-leap/dev/lex"
	"github.com/metaleap/atmo/old/v2019/ast"
)

func mainFmt() {
	var astfile atmoast.AstFile

	if len(os.Args) > 2 {
		astfile.SrcFilePath = os.Args[2]
	}
	udevlex.RestrictedWhitespaceRewriter = func(char rune) int {
		if char == '\t' || char == '\v' {
			return 4
		}
		return 1
	}

	astfile.LexAndParseFile(false, true, nil)
	if errs := astfile.Errors(); len(errs) > 0 {
		for _, e := range errs {
			println(e.Error())
		}
	} else {
		_, _ = os.Stdout.Write(astfile.Print(&atmoast.PrintFmtPretty{}))
		println("\n\n===\nTHIS feature isn't really done, rather it was postponed until it can be written in atmo itself. Above is only internal semi-pretty-printer.")
	}
}
