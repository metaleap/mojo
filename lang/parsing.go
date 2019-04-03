package odlang

import (
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
	ctxParseTopLevelDef struct {
		file *AstFile
		def  IAstDef
		mto  map[*udevlex.Token]int   // maps comments-stripped Tokens to orig Tokens
		mtc  map[*udevlex.Token][]int // maps comments-stripped Tokens to comment Tokens in orig
	}
)

func init() {
	udevlex.RestrictedWhitespace, udevlex.StandaloneSeps = true, []string{"(", ")"}
}

func (me *AstFile) parse(this *AstFileTopLevelChunk) {
	toks := this.Tokens
	if this.AstTopLevel.Comments, toks = me.parseTopLevelLeadingComments(toks); len(toks) > 0 {
		this.AstTopLevel.Def, this.errs.parsing = me.parseTopLevelDef(toks)
	}
}

func (*AstFile) parseTopLevelLeadingComments(toks udevlex.Tokens) (ret []*AstComment, rest udevlex.Tokens) {
	for len(toks) > 0 && toks[0].Kind() == udevlex.TOKEN_COMMENT {
		toks, ret = toks[1:], append(ret, newAstComment(toks, 0))
	}
	rest = toks
	return
}

func (me *AstFile) parseTopLevelDef(tokens udevlex.Tokens) (def IAstDef, err *Error) {
	ctx := ctxParseTopLevelDef{file: me, mtc: make(map[*udevlex.Token][]int), mto: make(map[*udevlex.Token]int)}
	return ctx.parseDef(tokens, true)
}

func (me *ctxParseTopLevelDef) parseDef(tokens udevlex.Tokens, topLevel bool) (def IAstDef, err *Error) {
	toks := tokens
	if topLevel {
		toks = tokens.SansComments(me.mtc, me.mto)
	}
	tokshead, toksbody := toks.BreakOnOther(":=")
	if len(toksbody) == 0 {
		err = errAt(&tokens[0], "missing: definition body following `:=`")
	} else if len(tokshead) == 0 {
		err = errAt(&tokens[0], "missing: definition name preceding `:=`")
	} else if toksheads := tokshead.Chunked(",", "(", ")"); len(toksheads[0]) == 0 {
		err = errAt(&tokens[0], "missing: definition name preceding `,`")
	} else {
		toksheadsig := toksheads[0]
		var namepos int
		if len(toksheadsig) > 1 {
			if me.file.Options.ApplStyle == APPLSTYLE_SVO {
				namepos = 1
			} else if me.file.Options.ApplStyle == APPLSTYLE_SOV {
				namepos = len(toksheadsig) - 1
			}
		}
		var defbase *AstDefBase
		if isdeftype := ustr.BeginsUpper(toksheadsig[namepos].Str); isdeftype {
			var deftype AstDefType
			def, defbase, deftype.AstDefBase.IsDefType = &deftype, &deftype.AstDefBase, true
			if len(toksheads) > 1 {
				err = errAt(&tokens[len(toksheadsig)], "unexpected comma")
				goto end
			}
		} else {
			var deffunc AstDefFunc
			def, defbase = &deffunc, &deffunc.AstDefBase
		}
		if topLevel {
			me.def = def
		}
		if topLevel {
			defbase.Tokens = tokens
		} else {
			me.setTokensFor(&defbase.AstBaseTokens, toks, nil)
		}
		if err = defbase.newIdent(me, -1, toksheadsig, namepos); err == nil {
			defbase.ensureArgsLen(len(toksheadsig) - 1)
			for i, a := 0, 0; i < len(toksheadsig); i++ {
				if i != namepos {
					if err = defbase.newIdent(me, a, toksheadsig, i); err != nil {
						goto end
					}
					a++
				}
			}
			if err = def.parseDefBody(me, toksbody); err == nil && len(toksheads) > 1 {
				defbase.Meta = make([]IAstExpr, len(toksheads)-1)
				for i := range toksheads {
					if i > 0 {
						if defbase.Meta[i-1], err = me.parseExpr(toksheads[i]); err != nil {
							goto end
						}
					}
				}
			}
		}
	}
end:
	if err != nil {
		if def = nil; topLevel {
			me.def = nil
		}
	}
	return
}

func (me *AstDefFunc) parseDefBody(ctx *ctxParseTopLevelDef, toks udevlex.Tokens) (err *Error) {
	me.Body, err = ctx.parseExpr(toks)
	return
}

