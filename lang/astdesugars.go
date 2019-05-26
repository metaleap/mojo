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
		def := B.Def("┬"+prefix(), &check, "specimen")
		check.Scrutinee, check.Alts[0].Body = def.Args[0].NameOrConstVal, def.Args[0].NameOrConstVal
		return B.Let(&def.Name, *def)
	}
	return nil
}

func (me *AstExprAppl) desugarToLetExprIfPlaceholders(prefix func() string) *AstExprLet {
	var num int
	var lamc string
	var lama []string
	if ident, _ := me.Callee.(*AstIdent); ident != nil && ustr.IsRepeat(ident.Val, '_') {
		num, lamc = num+1, ustr.Int(len(ident.Val)-1)+"_"
	}
	for i := range me.Args {
		if ident, _ := me.Args[i].(*AstIdent); ident != nil && ustr.IsRepeat(ident.Val, '_') {
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
		def.Args[i].NameOrConstVal = B.Ident(ustr.Int(i) + "_")
	}
	var appl AstExprAppl
	appl.Callee, appl.Args = me.Callee, make([]IAstExpr, len(me.Args))
	if lamc != "" {
		appl.Callee = B.Ident(lamc)
	}
	for i := range appl.Args {
		if la := lama[i]; la != "" {
			appl.Args[i] = B.Ident(la)
		} else {
			appl.Args[i] = me.Args[i]
		}
	}
	def.Body = &appl
	return B.Let(&def.Name, def)
}

func (me *AstExprCases) Desugared(prefix func() string) (expr IAstExpr, errs atmo.Errors) {
	havescrut := (me.Scrutinee != nil)
	var opeq, scrut *AstIdent
	var let *AstExprLet
	if havescrut {
		let = &AstExprLet{Defs: []AstDef{{Body: me.Scrutinee}}}
		let.AstBaseTokens, let.AstBaseComments, opeq, scrut, let.Defs[0].Tokens, let.Defs[0].Name.Val, let.Defs[0].Name.Tokens =
			me.AstBaseTokens, me.AstBaseComments, B.Ident("=="), &let.Defs[0].Name, me.Scrutinee.Toks(), prefix()+"scrut", me.Scrutinee.Toks()
	}
	var appl, applcur *AstExprAppl
	var defcase IAstExpr
	for i := range me.Alts {
		if alt := me.Alts[i]; alt.Body == nil {
			errs.AddSyn(&alt.Tokens[0], "malformed branching: case has no result expression (or nested branchings should be parenthesized)")
		} else if len(alt.Conds) == 0 {
			defcase = alt.Body
		} else {
			alt.Conds = make([]IAstExpr, len(alt.Conds)) // copying slice, too
			if havescrut {
				for c := range alt.Conds {
					alt.Conds[c] = B.Appl(opeq, scrut, me.Alts[i].Conds[c])
				}
			} else {
				copy(alt.Conds, me.Alts[i].Conds)
			}
			for len(alt.Conds) > 1 {
				cond0, cond1, opor := alt.Conds[0], alt.Conds[1], B.Ident("||")
				cond := B.Appl(opor, cond0, cond1)
				cond.Tokens, opor.Tokens = alt.Tokens.FromUntil(cond0.Toks().First(nil), cond1.Toks().Last(nil), true), alt.Tokens.Between(cond0.Toks().Last(nil), cond1.Toks().First(nil))
				alt.Conds = append([]IAstExpr{cond}, alt.Conds[2:]...)
			}
			ite := B.Appl(B.Ident("?:"), alt.Conds[0], alt.Body, nil)
			if ite.AstBaseTokens = alt.AstBaseTokens; applcur != nil {
				applcur.Args[2] = ite
			}
			if applcur = ite; appl == nil {
				appl = applcur
			}
		}
	}
	if defcase == nil {
		defcase = B.Ident("()")
	}
	if appl.Args[2] = defcase; havescrut {
		expr, let.Body = let, appl
	} else {
		expr = appl
	}
	return
}
