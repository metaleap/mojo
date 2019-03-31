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

	// either one or both nil:
	defType *AstDefType
	defFunc *AstDefFunc
}

type AstComment struct {
	astNode
	ContentText    string
	SelfTerminates bool // if comment is of /**/ form
}

type astDefBase struct {
	Name string
	Args []string
}

type AstDefType struct {
	astNode
	astDefBase
}

type AstDefFunc struct {
	astNode
	astDefBase
}
