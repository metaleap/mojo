package atmoil

import (
	"github.com/go-leap/str"
	. "github.com/metaleap/atmo"
	. "github.com/metaleap/atmo/ast"
)

func (me *IrDef) initFrom(ctx *ctxIrFromAst, orig *AstDef) (errs Errors) {
	me.Orig = orig
	origUnary := orig.ToUnary()
	// the ordering of these 4 matters: the latter ones depend on work done by the former ones
	errs.Add(me.initName(ctx, origUnary)...)
	errs.Add(me.initArg(ctx, origUnary)...)
	errs.Add(me.initMetas(ctx, origUnary)...)
	errs.Add(me.initBody(ctx, origUnary)...)
	return
}

func (me *IrDef) initName(ctx *ctxIrFromAst, origDefUnary *AstDef) (errs Errors) {
	// even if our name is erroneous as detected further down below:
	// don't want this to stay empty, generally speaking
	me.Name.Val = origDefUnary.Name.Val

	tok := origDefUnary.Name.Tokens.First1() // could have none so dont just Tokens[0]
	if tok == nil {
		if tok = origDefUnary.Tokens.First1(); tok == nil {
			tok = origDefUnary.Body.Toks().First1()
		}
	}
	var ident IIrExpr
	ident, errs = ctx.newExprFromIdent(&origDefUnary.Name)
	if name, _ := ident.(*IrIdentName); name == nil {
		errs.AddNaming(ErrFromAst_DefNameInvalidIdent, tok, "invalid def name: `"+tok.String()+"`") // some non-name ident: Tag or Undef or placeholder etc..
	} else {
		me.Name.IrIdentBase = name.IrIdentBase
	}
	if origDefUnary.NameAffix != nil {
		ctx.addCoercion(me, errs.AddVia(ctx.newExprFrom(origDefUnary.NameAffix)).(IIrExpr))
	}
	return
}

func (me *IrDef) initArg(ctx *ctxIrFromAst, origDefUnary *AstDef) (errs Errors) {
	if len(origDefUnary.Args) != 0 { // can only be 0 or 1 as toUnary-zation happened before here
		var arg IrArg
		errs.Add(arg.initFrom(ctx, &origDefUnary.Args[0])...)
		ctx.defArgs[me] = &arg
	}
	return
}

func (me *IrDef) initBody(ctx *ctxIrFromAst, origDefUnary *AstDef) (errs Errors) {
	defarg := ctx.defArgs[me]
	if defarg != nil {
		if ctx.absIdx++; ctx.absIdx > ctx.absMax {
			ctx.absMax = ctx.absIdx
		}
	}

	// fast-track special-casing for a def-body of mere-underscore
	if ident, _ := origDefUnary.Body.(*AstIdent); ident != nil && ident.IsPlaceholder() {
		tag := BuildIr.IdentTag(me.Name.Val)
		tag.Orig, me.Body = ident, tag
	} else {
		me.Body, errs = ctx.newExprFrom(origDefUnary.Body)
	}

	if len(ctx.coerceCallables) != 0 {
		// each takes the arg val (or ret val) and returns either it or undef
		if defarg != nil {
			if coerce, ok := ctx.coerceCallables[defarg]; ok {
				me.Body = ctx.bodyWithCoercion(coerce, me.Body,
					func() IIrExpr { return BuildIr.IdentNameCopy(&defarg.IrIdentBase) })
			}
		}
		if coerce, ok := ctx.coerceCallables[me]; ok {
			me.Body = ctx.bodyWithCoercion(coerce, me.Body, nil)
		}
	}

	if defarg != nil {
		abs := IrAbs{Arg: *defarg, Body: me.Body}
		abs.Orig, abs.Arg.Anns.AbsIdx, me.Body = me.Orig, ctx.absIdx, &abs
		if ctx.absIdx == 0 {
			abs.Arg.Anns.AbsIdx = -ctx.absMax
			ctx.absMax = 0
		}
		ctx.absIdx--
	}
	return
}

func (me *IrDef) initMetas(ctx *ctxIrFromAst, origDefUnary *AstDef) (errs Errors) {
	if len(origDefUnary.Meta) != 0 {
		errs.AddTodo(0, origDefUnary.Meta[0].Toks(), "def metas")
		for i := range origDefUnary.Meta {
			_ = errs.AddVia(ctx.newExprFrom(origDefUnary.Meta[i]))
		}
	}
	return
}

func (me *IrArg) initFrom(ctx *ctxIrFromAst, orig *AstDefArg) (errs Errors) {
	me.Orig, me.Anns.AbsIdx = orig, -1

	isexpr := true
	switch v := orig.NameOrConstVal.(type) {
	case *AstIdent:
		if !(v.IsTag || v.IsVar()) {
			if v.IsPlaceholder() {
				isexpr, me.IrIdentBase.Val, me.IrIdentBase.Orig =
					false, ustr.Int(len(v.Val))+"_"+ctx.nextPrefix(), v
				if len(v.Val) > 1 {
					errs.AddNaming(ErrFromAst_DefArgNameMultipleUnderscores, &v.Tokens[0], "invalid def-arg name: use 1 underscore for discards")
				}
			} else if cxn, ok1 := errs.AddVia(ctx.newExprFromIdent(v)).(*IrIdentName); ok1 {
				isexpr, me.IrIdentBase = false, cxn.IrIdentBase
			}
		}
	}

	if isexpr {
		me.IrIdentBase.Val = ctx.nextPrefix() + orig.NameOrConstVal.Toks()[0].Lexeme
		appl := BuildIr.Appl1(BuildIr.IdentName(KnownIdentCoerce), errs.AddVia(ctx.newExprFrom(orig.NameOrConstVal)).(IIrExpr))
		appl.IrExprBase.Orig = orig.NameOrConstVal
		ctx.addCoercion(me, appl)
	}
	if orig.Affix != nil {
		ctx.addCoercion(me, errs.AddVia(ctx.newExprFrom(orig.Affix)).(IIrExpr))
	}
	return
}

func (me *irLitBase) initFrom(ctx *ctxIrFromAst, orig IAstExprAtomic) {
	me.Orig = orig
}

func (me *IrLitFloat) initFrom(ctx *ctxIrFromAst, orig *AstExprLitFloat) {
	me.irLitBase.initFrom(ctx, orig)
	me.Val = orig.Val
}

func (me *IrLitUint) initFrom(ctx *ctxIrFromAst, orig *AstExprLitUint) {
	me.irLitBase.initFrom(ctx, orig)
	me.Val = orig.Val
}