func (me *AstDefType) parseDefBody(ctx *ctxParseTopLevelDef, toks udevlex.Tokens) (err *Error) {
	opts := toks.Chunked("|", "(", ")")
	if len(opts[0]) == 0 {
		err = errAt(&toks[0], "unexpected `|`")
	} else if len(opts) == 1 {
		me.Expr, err = ctx.parseTypeExpr(toks)
	} else {

	}
	return
}

func (me *ctxParseTopLevelDef) parseExpr(toks udevlex.Tokens) (ret IAstExpr, err *Error) {
	if chunks := toks.IndentBasedChunks(toks[0].Meta.Position.Column - 1); len(chunks) > 1 {
		ret, err = me.parseExprLetOuter(toks, chunks)

	} else {
		alltoks, accum := toks, make([]IAstExpr, 0, len(toks))
		for len(toks) > 0 {
			var exprcur IAstExpr
			switch k := toks[0].Kind(); k {
			case udevlex.TOKEN_FLOAT:
				exprcur = me.newExprLitFloat(toks)
				toks = toks[1:]
			case udevlex.TOKEN_UINT:
				exprcur = me.newExprLitUint(toks)
				toks = toks[1:]
			case udevlex.TOKEN_RUNE:
				exprcur = me.newExprLitRune(toks)
				toks = toks[1:]
			case udevlex.TOKEN_STR:
				exprcur = me.newExprLitStr(toks)
				toks = toks[1:]
			case udevlex.TOKEN_IDENT, udevlex.TOKEN_OTHER:
				switch toks[0].Str {
				case ",":
					exprcur, toks, err = me.parseExprLetInner(toks, accum, alltoks)
					accum = accum[0:0]
				case "?":
					exprcur, toks, err = me.parseExprCase(toks, accum, alltoks)
					accum = accum[0:0]
				default:
					exprcur = me.newExprIdent(toks)
					toks = toks[1:]
				}
			case udevlex.TOKEN_SEP:
				if toks[0].Str == ")" {
					err = errAt(&toks[0], "closing parenthesis without matching opening")
				} else if sub, rest, numunclosed := toks.Sub("(", ")"); len(sub) == 0 {
					if numunclosed == 0 {
						err = errAt(&toks[0], "empty parentheses")
					} else {
						err = errAt(&toks[0], "unclosed parenthesis")
					}
				} else if exprcur, err = me.parseExpr(sub); err == nil {
					toks = rest
				}
			default:
				panic(k)
			}
			if err != nil {
				return
			}
			accum = append(accum, exprcur)
		}
		ret, err = me.parseExprFinalize(accum, alltoks, nil)
	}
	return
}

func (me *ctxParseTopLevelDef) parseExprFinalize(accum []IAstExpr, allToks udevlex.Tokens, untilTok *udevlex.Token) (ret IAstExpr, err *Error) {
	if len(accum) == 1 {
		ret = accum[0]
	} else {
		var appl AstExprAppl
		me.setTokensFor(&appl.AstBaseTokens, allToks, untilTok)
		l := len(accum) - 1
		switch me.file.Options.ApplStyle {
		case APPLSTYLE_SVO:
			appl.Callee = accum[1]
			appl.Args = append(accum[0:1], accum[2:]...)
		case APPLSTYLE_VSO:
			appl.Callee = accum[0]
			appl.Args = accum[1:]
		case APPLSTYLE_SOV:
			appl.Callee = accum[l]
			appl.Args = accum[:l]
		}
		ret = &appl
	}
	return
}

func (me *ctxParseTopLevelDef) parseExprCase(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, rest udevlex.Tokens, err *Error) {
	var scrutinee IAstExpr
	if len(accum) > 0 {
		scrutinee, err = me.parseExprFinalize(accum, allToks, &toks[0])
	}
	if err == nil {
		var caseof AstExprCase
		caseof.Scrutinee, caseof.defaultIndex = scrutinee, -1
		me.setTokensFor(&caseof.AstBaseTokens, allToks, nil)
		toks, rest = toks[1:].BreakOnIndent(allToks[0].Meta.LineIndent)
		alts := toks.Chunked("|", "(", ")")
		if caseof.Alts = make([]AstCaseAlt, len(alts)); len(alts[0]) == 0 {
			err = errAt(&toks[0], "unexpected: `|`")
			return
		}
		for i := range alts {
			if len(alts[i]) == 0 {
				err = errAt(&toks[0], "malformed `?` branching: empty case")
			} else if ifthen := alts[i].Chunked(":", "(", ")"); len(ifthen) != 2 {
				err = errAt(&alts[i][0], "malformed `?` branching: each case needs exactly one corresponding `:` with subsequent expression")
			} else if me.setTokensFor(&caseof.Alts[i].AstBaseTokens, alts[i], nil); len(ifthen[0]) == 0 {
				if caseof.Alts[i].Body, err = me.parseExpr(ifthen[1]); caseof.defaultIndex >= 0 {
					err = errAt(&ifthen[0][0], "malformed `?` branching: encountered a second default case, only at most one is permissible")
				} else {
					caseof.defaultIndex = i
				}
			} else if caseof.Alts[i].Cond, err = me.parseExpr(ifthen[0]); err == nil {
				caseof.Alts[i].Body, err = me.parseExpr(ifthen[1])
			}
			if err != nil {
				return
			}
		}
		ret = &caseof
	}
	return
}

