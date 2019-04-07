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
		file        *AstFile
		def         IAstDef
		mto         map[*udevlex.Token]int   // maps comments-stripped Tokens to orig Tokens
		mtc         map[*udevlex.Token][]int // maps comments-stripped Tokens to comment Tokens in orig
		parensLevel int
	}
)

var (
	langReservedOps = []string{"&", "?", ":", "|", ",", ":=", "="}
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

func (me *ctxParseTopLevelDef) parseDef(tokens udevlex.Tokens, isTopLevel bool) (def IAstDef, err *Error) {
	toks := tokens
	if isTopLevel {
		toks = tokens.SansComments(me.mtc, me.mto)
	}
	tokshead, toksbody := toks.BreakOnOpish(":=")
	if len(toksbody) == 0 {
		err = errAt(&tokens[0], ErrCatSyntax, "missing: definition body following `:=`")
	} else if len(tokshead) == 0 {
		err = errAt(&tokens[0], ErrCatSyntax, "missing: definition name preceding `:=`")
	} else if toksheads := tokshead.Chunked(",", "(", ")"); len(toksheads[0]) == 0 {
		err = errAt(&tokens[0], ErrCatSyntax, "missing: definition name preceding `,`")
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
		} else {
			var deffunc AstDefFunc
			def, defbase = &deffunc, &deffunc.AstDefBase
		}
		if isTopLevel {
			me.def, defbase.Tokens = def, tokens
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
				defbase.Meta, err = me.parseMetas(toksheads[1:], false)
			}
		}
	}
end:
	if err != nil {
		if def = nil; isTopLevel {
			me.def = nil
		}
	}
	return
}

func (me *ctxParseTopLevelDef) parseMetas(chunks []udevlex.Tokens, typeExpr bool) (metas []IAstExpr, err *Error) {
	metas = make([]IAstExpr, 0, len(chunks))
	var meta IAstExpr
	for i := range chunks {
		if len(chunks[i]) > 0 {
			if meta, err = me.parseExpr(chunks[i], typeExpr); err != nil {
				metas = nil
				return
			} else {
				metas = append(metas, meta)
			}
		}
	}
	return
}

func (me *AstDefFunc) parseDefBody(ctx *ctxParseTopLevelDef, toks udevlex.Tokens) (err *Error) {
	me.Body, err = ctx.parseExpr(toks, false)
	return
}

func (me *AstDefType) parseDefBody(ctx *ctxParseTopLevelDef, toks udevlex.Tokens) (err *Error) {
	tags := toks.Chunked("|", "(", ")")
	if len(tags[0]) == 0 {
		err = errAt(&toks[0], ErrCatSyntax, "unexpected `|`")

	} else if istagged := len(tags) > 1 || (len(tags[0]) > 2 &&
		tags[0][1].Kind() == udevlex.TOKEN_OPISH && tags[0][0].Kind() == udevlex.TOKEN_IDENT &&
		tags[0][1].Str == ":" && ustr.BeginsUpper(tags[0][0].Str)); istagged {
		me.Tags = make([]AstDefTypeTag, len(tags))
		for i := range tags {
			if t := tags[i]; len(t) == 0 {
				err = errAt(&toks[0], ErrCatSyntax, "type definition `"+me.Name.Val()+"` is missing tag details in between two `|` operators")
			} else if t[0].Kind() != udevlex.TOKEN_IDENT || !ustr.BeginsUpper(t[0].Str) {
				err = errAt(&t[0], ErrCatSyntax, "malformed tag name `"+t[0].Meta.Orig+"`, should be upper-case")
			} else if len(t) > 1 && (t[1].Kind() != udevlex.TOKEN_OPISH || t[1].Str != ":") {
				err = errAt(&t[1], ErrCatSyntax, "expected `:` following tag name `"+t[0].Str+"`")
			} else if ctx.setTokenAndCommentsFor(&me.Tags[i].Name.AstBaseTokens, &me.Tags[i].Name.AstBaseComments, t, 0); len(t) > 2 {
				me.Tags[i].Expr, err = ctx.parseTypeExpr(t[2:])
			}
			if err != nil {
				return
			}
		}

	} else {
		me.Expr, err = ctx.parseTypeExpr(toks)
	}
	return
}

