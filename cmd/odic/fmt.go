package main

import (
	"io/ioutil"
	"os"

	"github.com/metaleap/odic/lang"
)

func mainFmt() {
	var src []byte
	var err error
	srcfilepath := "<stdin>"
	if len(os.Args) <= 2 {
		src, err = ioutil.ReadAll(os.Stdin)
	} else {
		srcfilepath = os.Args[2]
		src, err = ioutil.ReadFile(srcfilepath)
	}
	if err == nil {
		defs, errs := odlang.LexAndParseDefs(srcfilepath, string(src))
		for _, e := range errs {
			err = e
			break
		}
		if err == nil {
			println(len(defs))
		}
	}

	if err != nil {
		println(err.Error())
	}
}
