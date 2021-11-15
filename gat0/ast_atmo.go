package main

type AstFile struct {
	srcFilePath string
	origSrc     string
	toks        Tokens
	topLevel    []AstNode
}

type AstNode interface {
	base() AstNodeBase
	String(int) string
}

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

type AstNodePair struct {
	AstNodeBase
	sep string
	lhs AstNode
	rhs AstNode
}

type AstNodeAtom struct {
	AstNodeBase
}

func (me *AstFile) buildIr() (ret IrModule) {
	ret.ast = me
	return
}

func (me AstNodeAtom) String(indent int) string {
	return "<ATOM tk='" + me.toks[0].kind.String() + "'>" + me.toks.String("", "") + "</ATOM>"
}

func (me AstNodePair) String(indent int) string {
	return "<PAIR>" + me.lhs.String(indent) + me.sep + me.rhs.String(indent) + "</PAIR>"
}

func (me AstNodeList) String(indent int) (s string) {
	s = "<LIST sep='" + me.sep + "'>"
	for i, node := range me.nodes {
		s += "<" + itoa(i) + ">" + node.String(indent) + "</" + itoa(i) + ">"
	}
	s += "</LIST>"
	return
}

func (me AstNodeBraced) String(indent int) (s string) {
	s = ifStr(me.curly, "{", ifStr(me.square, "[", "("))
	s += me.list.String(indent)
	s += ifStr(me.curly, "}", ifStr(me.square, "]", ")"))
	return
}
