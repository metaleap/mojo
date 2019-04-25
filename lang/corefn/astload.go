package atmocorefn

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func (me *AstDef) initFrom(orig *atmolang.AstDef) {
	me.state.genNamePrefs = []string{orig.Name.Val}
	me.Locals = make(astDefs, 0, 16)
	me.Errs.Add(me.AstDefBase.initFrom(me, orig))
	println("LOCS:", len(me.Locals))
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
	tok := &me.Orig.Name.Tokens[0]
	if me.Name, errs = ctx.newAstIdentFrom(&me.Orig.Name); me.Name != nil {
		switch name := me.Name.(type) {
		case *AstIdentName:
			// all ok
		case *AstIdentOp:
			if name.Val == "" || ustr.In(name.Val, langReservedOps...) {
				errs.AddFrom(atmo.ErrCatNaming, tok, "reserved token not permissible as def name: `"+tok.Meta.Orig+"`")
			}
		case *AstIdentTag:
			errs.AddFrom(atmo.ErrCatNaming, tok, "invalid def name: `"+name.Val+"` is upper-case, this is reserved for tags")
		case *AstIdentVar:
			errs.AddFrom(atmo.ErrCatNaming, tok, "invalid def name: `"+tok.Meta.Orig+"` (begins with multiple underscores)")
		default:
			errs.AddFrom(atmo.ErrCatNaming, tok, "invalid def name: `"+tok.Meta.Orig+"`")
		}
	}
	return
}

func (me *AstDefBase) initBody(ctx *AstDef) (errs atmo.Errors) {
	me.Body, errs = ctx.newAstExprFrom(me.Orig.Body)
	return
}

func (me *AstDefBase) initArgs(ctx *AstDef) (errs atmo.Errors) {
	me.Args = make([]AstDefArg, len(me.Orig.Args))
	for i := range me.Orig.Args {
		errs.Add(me.Args[i].initFrom(ctx, &me.Orig.Args[i], i))
	}
	return
}

func (me *AstDefBase) initMetas(ctx *AstDef) (errs atmo.Errors) {
	if len(me.Orig.Meta) > 0 {
		errs.AddTodo(&me.Orig.Meta[0].Toks()[0], "def metas")
	}
	return
}

func (me *AstDef) newAstIdentFrom(orig *atmolang.AstIdent) (ret IAstIdent, errs atmo.Errors) {
	if t1, t2 := orig.IsTag, ustr.BeginsUpper(orig.Val); t1 && t2 {
		var ident AstIdentTag
		ret, ident.Val = &ident, orig.Val
	} else if t1 != t2 {
		panic("bug in `atmo/lang`: an `atmolang.AstIdent` had wrong `IsTag` value for its `Val` casing")

	} else if orig.IsOpish {
		if orig.Val == "()" {
			var ident AstIdentEmptyParens
			ret, ident.Val = &ident, orig.Val
		} else {
			var ident AstIdentOp
			ret, ident.Val = &ident, orig.Val
		}

	} else if orig.Val[0] != '_' {
		var ident AstIdentName
		ret, ident.Val = &ident, orig.Val

	} else if ustr.IsRepeat(orig.Val) {
		var ident AstIdentUnderscores
		ret, ident.Val = &ident, orig.Val

	} else if orig.Val[1] != '_' {
		var ident AstIdentVar
		ret, ident.Val = &ident, orig.Val

	} else {
		errs.AddFrom(atmo.ErrCatNaming, &orig.Tokens[0], "invalid identifier: begins with multiple underscores")
	}
	return
}

func (me *AstDef) newAstExprAtomicFrom(orig atmolang.IAstExprAtomic) (expr IAstExprAtomic, errs atmo.Errors) {
	x, e := me.newAstExprFrom(orig)
	errs.Add(e)
	if expr, _ = x.(IAstExprAtomic); expr == nil {
		errs.AddSyn(&orig.Toks()[0], "expected: atomic expression")
	}
	return
}

func (me *AstDef) newAstExprFrom(orig atmolang.IAstExpr) (expr IAstExpr, errs atmo.Errors) {
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
		var appl AstAppl
		errs = appl.initFrom(me, o)
		expr = &appl
	case *atmolang.AstExprCase:
		errs.AddTodo(&orig.Toks()[0], "case exprs")
	default:
		panic(o)
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
		me.AstIdentName.Val = "~arg~" + ustr.Int(argIdx)
		errs.AddTodo(&orig.NameOrConstVal.Toks()[0], "def arg const-exprs")
	}

	if orig.Affix != nil {
		errs.AddTodo(&orig.Affix.Toks()[0], "def arg affixes")
	}
	return
}

func (me *AstAppl) initFrom(ctx *AstDef, orig *atmolang.AstExprAppl) (errs atmo.Errors) {
	var e atmo.Errors
	if oc, _ := orig.Callee.(atmolang.IAstExprAtomic); oc != nil {
		me.Callee, e = ctx.newAstExprAtomicFrom(oc)
		errs.Add(e)
	} else {
		// var def AstDefBase
		// def.Name = &AstIdentName{Val: ""}
		// errs.Add(def.initFrom(me, &o.Defs[i]))
		// me.Locals = append(me.Locals, def)

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
