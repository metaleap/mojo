package atmoast

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
	. "github.com/metaleap/atmo"
)

func (me *AstFile) parse(this *AstFileChunk) (freshErrs Errors) {
	toks := this.Ast.Tokens
	if this.Ast.comments.Leading, toks = parseLeadingComments(toks); len(toks) != 0 {
		if this.Ast.Def.Orig, this.Ast.Def.NameIfErr, this.errs.parsing = this.parseTopLevelDef(toks); this.errs.parsing != nil {
			freshErrs.Add(this.errs.parsing)
		} else if this.Ast.Def.IsUnexported = (this.Ast.Def.Orig.Name.Val[0] == '_' && len(this.Ast.Def.Orig.Name.Val) > 1); this.Ast.Def.IsUnexported {
			this.Ast.Def.Orig.Name.Val = this.Ast.Def.Orig.Name.Val[1:]
		}
	}
	return
}

func parseLeadingComments(toks udevlex.Tokens) (ret []AstComment, rest udevlex.Tokens) {
	toks, rest = toks.BreakOnLeadingComments()
	if len(toks) != 0 {
		ret = make([]AstComment, len(toks))
		for i := range ret {
			ret[i].initFrom(toks, i)
		}
	}
	return
}

func (me *AstFileChunk) parseTopLevelDef(tokens udevlex.Tokens) (def *AstDef, nameOnlyIfErr string, err *Error) {
	astdef := AstDef{IsTopLevel: true}
	ctx := ctxTldParse{curTopLevel: me, curTopDef: &astdef, bracketsHalfIdx: len(udevlex.SepsGroupers) / 2}
	if err = ctx.parseDef(tokens, &astdef); err == nil {
		def = &astdef
	} else {
		nameOnlyIfErr = astdef.Name.Val
	}
	return
}

func (me *ctxTldParse) parseDef(toks udevlex.Tokens, def *AstDef) (err *Error) {
	def.Tokens = toks
	if tokshead, tokheadbodysep, toksbody := toks.BreakOnOpish(KnownIdentDecl); len(toksbody) == 0 {
		if t := tokheadbodysep.Or(&toks[0]); toks[0].Pos.Col1 == 1 {
			err = ErrSyn(ErrParsing_DefBodyMissing, t, "expected definition body following `:=`")
		} else {
			err = ErrSyn(ErrParsing_DefMissing, t, "expected a def at this indentation level")
		}
	} else if len(tokshead) == 0 {
		err = ErrSyn(ErrParsing_DefHeaderMissing, &toks[0], "expected definition name preceding `:=`")
	} else if toksheads := tokshead.Chunked(",", ""); len(toksheads[0]) == 0 {
		err = ErrSyn(ErrParsing_DefHeaderMalformed, &toks[0], "expected definition name preceding `,`")
	} else if err = me.parseDefHeadSig(toksheads[0], def); err == nil {
		if err = me.parseDefBodyExpr(toksbody, def); err == nil {
			if len(toksheads) > 1 {
				def.Meta, err = me.parseMetas(toksheads[1:])
			}
		}
	}
	return
}

func (me *ctxTldParse) parseDefHeadSig(toksHeadSig udevlex.Tokens, def *AstDef) (err *Error) {
	parseaffix := func(colon *AstExprAppl) (IAstExpr, *Error) {
		var tsub udevlex.Tokens
		if len(colon.Args) > 2 {
			tsub = toksHeadSig.FindSub(colon.Args[1].Toks(), colon.Args[len(colon.Args)-1].Toks())
		}
		return me.parseExprApplOrIdent(colon.Args[1:], tsub)
	}

	var exprsig IAstExpr
	if exprsig, err = me.parseExpr(toksHeadSig); err == nil {
		switch sig := exprsig.(type) {
		case *AstIdent:
			def.Name, def.Args = *sig, nil
		case *AstExprAppl:
			switch nx := sig.Callee.(type) {
			case *AstIdent:
				def.Name = *nx
			case *AstExprAppl:
				var ok bool
				if colon, ok1 := nx.Callee.(*AstIdent); ok1 && colon.Val == ":" && len(nx.Args) >= 2 {
					if ident, ok2 := nx.Args[0].(*AstIdent); ok2 {
						ok, def.Name = true, *ident
						if def.NameAffix, err = parseaffix(nx); err != nil {
							return
						}
					}
				}
				if !ok {
					err = ErrSyn(ErrParsing_DefNameAffixMalformed, &nx.Toks()[0], "malformed affix in def name: `"+nx.Tokens.Orig()+"`")
				}
			default:
				err = ErrSyn(ErrParsing_DefNameMalformed, &nx.Toks()[0], "expected def name instead of expression `"+nx.Toks().Orig()+"`")
			}
			if err == nil {
				def.Args = make([]AstDefArg, len(sig.Args))
				for i := range sig.Args {
					def.Args[i].NameOrConstVal, def.Args[i].Tokens = sig.Args[i], sig.Args[i].Toks()
					if appl, oka := sig.Args[i].(*AstExprAppl); oka {
						if colon, okc := appl.Callee.(*AstIdent); okc && colon.Val == ":" {
							if len(appl.Args) >= 2 {
								def.Args[i].NameOrConstVal, def.Args[i].Tokens = appl.Args[0], appl.Tokens
								def.Args[i].Affix, err = parseaffix(appl)
							} else {
								err = ErrSyn(ErrParsing_DefArgAffixMalformed, &sig.Args[i].Toks()[0], "invalid def-arg affixing: expected 2 operand expressions surrounding `:`")
							}
						}
					}
					if err != nil {
						return
					}
				}
			}
		default:
			err = ErrSyn(ErrParsing_DefNameMalformed, &sig.Toks()[0], "expected def name instead of expression `"+sig.Toks().Orig()+"`")
		}
	}
	return
}

