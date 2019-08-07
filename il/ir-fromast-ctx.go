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

func (me *ctxIrFromAst) appl(ensureAtomic bool, orig IAstExpr, applExpr IIrExpr) IIrExpr {
	if applExpr.exprBase().Orig = orig; ensureAtomic {
		applExpr = me.ensureAtomic(applExpr)
		applExpr.exprBase().Orig = orig
	}
	return applExpr
}

func (me *ctxIrFromAst) bodyWithCoercion(coerce IIrExpr, body IIrExpr, coerceArg func() IIrExpr) IIrExpr {
	corig, opeq := coerce.exprBase().Orig.(IAstExpr), BuildIr.IdentName(KnownIdentEq)
	if coerceArg == nil {
		coerceArg = func() IIrExpr { return body }
	}
	coerceappl := me.appl(requireAtomicCalleeAndCallArg, corig, BuildIr.Appl1(me.ensureAtomic(coerce), coerceArg()))
	cmpeq := me.appl(requireAtomicCalleeAndCallArg, corig, BuildIr.ApplN(me, opeq, coerceappl, coerceArg()))
	return me.appl(false, corig, BuildIr.ApplN(me, cmpeq, body, BuildIr.Undef()))
}

func (me *ctxIrFromAst) ensureAtomic(expr IIrExpr) IIrExpr {
	if (!requireAtomicCalleeAndCallArg) || expr.IsAtomic() {
		return expr
	}
	return me.addLocalDefToOwnScope(me.nextPrefix(), expr)
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
		ret, expr.Val, expr.Orig = &expr, orig.Val, orig
		if orig.IsVar() {
			expr.Val = expr.Val[1:]
		}
		if me.curTopLevelDef != nil {
			me.curTopLevelDef.refersTo[expr.Val] = true
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
		if 0 > 1 {
			expr = errs.AddVia(me.newExprFrom(origdes.Body)).(IIrExpr)
			for i := range origdes.Defs {
				astdef := &origdes.Defs[i]
				lam := IrLam{Body: expr}
				lam.Orig, lam.Arg.exprBase().Orig, lam.Arg.Val = astdef, &astdef.Name, astdef.Name.Val
				appl := IrAppl{Callee: &lam, CallArg: errs.AddVia(me.newExprFrom(astdef.Body)).(IIrExpr)}
				expr = &appl
			}
		}
		if 1 > 0 {
			oldscope, sidedefs, let :=
				me.defsScope, IrDefs{}, IrExprLetBase{letOrig: origdes, letPrefix: me.nextPrefix(), Defs: make(IrDefs, len(origdes.Defs))}
			me.defsScope = &sidedefs
			for i := range origdes.Defs {
				errs.Add(let.Defs[i].initFrom(me, &origdes.Defs[i])...)
			}
			expr = errs.AddVia(me.newExprFrom(origdes.Body)).(IIrExpr)
			let.Defs = append(let.Defs, sidedefs...)
			errs.Add(me.addLetDefsToNode(origdes.Body, expr, &let)...)
			me.defsScope = oldscope
		}
	case *AstExprAppl:
		origdes = origdes.ToUnary()
		var appl IrAppl
		appl.Orig = origdes
		isatomiccallee, isatomicarg := (!requireAtomicCalleeAndCallArg) || origdes.Callee.IsAtomic(), (!requireAtomicCalleeAndCallArg) || origdes.Args[0].IsAtomic()
		if isatomiccallee {
			appl.Callee = errs.AddVia(me.newExprFrom(origdes.Callee)).(IIrExpr)
		}
		if isatomicarg {
			appl.CallArg = errs.AddVia(me.newExprFrom(origdes.Args[0])).(IIrExpr)
		}
		if expr = &appl; !(isatomiccallee && isatomicarg) {
			oldscope, toatomic := me.defsScope, func(from IAstExpr) IIrExpr {
				body := errs.AddVia(me.newExprFrom(from)).(IIrExpr)
				return me.addLocalDefToOwnScope(appl.letPrefix+me.nextPrefix(), body)
			}
			me.defsScope, appl.letPrefix = &appl.Defs, me.nextPrefix()
			if !isatomiccallee {
				appl.Callee = toatomic(origdes.Callee)
			}
			if !isatomicarg {
				appl.CallArg = toatomic(origdes.Args[0])
			}
			me.defsScope = oldscope
		}
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

func (me *ctxIrFromAst) addLetDefsToNode(origBody IAstExpr, letBody IIrExpr, let *IrExprLetBase) (errs Errors) {
	if dst := letBody.Let(); dst == nil {
		toks := origBody.Toks()
		if leto := let.letOrig; leto != nil {
			if toks = leto.Tokens; len(leto.Defs) > 0 {
				toks = toks.FromUntil(leto.Defs[0].Tokens.First1(), toks.Last1(), true)
			}
		}
	} else {
		if dst.letPrefix == "" {
			dst.letPrefix = me.nextPrefix()
		}
		if dst.Defs = append(dst.Defs, let.Defs...); dst.letOrig == nil {
			dst.letOrig = let.letOrig
		}
	}
	return
}

func (me *ctxIrFromAst) addLocalDefToOwnScope(name string, body IIrExpr) *IrIdentName {
	var ident IrIdentName
	oldscope := me.defsScope
	me.defsScope = &ident.Defs
	ident.IrIdentBase = me.addLocalDefToScope(name, body).Name.IrIdentBase
	me.defsScope = oldscope
	return &ident
}

func (me *ctxIrFromAst) addLocalDefToScope(name string, body IIrExpr) *IrDef {
	return me.defsScope.add(name, body)
}
