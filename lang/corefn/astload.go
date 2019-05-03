package atmocorefn

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func (me *AstDef) nameInScope(name *atmolang.AstIdent) *atmolang.AstIdent {
	for i := range me.state.namesInScope {
		if me.state.namesInScope[i].Val == name.Val {
			return me.state.namesInScope[i]
		}
	}
	return nil
}

func (me *AstDef) namesInScopeAdd(errs *atmo.Errors, names ...*atmolang.AstIdent) (ok int) {
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
			ok, me.state.namesInScope = ok+1, append(me.state.namesInScope, names[i])
		}
	}
	return
}

func (me *AstDef) namesInScopeDrop(num int) {
	me.state.namesInScope = me.state.namesInScope[:len(me.state.namesInScope)-num]
}

func (me *AstDefBase) initFrom(ctx *AstDef, orig *atmolang.AstDef) (errs atmo.Errors) {
	me.Orig = orig
	var numnewnames int
	for i := range me.Orig.Args {
		if name, _ := me.Orig.Args[i].NameOrConstVal.(*atmolang.AstIdent); name != nil {
			numnewnames += ctx.namesInScopeAdd(&errs, name)
		}
	}
	errs.Add(me.initName(ctx))
	errs.Add(me.initArgs(ctx))
	errs.Add(me.initMetas(ctx))
	errs.Add(me.initBody(ctx))
	if !me.Orig.IsTopLevel {
		ctx.dynNameDrop(me.Orig.Name.Val)
	}
	ctx.namesInScopeDrop(numnewnames)
	return
}

func (me *AstDefBase) initName(ctx *AstDef) (errs atmo.Errors) {
	tok := me.Orig.Name.Tokens.First(nil) // could have none so dont just Tokens[0]
	var ident IAstExprAtomic
	ident, errs = ctx.newAstIdentFrom(&me.Orig.Name, true)
	if name, _ := ident.(*AstIdentName); name == nil {
		errs.AddNaming(tok, "invalid def name: `"+tok.Meta.Orig+"`") // Tag or EmptyParens or Placeholder etc..
	} else if me.Name = *name; name.Val == "" /*|| ustr.In(name.Val, langReservedOps...)*/ {
		errs.AddNaming(tok, "reserved token not permissible as def name: `"+tok.Meta.Orig+"`")
	}
	if !me.Orig.IsTopLevel {
		ctx.dynNameAdd(me.Orig.Name.Val)
	}
	if me.Orig.NameAffix != nil {
		me.nameCoerceFunc = errs.AddVia(ctx.newAstExprFrom(me.Orig.NameAffix)).(IAstExpr)
	}
	return
}

func (me *AstDefBase) initBody(ctx *AstDef) (errs atmo.Errors) {
	me.Body, errs = ctx.newAstExprFrom(me.Orig.Body)

	opeq := ctx.b.IdentName("==")
	for i := range me.Args {
		if me.Args[i].coerceValue != nil {
			me.Body = ctx.b.Case(ctx.b.Appls(ctx, opeq, &me.Args[i].AstIdentName, me.Args[i].coerceValue), me.Body)
		}
		if me.Args[i].coerceFunc != nil {
			appl := ctx.b.Appl(ctx.ensureAstAtomFor(me.Args[i].coerceFunc), &me.Args[i].AstIdentName)
			me.Body = ctx.b.Case(ctx.b.Appls(ctx, opeq, &me.Args[i].AstIdentName, ctx.ensureAstAtomFor(appl)), me.Body)
		}
	}
	if me.nameCoerceFunc != nil {
		appl := ctx.b.Appl(ctx.ensureAstAtomFor(me.nameCoerceFunc), &me.Name)
		me.Body = ctx.b.Case(ctx.b.Appls(ctx, opeq, &me.Name, ctx.ensureAstAtomFor(appl)), me.Body)
	}

	return
}

