package atmoil

import (
	"github.com/go-leap/str"
	. "github.com/metaleap/atmo"
	. "github.com/metaleap/atmo/ast"
)

func ExprFrom(orig IAstExpr) (IIrExpr, Errors) {
	var ctx ctxIrFromAst
	return ctx.newExprFrom(orig)
}

func (me *ctxIrFromAst) addCoercion(on IIrNode, coerce IIrExpr) {
	if coerce != nil {
		if me.coerceCallables == nil {
			me.coerceCallables = map[IIrNode]IIrExpr{on: coerce}
		} else {
			me.coerceCallables[on] = coerce
		}
	}
}

func (me *ctxIrFromAst) bodyWithCoercion(coerce IIrExpr, body IIrExpr, coerceArg func() IIrExpr) IIrExpr {
	corig, opeq := coerce.exprBase().Orig.(IAstExpr), BuildIr.IdentName(KnownIdentEq)
	if coerceArg == nil {
		coerceArg = func() IIrExpr { return body }
	}
	coerceappl := BuildIr.Appl1(coerce, coerceArg())
	cmpeq := BuildIr.ApplN(me, opeq, coerceappl, coerceArg())
	ret := BuildIr.ApplN(me, cmpeq, body, BuildIr.Undef())
	coerceappl.Orig, cmpeq.Orig, ret.Orig = corig, corig, corig
	return ret
}

func (me *ctxIrFromAst) newExprFromIdent(orig *AstIdent) (ret IIrExpr, errs Errors) {
	if orig.IsTag {
		var expr IrLitTag
		ret, expr.Val, expr.Orig = &expr, orig.Val, orig

	} else if orig.IsPlaceholder() {
		var expr IrNonValue // still return an arguably nonsensical but non-nil value, this allows other errors further down to still be collected as well
		ret, expr.OneOf.LeftoverPlaceholder, expr.Orig = &expr, true, orig
		errs.AddBug(ErrFromAst_UnhandledStandaloneUnderscores, orig.Tokens, "misplaced placeholder: only legal in def-args or call expressions")

	} else if orig.Val == KnownIdentUndef {
		var expr IrNonValue
		ret, expr.OneOf.Undefined, expr.Orig = &expr, true, orig

	} else {
		var expr IrIdentName
		ret, expr.Name, expr.Orig = &expr, orig.Val, orig
		if orig.IsVar() {
			expr.Name = expr.Name[1:]
		}
		if me.curTopLevelDef != nil {
			me.curTopLevelDef.refersTo[expr.Name] = true
		}

	}
	return
}

func (me *ctxIrFromAst) newExprFrom(origin IAstExpr) (expr IIrExpr, errs Errors) {
	var origdesugared IAstExpr
	var wasdesugared bool
	if origdesugared, errs = origin.Desugared(me.nextPrefix); origdesugared == nil {
		origdesugared = origin
	} else {
		wasdesugared = true
		for des := origdesugared; des != nil; {
			maybedes := errs.AddVia(des.Desugared(me.nextPrefix))
			if des, _ = maybedes.(IAstExpr); des != nil {
				origdesugared = des
			}
		}
	}

	switch origdes := origdesugared.(type) {
	case *AstIdent:
		expr = errs.AddVia(me.newExprFromIdent(origdes)).(IIrExpr)
	case *AstExprLitFloat:
		var lit IrLitFloat
		lit.initFrom(me, origdes)
		expr = &lit
	case *AstExprLitUint:
		var lit IrLitUint
		lit.initFrom(me, origdes)
		expr = &lit
	case *AstExprLitStr:
		var lit IrNonValue
		lit.Orig, lit.OneOf.TempStrLit = origdes, true
		expr = &lit
	case *AstExprLet:
		expr = errs.AddVia(me.newExprFrom(origdes.Body)).(IIrExpr)
		for i := range origdes.Defs {
			astdef, abs := &origdes.Defs[i], IrAbs{Body: expr}
			abs.Orig, abs.Arg.Orig, abs.Arg.Name = astdef, &astdef.Name, astdef.Name.Val
			appl := IrAppl{Callee: &abs}
			if len(astdef.Args) == 0 && astdef.NameAffix == nil && len(astdef.Meta) == 0 {
				appl.CallArg = errs.AddVia(me.newExprFrom(astdef.Body)).(IIrExpr)
			} else {
				var def IrDef
				errs.Add(def.initFrom(me, astdef)...)
				appl.CallArg = def.Body
			}
			expr = &appl
		}
	case *AstExprAppl:
		origdes = origdes.ToUnary()
		appl := IrAppl{Callee: errs.AddVia(me.newExprFrom(origdes.Callee)).(IIrExpr),
			CallArg: errs.AddVia(me.newExprFrom(origdes.Args[0])).(IIrExpr)}
		appl.Orig, expr = origdes, &appl
	default:
		if tok := origin.Toks().First1(); tok != nil {
			panic(tok.Pos.String())
		}
		panic(origdes)
	}
	if exprbase := expr.exprBase(); wasdesugared || exprbase.Orig == nil {
		exprbase.Orig = origin
	}
	return
}

func (me *ctxIrFromAst) nextPrefix() string {
	if me.counter.val == 122 || me.counter.val == 0 {
		me.counter.val, me.counter.times = 96, me.counter.times+1
	}
	me.counter.val++
	return "__" + string(ustr.RepeatB(me.counter.val, me.counter.times))
}
