package atmolang

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
)

func (*AstBaseExpr) Desugared(func() string) (IAstExpr, atmo.Errors) { return nil, nil }

func (me *AstExprAppl) Desugared(prefix func() string) (IAstExpr, atmo.Errors) {
	if lamb := me.desugarToLetExprIfPlaceholders(prefix); lamb != nil {
		return lamb, nil
	} else if lamb := me.desugarToLetExprIfUnionTest(prefix); lamb != nil {
		return lamb, nil
	}
	return nil, nil
}

func (me *AstExprAppl) desugarToLetExprIfUnionTest(prefix func() string) *AstExprLet {
	if ident, _ := me.Callee.(*AstIdent); ident != nil && ident.Val == "," {
		check := AstExprCases{Alts: []AstCase{{Conds: make([]IAstExpr, len(me.Args))}}}
		for i := range me.Args {
			check.Alts[0].Conds[i] = me.Args[i]
		}
		def := Build.Def("┬"+prefix(), &check, prefix())
		check.Scrutinee, check.Alts[0].Body = def.Args[0].NameOrConstVal, def.Args[0].NameOrConstVal
		return Build.Let(&def.Name, *def)
	}
	return nil
}

func (me *AstExprAppl) desugarToLetExprIfPlaceholders(prefix func() string) *AstExprLet {
	var num int
	var lamc string
	var lama []string
	if ident, _ := me.Callee.(*AstIdent); ident != nil && ident.IsPlaceholder() {
		num, lamc = num+1, ustr.Int(len(ident.Val)-1)+"_"
	}
	for i := range me.Args {
		if ident, _ := me.Args[i].(*AstIdent); ident != nil && ident.IsPlaceholder() {
			if lama == nil {
				lama = make([]string, len(me.Args))
			}
			num, lama[i] = num+1, ustr.Int(len(ident.Val)-1)+"_"
		}
	}
	if lamc == "" && lama == nil {
		return nil
	}

	def := AstDef{Name: AstIdent{Val: prefix() + "┌"}, Args: make([]AstDefArg, num)}
	for i := range def.Args {
		def.Args[i].NameOrConstVal = Build.Ident(ustr.Int(i) + "_")
	}
	var appl AstExprAppl
	if appl.Callee = me.Callee; lamc != "" {
		appl.Callee = Build.Ident(lamc)
	}
	if appl.Args = me.Args; lama != nil {
		appl.Args = make([]IAstExpr, len(me.Args))
		for i := range appl.Args {
			if la := lama[i]; la != "" {
				appl.Args[i] = Build.Ident(la)
			} else {
				appl.Args[i] = me.Args[i]
			}
		}
	}
	def.Body = &appl
	return Build.Let(&def.Name, def)
}

func (me *AstExprCases) Desugared(prefix func() string) (expr IAstExpr, errs atmo.Errors) {
	havescrut := (me.Scrutinee != nil)
	var opeq, scrut *AstIdent
	var let *AstExprLet
	if havescrut {
		let = &AstExprLet{Defs: []AstDef{{Body: me.Scrutinee}}}
		let.AstBaseTokens, let.AstBaseComments, opeq, scrut, let.Defs[0].Tokens, let.Defs[0].Name.Val, let.Defs[0].Name.Tokens =
			me.AstBaseTokens, me.AstBaseComments, Build.Ident(atmo.KnownIdentEq), &let.Defs[0].Name, me.Scrutinee.Toks(), prefix()+"scrut", me.Scrutinee.Toks()
	}
	var appl, applcur *AstExprAppl
	var defcase IAstExpr
	for i := range me.Alts {
		if alt := me.Alts[i]; alt.Body == nil {
			errs.AddSyn(alt.Tokens, "malformed branching: case has no result expression (or nested branchings should be parenthesized)")
		} else if len(alt.Conds) == 0 {
			defcase = alt.Body
		} else {
			alt.Conds = make([]IAstExpr, len(alt.Conds)) // copying slice, too
			if havescrut {
				for c := range alt.Conds {
					alt.Conds[c] = Build.Appl(opeq, scrut, me.Alts[i].Conds[c])
				}
			} else {
				copy(alt.Conds, me.Alts[i].Conds)
			}
			for len(alt.Conds) > 1 {
				cond0, cond1, opor := alt.Conds[0], alt.Conds[1], Build.Ident(atmo.KnownIdentOpOr)
				cond := Build.Appl(opor, cond0, cond1)
				cond.Tokens, opor.Tokens = alt.Tokens.FromUntil(cond0.Toks().First1(), cond1.Toks().Last1(), true), alt.Tokens.Between(cond0.Toks().Last1(), cond1.Toks().First1())
				alt.Conds = append([]IAstExpr{cond}, alt.Conds[2:]...)
			}
			ifthenelse := Build.Appl(alt.Conds[0], alt.Body, nil)
			if ifthenelse.AstBaseTokens = alt.AstBaseTokens; applcur != nil {
				applcur.Args[1] = ifthenelse
			}
			if applcur = ifthenelse; appl == nil {
				appl = applcur
			}
		}
	}
	if defcase == nil {
		defcase = Build.Ident(atmo.KnownIdentUndef)
	}
	if appl.Args[1] = defcase; havescrut {
		expr, let.Body = let, appl
	} else {
		expr = appl
	}
	return
}