func (me *AstDefBase) initArgs(ctx *AstDef) (errs atmo.Errors) {
	if len(me.Orig.Args) > 0 {
		args := make([]AstDefArg, len(me.Orig.Args))
		for i := range me.Orig.Args {
			errs.Add(args[i].initFrom(ctx, &me.Orig.Args[i], i))
		}
		me.Args = args
	}
	return
}

func (me *AstDefBase) initMetas(ctx *AstDef) (errs atmo.Errors) {
	if len(me.Orig.Meta) > 0 {
		errs.AddTodo(&me.Orig.Meta[0].Toks()[0], "def metas")
		for i := range me.Orig.Meta {
			_ = errs.AddVia(ctx.newAstExprFrom(me.Orig.Meta[i]))
		}
	}
	return
}

func (me *AstDefArg) initFrom(ctx *AstDef, orig *atmolang.AstDefArg, argIdx int) (errs atmo.Errors) {
	me.Orig = orig

	var constexpr IAstExprAtomic
	switch v := orig.NameOrConstVal.(type) {
	case *atmolang.AstIdent:
		if constexpr, errs = ctx.newAstIdentFrom(v, true); constexpr != nil {
			if cxn, ok1 := constexpr.(*AstIdentName); ok1 {
				constexpr, me.AstIdentName = nil, *cxn
			} else if cxu, ok2 := constexpr.(*AstIdentPlaceholder); ok2 {
				if constexpr, me.AstIdentName.Val, me.AstIdentName.Orig = nil, ustr.Int(argIdx)+"ª", v; cxu.Num() > 1 {
					errs.AddNaming(&v.Tokens[0], "invalid def arg name")
				}
			}
		}
	case *atmolang.AstExprLitFloat, *atmolang.AstExprLitUint, *atmolang.AstExprLitRune, *atmolang.AstExprLitStr:
		constexpr, errs = ctx.newAstExprAtomicFrom(v)
	default:
		panic(v)
	}
	if constexpr != nil {
		me.AstIdentName.Val = ustr.Int(argIdx) + "ª"
		me.coerceValue = constexpr
	}

	if orig.Affix != nil {
		me.coerceFunc = errs.AddVia(ctx.newAstExprFrom(orig.Affix)).(IAstExpr)
	}
	return
}

func (me *AstCases) initFrom(ctx *AstDef, orig *atmolang.AstExprCases) (errs atmo.Errors) {
	me.Orig = orig

	var scrut IAstExpr
	if orig.Scrutinee != nil {
		scrut = errs.AddVia(ctx.newAstExprFrom(orig.Scrutinee)).(IAstExpr)
	} else {
		scrut = ctx.b.IdentTagTrue()
	}
	scrut = ctx.b.Appl(ctx.b.IdentName("=="), ctx.ensureAstAtomFor(scrut))
	scrutatomic := ctx.ensureAstAtomFor(scrut).(IAstExprAtomic)

	me.Ifs, me.Thens = make([][]IAstExpr, len(orig.Alts)), make([]IAstExpr, len(orig.Alts))
	for i := range orig.Alts {
		alt := &orig.Alts[i]
		if alt.Body == nil {
			panic("bug in atmo/lang: received an `AstCase` with a `nil` `Body`")
		} else {
			me.Thens[i] = errs.AddVia(ctx.newAstExprFrom(alt.Body)).(IAstExpr)
		}
		me.Ifs[i] = make([]IAstExpr, len(alt.Conds))
		for c, cond := range alt.Conds {
			me.Ifs[i][c] = errs.AddVia(ctx.newAstExprFrom(cond)).(IAstExpr)
			me.Ifs[i][c] = ctx.b.Appl(scrutatomic, ctx.ensureAstAtomFor(me.Ifs[i][c]))
		}
	}
	return
}

func (me *AstLitBase) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.Orig = orig
}

func (me *AstLitFloat) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Float
}

