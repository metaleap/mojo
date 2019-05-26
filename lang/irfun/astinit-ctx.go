package atmolang_irfun

import (
	"strconv"

	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type ctxAstInit struct {
	curTopLevelDef *AstDefTop
	defsScope      *AstDefs
	coerceFuncs    map[IAstNode]IAstExpr
	counter        struct {
		val   byte
		times int
	}
}

func ExprFrom(orig atmolang.IAstExpr) (IAstExpr, atmo.Errors) {
	var ctx ctxAstInit
	return ctx.newAstExprFrom(orig)
}

func (me *ctxAstInit) addCoercion(on IAstNode, coerce IAstExpr) {
	if me.coerceFuncs == nil {
		me.coerceFuncs = map[IAstNode]IAstExpr{on: coerce}
	} else {
		me.coerceFuncs[on] = coerce
	}
}

func (me *ctxAstInit) ensureAstAtomFor(expr IAstExpr) IAstExpr {
	if expr.IsAtomic() {
		return expr
	}
	return &me.addLocalDefToScope(expr, me.nextPrefix()).Name
}

func (me *ctxAstInit) newAstIdentFrom(orig *atmolang.AstIdent) (ret IAstExpr, errs atmo.Errors) {
	if t1, t2 := orig.IsTag, ustr.BeginsUpper(orig.Val); t1 && t2 {
		var ident AstIdentTag
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig
	} else if t1 != t2 {
		panic("bug in `atmo/lang`: an `atmolang.AstIdent` had wrong `IsTag` value for its `Val` casing (Val: " + strconv.Quote(orig.Val) + " at " + ustr.If(len(orig.Tokens) == 0, "<dyn>", orig.Tokens[0].Meta.Position.String()) + ")")

	} else if orig.IsOpish && orig.Val == "()" {
		var ident AstIdentEmptyParens
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	} else if ustr.IsRepeat(orig.Val, '_') {
		var ident AstIdentVar // still return an arguably nonsensical but non-nil value, this allows other errors further down to still be found as well
		errs.AddSyn(&orig.Tokens[0], "illegal placeholder placement: valid in def-args or call expressions")
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	} else if orig.Val[0] == '_' && orig.Val[1] != '_' {
		var ident AstIdentVar
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	} else {
		var ident AstIdentName
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig
		// me.curTopLevelDef.refersTo[ident.Val] = true

	}
	return
}

func (me *ctxAstInit) newAstExprFrom(origin atmolang.IAstExpr) (expr IAstExpr, errs atmo.Errors) {
	origdesugared := origin.Desugared(me.nextPrefix)
	for des := origdesugared; des != nil; {
		if des = des.Desugared(me.nextPrefix); des != nil {
			origdesugared = des
		}
	}
	if origdesugared == nil {
		origdesugared = origin
	}

	switch origdes := origdesugared.(type) {
	case *atmolang.AstIdent:
		expr, errs = me.newAstIdentFrom(origdes)
	case *atmolang.AstExprLitFloat:
		var lit AstLitFloat
		lit.initFrom(me, origdes)
		expr = &lit
	case *atmolang.AstExprLitUint:
		var lit AstLitUint
		lit.initFrom(me, origdes)
		expr = &lit
	case *atmolang.AstExprLitRune:
		var lit AstLitRune
		lit.initFrom(me, origdes)
		expr = &lit
	case *atmolang.AstExprLitStr:
		var lit AstLitStr
		lit.initFrom(me, origdes)
		expr = &lit
	case *atmolang.AstExprLet:
		oldscope, sidedefs, let :=
			me.defsScope, AstDefs{}, AstExprLetBase{letOrig: origdes, letPrefix: me.nextPrefix(), letDefs: make(AstDefs, len(origdes.Defs))}
		me.defsScope = &sidedefs
		for i := range origdes.Defs {
			errs.Add(let.letDefs[i].initFrom(me, &origdes.Defs[i]))
		}
		expr = errs.AddVia(me.newAstExprFrom(origdes.Body)).(IAstExpr)
		let.letDefs = append(let.letDefs, sidedefs...)
		errs.Add(me.addLetDefsToNode(origdes.Body, expr, &let))
		me.defsScope = oldscope
	case *atmolang.AstExprAppl:
		origdes = origdes.ToUnary()
		appl, atc, ata := AstAppl{Orig: origdes}, origdes.Callee.IsAtomic(), origdes.Args[0].IsAtomic()
		if atc {
			appl.AtomicCallee = errs.AddVia(me.newAstExprFrom(origdes.Callee)).(IAstExpr)
		}
		if ata {
			appl.AtomicArg = errs.AddVia(me.newAstExprFrom(origdes.Args[0])).(IAstExpr)
		}
		if expr = &appl; !(atc && ata) {
			oldscope, toatomic := me.defsScope, func(from atmolang.IAstExpr) IAstExpr {
				body := errs.AddVia(me.newAstExprFrom(from)).(IAstExpr)
				return &me.addLocalDefToScope(body, appl.letPrefix+me.nextPrefix()).Name
			}
			me.defsScope, appl.letPrefix = &appl.letDefs, me.nextPrefix()
			if !atc {
				appl.AtomicCallee = toatomic(origdes.Callee)
			}
			if !ata {
				appl.AtomicArg = toatomic(origdes.Args[0])
			}
			me.defsScope = oldscope
		}
	default:
		panic(origdes)
	}
	expr.astExprBase().Orig = origin
	return
}

func (me *ctxAstInit) nextPrefix() string {
	if me.counter.val == 122 || me.counter.val == 0 {
		me.counter.val, me.counter.times = 96, me.counter.times+1
	}
	me.counter.val++
	return "__" + string(ustr.RepeatB(me.counter.val, me.counter.times))
}

func (me *ctxAstInit) addLetDefsToNode(origBody atmolang.IAstExpr, letBody IAstExpr, letDefs *AstExprLetBase) (errs atmo.Errors) {
	if letbase, _ := letBody.(IAstExprWithLetDefs); letbase == nil {
		errs.AddSyn(&origBody.Toks()[0], "cannot declare local defs for `"+origBody.Toks()[0].Meta.Orig+"`")
	} else {
		dst := letbase.astExprLetBase()
		if dst.letPrefix == "" {
			dst.letPrefix = me.nextPrefix()
		}
		if dst.letDefs = append(dst.letDefs, letDefs.letDefs...); dst.letOrig == nil {
			dst.letOrig = letDefs.letOrig
		}
	}
	return
}

func (me *ctxAstInit) addLocalDefToScope(body IAstExpr, name string) (def *AstDef) {
	def = me.defsScope.add(body)
	def.Name.Val = name
	return
}
