package atmoil

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func (me *IrDef) initFrom(ctx *ctxIrInit, orig *atmolang.AstDef) (errs atmo.Errors) {
	me.OrigDef = orig.ToUnary()
	errs.Add(me.initName(ctx)...)
	errs.Add(me.initArg(ctx)...)
	errs.Add(me.initMetas(ctx)...)
	errs.Add(me.initBody(ctx)...)

	// all inits worked off the orig-unary-fied, but for all
	// post-init usage we restore the real source orig-def:
	me.OrigDef = orig
	return
}

func (me *IrDef) initName(ctx *ctxIrInit) (errs atmo.Errors) {
	// even if our name is erroneous as detected further down below:
	// don't want this to stay empty, generally speaking
	me.Name.Val = me.OrigDef.Name.Val

	tok := me.OrigDef.Name.Tokens.First1() // could have none so dont just Tokens[0]
	if tok == nil {
		if tok = me.OrigDef.Tokens.First1(); tok == nil {
			tok = me.OrigDef.Body.Toks().First1()
		}
	}
	var ident IExpr
	ident, errs = ctx.newExprFromIdent(&me.OrigDef.Name)
	if name, _ := ident.(*IrIdentName); name == nil && tok != nil {
		errs.AddNaming(ErrInit_DefNameInvalidIdent, tok, "invalid def name: `"+tok.Lexeme+"`") // some non-name ident: Tag or Undef or placeholder etc..
	} else if me.Name.IrIdentBase = name.IrIdentBase; name.Val == "" && tok != nil {
		errs.AddNaming(ErrInit_DefNameReserved, tok, "reserved token not permissible as def name: `"+tok.Lexeme+"`")
	}
	if me.OrigDef.NameAffix != nil {
		ctx.addCoercion(me, errs.AddVia(ctx.newExprFrom(me.OrigDef.NameAffix)).(IExpr))
	}
	return
}

func (me *IrDef) initBody(ctx *ctxIrInit) (errs atmo.Errors) {
	// fast-track special-casing for a def-body of mere-underscore
	if ident, _ := me.OrigDef.Body.(*atmolang.AstIdent); ident != nil && ident.IsPlaceholder() {
		tag := Build.IdentTag(me.Name.Val)
		tag.Orig, me.Body = ident, tag
	} else {
		me.Body, errs = ctx.newExprFrom(me.OrigDef.Body)
	}
	if len(ctx.coerceCallables) > 0 {
		// each takes the arg val (or ret val) and returns either it or undef

		if me.Arg != nil {
			if coerce, ok := ctx.coerceCallables[me.Arg]; ok {
				me.Body = ctx.bodyWithCoercion(coerce, ctx.ensureAtomic(me.Body),
					func() IExpr { return Build.IdentNameCopy(&me.Arg.IrIdentBase) })
			}
		}
		if coerce, ok := ctx.coerceCallables[me]; ok {
			me.Body = ctx.bodyWithCoercion(coerce, ctx.ensureAtomic(me.Body), nil)
		}
	}
	return
}

func (me *IrDef) initArg(ctx *ctxIrInit) (errs atmo.Errors) {
	if len(me.OrigDef.Args) == 1 { // can only be 0 or 1 as toUnary-zation happened before here
		var arg IrDefArg
		errs.Add(arg.initFrom(ctx, &me.OrigDef.Args[0])...)
		me.Arg = &arg
	}
	return
}

func (me *IrDef) initMetas(ctx *ctxIrInit) (errs atmo.Errors) {
	if len(me.OrigDef.Meta) > 0 {
		errs.AddTodo(0, me.OrigDef.Meta[0].Toks(), "def metas")
		for i := range me.OrigDef.Meta {
			_ = errs.AddVia(ctx.newExprFrom(me.OrigDef.Meta[i]))
		}
	}
	return
}

func (me *IrDefArg) initFrom(ctx *ctxIrInit, orig *atmolang.AstDefArg) (errs atmo.Errors) {
	me.Orig = orig

	isexpr := true
	switch v := orig.NameOrConstVal.(type) {
	case *atmolang.AstIdent:
		if !(v.IsTag || v.IsVar()) {
			if v.IsPlaceholder() {
				isexpr, me.IrIdentBase.Val, me.IrIdentBase.Orig =
					false, ustr.Int(len(v.Val))+"_"+ctx.nextPrefix(), v
				if len(v.Val) > 1 {
					errs.AddNaming(ErrInit_DefArgNameUnderscores, &v.Tokens[0], "invalid def-arg name: use 1 underscore for discards")
				}
			} else if cxn, ok1 := errs.AddVia(ctx.newExprFromIdent(v)).(*IrIdentName); ok1 {
				isexpr, me.IrIdentBase = false, cxn.IrIdentBase
			}
		}
	}

	if isexpr {
		me.IrIdentBase.Val = ctx.nextPrefix() + orig.NameOrConstVal.Toks()[0].Lexeme
		appl := Build.Appl1(Build.IdentName(atmo.KnownIdentCoerce), ctx.ensureAtomic(errs.AddVia(ctx.newExprFrom(orig.NameOrConstVal)).(IExpr)))
		appl.IrExprBase.Orig = orig.NameOrConstVal
		ctx.addCoercion(me, appl)
	}
	if orig.Affix != nil {
		ctx.addCoercion(me, errs.AddVia(ctx.newExprFrom(orig.Affix)).(IExpr))
	}
	return
}

func (me *irLitBase) initFrom(ctx *ctxIrInit, orig atmolang.IAstExprAtomic) {
	me.Orig = orig
}

func (me *IrLitFloat) initFrom(ctx *ctxIrInit, orig *atmolang.AstExprLitFloat) {
	me.irLitBase.initFrom(ctx, orig)
	me.Val = orig.Val
}

func (me *IrLitUint) initFrom(ctx *ctxIrInit, orig *atmolang.AstExprLitUint) {
	me.irLitBase.initFrom(ctx, orig)
	me.Val = orig.Val
}

func (me *IrLitStr) initFrom(ctx *ctxIrInit, orig *atmolang.AstExprLitStr) {
	me.irLitBase.initFrom(ctx, orig)
	me.Val = orig.Val
}
