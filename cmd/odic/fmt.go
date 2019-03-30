package main

import (
	"os"

	"github.com/metaleap/odic/lang"
)

func mainFmt() {
	var err error
	var astfile odlang.AstFile
	if len(os.Args) > 2 {
		astfile.SrcFilePath = os.Args[2]
	}
	if !astfile.LexAndParseFile("<stdin>") {
		err = astfile.Err()
	} else {
		writeLn("\n\n" + astfile.Src().String() + "\n\n")
	}

	if err != nil {
		println(err.Error())
	}
}
