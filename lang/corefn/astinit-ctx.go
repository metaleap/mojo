package atmocorefn

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type ctxAstInit struct {
	dynNamePrefs   string
	nameReferences map[string]bool
	namesInScope   []*atmolang.AstIdent
	defsScope      *astDefs
	counter        struct {
		val   byte
		times int
	}
}

func (me *ctxAstInit) nameInScope(name *atmolang.AstIdent) *atmolang.AstIdent {
	for i := range me.namesInScope {
		if me.namesInScope[i].Val == name.Val {
			return me.namesInScope[i]
		}
	}
	return nil
}

func (me *ctxAstInit) namesInScopeAdd(errs *atmo.Errors, names ...*atmolang.AstIdent) (ok int) {
	ndone := make(map[string]bool, len(names))
	for i := range names {
		if ident := me.nameInScope(names[i]); ident != nil {
			errs.AddNaming(&names[i].Tokens[0], "name `"+names[i].Val+"` already taken by "+ident.Tokens[0].Meta.Position.String())
			names[i] = nil
		} else if ndone[names[i].Val] {
			names[i] = nil
		} else {
			ndone[names[i].Val] = true
		}
	}
	for i := range names {
		if names[i] != nil {
			ok, me.namesInScope = ok+1, append(me.namesInScope, names[i])
		}
	}
	return
}

func (me *ctxAstInit) namesInScopeDrop(num int) {
	me.namesInScope = me.namesInScope[:len(me.namesInScope)-num]
}

func (me *ctxAstInit) ensureAstAtomFor(expr IAstExpr) IAstExprAtomic {
	if xat, _ := expr.(IAstExprAtomic); xat != nil {
		return xat
	}
	return &me.addLocalDefToScope(expr, me.dynName(expr)).Name
}

func (me *ctxAstInit) newAstIdentFrom(orig *atmolang.AstIdent, isDecl bool) (ret IAstExprAtomic, errs atmo.Errors) {
	if t1, t2 := orig.IsTag, ustr.BeginsUpper(orig.Val); t1 && t2 {
		var ident AstIdentTag
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig
	} else if t1 != t2 {
		panic("bug in `atmo/lang`: an `atmolang.AstIdent` had wrong `IsTag` value for its `Val` casing")

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
	if !isDecl {
		switch ret.(type) {
		case *AstIdentPlaceholder:
			errs.AddSyn(&orig.Tokens[0], "lone placeholder illegal here: only permissible in def args or calls")
		case *AstIdentName:
			if me.nameInScope(orig) == nil {
				errs.AddNaming(&orig.Tokens[0], "name unknown: `"+orig.Val+"`")
			}
		}
	}
	return
}

func (me *ctxAstInit) newAstExprFrom(orig atmolang.IAstExpr) (expr IAstExpr, errs atmo.Errors) {
	switch o := orig.(type) {
	case *atmolang.AstIdent:
		expr, errs = me.newAstIdentFrom(o, false)
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
		oldscope, newdefs, newnamesinscope, let :=
			me.defsScope, astDefs{}, make([]*atmolang.AstIdent, len(o.Defs)), AstLet{Orig: o, prefix: me.nextPrefix(), Defs: make(astDefs, len(o.Defs))}
		for i := range o.Defs {
			newnamesinscope[i] = &o.Defs[i].Name
		}
		numnames := me.namesInScopeAdd(&errs, newnamesinscope...)

		me.defsScope = &newdefs
		for i := range o.Defs {
			errs.Add(let.Defs[i].initFrom(me, &o.Defs[i]))
		}

		let.Body = errs.AddVia(me.newAstExprFrom(o.Body)).(IAstExpr)
		let.Defs = append(let.Defs, newdefs...)
		me.namesInScopeDrop(numnames)
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
