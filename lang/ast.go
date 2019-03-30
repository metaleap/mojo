package odlang

import (
	"github.com/go-leap/dev/lex"
)

type IAstNode interface {
	Src() udevlex.Tokens
}

type astNode struct {
	toks udevlex.Tokens
}

func (me *astNode) Src() udevlex.Tokens { return me.toks }

type AstComments struct {
	astNode
}