func (me *ctxParseTopLevelDef) parseExpr(toks udevlex.Tokens, typeExpr bool) (ret IAstExpr, err *Error) {
	var chunks []udevlex.Tokens
	if !typeExpr {
		chunks = toks.IndentBasedChunks(toks[0].Meta.Position.Column - 1)
	}
	if len(chunks) > 1 {
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
			case udevlex.TOKEN_IDENT, udevlex.TOKEN_OPISH:
				switch toks[0].Str {
				case ",":
					if !typeExpr {
						exprcur, toks, err = me.parseExprLetInner(toks, accum, alltoks)
					} else {
						exprcur, toks, err = me.parseTypeExprMeta(toks, accum, alltoks)
					}
					accum = accum[0:0]
				case "?":
					exprcur, toks, err = me.parseExprCase(toks, accum, alltoks, typeExpr)
					accum = accum[0:0]
				default:
					if !typeExpr {
						exprcur = me.newExprIdent(toks)
					} else {
						exprcur = me.newTypeExprIdent(toks)
					}
					toks = toks[1:]
				}
			case udevlex.TOKEN_SEPISH:
				if sub, rest, e := me.parseParens(toks); e != nil {
					err = e
				} else if exprcur, err = me.parseExprInParens(sub, typeExpr); err == nil {
					toks = rest
				}
			default:
				err = errAt(&toks[0], ErrCatSyntax, "the impossible: unrecognized token (new bug in parser, parseExpr needs updating)")
			}
			if err != nil {
				return
			}
			accum = append(accum, exprcur)
		}
		ret = me.parseExprFinalize(accum, alltoks, nil, typeExpr)
	}
	return
}

func (me *ctxParseTopLevelDef) parseExprFinalize(accum []IAstExpr, allToks udevlex.Tokens, untilTok *udevlex.Token, typeExpr bool) (ret IAstExpr) {
	if len(accum) == 1 {
		ret = accum[0]
	} else {
		// accum = me.parseExprClaspOperators(accum, allToks, untilTok, typeExpr)
		var abt *AstBaseTokens
		var fcallee *IAstExpr
		var fargs *[]IAstExpr
		if typeExpr {
			var appl AstTypeExprAppl
			abt, fcallee, fargs = &appl.AstBaseTokens, &appl.Callee, &appl.Args
			ret = &appl
		} else {
			var appl AstExprAppl
			abt, fcallee, fargs = &appl.AstBaseTokens, &appl.Callee, &appl.Args
			ret = &appl
		}
		me.setTokensFor(abt, allToks, untilTok)
		switch me.file.Options.ApplStyle {
		case APPLSTYLE_SVO:
			*fcallee, *fargs = accum[1], append(accum[0:1], accum[2:]...)
		case APPLSTYLE_VSO:
			*fcallee, *fargs = accum[0], accum[1:]
		case APPLSTYLE_SOV:
			l := len(accum) - 1
			*fcallee, *fargs = accum[l], accum[:l]
		}
	}
	return
}

func (me *ctxParseTopLevelDef) parseExprClaspOperators(accum []IAstExpr, allToks udevlex.Tokens, untilTok *udevlex.Token, typeExpr bool) []IAstExpr {
	iuntil := -1
	clasp := func(ifrom int) {
		if xp, _ := me.parseExprFinalize(accum[ifrom:iuntil+1], allToks, untilTok, typeExpr).(*AstExprAppl); xp != nil {
			println(xp.Callee.ExprBase().Tokens.Pos().String(), xp.Callee.ExprBase().Tokens.String(), len(xp.Args))
		}
	}

	for i, curisop := len(accum)-1, accum[len(accum)-1].ExprBase().IsOp(langReservedOps...); i > 0; i-- {
		cur, prev := accum[i].ExprBase(), accum[i-1].ExprBase()
		previsop := prev.IsOp(langReservedOps...)
		nope := curisop || previsop || cur.Tokens.DistanceTo(prev.Tokens) > 0
		if nope {
			if iuntil > 0 {
				clasp(i)
				iuntil = -1
			}
		} else if iuntil < 0 {
			iuntil = i
		}
		curisop = previsop
	}
	if iuntil >= 0 && iuntil < len(accum)-1 {
		clasp(0)
	}
	return accum
}

