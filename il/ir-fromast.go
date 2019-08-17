package atmoil

import (
	"github.com/go-leap/str"
	. "github.com/metaleap/atmo"
	. "github.com/metaleap/atmo/ast"
)

func (me *IrDef) initFrom(ctx *ctxIrFromAst, origDef *AstDef) (errs Errors) {
	me.Orig = origDef
	origunary := origDef.ToUnary()

	errs.Add(me.initName(ctx, origunary)...)

	var maybearg *IrArg
	if len(origunary.Args) != 0 {
		maybearg = &IrArg{}
		errs.Add(maybearg.initFrom(ctx, &origunary.Args[0])...)
	}

	errs.Add(me.initMetas(ctx, origunary)...)
	errs.Add(me.initBody(ctx, origunary, maybearg)...)
	return
}

func (me *IrDef) initName(ctx *ctxIrFromAst, origDefUnary *AstDef) (errs Errors) {
	// even if our name is erroneous as detected further down below:
	// don't want this to stay empty, generally speaking
	me.Ident.Name = origDefUnary.Name.Val

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
		me.Ident.IrIdentBase = name.IrIdentBase
	}
	if origDefUnary.NameAffix != nil {
		ctx.addCoercion(me, errs.AddVia(ctx.newExprFrom(origDefUnary.NameAffix)).(IIrExpr))
	}
	return
}

func (me *IrDef) initBody(ctx *ctxIrFromAst, origDefUnary *AstDef, maybeArg *IrArg) (errs Errors) {
	if maybeArg != nil {
		if ctx.absIdx++; ctx.absIdx > ctx.absMax {
			ctx.absMax = ctx.absIdx
		}
	}

	// fast-track special-casing for a def-body of mere-underscore
	if ident, _ := origDefUnary.Body.(*AstIdent); ident != nil && ident.IsPlaceholder() {
		tag := BuildIr.LitTag(me.Ident.Name)
		tag.Orig, me.Body = ident, tag
	} else {
		me.Body, errs = ctx.newExprFrom(origDefUnary.Body)
	}

	if len(ctx.coerceCallables) != 0 {
		// each takes the arg val (or ret val) and returns either it or undef
		if maybeArg != nil {
			if coerce, ok := ctx.coerceCallables[maybeArg]; ok {
				me.Body = ctx.bodyWithCoercion(coerce, me.Body,
					func() IIrExpr { return BuildIr.IdentNameCopy(&maybeArg.IrIdentBase) })
			}
		}
		if coerce, ok := ctx.coerceCallables[me]; ok {
			me.Body = ctx.bodyWithCoercion(coerce, me.Body, nil)
		}
	}

	if maybeArg != nil {
		abs := IrAbs{Arg: *maybeArg, Body: me.Body}
		abs.Orig, abs.Ann.AbsIdx, abs.Arg.ownerAbs, me.Body = me.Orig, ctx.absIdx, &abs, &abs
		if ctx.absIdx == 0 {
			abs.Ann.AbsIdx = -ctx.absMax
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
	me.Orig = orig

	isexpr := true
	if ident, _ := orig.NameOrConstVal.(*AstIdent); ident != nil {
		if !(ident.IsTag || ident.IsVar()) {
			if ident.IsPlaceholder() {
				isexpr, me.Name, me.Orig =
					false, ustr.Int(len(ident.Val))+"_"+ctx.nextPrefix(), ident
				if len(ident.Val) > 1 {
					errs.AddNaming(ErrFromAst_DefArgNameMultipleUnderscores, &ident.Tokens[0], "invalid def-arg name: use 1 underscore for discards")
				}
			} else if cxn, ok := errs.AddVia(ctx.newExprFromIdent(ident)).(*IrIdentName); ok {
				isexpr, me.IrIdentBase = false, cxn.IrIdentBase
			}
		}
	}

	if isexpr {
		me.Name = ctx.nextPrefix() + orig.NameOrConstVal.Toks()[0].Lexeme
		appl := BuildIr.Appl1(BuildIr.IdentName(KnownIdentCoerce), errs.AddVia(ctx.newExprFrom(orig.NameOrConstVal)).(IIrExpr))
		appl.IrExprBase.Orig = orig.NameOrConstVal
		ctx.addCoercion(me, appl)
	}
	if orig.Affix != nil {
		ctx.addCoercion(me, errs.AddVia(ctx.newExprFrom(orig.Affix)).(IIrExpr))
	}
	return
}

func (me *irLitBase) initFrom(orig IAstExprAtomic) {
	me.Orig = orig
}

func (me *IrLitFloat) initFrom(orig *AstExprLitFloat) {
	me.irLitBase.initFrom(orig)
	me.Val = orig.Val
}

func (me *IrLitUint) initFrom(orig *AstExprLitUint) {
	me.irLitBase.initFrom(orig)
	me.Val = orig.Val
}
