package atmolang_irfun

import (
	"strconv"
	"text/scanner"

	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type ctxAstInit struct {
	curTopLevel     *atmolang.AstDef
	dynNamePrefs    string
	namesInScopeOld map[string]atmolang.IAstNode
	namesInScope    []astNameInScope
	nameReferences  map[string]bool
	defsScope       *AstDefs
	coerceFuncs     map[IAstNode]IAstExpr
	counter         struct {
		val   byte
		times int
	}
}

type astNameInScope struct {
	name  string
	pos   scanner.Position
	arity bool
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
	return &me.addLocalDefToScope(expr, me.dynName(expr)).Name
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
	for des := orig.Desugared(me.nextPrefix); des != nil; des = orig.Desugared(me.nextPrefix) {
		orig = des
	}
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
		numnames, oldscope, sidedefs, let :=
			0, me.defsScope, AstDefs{}, AstExprLetBase{letOrig: o, letPrefix: me.nextPrefix(), letDefs: make(AstDefs, len(o.Defs))}
		me.defsScope = &sidedefs
		for i := range o.Defs {
			numnames += me.namesInScopeAdd(&errs, &o.Defs[i])
		}
		for i := range o.Defs {
			errs.Add(let.letDefs[i].initFrom(me, &o.Defs[i]))
		}
		expr = errs.AddVia(me.newAstExprFrom(o.Body)).(IAstExpr)
		let.letDefs = append(let.letDefs, sidedefs...)
		errs.Add(me.addLetDefsToNode(o.Body, expr, &let))
		me.defsScope = oldscope
		me.namesInScopeDrop(numnames)
	case *atmolang.AstExprAppl:
		o = o.ToUnary()
		appl, atc, ata := AstAppl{Orig: o}, o.Callee.IsAtomic(), o.Args[0].IsAtomic()
		if atc {
			appl.AtomicCallee = errs.AddVia(me.newAstExprFrom(o.Callee)).(IAstExpr)
		}
		if ata {
			appl.AtomicArg = errs.AddVia(me.newAstExprFrom(o.Args[0])).(IAstExpr)
		}
		if expr = &appl; !(atc && ata) {
			oldscope, toatomic := me.defsScope, func(from atmolang.IAstExpr) IAstExpr {
				body := errs.AddVia(me.newAstExprFrom(from)).(IAstExpr)
				return &me.addLocalDefToScope(body, appl.letPrefix+body.dynName()).Name
			}
			me.defsScope, appl.letPrefix = &appl.letDefs, me.nextPrefix()
			if !atc {
				appl.AtomicCallee = toatomic(o.Callee)
			}
			if !ata {
				appl.AtomicArg = toatomic(o.Args[0])
			}
			me.defsScope = oldscope
		}
	case *atmolang.AstExprCases:
		var cases AstCases
		errs = cases.initFrom(me, o)
		expr = &cases
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
	return me.dynNamePrefs + expr.dynName()
}

func (me *ctxAstInit) nextPrefix() string {
	if me.counter.val == 122 || me.counter.val == 0 {
		me.counter.val, me.counter.times = 96, me.counter.times+1
	}
	me.counter.val++
	return string(ustr.RepeatB(me.counter.val, me.counter.times))
}

func (me *ctxAstInit) addLetDefsToNode(origBody atmolang.IAstExpr, letBody IAstExpr, letDefs *AstExprLetBase) (errs atmo.Errors) {
	var dst *AstExprLetBase
	switch body := letBody.(type) {
	case *AstIdentName:
		dst = &body.AstExprLetBase
	case *AstAppl:
		dst = &body.AstExprLetBase
	case *AstCases:
		dst = &body.AstExprLetBase
	}
	if dst == nil {
		errs.AddSyn(&origBody.Toks()[0], "cannot declare local defs for `"+origBody.Toks()[0].Meta.Orig+"`")
	} else {
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

func (me *ctxAstInit) nameInScope(name *atmolang.AstIdent) *astNameInScope {
	for i := range me.namesInScope {
		if me.namesInScope[i].name == name.Val {
			return &me.namesInScope[i]
		}
	}
	return nil
}

func (me *ctxAstInit) namesInScopeAdd(errs *atmo.Errors, names ...atmolang.IAstNode) (ok int) {
	for i := range names {
		if len(names[i].Toks()) > 0 {
			var arity bool
			name, _ := names[i].(*atmolang.AstIdent)
			if name == nil {
				def := names[i].(*atmolang.AstDef)
				name, arity = &def.Name, len(def.Args) > 0
			}
			if len(name.Tokens) > 0 { // could have been def so different than seemingly-same above check
				if existing := me.nameInScope(name); existing != nil && !(arity && existing.arity) {
					errs.AddNaming(&name.Tokens[0], "name `"+name.Val+"` already taken by "+existing.pos.String())
				} else {
					ok, me.namesInScope = ok+1, append(me.namesInScope, astNameInScope{name.Val, name.Tokens[0].Meta.Position, arity})
				}
			}
		}
	}
	return
}

func (me *ctxAstInit) namesInScopeDrop(num int) {
	me.namesInScope = me.namesInScope[:len(me.namesInScope)-num]
}
