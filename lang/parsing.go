package atmolang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
)

const ErrMsgDefNotExpr = "unexpected `:=` in expression, only permissible for defs (forgot a comma?)"

type ApplStyle int

const (
	APPLSTYLE_SVO ApplStyle = iota
	APPLSTYLE_VSO
	APPLSTYLE_SOV
)

func init() {
	udevlex.SepsGroupers, udevlex.SepsOthers, udevlex.RestrictedWhitespace, udevlex.SanitizeDirtyFloatsNextToDotOpishs =
		"([{}])", ",", true, true
}

func (me *AstFile) parse(this *SrcTopChunk) (freshErrs []error) {
	toks := this.Ast.Tokens
	if this.Ast.comments.Leading, toks = me.parseTopLevelLeadingComments(toks); len(toks) > 0 {
		if this.Ast.Def.Orig, this.Ast.Def.NameIfErr, this.errs.parsing = me.parseTopLevelDef(toks); this.errs.parsing != nil {
			freshErrs = append(freshErrs, this.errs.parsing)
		} else if this.Ast.Def.IsUnexported = (this.Ast.Def.Orig.Name.Val[0] == '_' && len(this.Ast.Def.Orig.Name.Val) > 1); this.Ast.Def.IsUnexported {
			this.Ast.Def.Orig.Name.Val = this.Ast.Def.Orig.Name.Val[1:]
		}
	}
	return
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

func (me *AstFile) parseTopLevelDef(tokens udevlex.Tokens) (def *AstDef, nameOnlyIfErr string, err *atmo.Error) {
	ctx := ctxTldParse{file: me}
	astdef := AstDef{IsTopLevel: true}
	if err = ctx.parseDef(tokens, &astdef); err == nil {
		def = &astdef
	} else {
		nameOnlyIfErr = astdef.Name.Val
	}
	return
}

func (me *ctxTldParse) parseDef(toks udevlex.Tokens, def *AstDef) (err *atmo.Error) {
	def.Tokens = toks
	if tokshead, tokheadbodysep, toksbody := toks.BreakOnOpish(":="); len(toksbody) == 0 {
		if t := tokheadbodysep.Or(&toks[0]); toks[0].Meta.Pos.Column == 1 {
			err = atmo.ErrSyn(t, "missing: definition body following `:=`")
		} else {
			err = atmo.ErrSyn(t, "at this indentation level, expected a def")
		}
	} else if len(tokshead) == 0 {
		err = atmo.ErrSyn(&toks[0], "missing: definition name preceding `:=`")
	} else if toksheads := tokshead.Chunked(",", ""); len(toksheads[0]) == 0 {
		err = atmo.ErrSyn(&toks[0], "missing: definition name preceding `,`")
	} else if err = me.parseDefHeadSig(toksheads[0], def); err == nil {
		me.exprWillBeDefBody, toksbody = toksbody.BreakOnLeadingComments()
		if me.indentHintForLet = 0; toksbody[0].Meta.Pos.Line == tokheadbodysep.Meta.Pos.Line {
			me.indentHintForLet = toksbody[0].Meta.Pos.Column - 1
		}
		if def.Body, err = me.parseExpr(toksbody); err == nil {
			if len(toksheads) > 1 {
				def.Meta, err = me.parseMetas(toksheads[1:])
			}
		}
	}
	return
}

func (me *ctxTldParse) parseDefHeadSig(toksHeadSig udevlex.Tokens, def *AstDef) (err *atmo.Error) {
	parseaffix := func(appl *AstExprAppl) IAstExpr {
		var tsub udevlex.Tokens
		if len(appl.Args) > 2 {
			tsub = toksHeadSig.FindSub(appl.Args[1].Toks(), appl.Args[len(appl.Args)-1].Toks())
		}
		return me.parseExprApplOrIdent(appl.Args[1:], tsub)
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
						ok, def.Name, def.NameAffix = true, *ident, parseaffix(nx)
					}
				}
				if !ok {
					err = atmo.ErrSyn(&nx.Toks()[0], "invalid def name")
				}
			default:
				err = atmo.ErrSyn(&nx.Toks()[0], "invalid def name")
			}
			if err == nil {
				def.Args = make([]AstDefArg, len(sig.Args))
				for i := range sig.Args {
					if atom, okatom := sig.Args[i].(IAstExprAtomic); okatom {
						def.Args[i].Tokens, def.Args[i].NameOrConstVal = atom.Toks(), atom
					} else {
						if appl, oka := sig.Args[i].(*AstExprAppl); oka {
							if colon, okc := appl.Callee.(*AstIdent); okc && colon.Val == ":" && len(appl.Args) >= 2 {
								if atom, okatom = appl.Args[0].(IAstExprAtomic); okatom {
									def.Args[i].Tokens, def.Args[i].NameOrConstVal, def.Args[i].Affix =
										appl.Tokens, atom, parseaffix(appl)
								}
							}
						}
						if !okatom {
							err = atmo.ErrSyn(&sig.Args[i].Toks()[0], "invalid def arg: needs to be atomic, or atomic:some-qualifying-expression")
							return
						}
					}
				}
			}
		default:
			err = atmo.ErrSyn(&sig.Toks()[0], "expected: def name")
		}
	}
	return
}

