package atmolang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
)

func (me *AstDef) initIdent(ctx *ctxParseTld, arg int, ttmp udevlex.Tokens, at int, affixIndices map[int]int) (err *atmo.Error) {
	tok, this, isarg := &ttmp[at], &me.Name, arg > -1
	if isarg {
		this = &me.Args[arg]
	}
	s, k, affix := tok.Meta.Orig, tok.Kind(), ""
	if affixIndices != nil {
		if idx, ok := affixIndices[at]; ok {
			s, affix = s[:idx], s[idx+1:]
		}
	}
	if (!isarg) && k != udevlex.TOKEN_IDENT && k != udevlex.TOKEN_OPISH {
		err = atmo.ErrSyn(tok, "expected def name instead of `"+tok.Meta.Orig+"`")
	} else {
		ctx.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, ttmp, at)
		this.Val, this.IsOpish, this.IsTag, this.Affix =
			s, (k == udevlex.TOKEN_OPISH), isarg && ustr.BeginsUpper(s), affix
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

func (me *AstExprCase) removeAltAt(idx int) {
	for i := idx; i < len(me.Alts)-1; i++ {
		me.Alts[i] = me.Alts[i+1]
	}
	me.Alts = me.Alts[:len(me.Alts)-1]
}
