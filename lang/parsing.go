package atmolang

import (
	"github.com/go-leap/dev/lex"
	"github.com/metaleap/atmo"
)

type ApplStyle int

const (
	APPLSTYLE_SVO ApplStyle = iota
	APPLSTYLE_VSO
	APPLSTYLE_SOV
)

func init() {
	udevlex.RestrictedWhitespace, udevlex.StandaloneSeps, udevlex.SepsForChunking =
		true, []string{"(", ")"}, "([{}])"
}

func (me *AstFile) parse(this *AstFileTopLevelChunk) {
	toks := this.Ast.Tokens
	if this.Ast.Comments, toks = me.parseTopLevelLeadingComments(toks); len(toks) > 0 {
		if this.Ast.Def, this.errs.parsing = me.parseTopLevelDef(toks); this.errs.parsing == nil {
			if this.Ast.DefIsUnexported = (this.Ast.Def.Name.Val[0] == '_' && len(this.Ast.Def.Name.Val) > 1); this.Ast.DefIsUnexported {
				this.Ast.Def.Name.Val = this.Ast.Def.Name.Val[1:]
			}
		}
	}
}

func (*AstFile) parseTopLevelLeadingComments(toks udevlex.Tokens) (ret []AstComment, rest udevlex.Tokens) {
	rest = toks
	for len(rest) > 0 && rest[0].Kind() == udevlex.TOKEN_COMMENT {
		rest = rest[1:]
	}
	if count := len(toks) - len(rest); count > 0 {
		ret = make([]AstComment, count)
		for i := range ret {
			ret[i].initFrom(toks, i)
		}
	}
	return
}

func (me *AstFile) parseTopLevelDef(tokens udevlex.Tokens) (def *AstDef, err *atmo.Error) {
	ctx := ctxParseTld{file: me, mtc: make(map[*udevlex.Token][]int), mto: make(map[*udevlex.Token]int)}
	var astdef AstDef
	if err = ctx.parseDef(tokens, true, &astdef); err == nil {
		def = &astdef
	}
	return
}

func (me *ctxParseTld) parseDef(tokens udevlex.Tokens, isTopLevel bool, def *AstDef) (err *atmo.Error) {
	toks := tokens
	if isTopLevel {
		toks = tokens.SansComments(me.mtc, me.mto)
	}
	if tokshead, tokheadbodysep, toksbody := toks.BreakOnOpish(":="); len(toksbody) == 0 {
		if tokens[0].Meta.Position.Column == 1 {
			err = atmo.ErrSyn(&tokens[0], "missing: definition body following `:=`")
		} else {
			err = atmo.ErrSyn(&tokens[0], "subsequent line in top-level expression needs more indentation")
		}
	} else if len(tokshead) == 0 {
		err = atmo.ErrSyn(&tokens[0], "missing: definition name preceding `:=`")
	} else if toksheads := tokshead.Chunked(","); len(toksheads[0]) == 0 {
		err = atmo.ErrSyn(&tokens[0], "missing: definition name preceding `,`")
	} else {
		if isTopLevel {
			me.curDef, def.Tokens, def.IsTopLevel = def, tokens, true
		} else {
			me.setTokensFor(&def.AstBaseTokens, toks, nil)
		}

		toksheadsig := toksheads[0]
		var exprsig IAstExpr
		if exprsig, err = me.parseExpr(toksheadsig); err == nil {
			switch sig := exprsig.(type) {
			case *AstIdent:
				def.Name, def.Args = *sig, nil
			case *AstExprAppl:
				switch nx := sig.Callee.(type) {
				case *AstIdent:
					def.Name, def.Args = *nx, make([]AstDefArg, len(sig.Args))
					for i := range sig.Args {
						if atom, ok1 := sig.Args[i].(IAstExprAtom); ok1 {
							def.Args[i].NameOrConstVal = atom
						} else {
							if appl, ok2 := sig.Args[i].(*AstExprAppl); ok2 {
								if colon, ok3 := appl.Callee.(*AstIdent); ok3 && colon.Val == ":" && len(appl.Args) == 2 {
									if atom, ok1 = appl.Args[0].(IAstExprAtom); ok1 {
										def.Args[i].NameOrConstVal, def.Args[i].Affix = atom, appl.Args[1]
									}
								}
							}
							if !ok1 {
								err = atmo.ErrSyn(&sig.Args[i].BaseTokens().Tokens[0], "invalid def arg: not atomic")
								return
							}
						}
					}
				default:
					err = atmo.ErrSyn(&nx.BaseTokens().Tokens[0], "invalid def name")
				}
			default:
				err = atmo.ErrSyn(&sig.BaseTokens().Tokens[0], "invalid def signature: must be lone ident or func-appl form")
			}
			if err == nil {
				if def.Body, err = me.parseExpr(toksbody); err == nil && len(toksheads) > 1 {
					if def.Meta, err = me.parseMetas(toksheads[1:]); err == nil {
						if me.indentHint = 0; toksbody[0].Meta.Position.Line == tokheadbodysep.Meta.Line {
							me.indentHint = tokheadbodysep.Meta.Position.Column - 1
						}
					}
				}
			}
		}
	}
	return
}

