package main

import (
	"bufio"
	"os"

	"github.com/go-leap/str"
)

var (
	replDirectives = map[string]func(string){
		"i": onNameInfo,
		"t": onTypeInfo,
		"k": onKindInfo,
		"m": onGotoMod,
	}
)

func mainRepl() {
	writeLn("mojo repl:")
	writeLn("— directives are prefixed with `:`")
	writeLn("— a line ending in `…` either begins\n  or ends a multi-line input")
	writeLn("— enter any mojo definition or expression")
	multiln, repl := "", bufio.NewScanner(os.Stdin)
	for repl.Scan() {
		if readln := ustr.Trim(repl.Text()); readln != "" {
			if ustr.Suff(readln, "…") {
				if multiln == "" {
					multiln = readln[:len(readln)-len("…")] + "\n  "
					continue
				} else {
					readln, multiln = ustr.Trim(multiln+readln[:len(readln)-len("…")]), ""
				}
			}
			switch {
			case multiln != "":
				multiln += readln + "\n  "
			case readln[0] == ':':
				directive, arg := ustr.BreakOnFirstOrPref(readln[1:], " ")
				if do := replDirectives[directive]; do != nil {
					do(arg)
				} else if directive == "q" {
					writeLn("…and we're done.")
					return
				} else {
					writeLn("unknown directive: `:" + directive + "` — try: ")
					writeLn("\t:q — quit")
					writeLn("\t:m — go to given module")
					writeLn("\t:i — info about given name")
					writeLn("\t:t — type info")
					writeLn("\t:k — kind info")
				}
			default:
				if out, err := ctx.ReadEvalPrint(readln); err != nil {
					println(err.Error())
				} else {
					writeLn(out.String())
				}
			}
		}
	}
}

func onGotoMod(arg string) {
	writeLn("goes to module: " + arg)
}

func onKindInfo(arg string) {
	writeLn("kind of: " + arg)
}

func onNameInfo(arg string) {
	writeLn("info for: " + arg)
}

func onTypeInfo(arg string) {
	writeLn("type of: " + arg)
}