func (me *ctxTldParse) parseExpr(toks udevlex.Tokens) (ret IAstExpr, err *atmo.Error) {
	indhint := toks[0].Meta.Pos.Column - 1
	if me.indentHintForLet != 0 {
		indhint, me.indentHintForLet = me.indentHintForLet, 0
	}
	if me.exprWillBeDefBody != nil {
		me.exprWillBeDefBody = nil
		if chunks := toks.IndentBasedChunks(indhint); len(chunks) > 1 {
			ret, err = me.parseExprLetOuter(toks, chunks)
			return
		}
	}

	alltoks, accum, greeds := toks, make([]IAstExpr, 0, len(toks)), toks.CrampedOnes(func(i int) (isbreaker bool) {
		isbreaker = (toks[i].Meta.Orig == ",")
		if (!isbreaker) && toks[i].Meta.Orig == ":" { /* special exception for ergonomic `fn: arg arg` --- suspends language integrity but if (fn:) was really wanted as standalone expression (hardly ever), can still parenthesize */
			return i < len(toks)-1 && toks[i+1].Meta.Pos.Offset > toks[i].Meta.Pos.Offset+1
		}
		return
	})
	var exprcur IAstExpr
	var accumcomments []udevlex.Tokens
	for greed, hasgreeds := 0, greeds != nil; err == nil && len(toks) > 0; exprcur = nil {
		if hasgreeds {
			greed = greeds[&toks[0]]
		}
		if greed > 1 {
			exprcur, err = me.parseExpr(toks[:greed])
			toks = toks[greed:]
		} else {
			switch tkind := toks[0].Kind(); tkind {
			case udevlex.TOKEN_COMMENT:
				accumcomments = append(accumcomments, toks[0:1])
				toks = toks[1:]
			case udevlex.TOKEN_FLOAT:
				exprcur = me.parseExprLitFloat(toks)
				toks = toks[1:]
			case udevlex.TOKEN_UINT:
				exprcur = me.parseExprLitUint(toks)
				toks = toks[1:]
			case udevlex.TOKEN_RUNE:
				exprcur = me.parseExprLitRune(toks)
				toks = toks[1:]
			case udevlex.TOKEN_STR:
				exprcur = me.parseExprLitStr(toks)
				toks = toks[1:]
			case udevlex.TOKEN_SEPISH:
				switch toks[0].Str {
				case ",":
					exprcur, toks, err = me.parseCommaSeparated(toks, accum, alltoks)
					accum = accum[:0]
				default:

					if sub, rest, e := me.parseParens(toks); e != nil {
						err = e
					} else if len(sub) == 0 { // empty parens are otherwise useless so we'll use it as some builtin ident
						exprcur = me.parseExprIdent(toks[:2], true)
						toks = rest
					} else if exprcur, err = me.parseExprInParens(sub); err == nil {
						toks = rest
					}
				}
			case udevlex.TOKEN_IDENT, udevlex.TOKEN_OPISH:
				switch toks[0].Str {
				case "?":
					exprcur, toks, err = me.parseExprCase(toks, accum, alltoks)
					accum = accum[:0]
				case ":=":
					err = atmo.ErrSyn(&toks[0], ErrMsgDefNotExpr)
				default:
					ident := me.parseExprIdent(toks, false)
					if len(ident.Val) > 1 && ident.Val[0] == '_' && ident.Val[1] == '_' && !ident.IsPlaceholder() {
						err = atmo.ErrNaming(&toks[0], ustr.Int(ustr.CountPrefixRunes(ident.Val, '_'))+" leading underscores in identifier: this is reserved for internal use")
					}
					exprcur = ident
					toks = toks[1:]
				}
			default:
				panic("the impossible: unrecognized token (new bug in parser, parseExpr needs updating) at " + toks[0].Meta.Pos.String() + ", `" + toks[0].Meta.Orig + "`")
			}
		}
		if err == nil && exprcur != nil {
			if accum = append(accum, exprcur); len(accumcomments) > 0 {
				exprcur.Comments().Leading.initFrom(accumcomments)
				accumcomments = accumcomments[0:0]
			}
		}
	}
	if err == nil {
		if ret = me.parseExprApplOrIdent(accum, alltoks); len(accumcomments) > 0 {
			ret.Comments().Trailing.initFrom(accumcomments)
		}
	}
	return
}