func (me *ctxParseTld) parseExpr(toks udevlex.Tokens) (ret IAstExpr, err *atmo.Error) {
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
				case "|":
					exprcur, toks, err = me.parseExprCase(toks, accum, alltoks)
					accum = accum[:0]
				default:
					exprcur = me.newExprIdent(toks)
					toks = toks[1:]
				}
			case udevlex.TOKEN_SEPISH:
				if sub, rest, e := me.parseParens(toks); e != nil {
					err = e
				} else if len(sub) == 0 { // empty parens are otherwise useless so we'll use it as some builtin ident
					exprcur = me.newExprIdent(toks[:2])
					toks = rest
				} else if exprcur, err = me.parseExprInParens(sub); err == nil {
					toks = rest
				}
			default:
				err = atmo.ErrSyn(&toks[0], "the impossible: unrecognized token (new bug in parser, parseExpr needs updating)")
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
		accum = me.parseExprFinalizeClasp(accum, allToks, untilTok)
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

func (me *ctxParseTld) parseExprFinalizeClasp(accum []IAstExpr, allToks udevlex.Tokens, untilTok *udevlex.Token) []IAstExpr {
	var claspstart int
	for i := 1; i < len(accum); i++ {
		dist := accum[i].BaseTokens().Tokens.NumberOfCharsBetweenFirstAndLastOf(accum[i-1].BaseTokens().Tokens)
		if dist > 0 {
			if clasp := accum[claspstart:i]; len(clasp) > 1 {
				prefix, suffix := accum[:claspstart], accum[i:]
				accum = append(prefix, me.parseExprFinalize(clasp, allToks, untilTok))
				i = len(accum)
				accum = append(accum, suffix...)
			}
			claspstart = i
		}
	}
	if claspstart > 0 {
		if clasp := accum[claspstart:]; len(clasp) > 1 {
			accum = append(accum[:claspstart], me.parseExprFinalize(clasp, allToks, untilTok))
		}
	}
	return accum
}

func (me *ctxParseTld) parseExprInParens(toks udevlex.Tokens) (ret IAstExpr, err *atmo.Error) {
	me.parensLevel++
	ret, err = me.parseExpr(toks)
	me.parensLevel--
	return
}

