package main

import (
    "os"
)

func main() {
    data, _ := os.ReadFile(os.Args[1])
    origsrc := string(data)
    toks := tokenizer.tokenize(origsrc)
    ast := parse(toks, origsrc, os.Args[1])

    // for _, tlc := range toks.indentLevelChunks(0) {
    // 	println(tlc.String(ast.origSrc, "") + "\n\n===================\n\n")
    // }

    for _, node := range ast.topLevel {
        println(node.String(0) + "\n\n===================\n\n")
    }

    // ir := ast.buildIr()
    // llvmir := ir.buildLLvmIr()
    // llIrSrc(os.Stdout, &llvmir)
}
