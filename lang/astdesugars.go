package atmolang

import (
	"github.com/go-leap/str"
)

func (me *AstExprAppl) Desugared(prefix func() string) IAstExpr {
	if lamb := me.desugarToLetExprIfPlaceholders(prefix); lamb != nil {
		return lamb
	} else if lamb := me.desugarToLetExprIfUnionTest(prefix); lamb != nil {
		return lamb
	}
	return nil
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
	if !me.HasPlaceholders {
		return nil
	}
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
