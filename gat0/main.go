package main

import (
	"os"
)

func main() {
	data, _ := os.ReadFile(os.Args[1])
	toks := tokenizer.tokenize(string(data))
	for _, tok := range toks {
		println(tok.String())
	}
}
