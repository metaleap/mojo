package atmoil

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func ExprFrom(orig atmolang.IAstExpr) (IExpr, atmo.Errors) {
	var ctx ctxIrFromAst
	return ctx.newExprFrom(orig)
}

func (me *ctxIrFromAst) addCoercion(on INode, coerce IExpr) {
	if coerce != nil {
		if me.coerceCallables == nil {
			me.coerceCallables = map[INode]IExpr{on: coerce}
		} else {
			me.coerceCallables[on] = coerce
		}
	}
}

func (me *ctxIrFromAst) appl(ensureAtomic bool, orig atmolang.IAstExpr, applExpr IExpr) IExpr {
	if applExpr.exprBase().Orig = orig; ensureAtomic {
		applExpr = me.ensureAtomic(applExpr)
		applExpr.exprBase().Orig = orig
	}
	return applExpr
}

func (me *ctxIrFromAst) bodyWithCoercion(coerce IExpr, body IExpr, coerceArg func() IExpr) IExpr {
	corig, opeq := coerce.exprBase().Orig, Build.IdentName(atmo.KnownIdentEq)
	if coerceArg == nil {
		coerceArg = func() IExpr { return body }
	}
	coerceappl := me.appl(requireAtomicCalleeAndCallArg, corig, Build.Appl1(me.ensureAtomic(coerce), coerceArg()))
	cmpeq := me.appl(requireAtomicCalleeAndCallArg, corig, Build.ApplN(me, opeq, coerceappl, coerceArg()))
	return me.appl(false, corig, Build.ApplN(me, cmpeq, body, Build.Undef()))
}

func (me *ctxIrFromAst) ensureAtomic(expr IExpr) IExpr {
	if (!requireAtomicCalleeAndCallArg) || expr.IsAtomic() {
		return expr
	}
	return me.addLocalDefToOwnScope(me.nextPrefix(), expr)
}

func (me *ctxIrFromAst) newExprFromIdent(orig *atmolang.AstIdent) (ret IExpr, errs atmo.Errors) {
	if orig.IsTag {
		var ident IrIdentTag
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	} else if orig.IsPlaceholder() {
		var ident IrNonValue // still return an arguably nonsensical but non-nil value, this allows other errors further down to still be collected as well
		ret, ident.OneOf.LeftoverPlaceholder, ident.Orig = &ident, true, orig
		errs.AddBug(ErrFromAst_UnhandledStandaloneUnderscores, orig.Tokens, "misplaced placeholder: only legal in def-args or call expressions")

	} else if orig.Val == atmo.KnownIdentUndef {
		var ident IrNonValue
		ret, ident.OneOf.Undefined, ident.Orig = &ident, true, orig

	} else {
		var ident IrIdentName
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig
		if orig.IsVar() {
			ident.Val = ident.Val[1:]
		}
		if me.curTopLevelDef != nil {
			me.curTopLevelDef.refersTo[ident.Val] = true
		}

	}
	return
}

func (me *ctxIrFromAst) newExprFrom(origin atmolang.IAstExpr) (expr IExpr, errs atmo.Errors) {
	var origdesugared atmolang.IAstExpr
	var wasdesugared bool
	if origdesugared, errs = origin.Desugared(me.nextPrefix); origdesugared == nil {
		origdesugared = origin
	} else {
		wasdesugared = true
		for des := origdesugared; des != nil; {
			maybedes := errs.AddVia(des.Desugared(me.nextPrefix))
			if des, _ = maybedes.(atmolang.IAstExpr); des != nil {
				origdesugared = des
			}
		}
	}

	switch origdes := origdesugared.(type) {
	case *atmolang.AstIdent:
		expr = errs.AddVia(me.newExprFromIdent(origdes)).(IExpr)
	case *atmolang.AstExprLitFloat:
		var lit IrLitFloat
		lit.initFrom(me, origdes)
		expr = &lit
	case *atmolang.AstExprLitUint:
		var lit IrLitUint
		lit.initFrom(me, origdes)
		expr = &lit
	case *atmolang.AstExprLitStr:
		var lit IrLitStr
		lit.initFrom(me, origdes)
		expr = &lit
	case *atmolang.AstExprLet:
		oldscope, sidedefs, let :=
			me.defsScope, IrDefs{}, IrExprLetBase{letOrig: origdes, letPrefix: me.nextPrefix(), Defs: make(IrDefs, len(origdes.Defs))}
		me.defsScope = &sidedefs
		for i := range origdes.Defs {
			errs.Add(let.Defs[i].initFrom(me, &origdes.Defs[i])...)
		}
		expr = errs.AddVia(me.newExprFrom(origdes.Body)).(IExpr)
		let.Defs = append(let.Defs, sidedefs...)
		errs.Add(me.addLetDefsToNode(origdes.Body, expr, &let)...)
		me.defsScope = oldscope
	case *atmolang.AstExprAppl:
		origdes = origdes.ToUnary()
		appl, isatomiccallee, isatomicarg := IrAppl{Orig: origdes}, (!requireAtomicCalleeAndCallArg) || origdes.Callee.IsAtomic(), (!requireAtomicCalleeAndCallArg) || origdes.Args[0].IsAtomic()
		if isatomiccallee {
			appl.Callee = errs.AddVia(me.newExprFrom(origdes.Callee)).(IExpr)
		}
		if isatomicarg {
			appl.CallArg = errs.AddVia(me.newExprFrom(origdes.Args[0])).(IExpr)
		}
		if expr = &appl; !(isatomiccallee && isatomicarg) {
			oldscope, toatomic := me.defsScope, func(from atmolang.IAstExpr) IExpr {
				body := errs.AddVia(me.newExprFrom(from)).(IExpr)
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

func (me *ctxIrFromAst) addLetDefsToNode(origBody atmolang.IAstExpr, letBody IExpr, let *IrExprLetBase) (errs atmo.Errors) {
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

func (me *ctxIrFromAst) addLocalDefToOwnScope(name string, body IExpr) *IrIdentName {
	var ident IrIdentName
	oldscope := me.defsScope
	me.defsScope = &ident.Defs
	ident.IrIdentBase = me.addLocalDefToScope(name, body).Name.IrIdentBase
	me.defsScope = oldscope
	return &ident
}

func (me *ctxIrFromAst) addLocalDefToScope(name string, body IExpr) *IrDef {
	return me.defsScope.add(name, body)
}
