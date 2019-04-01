package odlang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

type AstBaseComments struct {
	Comments []*AstComment
}

type AstBaseTokens struct {
	Tokens udevlex.Tokens
}

func (me *AstBaseTokens) self() *AstBaseTokens { return me }

type AstTopLevel struct {
	AstBaseTokens
	AstBaseComments

	// either one or both nil:
	DefType *AstDefType
	DefFunc *AstDefFunc
}

type AstComment struct {
	AstBaseTokens
}

func (me *AstComment) ContentText() string     { return me.Tokens[0].Str }
func (me *AstComment) IsSelfTerminating() bool { return me.Tokens[0].IsCommentSelfTerminating() }

type AstIdent struct {
	AstBaseTokens
	AstBaseComments
}

func (me *AstIdent) BeginsLower() bool { return ustr.BeginsLower(me.Tokens[0].Str) }
func (me *AstIdent) BeginsUpper() bool { return ustr.BeginsUpper(me.Tokens[0].Str) }
func (me *AstIdent) String() string    { return me.Tokens[0].Str }
func (me *AstIdent) IsOpish() bool     { return me.Tokens[0].Kind() == udevlex.TOKEN_OTHER }

type AstDefBase struct {
	AstBaseTokens

	Name AstIdent
	Args []AstIdent
}

type AstDefType struct {
	AstDefBase
}

type AstDefFunc struct {
	AstDefBase
}
