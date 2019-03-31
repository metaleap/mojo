package odlang

import (
	"github.com/go-leap/dev/lex"
)

type IAstNode interface {
	self() *astNode
	Src() udevlex.Tokens
}

type astNode struct {
	toks udevlex.Tokens
}

func (me *astNode) self() *astNode      { return me }
func (me *astNode) Src() udevlex.Tokens { return me.toks }

type AstTopLevel struct {
	astNode
	comments []*AstComment
	defType  *AstDefType
	defFunc  *AstDefFunc
}

type AstComment struct {
	astNode
	Text           string
	SelfTerminates bool
}

type AstDefType struct {
	astNode
}

type AstDefFunc struct {
	astNode
}
