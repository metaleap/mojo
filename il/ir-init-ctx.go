package atmoil

import (
	"strconv"

	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type ctxIrInit struct {
	curTopLevelDef  *IrDefTop
	defsScope       *IrDefs
	coerceCallables map[IIrNode]IIrExpr
	counter         struct {
		val   byte
		times int
	}
}

func ExprFrom(orig atmolang.IAstExpr) (IIrExpr, atmo.Errors) {
	var ctx ctxIrInit
	return ctx.newExprFrom(orig)
}

func (me *ctxIrInit) addCoercion(on IIrNode, coerce IIrExpr) {
	if me.coerceCallables == nil {
		me.coerceCallables = map[IIrNode]IIrExpr{on: coerce}
	} else {
		me.coerceCallables[on] = coerce
	}
}
func (me *ctxIrInit) ensureAtomic(expr IIrExpr) IIrExpr {
	if expr.IsAtomic() {
		return expr
	}
	return me.addLocalDefToOwnScope(me.nextPrefix(), expr)
}

func (me *ctxIrInit) newExprFromIdent(orig *atmolang.AstIdent) (ret IIrExpr, errs atmo.Errors) {
	if t1, t2 := orig.IsTag, ustr.BeginsUpper(orig.Val); t1 && t2 {
		var ident IrIdentTag
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig
	} else if t1 != t2 {
		panic("bug in `atmo/lang`: an `atmolang.AstIdent` had wrong `IsTag` value for its `Val` casing (Val: " + strconv.Quote(orig.Val) + " at " + orig.Tokens.First(nil).Meta.Pos.String() + ")")

	} else if orig.IsPlaceholder() {
		var ident IrSpecial // still return an arguably nonsensical but non-nil value, this allows other errors further down to still be collected as well
		ret, ident.OneOf.InvalidToken, ident.Orig = &ident, true, orig
		errs.AddSyn(&orig.Tokens[0], "misplaced placeholder: only legal in def-args or call expressions")

	} else if orig.IsVar() {
		var ident IrSpecial
		ret, ident.OneOf.InvalidToken, ident.Orig = &ident, true, orig
		errs.AddSyn(&orig.Tokens[0], "our bug, not your fault: encountered a var expression that wasn't desugared")

	} else if orig.Val == atmo.KnownIdentUndef {
		var ident IrSpecial
		ret, ident.OneOf.Undefined, ident.Orig = &ident, true, orig

	} else {
		var ident IrIdentName
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig
		// me.curTopLevelDef.refersTo[ident.Val] = true

	}
	return
}

func (me *ctxIrInit) newExprFrom(origin atmolang.IAstExpr) (expr IIrExpr, errs atmo.Errors) {
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
		expr = errs.AddVia(me.newExprFromIdent(origdes)).(IIrExpr)
	case *atmolang.AstExprLitFloat:
		var lit IrLitFloat
		lit.initFrom(me, origdes)
		expr = &lit
	case *atmolang.AstExprLitUint:
		var lit IrLitUint
		lit.initFrom(me, origdes)
		expr = &lit
	case *atmolang.AstExprLitRune:
		var lit IrLitRune
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
			errs.Add(let.Defs[i].initFrom(me, &origdes.Defs[i]))
		}
		expr = errs.AddVia(me.newExprFrom(origdes.Body)).(IIrExpr)
		let.Defs = append(let.Defs, sidedefs...)
		errs.Add(me.addLetDefsToNode(origdes.Body, expr, &let))
		me.defsScope = oldscope
	case *atmolang.AstExprAppl:
		origdes = origdes.ToUnary()
		appl, isatomiccallee, isatomicarg := IrAppl{Orig: origdes}, origdes.Callee.IsAtomic(), origdes.Args[0].IsAtomic()
		if isatomiccallee {
			appl.AtomicCallee = errs.AddVia(me.newExprFrom(origdes.Callee)).(IIrExpr)
		}
		if isatomicarg {
			appl.AtomicArg = errs.AddVia(me.newExprFrom(origdes.Args[0])).(IIrExpr)
		}
		if expr = &appl; !(isatomiccallee && isatomicarg) {
			oldscope, toatomic := me.defsScope, func(from atmolang.IAstExpr) IIrExpr {
				body := errs.AddVia(me.newExprFrom(from)).(IIrExpr)
				return me.addLocalDefToOwnScope(appl.letPrefix+me.nextPrefix(), body)
			}
			me.defsScope, appl.letPrefix = &appl.Defs, me.nextPrefix()
			if !isatomiccallee {
				appl.AtomicCallee = toatomic(origdes.Callee)
			}
			if !isatomicarg {
				appl.AtomicArg = toatomic(origdes.Args[0])
			}
			me.defsScope = oldscope
		}
	default:
		if tok := origin.Toks().First(nil); tok != nil {
			panic(tok.Meta.Pos.String())
		}
		panic(origdes)
	}
	if exprbase := expr.exprBase(); wasdesugared || exprbase.Orig == nil {
		exprbase.Orig = origin
	}
	return
}

func (me *ctxIrInit) nextPrefix() string {
	if me.counter.val == 122 || me.counter.val == 0 {
		me.counter.val, me.counter.times = 96, me.counter.times+1
	}
	me.counter.val++
	return "__" + string(ustr.RepeatB(me.counter.val, me.counter.times))
}

func (me *ctxIrInit) addLetDefsToNode(origBody atmolang.IAstExpr, letBody IIrExpr, let *IrExprLetBase) (errs atmo.Errors) {
	if dst := letBody.Let(); dst == nil {
		tok := origBody.Toks().First(nil)
		tokd, tokl := tok, 0
		if let.letOrig != nil && len(let.letOrig.Defs) > 0 {
			tokd = let.letOrig.Defs[0].Tokens.First(nil)
			toklast := let.letOrig.Defs[len(let.letOrig.Defs)-1].Tokens.Last(nil)
			tokl = (toklast.Meta.Pos.Offset + len(toklast.Meta.Orig)) - tokd.Meta.Pos.Offset
		}
		errs.AddUnreach(tokd, "can never be used: "+ustr.Plu(len(let.Defs), "local def")+" scoped only for `"+tok.Meta.Orig+"`", tokl)
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

func (me *ctxIrInit) addLocalDefToOwnScope(name string, body IIrExpr) *IrIdentName {
	var ident IrIdentName
	oldscope := me.defsScope
	me.defsScope = &ident.Defs
	ident.IrIdentBase = me.addLocalDefToScope(name, body).Name.IrIdentBase
	me.defsScope = oldscope
	return &ident
}

func (me *ctxIrInit) addLocalDefToScope(name string, body IIrExpr) (def *IrDef) {
	def = me.defsScope.add(body)
	def.Name.Val = name
	return
}
