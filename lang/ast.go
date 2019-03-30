package odlang

import (
	"github.com/go-leap/dev/lex"
)

type IAstNode interface {
	Toks() udevlex.Tokens
}

type astNode struct {
	toks udevlex.Tokens
}

func (me *astNode) Toks() udevlex.Tokens { return me.toks }

type AstFile struct {
	astNode
	Nodes []IAstNode
}

type AstComments struct {
	astNode
}