func (me *ctxTldParse) parseDefBodyExpr(toksBody udevlex.Tokens, def *AstDef) (err *Error) {
	var chunks []udevlex.Tokens
	if toksBody.MultipleLines() {
		chunks = toksBody.ChunkedByIndent(true, true)
	}
	if len(chunks) > 1 {
		def.Body, err = me.parseExprLetOuter(toksBody, chunks)
	} else {
		def.Body, err = me.parseExpr(toksBody)
	}
	return
}

func (me *ctxTldParse) parseExpr(toks udevlex.Tokens) (ret IAstExpr, err *Error) {
	alltoks, accum, greeds := toks, make([]IAstExpr, 0, len(toks)), toks.Cliques(func(i int, idxlast int) (isbreaker bool) {
		isbreaker = (toks[i].Lexeme == ",")
		if (!isbreaker) && i < idxlast && toks[i].Lexeme == ":" { /* special exception for ergonomic `fn: arg arg` --- suspends language integrity but if (fn:) was really wanted as standalone expression (hardly ever), can still parenthesize */
			isbreaker = toks[i+1].Pos.Off0 > toks[i].Pos.Off0+1
		}
		return
	})
	var exprcur IAstExpr
	var accumcomments []udevlex.Tokens
	for greed, hasgreeds := 0, len(greeds) != 0; err == nil && len(toks) != 0; exprcur = nil {
		if hasgreeds {
			greed = greeds[&toks[0]]
		}
		if greed > 1 {
			exprcur, err = me.parseExpr(toks[:greed])
			toks = toks[greed:]
		} else {
			switch toks[0].Kind {
			case udevlex.TOKEN_COMMENT:
				accumcomments = append(accumcomments, toks[0:1])
				toks = toks[1:]
			case udevlex.TOKEN_FLOAT:
				exprcur = me.parseExprLitFloat(toks)
				toks = toks[1:]
			case udevlex.TOKEN_UINT:
				exprcur = me.parseExprLitUint(toks)
				toks = toks[1:]
			case udevlex.TOKEN_STR:
				exprcur = me.parseExprLitStr(toks)
				toks = toks[1:]
			case udevlex.TOKEN_SEPISH:
				switch tok := toks[0].Lexeme; tok {
				case ",":
					exprcur, toks, err = me.parseCommaSeparated(toks, accum, alltoks)
					accum = accum[:0]
				default:
					if idx := ustr.IdxB(udevlex.SepsGroupers, tok[0]); idx < 0 {
						err = ErrBug(ErrParsing_TokenUnexpected_Separator, toks[0:1], "unexpected separator: `"+tok+"`")
					} else if sub, rest, e := me.parseBrackets(toks, idx); e != nil {
						err = e
					} else if len(sub) == 0 { // empty brackets turn into built-in idents with predefined meanings
						exprcur = me.parseExprIdent(toks[:2], true)
						toks = rest
					} else {
						me.brackets = append(me.brackets, tok[0])
						switch tok[0] {
						case '(':
							if exprcur, err = me.parseExprInParens(sub); err == nil {
								toks = rest
							}
						default:
							err = ErrTodo(0, sub, "not yet implemented: non-paren brackets such as `"+tok+"`")
						}
						me.brackets = me.brackets[:len(me.brackets)-1]
					}
				}
			case udevlex.TOKEN_IDENT, udevlex.TOKEN_OPISH:
				switch toks[0].Lexeme {
				case "?":
					exprcur, toks, err = me.parseExprCase(toks, accum, alltoks)
					accum = accum[:0]
				case KnownIdentDecl:
					err = ErrSyn(ErrParsing_TokenUnexpected_DefDecl, &toks[0], "unexpected `"+KnownIdentDecl+"` in expression (forgot a comma?)")
				default:
					ident := me.parseExprIdent(toks, false)
					if len(ident.Val) > 1 && ident.Val[0] == '_' && !ident.IsPlaceholder() {
						if ident.Val[1] == '_' {
							err = ErrNaming(ErrParsing_TokenUnexpected_Underscores, &toks[0], "multiple leading underscores: reserved for internal use")
						} else if !ustr.BeginsLower(ident.Val[1:]) {
							err = ErrNaming(ErrParsing_IdentExpected, &toks[0], "identifier expected to follow underscore")
						}
					}
					exprcur = ident
					toks = toks[1:]
				}
			default:
				panic(toks[0].Lexeme)
			}
		}
		if err == nil && exprcur != nil {
			if accum = append(accum, exprcur); len(accumcomments) != 0 {
				exprcur.Comments().Leading.initFrom(accumcomments)
				accumcomments = accumcomments[0:0]
			}
		}
	}
	if err == nil {
		if ret, err = me.parseExprApplOrIdent(accum, alltoks); err == nil && len(accumcomments) != 0 {
			ret.Comments().Trailing.initFrom(accumcomments)
		}
	}
	return
}

