package atemlang

import (
	"github.com/go-leap/dev/lex"
)

type ApplStyle int

const (
	APPLSTYLE_SVO ApplStyle = iota
	APPLSTYLE_VSO
	APPLSTYLE_SOV
)

type (
	ctxParseTld struct {
		file        *AstFile
		cur         *AstDef
		indentHint  int
		mto         map[*udevlex.Token]int   // maps comments-stripped Tokens to orig Tokens
		mtc         map[*udevlex.Token][]int // maps comments-stripped Tokens to comment Tokens in orig
		parensLevel int
	}
)

var (
	langReservedOps = []string{"&", "?", ":", "|", ",", ":=", "=", "!"}
)

func init() {
	udevlex.RestrictedWhitespace, udevlex.StandaloneSeps = true, []string{"(", ")"}
}

func (me *AstFile) parse(this *AstFileTopLevelChunk) {
	toks := this.Ast.Tokens
	if this.Ast.Comments, toks = me.parseTopLevelLeadingComments(toks); len(toks) > 0 {
		this.Ast.Def, this.errs.parsing = me.parseTopLevelDef(toks)
	}
}

func (*AstFile) parseTopLevelLeadingComments(toks udevlex.Tokens) (ret []*AstComment, rest udevlex.Tokens) {
	for len(toks) > 0 && toks[0].Kind() == udevlex.TOKEN_COMMENT {
		toks, ret = toks[1:], append(ret, newAstComment(toks, 0))
	}
	rest = toks
	return
}

func (me *AstFile) parseTopLevelDef(tokens udevlex.Tokens) (def *AstDef, err *Error) {
	ctx := ctxParseTld{file: me, mtc: make(map[*udevlex.Token][]int), mto: make(map[*udevlex.Token]int)}
	var astdef AstDef
	if err = ctx.parseDef(tokens, true, &astdef); err == nil {
		def = &astdef
	}
	return
}

func (me *ctxParseTld) parseDef(tokens udevlex.Tokens, isTopLevel bool, def *AstDef) (err *Error) {
	toks := tokens
	if isTopLevel {
		toks = tokens.SansComments(me.mtc, me.mto)
	}
	if tokshead, tokheadbodysep, toksbody := toks.BreakOnOpish(":="); len(toksbody) == 0 {
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

		if isTopLevel {
			me.cur, def.Tokens, def.IsTopLevel = def, tokens, true
		} else {
			me.setTokensFor(&def.AstBaseTokens, toks, nil)
		}
		if err = def.newIdent(me, -1, toksheadsig, namepos); err == nil {
			def.ensureArgsLen(len(toksheadsig) - 1)
			for i, a := 0, 0; i < len(toksheadsig); i++ {
				if i != namepos {
					if err = def.newIdent(me, a, toksheadsig, i); err != nil {
						return
					}
					a++
				}
			}
			if me.indentHint = 0; toksbody[0].Meta.Position.Line == tokheadbodysep.Meta.Line {
				me.indentHint = tokheadbodysep.Meta.Position.Column - 1
			}
			if def.Body, err = me.parseExpr(toksbody); err == nil && len(toksheads) > 1 {
				def.Meta, err = me.parseMetas(toksheads[1:])
			}
		}
	}
	return
}

func (me *ctxParseTld) parseExpr(toks udevlex.Tokens) (ret IAstExpr, err *Error) {
	indhint := toks[0].Meta.Position.Column
	if me.indentHint != 0 {
		indhint, me.indentHint = me.indentHint, 0
	}
	if chunks := toks.IndentBasedChunks(indhint); len(chunks) > 1 {
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
					exprcur, toks, err = me.parseExprLetInner(toks, accum, alltoks)
					accum = accum[:0]
				case "?":
					exprcur, toks, err = me.parseExprCase(toks, accum, alltoks)
					accum = accum[:0]
				default:
					exprcur = me.newExprIdent(toks)
					toks = toks[1:]
				}
			case udevlex.TOKEN_SEPISH:
				if sub, rest, e := me.parseParens(toks); e != nil {
					err = e
				} else if exprcur, err = me.parseExprInParens(sub); err == nil {
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
		ret = me.parseExprFinalize(accum, alltoks, nil)
	}
	return
}

