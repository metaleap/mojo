package main

type AstFile struct {
	srcFilePath string
	origSrc     string
	toks        Tokens
	topLevel    []AstNode
}

type AstNode interface{}

type AstNodeBase struct {
	toks Tokens
}

func (me *AstNodeBase) base() *AstNodeBase { return me }

type AstNodeCommaSeparated struct {
	AstNodeBase
	nodes []AstNode
}

type AstNodeBraces struct {
	AstNodeBase
	square bool
	curly  bool
	list   AstNodeCommaSeparated
}

func (me *AstFile) buildIr() (ret IrModule) {
	ret.ast = me
	return
}