func (me *ctxTldParse) parseExprApplOrIdent(accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, err *Error) {
	if len(accum) == 0 {
		err = ErrSyn(ErrParsing_ExpressionMissing_Accum, &allToks[0], "expression expected")
	} else if len(accum) == 1 {
		ret = accum[0]
	} else {
		var appl AstExprAppl
		ret = &appl
		if allToks != nil {
			appl.Tokens = allToks
		}
		args := make([]IAstExpr, 1, len(accum)-1)
		var applstyle ApplStyle
		if me.curTopLevel != nil && me.curTopLevel.SrcFile != nil {
			applstyle = me.curTopLevel.SrcFile.Options.ApplStyle
		}
		switch applstyle {
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

func (me *ctxTldParse) parseExprCase(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, rest udevlex.Tokens, err *Error) {
	if len(toks) == 1 {
		err = ErrSyn(ErrParsing_ExpressionMissing_Case, &toks[0], "expected expression following `?` branching")
		return
	}
	var cases AstExprCases
	if len(accum) != 0 {
		cases.Scrutinee, _ = me.parseExprApplOrIdent(accum, allToks.FromUntil(nil, &toks[0], false))
	}
	cases.Tokens, cases.defaultIndex = allToks, -1
	toks, rest = toks[1:].BreakOnIndent(allToks[0].LineIndent)
	alts := toks.Chunked("|", "?")
	cases.Alts = make([]AstCase, len(alts))
	var cond IAstExpr
	var hasmulticonds bool
	for i := range alts {
		if len(alts[i]) == 0 {
			err = ErrSyn(ErrParsing_CaseEmpty, &toks[0], "expected one or more cases following `?`")
		} else if ifthen := alts[i].Chunked("=>", "?"); len(ifthen) > 2 {
			err = ErrSyn(ErrParsing_CaseNoPair, &alts[i][0], "expected one result expression following `=>`, not more")
		} else if len(ifthen) > 1 && len(ifthen[1]) == 0 {
			err = ErrSyn(ErrParsing_CaseNoResult, &alts[i][0], "expected one result expression following `=>`, not less")
		} else if len(ifthen[0]) == 0 {
			// the branching's "default" case (empty between `|` and `=>`)
			if cases.Alts[i].Body, err = me.parseExpr(ifthen[1]); err == nil && cases.defaultIndex >= 0 {
				err = ErrSyn(ErrParsing_CaseSecondDefault, &alts[i][0], "expected at most one default case, not multiple")
			} else {
				cases.defaultIndex = i
			}
		} else if cond, err = me.parseExpr(ifthen[0]); err == nil {
			if cases.Alts[i].Conds = []IAstExpr{cond}; len(ifthen) > 1 {
				cases.Alts[i].Body, err = me.parseExpr(ifthen[1])
			} else {
				hasmulticonds = true
			}
		}
		if err == nil {
			cases.Alts[i].Tokens = alts[i]
		} else {
			return
		}
	}
	if hasmulticonds {
		for i := 0; i < len(cases.Alts); i++ {
			if ca := &cases.Alts[i]; ca.Body == nil {
				if i < len(cases.Alts)-1 {
					canext := &cases.Alts[i+1]
					canext.Conds = append(ca.Conds, canext.Conds...)
					canext.Tokens = allToks.FindSub(ca.Toks(), canext.Toks())
					cases.removeAltAt(i)
					i--
				} else {
					err = ErrSyn(ErrParsing_CaseDisjNoResult, &ca.Tokens[0], "expected one result expression following `=>`, not less")
				}
			}
		}
	}
	ret = &cases
	return
}

func (me *ctxTldParse) parseCommaSeparated(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, rest udevlex.Tokens, err *Error) {
	tokcomma := &toks[0]
	var precomma IAstExpr
	if precomma, err = me.parseExprApplOrIdent(accum, allToks.FromUntil(nil, &toks[0], false)); err != nil {
		return
	}
	toks, rest = toks[1:].BreakOnIndent(allToks[0].LineIndent)
	numdefs, numothers, chunks := 0, 0, toks.Chunked(",", "")
	for len(chunks) != 0 && len(chunks[len(chunks)-1]) == 0 { // allow & drop trailing commas
		chunks = chunks[:len(chunks)-1]
	}
	toklast := tokcomma
	for i := range chunks {
		if len(chunks[i]) == 0 {
			err = ErrSyn(ErrParsing_CommasConsecutive, toks.Next(toklast, true),
				"consecutive commas with nothing between")
			return
		} else if toklast = chunks[i].Last1(); chunks[i].Has(KnownIdentDecl, false) {
			numdefs++
		} else {
			numothers++
		}
	}
	if numdefs != 0 && numothers == 0 {
		ret, err = me.parseExprLetInner(precomma, chunks, allToks)
	} else if numdefs != 0 && numothers != 0 {
		err = ErrAtPos(ErrCatParsing, ErrParsing_CommasMixDefsAndExprs,
			&tokcomma.Pos, allToks.FromUntil(tokcomma, toks.Last1(), true).Length(),
			"cannot group expressions and defs together (parenthesize to disambiguate)")
	} else { // for now, a comma-sep'd grouping is an appl with callee `,` and all items as args --- to be further desugared down to meaning contextually in irfun
		appl := AstExprAppl{Callee: me.parseExprIdent(allToks.FromUntil(tokcomma, tokcomma, true), false), Args: make([]IAstExpr, 1, 1+len(chunks))}
		appl.Args[0], appl.Tokens = precomma, allToks
		for i := range chunks {
			var arg IAstExpr
			if arg, err = me.parseExpr(chunks[i]); err != nil {
				return
			}
			appl.Args = append(appl.Args, arg)
		}
		ret = &appl
	}
	return
}

func (me *ctxTldParse) parseExprLetInner(body IAstExpr, chunks []udevlex.Tokens, allToks udevlex.Tokens) (ret IAstExpr, err *Error) {
	var let AstExprLet
	let.Tokens, let.Body, let.Defs = allToks, body, make([]AstDef, len(chunks))
	for i := range chunks {
		if err = me.parseDef(chunks[i], &let.Defs[i]); err != nil {
			return
		}
	}
	ret = &let
	return
}

func (me *ctxTldParse) parseExprLetOuter(toks udevlex.Tokens, toksChunked []udevlex.Tokens) (ret *AstExprLet, err *Error) {
	var let AstExprLet
	let.Tokens, let.Defs = toks, make([]AstDef, len(toksChunked)-1)
	for i := range toksChunked {
		if i == 0 {
			let.Body, err = me.parseExpr(toksChunked[i])
		} else {
			err = me.parseDef(toksChunked[i], &let.Defs[i-1])
		}
		if err != nil {
			return
		}
	}
	ret = &let
	return
}

func (me *ctxTldParse) parseExprInParens(toks udevlex.Tokens) (ret IAstExpr, err *Error) {
	ret, err = me.parseExpr(toks)
	return
}

func (me *ctxTldParse) parseBrackets(toks udevlex.Tokens, idx int) (sub udevlex.Tokens, rest udevlex.Tokens, err *Error) {
	var numunclosed int
	if idx >= me.bracketsHalfIdx {
		err = ErrSyn(ErrParsing_BracketUnopened, &toks[0], "closing bracket `"+toks[0].Lexeme+"` without matching opening")
	} else if sub, rest, numunclosed = toks.Sub(udevlex.SepsGroupers[idx], udevlex.SepsGroupers[len(udevlex.SepsGroupers)-(1+idx)]); len(sub) == 0 && numunclosed != 0 {
		err = ErrSyn(ErrParsing_BracketUnclosed, &toks[0], "unclosed bracket")
	}
	return
}

func (me *ctxTldParse) parseMetas(chunks []udevlex.Tokens) (metas []IAstExpr, err *Error) {
	metas = make([]IAstExpr, 0, len(chunks))
	var meta IAstExpr
	for i := range chunks {
		if len(chunks[i]) != 0 {
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