func (me *ctxTldParse) parseExprApplOrIdent(accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr) {
	if len(accum) == 1 {
		ret = accum[0]
	} else {
		var appl AstExprAppl
		ret = &appl
		if allToks != nil {
			appl.Tokens = allToks
		}
		args := make([]IAstExpr, 1, len(accum)-1)
		var applstyle ApplStyle
		if me.file != nil {
			applstyle = me.file.Options.ApplStyle
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

func (me *ctxTldParse) parseExprCase(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, rest udevlex.Tokens, err *atmo.Error) {
	if len(toks) == 1 {
		err = atmo.ErrSyn(&toks[0], "missing expressions following `?` branching")
		return
	}
	var cases AstExprCases
	if len(accum) > 0 {
		cases.Scrutinee = me.parseExprApplOrIdent(accum, allToks.FromUntil(nil, &toks[0], false))
	}
	cases.Tokens, cases.defaultIndex = allToks, -1
	toks, rest = toks[1:].BreakOnIndent(allToks[0].Meta.LineIndent)
	alts := toks.Chunked("|", "?")
	cases.Alts = make([]AstCase, len(alts))
	var cond IAstExpr
	var hasmulticonds bool
	for i := range alts {
		if len(alts[i]) == 0 {
			err = atmo.ErrSyn(&toks[0], "malformed branching: empty case")
		} else if ifthen := alts[i].Chunked("=>", "?"); len(ifthen) > 2 {
			err = atmo.ErrSyn(&alts[i][0], "malformed branching: case has more than one result expression")
		} else if len(ifthen) > 1 && len(ifthen[1]) == 0 {
			err = atmo.ErrSyn(&alts[i][0], "malformed branching: case has no result expression")
		} else if len(ifthen[0]) == 0 {
			// the branching's "default" case (empty between `|` and `=>`)
			if cases.Alts[i].Body, err = me.parseExpr(ifthen[1]); err == nil && cases.defaultIndex >= 0 {
				err = atmo.ErrSyn(&alts[i][0], "malformed branching: encountered a second default case, only at most one is permissible")
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
					err = atmo.ErrSyn(&ca.Tokens[0], "malformed branching: case has no result expression")
				}
			}
		}
	}
	ret = &cases
	return
}

func (me *ctxTldParse) parseCommaSeparated(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, rest udevlex.Tokens, err *atmo.Error) {
	tokcomma := &toks[0]
	if len(accum) == 0 {
		err = atmo.ErrSyn(tokcomma, "this part cannot begin with a comma")
		return
	}
	precomma := me.parseExprApplOrIdent(accum, allToks.FromUntil(nil, &toks[0], false))
	toks, rest = toks[1:].BreakOnIndent(allToks[0].Meta.LineIndent)
	numdefs, numothers, chunks := 0, 0, toks.Chunked(",", "")
	for len(chunks) > 0 && len(chunks[len(chunks)-1]) == 0 { // allow & drop trailing commas
		chunks = chunks[:len(chunks)-1]
	}
	toklast := tokcomma
	for i := range chunks {
		if len(chunks[i]) == 0 {
			err = atmo.ErrSyn(toks.Next(toklast, true), "consecutive commas with nothing between")
			return
		} else if toklast = chunks[i].Last1(); chunks[i].Has(":=", false) {
			numdefs++
		} else {
			numothers++
		}
	}
	if numdefs > 0 && numothers == 0 {
		ret, err = me.parseExprLetInner(precomma, chunks, allToks)
	} else if numdefs > 0 && numothers > 0 {
		err = atmo.ErrAtPos(atmo.ErrCatParsing, &tokcomma.Meta.Pos,
			allToks.FromUntil(tokcomma, toks.Last1(), true).Length(),
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

func (me *ctxTldParse) parseExprLetInner(body IAstExpr, chunks []udevlex.Tokens, allToks udevlex.Tokens) (ret IAstExpr, err *atmo.Error) {
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

func (me *ctxTldParse) parseExprLetOuter(toks udevlex.Tokens, toksChunked []udevlex.Tokens) (ret *AstExprLet, err *atmo.Error) {
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

func (me *ctxTldParse) parseExprInParens(toks udevlex.Tokens) (ret IAstExpr, err *atmo.Error) {
	me.parensLevel++
	ret, err = me.parseExpr(toks)
	me.parensLevel--
	return
}

func (me *ctxTldParse) parseParens(toks udevlex.Tokens) (sub udevlex.Tokens, rest udevlex.Tokens, err *atmo.Error) {
	var numunclosed int
	if toks[0].Str == ")" {
		err = atmo.ErrSyn(&toks[0], "closing parenthesis without matching opening")
	} else if sub, rest, numunclosed = toks.Sub('(', ')'); len(sub) == 0 && numunclosed != 0 {
		err = atmo.ErrSyn(&toks[0], "unclosed parenthesis")
	}
	return
}

func (me *ctxTldParse) parseMetas(chunks []udevlex.Tokens) (metas []IAstExpr, err *atmo.Error) {
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
