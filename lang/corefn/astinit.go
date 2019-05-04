package atmocorefn

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func (me *AstDefBase) initFrom(ctx *ctxAstInit, orig *atmolang.AstDef) (errs atmo.Errors) {
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

func (me *AstDefBase) initName(ctx *ctxAstInit) (errs atmo.Errors) {
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

func (me *AstDefBase) initBody(ctx *ctxAstInit) (errs atmo.Errors) {
	me.Body, errs = ctx.newAstExprFrom(me.Orig.Body)

	opeq := B.IdentName("==")
	for i := range me.Args {
		if me.Args[i].coerceValue != nil {
			me.Body = B.Case(B.Appls(ctx, opeq, &me.Args[i].AstIdentName, me.Args[i].coerceValue), me.Body)
		}
		if me.Args[i].coerceFunc != nil {
			appl := B.Appl(ctx.ensureAstAtomFor(me.Args[i].coerceFunc), &me.Args[i].AstIdentName)
			me.Body = B.Case(B.Appls(ctx, opeq, &me.Args[i].AstIdentName, ctx.ensureAstAtomFor(appl)), me.Body)
		}
	}
	if me.nameCoerceFunc != nil {
		appl := B.Appl(ctx.ensureAstAtomFor(me.nameCoerceFunc), &me.Name)
		me.Body = B.Case(B.Appls(ctx, opeq, &me.Name, ctx.ensureAstAtomFor(appl)), me.Body)
	}

	return
}

func (me *AstDefBase) initArgs(ctx *ctxAstInit) (errs atmo.Errors) {
	if len(me.Orig.Args) > 0 {
		args := make([]AstDefArg, len(me.Orig.Args))
		for i := range me.Orig.Args {
			errs.Add(args[i].initFrom(ctx, &me.Orig.Args[i], i))
		}
		me.Args = args
	}
	return
}

func (me *AstDefBase) initMetas(ctx *ctxAstInit) (errs atmo.Errors) {
	if len(me.Orig.Meta) > 0 {
		errs.AddTodo(&me.Orig.Meta[0].Toks()[0], "def metas")
		for i := range me.Orig.Meta {
			_ = errs.AddVia(ctx.newAstExprFrom(me.Orig.Meta[i]))
		}
	}
	return
}

func (me *AstDefArg) initFrom(ctx *ctxAstInit, orig *atmolang.AstDefArg, argIdx int) (errs atmo.Errors) {
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
		constexpr = errs.AddVia(ctx.newAstExprFrom(v)).(IAstExprAtomic)
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

func (me *AstCases) initFrom(ctx *ctxAstInit, orig *atmolang.AstExprCases) (errs atmo.Errors) {
	me.Orig = orig

	var scrut IAstExpr
	if orig.Scrutinee != nil {
		scrut = errs.AddVia(ctx.newAstExprFrom(orig.Scrutinee)).(IAstExpr)
	} else {
		scrut = B.IdentTagTrue()
	}
	scrut = B.Appl(B.IdentName("=="), ctx.ensureAstAtomFor(scrut))
	scrutatomic, opeq := ctx.ensureAstAtomFor(scrut), B.IdentName("||")

	me.Ifs, me.Thens = make([]IAstExpr, len(orig.Alts)), make([]IAstExpr, len(orig.Alts))
	for i := range orig.Alts {
		alt := &orig.Alts[i]
		if alt.Body == nil {
			panic("bug in atmo/lang: received an `AstCase` with a `nil` `Body`")
		} else {
			me.Thens[i] = errs.AddVia(ctx.newAstExprFrom(alt.Body)).(IAstExpr)
		}
		for c, cond := range alt.Conds {
			if c == 0 {
				me.Ifs[i] = B.Appl(scrutatomic, ctx.ensureAstAtomFor(errs.AddVia(ctx.newAstExprFrom(cond)).(IAstExpr)))
			} else {
				me.Ifs[i] = B.Appls(ctx, opeq, ctx.ensureAstAtomFor(me.Ifs[i]), ctx.ensureAstAtomFor(
					B.Appl(scrutatomic, ctx.ensureAstAtomFor(errs.AddVia(ctx.newAstExprFrom(cond)).(IAstExpr)))))
			}
		}
	}
	return
}

func (me *AstLitBase) initFrom(ctx *ctxAstInit, orig atmolang.IAstExprAtomic) {
	me.Orig = orig
}

func (me *AstLitFloat) initFrom(ctx *ctxAstInit, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Float
}

func (me *AstLitUint) initFrom(ctx *ctxAstInit, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Uint
}

func (me *AstLitRune) initFrom(ctx *ctxAstInit, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Rune()
}

func (me *AstLitStr) initFrom(ctx *ctxAstInit, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Str
}
