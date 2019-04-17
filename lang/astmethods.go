package atmolang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

func (me *AstDef) initIdent(ctx *ctxParseTld, arg int, ttmp udevlex.Tokens, at int, affixIndices map[int]int) (err *Error) {
	tok, this, isarg := &ttmp[at], &me.Name, arg > -1
	if isarg {
		this = &me.Args[arg]
	}
	s, affix := tok.Meta.Orig, ""
	if affixIndices != nil {
		if idx, ok := affixIndices[at]; ok {
			s, affix = s[:idx], s[idx+1:]
		}
	}
	this.Val, this.IsOpish, this.IsTag, this.Affix =
		s, (tok.Kind() == udevlex.TOKEN_OPISH), isarg && ustr.BeginsUpper(s), affix
	ctx.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, ttmp, at)
	return
}

func (me *AstComment) initFrom(tokens udevlex.Tokens, at int) {
	me.Tokens = tokens[at : at+1]
	me.ContentText, me.IsSelfTerminating = me.Tokens[0].Str, me.Tokens[0].IsCommentSelfTerminating()
}

func (me *AstExprCase) Default() *AstCaseAlt {
	if me.defaultIndex < 0 {
		return nil
	}
	return &me.Alts[me.defaultIndex]
}
