package main

import (
	"os"

	"github.com/go-leap/dev/lex"
	"github.com/metaleap/odic/lang"
)

func mainFmt() {
	var err error
	var astfile odlang.AstFile

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
	} else if src, e := astfile.Print(&odlang.PrintFormatterMinimal{}); e != nil {
		err = e
	} else {
		_, err = os.Stdout.Write(src)
	}

	if err != nil {
		println(err.Error())
	}
}
