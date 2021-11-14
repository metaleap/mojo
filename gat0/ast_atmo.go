package main

type AstFile struct {
	srcFilePath string
	origSrc     string
	toks        Tokens
	topLevel    []AstNode
}

type AstNode interface{ base() AstNodeBase }

type AstNodeBase struct {
	toks Tokens
}

func (me AstNodeBase) base() AstNodeBase { return me }

type AstNodeBraced struct {
	AstNodeBase
	square bool
	curly  bool
	list   AstNodeList
}

type AstNodeList struct {
	AstNodeBase
	sep   string
	nodes []AstNode
}

func (me *AstFile) buildIr() (ret IrModule) {
	ret.ast = me
	return
}