func (me *AstLitUint) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Uint
}

func (me *AstLitRune) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Rune()
}

func (me *AstLitStr) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Str
}

func (me *AstDef) ensureAstAtomFor(expr IAstExpr) IAstExprAtomic {
	if xat, _ := expr.(IAstExprAtomic); xat != nil {
		return xat
	}
	return &me.addLocalDefToScope(expr, me.dynName(expr)).Name
}

func (me *AstDef) newAstIdentFrom(orig *atmolang.AstIdent, isDecl bool) (ret IAstExprAtomic, errs atmo.Errors) {
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

func (me *AstDef) newAstExprAtomicFrom(orig atmolang.IAstExprAtomic) (expr IAstExprAtomic, errs atmo.Errors) {
	expr = errs.AddVia(me.newAstExprFrom(orig)).(IAstExprAtomic)
	return
}

func (me *AstDef) newAstExprFrom(orig atmolang.IAstExpr) (expr IAstExpr, errs atmo.Errors) {
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
			me.state.defsScope, astDefs{}, make([]*atmolang.AstIdent, len(o.Defs)), AstLet{Orig: o, prefix: me.nextClet(), Defs: make(astDefs, len(o.Defs))}
		for i := range o.Defs {
			newnamesinscope[i] = &o.Defs[i].Name
		}
		numnames := me.namesInScopeAdd(&errs, newnamesinscope...)

		me.state.defsScope = &newdefs
		for i := range o.Defs {
			errs.Add(let.Defs[i].initFrom(me, &o.Defs[i]))
		}

		let.Body = errs.AddVia(me.newAstExprFrom(o.Body)).(IAstExpr)
		let.Defs = append(let.Defs, newdefs...)
		me.namesInScopeDrop(numnames)
		expr, me.state.defsScope = &let, oldscope
	case *atmolang.AstExprAppl:
		if lamb := o.ToLetExprIfPlaceholders(me.nextClamb()); lamb != nil {
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
				oldscope, let := me.state.defsScope, AstLet{prefix: me.nextClet(), Body: &appl}
				if me.state.defsScope = &let.Defs; !atc {
					def := let.Defs.add(errs.AddVia(me.newAstExprFrom(o.Callee)).(IAstExpr))
					def.Name.Val = ustr.Int(let.prefix) + "c" + def.Body.DynName()
					appl.Callee = &def.Name
				}
				if !ata {
					def := let.Defs.add(errs.AddVia(me.newAstExprFrom(o.Args[0])).(IAstExpr))
					def.Name.Val = ustr.Int(let.prefix) + "c" + def.Body.DynName()
					appl.Arg = &def.Name
				}
				expr, me.state.defsScope = &let, oldscope
			}
		}
	case *atmolang.AstExprCases:
		if let := o.ToLetIfUnionSugar(me.nextCsumt()); let == nil {
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

func (me *AstDef) dynNameAdd(s string) string {
	old := me.state.dynNamePrefs
	me.state.dynNamePrefs += s
	return old
}

func (me *AstDef) dynNameDrop(s string) string {
	me.state.dynNamePrefs = me.state.dynNamePrefs[:len(me.state.dynNamePrefs)-len(s)]
	return me.state.dynNamePrefs
}

func (me *AstDef) dynName(expr IAstExpr) string {
	return me.state.dynNamePrefs + expr.DynName()
}

func (me *AstDef) nextClamb() int {
	me.state.counters.lamb++
	return me.state.counters.lamb
}

func (me *AstDef) nextClet() int {
	me.state.counters.let++
	return me.state.counters.let
}

func (me *AstDef) nextCsumt() int {
	me.state.counters.sumt++
	return me.state.counters.sumt
}

func (me *AstDef) addLocalDefToScope(body IAstExpr, name string) (def *AstDefBase) {
	def = me.state.defsScope.add(body)
	def.Name.Val = name
	return
}
