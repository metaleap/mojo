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
		{
			prefixNameMetasWithIdxs()
			for again := true; again; {
				again = false
				fixFuncDefArgsUsageNumbers()
				for _, mayberewrite := range []func(Prog) (Prog, bool){
					rewrite_ditchUnusedFuncDefs,
					rewrite_ditchDuplicateDefs,
					rewrite_inlineNaryFuncAliases,
					rewrite_inlineCallsToArgRefFuncs,
					rewrite_argDropperCalls,
					rewrite_inlineArgCallers,
					rewrite_inlineArgsRearrangers,
					rewrite_primOpPreCalcs,
					rewrite_callsToGeqOrLeq,
					rewrite_minifyNeedlesslyElaborateBoolOpCalls,
					rewrite_inlineOnceCalleds,
					rewrite_inlineEverSameArgs,
					rewrite_preEvalArgRefLessCalls,
					rewrite_inlineNullaries,
					rewrite_commonSubExprs,
				} {
					if prog, again = mayberewrite(prog); again {
						break
					}
				}
			}
			prefixNameMetasWithIdxs()
		}
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
