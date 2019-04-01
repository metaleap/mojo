package odlang

import (
	"text/scanner"

	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

type ApplStyle int

const (
	APPLSTYLE_SVO ApplStyle = iota
	APPLSTYLE_VSO
	APPLSTYLE_SOV
)

type (
	mapTokCmnts   = map[*udevlex.Token][]int
	mapTokOldIdxs = map[*udevlex.Token]int
	ctxParseDef   struct {
		mapTokCmnts
		mapTokOldIdxs
	}
)

func init() {
	udevlex.RestrictedWhitespace, udevlex.StandaloneSeps = true, []string{"(", ")"}
}

type Error struct {
	Msg string
	Pos scanner.Position
	Len int
}

func errAt(tok *udevlex.Token, msg string) *Error {
	return &Error{Msg: msg, Pos: tok.Meta.Position, Len: len(tok.Meta.Orig)}
}

func (this *Error) Error() string {
	return this.Pos.String() + ": " + this.Msg
}

func (me *AstFile) parse(this *AstFileTopLevelChunk) {
	toks := this.Tokens
	if this.AstTopLevel.Comments, toks = me.parseTopLevelLeadingComments(toks); len(toks) > 0 {
		this.AstTopLevel.Def, this.errs.parsing = me.parseTopLevelDefinition(toks)
	}
}

func (*AstFile) parseTopLevelLeadingComments(toks udevlex.Tokens) (ret []*AstComment, rest udevlex.Tokens) {
	for len(toks) > 0 && toks[0].Kind() == udevlex.TOKEN_COMMENT {
		toks, ret = toks[1:], append(ret, newAstComment(toks, 0))
	}
	rest = toks
	return
}

func (me *AstFile) parseTopLevelDefinition(tokens udevlex.Tokens) (def IAstDef, err *Error) {
	var ctx ctxParseDef
	ctx.mapTokCmnts, ctx.mapTokOldIdxs = make(mapTokCmnts), make(mapTokOldIdxs)
	tokshead, toksbody := tokens.SansComments(ctx.mapTokCmnts, ctx.mapTokOldIdxs).BreakOnOther(":=")
	if len(toksbody) == 0 {
		err = errAt(&tokens[0], "missing: definition body following `:=`")
	} else if len(tokshead) == 0 {
		err = errAt(&tokens[0], "missing: definition name preceding `:=`")
	} else if toksheadsig, _ := tokshead.BreakOnOther(","); len(toksheadsig) == 0 {
		err = errAt(&tokens[0], "missing: definition name preceding `,`")
	} else {
		var namepos int
		if len(toksheadsig) > 1 {
			if me.Options.ApplStyle == APPLSTYLE_SVO {
				namepos = 1
			} else if me.Options.ApplStyle == APPLSTYLE_SOV {
				namepos = len(toksheadsig) - 1
			}
		}

		var defbase *AstDefBase
		if isdeftype := ustr.BeginsUpper(toksheadsig[namepos].Str); isdeftype {
			var deftype AstDefType
			def, defbase, deftype.AstDefBase.IsDefType = &deftype, &deftype.AstDefBase, true
		} else {
			var deffunc AstDefFunc
			def, defbase = &deffunc, &deffunc.AstDefBase
		}
		defbase.Tokens = tokens
		if err = defbase.newIdent(-1, toksheadsig, namepos, &ctx); err != nil {
			def = nil
		} else {
			defbase.ensureArgsLen(len(toksheadsig) - 1)
			for i, a := 0, 0; i < len(toksheadsig); i++ {
				if i != namepos {
					if err = defbase.newIdent(a, toksheadsig, i, &ctx); err != nil {
						def = nil
						return
					}
					a++
				}
			}
			if err = def.parseDefBody(toksbody, &ctx); err != nil {
				def = nil
			}
		}
	}
	return
}

func (me *AstDefFunc) parseDefBody(toks udevlex.Tokens, ctx *ctxParseDef) (err *Error) {
	// println(defbase.Name.Val(), len(toksbody.IndentBasedChunks(toksbody[0].Meta.LineIndent)))

	me.Body, err = me.parseExpr(toks, ctx)
	return
}

func (me *AstDefType) parseDefBody(toks udevlex.Tokens, ctx *ctxParseDef) *Error {
	return nil
}

func (me *AstDefBase) parseExpr(toks udevlex.Tokens, ctx *ctxParseDef) (r IAstExpr, err *Error) {
	if len(toks) == 0 {
		panic("bug in parseExpr")
	}
	for len(toks) > 0 {
		var this IAstExpr
		switch k := toks[0].Kind(); k {
		case udevlex.TOKEN_FLOAT:
			this = me.newExprLitFloat(toks, ctx)
			toks = toks[1:]
		case udevlex.TOKEN_UINT:
			this = me.newExprLitUint(toks, ctx)
			toks = toks[1:]
		case udevlex.TOKEN_RUNE:
			this = me.newExprLitRune(toks, ctx)
			toks = toks[1:]
		case udevlex.TOKEN_STR:
			this = me.newExprLitStr(toks, ctx)
			toks = toks[1:]
		case udevlex.TOKEN_IDENT, udevlex.TOKEN_OTHER:
			this = me.newExprIdent(toks, ctx)
			toks = toks[1:]
		case udevlex.TOKEN_SEP:
			if toks[0].Str == ")" {
				err = errAt(&toks[0], "closing parenthesis without matching opening")
			} else if sub, tail, numunclosed := toks.Sub("(", ")"); len(sub) == 0 {
				if numunclosed == 0 {
					err = errAt(&toks[0], "empty parentheses")
				} else {
					err = errAt(&toks[0], "unclosed parenthesis")
				}
			} else if this, err = me.parseExpr(sub, ctx); err == nil {
				toks = tail
			}
		default:
			panic(k)
		}
		if err != nil {
			return
		}
		if r == nil {
			r = this
		} else if rf, _ := r.(*AstExprCall); rf != nil {

		} else {

		}
	}
	return
}
