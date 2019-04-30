package atmocorefn

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func (me *AstDef) initFrom(orig *atmolang.AstDef) {
	const caplocals = 5
	me.Locals = make(astDefs, 0, caplocals)
	me.Errs.Add(me.AstDefBase.initFrom(me, orig))
	if len(me.Locals) > caplocals {
		println("LOCALDEFS", len(me.Locals))
	}
}

func (me *AstDef) NextName(prefix string) string {
	return prefix + ustr.Int(me.Next())
}

func (me *AstDef) Next() (counterIncr int) {
	counterIncr, me.state.counter = me.state.counter, me.state.counter+1
	return
}

func (me *AstDefBase) initFrom(ctx *AstDef, orig *atmolang.AstDef) (errs atmo.Errors) {
	me.Orig = orig
	errs.Add(me.initName(ctx))
	errs.Add(me.initBody(ctx))
	errs.Add(me.initArgs(ctx))
	errs.Add(me.initMetas(ctx))
	return
}

func (me *AstDefBase) initName(ctx *AstDef) (errs atmo.Errors) {
	tok := me.Orig.Name.Tokens.First(nil)
	if me.Name, errs = ctx.newAstIdentFrom(&me.Orig.Name); me.Name != nil {
		switch name := me.Name.(type) {
		case *AstIdentName:
			// all ok
		case *AstIdentOp:
			if name.Val == "" || ustr.In(name.Val, langReservedOps...) {
				errs.AddNaming(tok, "reserved token not permissible as def name: `"+tok.Meta.Orig+"`")
			}
		case *AstIdentTag:
			errs.AddNaming(tok, "invalid def name: `"+name.Val+"` is upper-case, this is reserved for tags")
		case *AstIdentVar:
			errs.AddNaming(tok, "invalid def name: `"+tok.Meta.Orig+"` (begins with multiple underscores)")
		case *AstIdentUnderscores:
			errs.AddNaming(tok, "invalid def name: `"+tok.Meta.Orig+"`")
		default:
			panic(name)
		}
	}
	if me.Orig.NameAffix != nil {
		errs.AddTodo(&me.Orig.NameAffix.Toks()[0], "def name affixes")
	}

	return
}

func (me *AstDefBase) initBody(ctx *AstDef) (errs atmo.Errors) {
	me.Body, errs = ctx.newAstExprFrom(me.Orig.Body)
	return
}

func (me *AstDefBase) initArgs(ctx *AstDef) (errs atmo.Errors) {
	if len(me.Orig.Args) > 0 {
		opeq, args := ctx.b.IdentOp("=="), make([]AstDefArg, len(me.Orig.Args))
		for i := range me.Orig.Args {
			if errs.Add(args[i].initFrom(ctx, &me.Orig.Args[i], i)); args[i].coerceValue != nil {
				me.Body = ctx.b.Case(ctx.b.Appls(ctx, opeq, &args[i].AstIdentName, args[i].coerceValue), me.Body)
			}
		}
		me.Args = args
	}
	return
}

func (me *AstDefBase) initMetas(ctx *AstDef) (errs atmo.Errors) {
	if len(me.Orig.Meta) > 0 {
		errs.AddTodo(&me.Orig.Meta[0].Toks()[0], "def metas")
	}
	return
}

func (me *AstDefArg) initFrom(ctx *AstDef, orig *atmolang.AstDefArg, argIdx int) (errs atmo.Errors) {
	me.Orig = orig

	var constexpr IAstExprAtomic
	switch v := orig.NameOrConstVal.(type) {
	case *atmolang.AstIdent:
		if constexpr, errs = ctx.newAstIdentFrom(v); constexpr != nil {
			if cx, ok := constexpr.(*AstIdentName); ok {
				constexpr, me.AstIdentName = nil, *cx
			}
		}
	case *atmolang.AstExprLitFloat, *atmolang.AstExprLitUint, *atmolang.AstExprLitRune, *atmolang.AstExprLitStr:
		constexpr, errs = ctx.newAstExprAtomicFrom(v)
	default:
		panic(v)
	}
	if constexpr != nil {
		me.AstIdentName.Val = "__arg__" + ustr.Int(argIdx)
		me.coerceValue = constexpr
	}

	if orig.Affix != nil {
		errs.AddTodo(&orig.Affix.Toks()[0], "def arg affixes")
	}
	return
}

func (me *AstAppl) initFrom(ctx *AstDef, orig *atmolang.AstExprAppl) (errs atmo.Errors) {
	if len(orig.Args) > 1 {
		errs = me.initFrom(ctx, orig.ToUnary())
	} else {
		c, e := ctx.ensureAstAtomFrom(orig.Callee, true)
		me.Arg, errs = ctx.ensureAstAtomFrom(orig.Args[0], false)
		me.Callee, errs = c.(IAstIdent), append(errs, e...)
	}
	me.Orig = orig
	return
}