func (me *ctxParseTopLevelDef) parseExprLetInner(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, rest udevlex.Tokens, err *Error) {
	var body IAstExpr
	if body, err = me.parseExprFinalize(accum, allToks, &toks[0]); err == nil {
		toks, rest = toks[1:].BreakOnIndent(allToks[0].Meta.LineIndent)
		if chunks := toks.Chunked(",", "(", ")"); len(chunks) > 0 {
			var let AstExprLet
			let.Body, let.Defs = body, make([]IAstDef, 0, len(chunks))
			me.setTokensFor(&let.AstBaseTokens, allToks, nil)
			var def IAstDef
			for i := range chunks {
				if def, err = me.parseDef(chunks[i], false); err != nil {
					return
				} else {
					let.Defs = append(let.Defs, def)
				}
			}
			ret = &let
		}
	}
	return
}

func (me *ctxParseTopLevelDef) parseExprLetOuter(toks udevlex.Tokens, toksChunked []udevlex.Tokens) (ret *AstExprLet, err *Error) {
	var let AstExprLet
	var def IAstDef
	me.setTokensFor(&let.AstBaseTokens, toks, nil)
	for i := range toksChunked {
		if i == 0 {
			let.Body, err = me.parseExpr(toksChunked[i])
		} else if def, err = me.parseDef(toksChunked[i], false); err == nil {
			let.Defs = append(let.Defs, def)
		}
		if err != nil {
			return
		}
	}
	ret = &let
	return
}

func (me *ctxParseTopLevelDef) parseTypeExpr(toks udevlex.Tokens) (ret IAstTypeExpr, err *Error) {
	alltoks, accum := toks, make([]IAstTypeExpr, 0, len(toks))
	// var tmetas []IAstExpr
	for len(toks) > 0 {
		var exprcur IAstTypeExpr
		switch k := toks[0].Kind(); k {
		case udevlex.TOKEN_FLOAT:
			err = errAt(&toks[0], "unexpected float literal")
		case udevlex.TOKEN_UINT:
			err = errAt(&toks[0], "unexpected integer literal")
		case udevlex.TOKEN_RUNE:
			err = errAt(&toks[0], "unexpected character literal")
		case udevlex.TOKEN_STR:
			err = errAt(&toks[0], "unexpected text literal")
		case udevlex.TOKEN_IDENT, udevlex.TOKEN_OTHER:
			switch toks[0].Str {
			case ":":
				return
			case "&":
				return
			case ",":
				return
				// metas := toks[1:].Chunked(",", "(", ")")
				// tmetas = make([]IAstExpr)
				// exprcur, toks, err = me.parseExprLetInner(toks, accum, alltoks)
				accum = accum[0:0]
				return
			default:
				exprcur = me.newTypeExprIdent(toks)
				toks = toks[1:]
			}
		case udevlex.TOKEN_SEP:
			if toks[0].Str == ")" {
				err = errAt(&toks[0], "closing parenthesis without matching opening")
			} else if sub, rest, numunclosed := toks.Sub("(", ")"); len(sub) == 0 {
				if numunclosed == 0 {
					err = errAt(&toks[0], "empty parentheses")
				} else {
					err = errAt(&toks[0], "unclosed parenthesis")
				}
			} else if exprcur, err = me.parseTypeExpr(sub); err == nil {
				toks = rest
			}
		default:
			panic(k)
		}
		if err != nil {
			goto end
		}
		accum = append(accum, exprcur)
	}
	if alltoks == nil {
	}
	ret, err = nil, nil //,me.parseExprFinalize(accum, alltoks, nil)

	goto end
end:
	if err != nil {
		ret = nil
	}
	return
}
