package main

import (
	"strings"
)

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

func (me *AstNodeList) treeifyByIndents(astFile *AstFile) {

}

func (me *AstFile) buildIr() (ret IrModule) {
	ret.ast = me
	return
}

func (me AstNodeAtom) String(indent int) (s string) {
	s = strings.Repeat("\t", indent) + "<ATOM tk='" + me.toks[0].kind.String() + "'>"
	s += me.toks.String("", "")
	s += "</ATOM>\n"
	return
}

func (me AstNodePair) String(indent int) (s string) {
	s = strings.Repeat("\t", indent) + "<PAIR>\n"
	s += me.lhs.String(indent + 1)
	s += strings.Repeat("\t", indent+1) + me.sep + "\n"
	s += me.rhs.String(indent + 1)
	s += strings.Repeat("\t", indent) + "</PAIR>\n"
	return
}

func (me AstNodeList) String(indent int) (s string) {
	s = strings.Repeat("\t", indent) + "<LIST sep='" + me.sep + "'>\n"
	for i, node := range me.nodes {
		s += strings.Repeat("\t", indent+1) + "<" + itoa(i) + ">\n"
		s += node.String(indent + 2)
		s += strings.Repeat("\t", indent+1) + "</" + itoa(i) + ">\n"
	}
	s += strings.Repeat("\t", indent) + "</LIST>\n"
	return
}

func (me AstNodeBraced) String(indent int) (s string) {
	s = strings.Repeat("\t", indent) + "<BR t='" + me.toks[0].src + "'>\n"
	s += me.list.String(indent + 1)
	s += strings.Repeat("\t", indent) + "</BR>\n"
	return
}
