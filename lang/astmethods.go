package atemlang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

func (me *AstDef) initIdent(ctx *ctxParseTld, arg int, ttmp udevlex.Tokens, at int, affixIndices map[int]int) (err *Error) {
	tok, this, isarg := &ttmp[at], &me.Name, arg > -1
	if isarg {
		this = &me.Args[arg]
	}

	k, s, affix, namedesc := tok.Kind(), tok.Meta.Orig, "", ustr.If(isarg, "argument", "definition")
	if affixIndices != nil {
		if idx, ok := affixIndices[at]; ok {
			s, affix = s[:idx], s[idx+1:]
		}
	}
	if isident, isopish := (k == udevlex.TOKEN_IDENT), (k == udevlex.TOKEN_OPISH); !(isident || isopish) {
		err = errSyntax(tok, "not a valid "+namedesc+" name: "+s)
	} else if _s := ustr.If(me.IsTopLevel && s[0] == '_', s[1:], s); (!isarg) && isident && !ustr.BeginsLower(_s) {
		err = errSyntax(tok, "not a valid "+namedesc+" name: `"+_s+"` should begin with lower-case letter")
	} else if isarg && s[0] == '_' && len(s) > 1 {
		err = errSyntax(tok, "not a valid "+namedesc+" name: `"+s+"` shouldn't begin with underscore")
	} else if tok.IsAnyOneOf(langReservedOps...) {
		err = errSyntax(tok, "not a valid "+namedesc+" name: `"+s+"` is reserved and cannot be overloaded")
	} else {
		this.Val, this.IsOpish, this.IsTag, this.Affix = s, isopish, isarg && ustr.BeginsUpper(s), affix
		ctx.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, ttmp, at)
	}
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
