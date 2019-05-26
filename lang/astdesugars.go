package atmolang

import (
	"github.com/go-leap/str"
)

func (*AstBaseExpr) Desugared(func() string) IAstExpr { return nil }

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

// func (me *AstExprCases) Desugared(prefix func() string) IAstExpr {
// 	if me.defaultIndex >= 0 {
// 		panic("to-be-handled: AstExprCases.defaultIndex")
// 	}
// 	if me.Scrutinee == nil {

// 	} else {

// 	}
// 	return nil
// }

// func (me *AstCases) initFrom(ctx *ctxAstInit, orig *atmolang.AstExprCases) (errs atmo.Errors) {
// 	me.Orig = orig

// 	var scrut IAstExpr
// 	if orig.Scrutinee != nil {
// 		scrut = errs.AddVia(ctx.newAstExprFrom(orig.Scrutinee)).(IAstExpr)
// 	} else {
// 		scrut = B.IdentTag("True")
// 	}
// 	scrut = B.Appl(B.IdentName("=="), ctx.ensureAstAtomFor(scrut))
// 	scrutatomic, opeq := ctx.ensureAstAtomFor(scrut), B.IdentName("||")

// 	me.Ifs, me.Thens = make([]IAstExpr, len(orig.Alts)), make([]IAstExpr, len(orig.Alts))
// 	for i := range orig.Alts {
// 		alt := &orig.Alts[i]
// 		if alt.Body == nil {
// 			errs.AddSyn(&alt.Tokens[0], "malformed branching: case has no result expression (or nested branchings should be parenthesized)")
// 		} else {
// 			me.Thens[i] = errs.AddVia(ctx.newAstExprFrom(alt.Body)).(IAstExpr)
// 		}
// 		if len(alt.Conds) == 0 {
// 			errs.AddTodo(&alt.Tokens[0], "deriving of default case")
// 		}
// 		for c, cond := range alt.Conds {
// 			if c == 0 {
// 				me.Ifs[i] = B.Appl(scrutatomic, ctx.ensureAstAtomFor(errs.AddVia(ctx.newAstExprFrom(cond)).(IAstExpr)))
// 			} else {
// 				me.Ifs[i] = B.Appls(ctx, opeq, ctx.ensureAstAtomFor(me.Ifs[i]), ctx.ensureAstAtomFor(
// 					B.Appl(scrutatomic, ctx.ensureAstAtomFor(errs.AddVia(ctx.newAstExprFrom(cond)).(IAstExpr)))))
// 			}
// 		}
// 	}
// 	return
// }
