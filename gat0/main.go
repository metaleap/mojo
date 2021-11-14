package main

import (
	"fmt"
	"os"
)

func main() {
	data, _ := os.ReadFile(os.Args[1])
	origsrc := string(data)
	toks := tokenizer.tokenize(origsrc)
	ast := parse(toks, origsrc, os.Args[1])
	fmt.Printf("%#v\n", ast.topLevel)
	// ir := ast.buildIr()
	// llvmir := ir.buildLLvmIr()
	// llIrSrc(os.Stdout, &llvmir)
}
