package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	. "github.com/metaleap/atmo/atem"
)

var prog Prog

func main() {
	src, err := ioutil.ReadAll(os.Stdin)
	if err == nil {
		prog = LoadFromJson(src)
		for i := range prog {
			prog[i].Body = convFrom(prog[i].Body)
		}
		prefixNameMetasWithIdxs()
		prog = optimize(prog)
		prefixNameMetasWithIdxs()
		for i := range prog {
			prog[i].Body = convTo(prog[i].Body)
		}
		_, err = os.Stdout.WriteString(prog.JsonSrc(false))
	}
	if err != nil {
		panic(err)
	}
}

func prefixNameMetasWithIdxs() {
	for i := 0; i < len(prog)-1; i++ {
		pos := strings.IndexByte(prog[i].Meta[0], ']')
		prog[i].Meta[0] = "[" + strconv.Itoa(i) + "]" + prog[i].Meta[0][pos+1:]
	}
}
