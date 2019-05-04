package atmocorefn

import (
	"strconv"

	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type ctxAstInit struct {
	dynNamePrefs   string
	nameReferences map[string]bool
	namesInScope   []*atmolang.AstIdent
	defsScope      *astDefs
	coerceFuncs    map[IAstNode]IAstExpr
	counter        struct {
		val   byte
		times int
	}
}

func (me *ctxAstInit) addCoercion(on IAstNode, coerce IAstExpr) {
	if me.coerceFuncs == nil {
		me.coerceFuncs = map[IAstNode]IAstExpr{on: coerce}
	} else {
		me.coerceFuncs[on] = coerce
	}
}

func (me *ctxAstInit) ensureAstAtomFor(expr IAstExpr) IAstExprAtomic {
	if xat, _ := expr.(IAstExprAtomic); xat != nil {
		return xat
	}
	return &me.addLocalDefToScope(expr, me.dynName(expr)).Name
}

func (me *ctxAstInit) newAstIdentFrom(orig *atmolang.AstIdent) (ret IAstExprAtomic, errs atmo.Errors) {
	if t1, t2 := orig.IsTag, ustr.BeginsUpper(orig.Val); t1 && t2 {
		var ident AstIdentTag
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig
	} else if t1 != t2 {
		panic("bug in `atmo/lang`: an `atmolang.AstIdent` had wrong `IsTag` value for its `Val` casing (Val: " + strconv.Quote(orig.Val) + " at " + ustr.If(len(orig.Tokens) == 0, "<dyn>", orig.Tokens[0].Meta.Position.String()) + ")")

	} else if orig.IsOpish && orig.Val == "()" {
		var ident AstIdentEmptyParens
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	} else if ustr.IsRepeat(orig.Val, '_') {
		var ident AstIdentPlaceholder
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	} else if orig.Val[0] == '_' {
		var ident AstIdentVar
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	} else {
		var ident AstIdentName
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	}
	return
}

func (me *ctxAstInit) newAstExprFrom(orig atmolang.IAstExpr) (expr IAstExpr, errs atmo.Errors) {
	switch o := orig.(type) {
	case *atmolang.AstIdent:
		expr, errs = me.newAstIdentFrom(o)
	case *atmolang.AstExprLitFloat:
		var lit AstLitFloat
		lit.initFrom(me, o)
		expr = &lit
	case *atmolang.AstExprLitUint:
		var lit AstLitUint
		lit.initFrom(me, o)
		expr = &lit
	case *atmolang.AstExprLitRune:
		var lit AstLitRune
		lit.initFrom(me, o)
		expr = &lit
	case *atmolang.AstExprLitStr:
		var lit AstLitStr
		lit.initFrom(me, o)
		expr = &lit
	case *atmolang.AstExprLet:
		oldscope, newdefs, let :=
			me.defsScope, astDefs{}, AstLet{Orig: o, prefix: me.nextPrefix(), Defs: make(astDefs, len(o.Defs))}

		me.defsScope = &newdefs
		for i := range o.Defs {
			errs.Add(let.Defs[i].initFrom(me, &o.Defs[i]))
		}

		let.Body = errs.AddVia(me.newAstExprFrom(o.Body)).(IAstExpr)
		let.Defs = append(let.Defs, newdefs...)
		expr, me.defsScope = &let, oldscope
	case *atmolang.AstExprAppl:
		if lamb := o.ToLetExprIfPlaceholders(me.nextPrefix()); lamb != nil {
			expr, errs = me.newAstExprFrom(lamb)
		} else {
			o = o.ToUnary()
			appl, atc, ata := AstAppl{Orig: o}, o.Callee.IsAtomic(), o.Args[0].IsAtomic()
			if atc {
				appl.Callee = errs.AddVia(me.newAstExprFrom(o.Callee)).(IAstExprAtomic)
			}
			if ata {
				appl.Arg = errs.AddVia(me.newAstExprFrom(o.Args[0])).(IAstExprAtomic)
			}
			if atc && ata {
				expr = &appl
			} else {
				// oldscope := me.defsScope
				let := AstLet{prefix: me.nextPrefix(), Body: &appl}
				// me.defsScope = &let.Defs
				if !atc {
					def := let.Defs.add(errs.AddVia(me.newAstExprFrom(o.Callee)).(IAstExpr))
					def.Name.Val = def.Body.DynName()
					appl.Callee = &def.Name
				}
				if !ata {
					def := let.Defs.add(errs.AddVia(me.newAstExprFrom(o.Args[0])).(IAstExpr))
					def.Name.Val = def.Body.DynName()
					appl.Arg = &def.Name
				}
				expr = &let
				// me.defsScope = oldscope
			}
		}
	case *atmolang.AstExprCases:
		if let := o.ToLetIfUnionSugar(me.nextPrefix()); let == nil {
			var cases AstCases
			errs = cases.initFrom(me, o)
			expr = &cases
		} else {
			expr, errs = me.newAstExprFrom(let)
		}
	default:
		panic(o)
	}
	return
}

func (me *ctxAstInit) dynNameAdd(s string) string {
	old := me.dynNamePrefs
	me.dynNamePrefs += s
	return old
}

func (me *ctxAstInit) dynNameDrop(s string) string {
	me.dynNamePrefs = me.dynNamePrefs[:len(me.dynNamePrefs)-len(s)]
	return me.dynNamePrefs
}

func (me *ctxAstInit) dynName(expr IAstExpr) string {
	return me.dynNamePrefs + expr.DynName()
}

func (me *ctxAstInit) nextPrefix() string {
	if me.counter.val == 122 || me.counter.val == 0 {
		me.counter.val, me.counter.times = 96, me.counter.times+1
	}
	me.counter.val++
	return string(ustr.RepeatB(me.counter.val, me.counter.times))
}

func (me *ctxAstInit) addLocalDefToScope(body IAstExpr, name string) (def *AstDefBase) {
	def = me.defsScope.add(body)
	def.Name.Val = name
	return
}