func (me *AstCases) initFrom(ctx *AstDef, orig *atmolang.AstExprCases) (errs atmo.Errors) {
	me.Orig = orig
	var e atmo.Errors

	var scrut IAstExpr
	if orig.Scrutinee != nil {
		scrut, e = ctx.newAstExprFrom(orig.Scrutinee)
		errs.Add(e)
	} else {
		scrut = ctx.b.IdentTagTrue()
	}
	scrut = ctx.b.Appl(ctx.b.IdentOp("=="), ctx.ensureAstAtomFor(scrut, false, "__scrut__"))
	scrutid := ctx.ensureAstAtomFor(scrut, true, "__scrut_eq__").(IAstIdent)

	me.Ifs, me.Thens = make([][]IAstExpr, len(orig.Alts)), make([]IAstExpr, len(orig.Alts))
	for i := range orig.Alts {
		alt := &orig.Alts[i]
		if alt.Body == nil {
			panic("bug in atmo/lang: received an `AstCase` with a `nil` `Body`")
		} else {
			me.Thens[i], e = ctx.newAstExprFrom(alt.Body)
			errs.Add(e)
		}
		me.Ifs[i] = make([]IAstExpr, len(alt.Conds))
		for c, cond := range alt.Conds {
			me.Ifs[i][c], e = ctx.newAstExprFrom(cond)
			me.Ifs[i][c] = ctx.b.Appl(scrutid, ctx.ensureAstAtomFor(me.Ifs[i][c], false, "__cond_"+ustr.Int(i)+"_"+ustr.Int(c)+"__"))
			errs.Add(e)
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

func (me *AstDef) ensureAstAtomFor(expr IAstExpr, retMustBeIAstIdent bool, dynNameIfNeeded string) IAstExprAtomic {
	if !retMustBeIAstIdent {
		if xat, _ := expr.(IAstExprAtomic); xat != nil {
			return xat
		}
	}
	if xid, _ := expr.(IAstIdent); xid != nil {
		return xid
	}
	var def AstDefBase
	def.Name = me.b.IdentName(dynNameIfNeeded)
	def.Body = expr
	me.Locals = append(me.Locals, def)
	return me.Locals[len(me.Locals)-1].Name
}

func (me *AstDef) ensureAstAtomFrom(orig atmolang.IAstExpr, retMustBeIAstIdent bool) (ret IAstExprAtomic, errs atmo.Errors) {
	if (!retMustBeIAstIdent) && orig.IsAtomic() {
		return me.newAstExprAtomicFrom(orig.(atmolang.IAstExprAtomic))
	}
	if oid, _ := orig.(*atmolang.AstIdent); oid != nil {
		ret, errs = me.newAstIdentFrom(oid)
	} else {
		var def AstDefBase
		def.Body, errs = me.newAstExprFrom(orig)
		def.Name = me.b.IdentName(def.Body.DynName())
		me.Locals = append(me.Locals, def)
		ret = me.Locals[len(me.Locals)-1].Name
	}
	return
}

func (me *AstDef) newAstIdentFrom(orig *atmolang.AstIdent) (ret IAstIdent, errs atmo.Errors) {
	if t1, t2 := orig.IsTag, ustr.BeginsUpper(orig.Val); t1 && t2 {
		var ident AstIdentTag
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig
	} else if t1 != t2 {
		panic("bug in `atmo/lang`: an `atmolang.AstIdent` had wrong `IsTag` value for its `Val` casing")

	} else if orig.IsOpish {
		if orig.Val == "()" {
			var ident AstIdentEmptyParens
			ret, ident.Val, ident.Orig = &ident, orig.Val, orig
		} else {
			var ident AstIdentOp
			ret, ident.Val, ident.Orig = &ident, orig.Val, orig
		}

	} else if orig.Val[0] != '_' {
		var ident AstIdentName
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	} else if ustr.IsRepeat(orig.Val, '_') {
		var ident AstIdentUnderscores
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	} else if orig.Val[1] != '_' {
		var ident AstIdentVar
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	} else {
		errs.AddNaming(&orig.Tokens[0], "invalid identifier: begins with multiple underscores")
	}
	return
}

func (me *AstDef) newAstExprAtomicFrom(orig atmolang.IAstExprAtomic) (expr IAstExprAtomic, errs atmo.Errors) {
	x, e := me.newAstExprFrom(orig)
	errs.Add(e)
	if expr, _ = x.(IAstExprAtomic); expr == nil {
		panic("should-be-impossible bug just occurred in atmo/lang/corefn.AstDefnewAstExprAtomicFrom")
	}
	return
}

func (me *AstDef) newAstExprFrom(orig atmolang.IAstExpr) (expr IAstExpr, errs atmo.Errors) {
	if orig != nil {
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
			expr, errs = me.newAstExprFrom(o.Body)
			for i := range o.Defs {
				var def AstDefBase
				errs.Add(def.initFrom(me, &o.Defs[i]))
				me.Locals = append(me.Locals, def)
			}
		case *atmolang.AstExprAppl:
			if let := o.ToLetExprIfUnderscores(); let == nil {
				var appl AstAppl
				errs = appl.initFrom(me, o)
				expr = &appl
			} else {
				expr, errs = me.newAstExprFrom(let)
			}
		case *atmolang.AstExprCases:
			if o.Desugared == nil {
				var cases AstCases
				errs = cases.initFrom(me, o)
				expr = &cases
			} else {
				expr, errs = me.newAstExprFrom(o.Desugared)
			}
		default:
			panic(o)
		}
	}
	return
}