func (me *ctxParseTld) parseExprCase(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, rest udevlex.Tokens, err *atmo.Error) {
	if len(toks) == 1 {
		err = atmo.ErrSyn(&toks[0], "missing expressions following `|` branching")
	}
	var scrutinee IAstExpr
	if len(accum) > 0 {
		scrutinee = me.parseExprFinalize(accum, allToks, &toks[0])
	}
	var caseof AstExprCase
	caseof.Scrutinee, caseof.defaultIndex = scrutinee, -1
	me.setTokensFor(&caseof.AstBaseTokens, allToks, nil)
	toks, rest = toks[1:].BreakOnIndent(allToks[0].Meta.LineIndent)
	alts := toks.Chunked("|")
	caseof.Alts = make([]AstCaseAlt, len(alts))
	var cond IAstExpr
	var hasmulticonds bool
	for i := range alts {
		if len(alts[i]) == 0 {
			err = atmo.ErrSyn(&toks[0], "malformed `|?` branching: empty case")
		} else if ifthen := alts[i].Chunked("?"); len(ifthen) > 2 {
			err = atmo.ErrSyn(&alts[i][0], "malformed `|?` branching: `|` case has more than one `?` result expression")
		} else if me.setTokensFor(&caseof.Alts[i].AstBaseTokens, alts[i], nil); len(ifthen[0]) == 0 {
			// the branching's "default" case (empty between `|` and `?`)
			if len(ifthen[1]) == 0 {
				err = atmo.ErrSyn(&alts[i][0], "malformed `|?` branching: default case has no result expression")
			} else if caseof.Alts[i].Body, err = me.parseExpr(ifthen[1]); err == nil && caseof.defaultIndex >= 0 {
				err = atmo.ErrSyn(&alts[i][0], "malformed `|?` branching: encountered a second default case, only at most one is permissible")
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
		for i := 0; i < len(caseof.Alts); i++ {
			if ca := &caseof.Alts[i]; ca.Body == nil {
				if i < len(caseof.Alts)-1 {
					canext := &caseof.Alts[i+1]
					canext.Conds = append(ca.Conds, canext.Conds...)
					caseof.removeAltAt(i)
					i--
				}
			}
		}
	}
	ret = &caseof
	return
}

func (me *ctxParseTld) parseExprLetInner(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, rest udevlex.Tokens, err *atmo.Error) {
	const errmsg = "missing definitions following `,` comma"
	if len(toks) == 1 {
		err = atmo.ErrSyn(&toks[0], errmsg)
		return
	}
	var body IAstExpr
	body = me.parseExprFinalize(accum, allToks, &toks[0])
	toks, rest = toks[1:].BreakOnIndent(allToks[0].Meta.LineIndent)
	if chunks := toks.Chunked(","); len(chunks) > 0 {
		var let AstExprLet
		let.Body, let.Defs = body, make([]AstDef, len(chunks))
		me.setTokensFor(&let.AstBaseTokens, allToks, nil)
		lasttokforerr := &toks[0]
		for i := range chunks {
			if len(chunks[i]) == 0 {
				err = atmo.ErrSyn(lasttokforerr, errmsg)
			} else if err = me.parseDef(chunks[i], false, &let.Defs[i]); err == nil {
				lasttokforerr = chunks[i].Last(nil)
			}
			if err != nil {
				return
			}
		}
		ret = &let
	}
	return
}

func (me *ctxParseTld) parseExprLetOuter(toks udevlex.Tokens, toksChunked []udevlex.Tokens) (ret *AstExprLet, err *atmo.Error) {
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

func (me *ctxParseTld) parseMetas(chunks []udevlex.Tokens) (metas []IAstExpr, err *atmo.Error) {
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

func (me *ctxParseTld) parseParens(toks udevlex.Tokens) (sub udevlex.Tokens, rest udevlex.Tokens, err *atmo.Error) {
	var numunclosed int
	if toks[0].Str == ")" {
		err = atmo.ErrSyn(&toks[0], "closing parenthesis without matching opening")
	} else if sub, rest, numunclosed = toks.Sub('(', ')'); len(sub) == 0 && numunclosed != 0 {
		err = atmo.ErrSyn(&toks[0], "unclosed parenthesis")
	}
	return
}
