package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/go-leap/str"
)

var (
	replCommands = map[string]func(string){
		"i": onNameInfo,
		"t": onTypeInfo,
		"k": onKindInfo,
		"m": onGotoMod,
	}
)

func mainRepl() {
	multiln, repl := "", bufio.NewScanner(os.Stdin)
	for repl.Scan() {
		if readln := strings.TrimSpace(repl.Text()); readln != "" {
			if readln == "…" && multiln != "" {
				readln, multiln = strings.TrimSpace(multiln), ""
			}
			switch {
			case strings.HasSuffix(readln, "…"):
				multiln = readln[:len(readln)-len("…")] + "\n  "
			case multiln != "":
				multiln += readln + "\n  "
			case readln[0] == ':':
				directive, arg := ustr.BreakOnFirstOrSuff(readln[1:], " ")
				if directive == "" {
					directive, arg = arg, ""
				}
				if do := replCommands[directive]; do != nil {
					do(arg)
				} else if directive == "q" {
					writeLn("…and we're done.")
					break
				} else {
					writeLn("unrecognized: `:" + directive + "` — try those: ")
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

func onGotoMod(string) {
	writeLn("mod-goto: to-do")
}

func onKindInfo(string) {
	writeLn("kind of to-do")
}

func onNameInfo(string) {
	writeLn("info: to-do")
}

func onTypeInfo(string) {
	writeLn("typical: to-do")
}