func (me *ctxParseTld) parseExprFinalize(accum []IAstExpr, allToks udevlex.Tokens, untilTok *udevlex.Token) (ret IAstExpr) {
	if len(accum) == 1 {
		ret = accum[0]
	} else {
		var appl AstExprAppl
		ret = &appl

		me.setTokensFor(&appl.AstBaseTokens, allToks, untilTok)
		args := make([]IAstExpr, 1, len(accum)-1)
		switch me.file.Options.ApplStyle {
		case APPLSTYLE_VSO:
			appl.Callee, args[0] = accum[0], accum[1]
			appl.Args = append(args, accum[2:]...)
		case APPLSTYLE_SVO:
			appl.Callee, args[0] = accum[1], accum[0]
			appl.Args = append(args, accum[2:]...)
		case APPLSTYLE_SOV:
			l := len(accum) - 1
			appl.Callee, args[0] = accum[l], accum[0]
			appl.Args = append(args, accum[1:l]...)
		}
	}
	return
}

func (me *ctxParseTld) parseExprClaspOperators(accum []IAstExpr, allToks udevlex.Tokens, untilTok *udevlex.Token) []IAstExpr {
	iuntil := -1
	clasp := func(ifrom int) {
		if xp, _ := me.parseExprFinalize(accum[ifrom:iuntil+1], allToks, untilTok).(*AstExprAppl); xp != nil {
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

func (me *ctxParseTld) parseExprInParens(toks udevlex.Tokens) (ret IAstExpr, err *Error) {
	me.parensLevel++
	ret, err = me.parseExpr(toks)
	me.parensLevel--
	return
}

func (me *ctxParseTld) parseExprCase(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, rest udevlex.Tokens, err *Error) {
	var scrutinee IAstExpr
	if len(accum) > 0 {
		scrutinee = me.parseExprFinalize(accum, allToks, &toks[0])
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
			} else if caseof.Alts[i].Body, err = me.parseExpr(ifthen[1]); caseof.defaultIndex >= 0 {
				err = errAt(&alts[i][0], ErrCatSyntax, "malformed `?` branching: encountered a second default case, only at most one is permissible")
			} else {
				caseof.defaultIndex = i
			}
		} else if cond, err = me.parseExpr(ifthen[0]); err == nil {
			if caseof.Alts[i].Conds = []IAstExpr{cond}; len(ifthen) > 1 {
				caseof.Alts[i].Body, err = me.parseExpr(ifthen[1])
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

func (me *ctxParseTld) parseExprLetInner(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, rest udevlex.Tokens, err *Error) {
	var body IAstExpr
	body = me.parseExprFinalize(accum, allToks, &toks[0])
	toks, rest = toks[1:].BreakOnIndent(allToks[0].Meta.LineIndent)
	if chunks := toks.Chunked(",", "(", ")"); len(chunks) > 0 {
		var let AstExprLet
		let.Body, let.Defs = body, make([]AstDef, len(chunks))
		me.setTokensFor(&let.AstBaseTokens, allToks, nil)
		lasttokforerr := &toks[0]
		for i := range chunks {
			if len(chunks[i]) == 0 {
				err = errAt(lasttokforerr, ErrCatSyntax, "unexpected empty space between two commas")
			} else if err = me.parseDef(chunks[i], false, &let.Defs[i]); err == nil {
				lasttokforerr = chunks[i].Last()
			}
			if err != nil {
				return
			}
		}
		ret = &let
	}
	return
}

func (me *ctxParseTld) parseExprLetOuter(toks udevlex.Tokens, toksChunked []udevlex.Tokens) (ret *AstExprLet, err *Error) {
	var let AstExprLet
	me.setTokensFor(&let.AstBaseTokens, toks, nil)
	let.Defs = make([]AstDef, len(toksChunked)-1)
	for i := range toksChunked {
		if i == 0 {
			let.Body, err = me.parseExpr(toksChunked[i])
		} else {
			err = me.parseDef(toksChunked[i], false, &let.Defs[i-1])
		}
		if err != nil {
			return
		}
	}
	ret = &let
	return
}

func (me *ctxParseTld) parseMetas(chunks []udevlex.Tokens) (metas []IAstExpr, err *Error) {
	metas = make([]IAstExpr, 0, len(chunks))
	var meta IAstExpr
	for i := range chunks {
		if len(chunks[i]) > 0 {
			if meta, err = me.parseExpr(chunks[i]); err != nil {
				metas = nil
				return
			} else {
				metas = append(metas, meta)
			}
		}
	}
	return
}

func (me *ctxParseTld) parseParens(toks udevlex.Tokens) (sub udevlex.Tokens, rest udevlex.Tokens, err *Error) {
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