func (me *ctxParseTopLevelDef) parseExprInParens(toks udevlex.Tokens, typeExpr bool) (ret IAstExpr, err *Error) {
	me.parensLevel++
	ret, err = me.parseExpr(toks, typeExpr)
	me.parensLevel--
	return
}

func (me *ctxParseTopLevelDef) parseExprCase(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens, typeExpr bool) (ret IAstExpr, rest udevlex.Tokens, err *Error) {
	var scrutinee IAstExpr
	if len(accum) > 0 {
		scrutinee = me.parseExprFinalize(accum, allToks, &toks[0], typeExpr)
	}
	var caseof AstExprCase
	caseof.Scrutinee, caseof.defaultIndex = scrutinee, -1
	me.setTokensFor(&caseof.AstBaseTokens, allToks, nil)
	toks, rest = toks[1:].BreakOnIndent(allToks[0].Meta.LineIndent)
	alts := toks.Chunked("|", "(", ")")
	caseof.Alts = make([]AstCaseAlt, len(alts))
	var cond IAstExpr
	var hasmulticonds bool
	for i := range alts {
		if len(alts[i]) == 0 {
			err = errAt(&toks[0], ErrCatSyntax, "malformed `?` branching: empty case")
		} else if ifthen := alts[i].Chunked(":", "(", ")"); len(ifthen) > 2 {
			err = errAt(&alts[i][0], ErrCatSyntax, "malformed `?` branching: each case needs exactly one corresponding `:` with subsequent result expression")
		} else if me.setTokensFor(&caseof.Alts[i].AstBaseTokens, alts[i], nil); len(ifthen[0]) == 0 {
			if len(ifthen[1]) == 0 {
				err = errAt(&alts[i][0], ErrCatSyntax, "malformed `?` branching: default case has no result expression")
			} else if caseof.Alts[i].Body, err = me.parseExpr(ifthen[1], false); caseof.defaultIndex >= 0 {
				err = errAt(&alts[i][0], ErrCatSyntax, "malformed `?` branching: encountered a second default case, only at most one is permissible")
			} else {
				caseof.defaultIndex = i
			}
		} else if cond, err = me.parseExpr(ifthen[0], false); err == nil {
			if caseof.Alts[i].Conds = []IAstExpr{cond}; len(ifthen) > 1 {
				caseof.Alts[i].Body, err = me.parseExpr(ifthen[1], false)
			} else {
				hasmulticonds = true
			}
		}
		if err != nil {
			return
		}
	}
	if hasmulticonds {
		if len(caseof.Alts) == 2 && caseof.Alts[0].Body == nil && caseof.Alts[1].Body == nil {
			caseof.Alts[0].IsShortForm, caseof.Alts[0].Body, caseof.Alts[0].Conds = true, caseof.Alts[0].Conds[0], nil
			caseof.Alts[1].IsShortForm, caseof.Alts[1].Body, caseof.Alts[1].Conds = true, caseof.Alts[1].Conds[0], nil
		} else {
			for i := 0; i < len(caseof.Alts); i++ {
				if ca := &caseof.Alts[i]; ca.Body == nil {
					if i < len(caseof.Alts)-1 {
						canext := &caseof.Alts[i+1]
						canext.Conds = append(canext.Conds, ca.Conds...)
						caseof.Alts = append(caseof.Alts[:i], caseof.Alts[i+1:]...)
						i--
					} else if caseof.defaultIndex < 0 && len(ca.Conds) == 1 {
						caseof.defaultIndex, ca.Body, ca.Conds, ca.IsShortForm = i, ca.Conds[0], nil, true
					} else {
						err = errAt(&ca.Tokens[0], ErrCatSyntax, "malformed `?` branching: case has no result expression")
						return
					}
				}
			}
		}
	}
	ret = &caseof
	return
}

