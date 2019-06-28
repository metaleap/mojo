package atmoil

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func (me *IrDef) initFrom(ctx *ctxIrInit, orig *atmolang.AstDef) (errs atmo.Errors) {
	me.OrigDef = orig.ToUnary()
	errs.Add(me.initName(ctx))
	errs.Add(me.initArg(ctx))
	errs.Add(me.initMetas(ctx))
	errs.Add(me.initBody(ctx))

	// all inits worked off the orig-unary-fied, but for all
	// post-init usage we restore the real source orig-def:
	me.OrigDef = orig
	return
}

func (me *IrDef) initName(ctx *ctxIrInit) (errs atmo.Errors) {
	tok := me.OrigDef.Name.Tokens.First1() // could have none so dont just Tokens[0]
	if tok == nil {
		if tok = me.OrigDef.Tokens.First1(); tok == nil {
			tok = me.OrigDef.Body.Toks().First1()
		}
	}
	var ident IExpr
	ident, errs = ctx.newExprFromIdent(&me.OrigDef.Name)
	if name, _ := ident.(*IrIdentName); name == nil && tok != nil {
		errs.AddNaming(tok, "invalid def name: `"+tok.Meta.Orig+"`") // Tag or Undef or placeholder etc..
	} else if me.Name.IrIdentBase = name.IrIdentBase; name.Val == "" && tok != nil {
		errs.AddNaming(tok, "reserved token not permissible as def name: `"+tok.Meta.Orig+"`")
	}
	if me.OrigDef.NameAffix != nil {
		ctx.addCoercion(me, errs.AddVia(ctx.newExprFrom(me.OrigDef.NameAffix)).(IExpr))
	}
	return
}

func (me *IrDef) initBody(ctx *ctxIrInit) (errs atmo.Errors) {
	// fast-track special case: "func signature expression" aka body-less def
	// (also for ffi / builtin-primops) acts as notation for a func type
	if ident, _ := me.OrigDef.Body.(*atmolang.AstIdent); ident != nil && ident.IsPlaceholder() {
		var body IrSpecial
		if me.Body, body.Orig, body.OneOf.DefArgfulButBodyless = &body, ident, me.Arg != nil; me.Arg == nil {
			errs.AddSyn(ident.Tokens, "illegal placeholder placement")
		}
	} else {
		me.Body, errs = ctx.newExprFrom(me.OrigDef.Body)
	}
	if len(ctx.coerceCallables) > 0 {
		opeq, appl := Build.IdentName(atmo.KnownIdentEq), func(applexpr IExpr, orig atmolang.IAstExpr, atomic bool) IExpr {
			if applexpr.exprBase().Orig = orig; atomic {
				applexpr = ctx.ensureAtomic(applexpr)
				applexpr.exprBase().Orig = orig
			}
			return applexpr
		}
		if me.Arg != nil {
			if coerce := ctx.coerceCallables[me.Arg]; coerce != nil {
				coerceorig := coerce.exprBase().Orig
				newbody := appl(Build.Appl1(ctx.ensureAtomic(coerce), &IrIdentName{IrIdentBase: me.Arg.IrIdentBase}), coerceorig, true)
				newbody = appl(Build.ApplN(ctx, opeq, &IrIdentName{IrIdentBase: me.Arg.IrIdentBase}, newbody), coerceorig, true)
				me.Body = appl(Build.ApplN(ctx, Build.IdentName(atmo.KnownIdentIf), newbody, ctx.ensureAtomic(me.Body), &IrSpecial{}), coerceorig, false)
			}
		}
		if coerce := ctx.coerceCallables[me]; coerce != nil {
			oldbody, coerceorig := ctx.ensureAtomic(me.Body), coerce.exprBase().Orig
			newbody := appl(Build.Appl1(ctx.ensureAtomic(coerce), oldbody), coerceorig, true)
			newbody = appl(Build.ApplN(ctx, opeq, oldbody, newbody), coerceorig, true)
			me.Body = appl(Build.ApplN(ctx, Build.IdentName(atmo.KnownIdentIf), newbody, oldbody, &IrSpecial{}), coerceorig, false)
		}
	}
	return
}

func (me *IrDef) initArg(ctx *ctxIrInit) (errs atmo.Errors) {
	if len(me.OrigDef.Args) == 1 { // can only be 0 or 1 as toUnary-zation happened before here
		var arg IrDefArg
		errs.Add(arg.initFrom(ctx, &me.OrigDef.Args[0]))
		me.Arg = &arg
	}
	return
}

func (me *IrDef) initMetas(ctx *ctxIrInit) (errs atmo.Errors) {
	if len(me.OrigDef.Meta) > 0 {
		errs.AddTodo(me.OrigDef.Meta[0].Toks(), "def metas")
		for i := range me.OrigDef.Meta {
			_ = errs.AddVia(ctx.newExprFrom(me.OrigDef.Meta[i]))
		}
	}
	return
}

func (me *IrDefArg) initFrom(ctx *ctxIrInit, orig *atmolang.AstDefArg) (errs atmo.Errors) {
	me.Orig = orig

	var isconstexpr bool
	switch v := orig.NameOrConstVal.(type) {
	case *atmolang.AstIdent:
		if isconstexpr = true; !(v.IsTag || v.IsVar()) {
			if v.IsPlaceholder() {
				if isconstexpr, me.IrIdentBase.Val, me.IrIdentBase.Orig = false, ustr.Int(len(v.Val))+"_"+ctx.nextPrefix(), v; len(v.Val) > 1 {
					errs.AddNaming(&v.Tokens[0], "invalid def arg name")
				}
			} else if cxn, ok1 := errs.AddVia(ctx.newExprFromIdent(v)).(*IrIdentName); ok1 {
				isconstexpr, me.IrIdentBase = false, cxn.IrIdentBase
			}
		}
	case *atmolang.AstExprLitFloat, *atmolang.AstExprLitUint, *atmolang.AstExprLitStr:
		isconstexpr = true
	}

	if isconstexpr && orig.Affix != nil {
		errs.AddSyn(orig.Affix.Toks(), "def-arg affix illegal where the arg is itself a constant value")
	}
	if orig.Affix != nil {
		ctx.addCoercion(me, errs.AddVia(ctx.newExprFrom(orig.Affix)).(IExpr))
	}
	if isconstexpr {
		me.IrIdentBase.Val = ctx.nextPrefix() + orig.NameOrConstVal.Toks()[0].Meta.Orig
		appl := Build.Appl1(Build.IdentName(atmo.KnownIdentCoerce), ctx.ensureAtomic(errs.AddVia(ctx.newExprFrom(orig.NameOrConstVal)).(IExpr)))
		appl.IrExprBase.Orig = orig.NameOrConstVal
		ctx.addCoercion(me, appl)
	}
	return
}

func (me *irLitBase) initFrom(ctx *ctxIrInit, orig atmolang.IAstExprAtomic) {
	me.Orig = orig
}

func (me *IrLitFloat) initFrom(ctx *ctxIrInit, orig atmolang.IAstExprAtomic) {
	me.irLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Float
}

func (me *IrLitUint) initFrom(ctx *ctxIrInit, orig atmolang.IAstExprAtomic) {
	me.irLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Uint
}

func (me *IrLitStr) initFrom(ctx *ctxIrInit, orig atmolang.IAstExprAtomic) {
	me.irLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Str
}