func (me *ctxParseTopLevelDef) parseExprLetInner(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, rest udevlex.Tokens, err *Error) {
	var body IAstExpr
	body = me.parseExprFinalize(accum, allToks, &toks[0], false)
	toks, rest = toks[1:].BreakOnIndent(allToks[0].Meta.LineIndent)
	if chunks := toks.Chunked(",", "(", ")"); len(chunks) > 0 {
		var let AstExprLet
		let.Body, let.Defs = body, make([]IAstDef, 0, len(chunks))
		me.setTokensFor(&let.AstBaseTokens, allToks, nil)
		var def IAstDef
		for i := range chunks {
			if len(chunks[i]) > 0 {
				if def, err = me.parseDef(chunks[i], false); err != nil {
					return
				} else {
					let.Defs = append(let.Defs, def)
				}
			}
		}
		ret = &let
	}
	return
}

func (me *ctxParseTopLevelDef) parseExprLetOuter(toks udevlex.Tokens, toksChunked []udevlex.Tokens) (ret *AstExprLet, err *Error) {
	var let AstExprLet
	var def IAstDef
	me.setTokensFor(&let.AstBaseTokens, toks, nil)
	for i := range toksChunked {
		if i == 0 {
			let.Body, err = me.parseExpr(toksChunked[i], false)
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
	// split by &
	// split by ,
	// labelled:
	// ident or appl form

	var expr IAstExpr
	if expr, err = me.parseExpr(toks, true); err == nil {
		if ret, _ = expr.(IAstTypeExpr); ret == nil {
			err = errAt(&toks[0], ErrCatSyntax, "expected type expression, not "+expr.Description())
		}
	}
	return
}

func (me *ctxParseTopLevelDef) parseTypeExprMeta(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, rest udevlex.Tokens, err *Error) {
	var texpr IAstTypeExpr
	tmp := me.parseExprFinalize(accum, allToks, &toks[0], true)
	if texpr, _ = tmp.(IAstTypeExpr); texpr == nil {
		err = errAt(&tmp.ExprBase().Tokens[0], ErrCatSyntax, "expected (prior to meta expressions) type expression, not "+tmp.Description())
	}
	if err == nil {
		toks, rest = toks[1:].BreakOnIndent(allToks[0].Meta.LineIndent)
		if chunks := toks.Chunked(",", "(", ")"); len(chunks) > 0 {
			if texpr.TypeExprBase().Meta, err = me.parseMetas(chunks, true); err == nil {
				ret = texpr
			}
		}
	}
	return
}

func (me *ctxParseTopLevelDef) parseTypeExprConj(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, rest udevlex.Tokens, err *Error) {
	// var conj AstTypeExprRec
	return
}

func (me *ctxParseTopLevelDef) parseParens(toks udevlex.Tokens) (sub udevlex.Tokens, rest udevlex.Tokens, err *Error) {
	var numunclosed int
	if toks[0].Str == ")" {
		err = errAt(&toks[0], ErrCatSyntax, "closing parenthesis without matching opening")
	} else if sub, rest, numunclosed = toks.Sub("(", ")"); len(sub) == 0 {
		if numunclosed == 0 {
			err = errAt(&toks[0], ErrCatSyntax, "empty parentheses")
		} else {
			err = errAt(&toks[0], ErrCatSyntax, "unclosed parenthesis")
		}
	}
	return
}

func (me *ctxParseTopLevelDef) parseTypeExprInParens(toks udevlex.Tokens) (ret IAstTypeExpr, err *Error) {
	me.parensLevel++
	ret, err = me.parseTypeExpr(toks)
	me.parensLevel--
	return
}
