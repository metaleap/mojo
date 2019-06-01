package atmolang_irfun

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func (me *AstDef) initFrom(ctx *ctxAstInit, orig *atmolang.AstDef) (errs atmo.Errors) {
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

func (me *AstDef) initName(ctx *ctxAstInit) (errs atmo.Errors) {
	tok := me.OrigDef.Name.Tokens.First(nil) // could have none so dont just Tokens[0]
	if tok == nil {
		tok = me.OrigToks().First(nil)
	}
	var ident IAstExpr
	ident, errs = ctx.newAstExprFromIdent(&me.OrigDef.Name)
	if name, _ := ident.(*AstIdentName); name == nil && tok != nil {
		errs.AddNaming(tok, "invalid def name: `"+tok.Meta.Orig+"`") // Tag or Undef or placeholder etc..
	} else if me.Name = name.AstIdentBase; name.Val == "" && tok != nil {
		errs.AddNaming(tok, "reserved token not permissible as def name: `"+tok.Meta.Orig+"`")
	}
	if me.OrigDef.NameAffix != nil {
		ctx.addCoercion(me, errs.AddVia(ctx.newAstExprFrom(me.OrigDef.NameAffix)).(IAstExpr))
	}
	return
}

func (me *AstDef) initBody(ctx *ctxAstInit) (errs atmo.Errors) {
	// fast-track special case: "func signature expression" aka body-less def acts as notation for a func type
	if toks := me.OrigDef.Body.Toks(); len(toks) == 1 && toks[0].Meta.Orig == "_" {
		// no-op: me.Body remains `nil`, this is preserved also in any `if`s from the below coerce-propagations, if any
	} else {
		me.Body, errs = ctx.newAstExprFrom(me.OrigDef.Body)
	}
	if len(ctx.coerceCallables) > 0 {
		opeq, appl := B.IdentName(atmo.Syn_Eq), func(applexpr IAstExpr, orig atmolang.IAstExpr, atomic bool) IAstExpr {
			if applexpr.astExprBase().Orig = orig; atomic {
				applexpr = ctx.ensureAstAtomFor(applexpr)
				applexpr.astExprBase().Orig = orig
			}
			return applexpr
		}
		if me.Arg != nil {
			if coerce := ctx.coerceCallables[me.Arg]; coerce != nil {
				coerceorig := coerce.astExprBase().Orig
				newbody := appl(B.Appl1(ctx.ensureAstAtomFor(coerce), &AstIdentName{AstIdentBase: me.Arg.AstIdentBase}), coerceorig, true)
				newbody = appl(B.ApplN(ctx, opeq, &AstIdentName{AstIdentBase: me.Arg.AstIdentBase}, newbody), coerceorig, true)
				me.Body = appl(B.ApplN(ctx, B.IdentName(atmo.Syn_If), newbody, me.Body, B.LitUndef()), coerceorig, false)
			}
		}
		if coerce := ctx.coerceCallables[me]; coerce != nil {
			oldbody, coerceorig := ctx.ensureAstAtomFor(me.Body), coerce.astExprBase().Orig
			newbody := appl(B.Appl1(ctx.ensureAstAtomFor(coerce), oldbody), coerceorig, true)
			newbody = appl(B.ApplN(ctx, opeq, oldbody, newbody), coerceorig, true)
			me.Body = appl(B.ApplN(ctx, B.IdentName(atmo.Syn_If), newbody, oldbody, B.LitUndef()), coerceorig, false)
		}
	}
	return
}

func (me *AstDef) initArg(ctx *ctxAstInit) (errs atmo.Errors) {
	if len(me.OrigDef.Args) == 1 { // can only be 0 or 1 as toUnary-zation happened before here
		var arg AstDefArg
		errs.Add(arg.initFrom(ctx, &me.OrigDef.Args[0]))
		me.Arg = &arg
	}
	return
}

func (me *AstDef) initMetas(ctx *ctxAstInit) (errs atmo.Errors) {
	if len(me.OrigDef.Meta) > 0 {
		errs.AddTodo(&me.OrigDef.Meta[0].Toks()[0], "def metas")
		for i := range me.OrigDef.Meta {
			_ = errs.AddVia(ctx.newAstExprFrom(me.OrigDef.Meta[i]))
		}
	}
	return
}

func (me *AstDefArg) initFrom(ctx *ctxAstInit, orig *atmolang.AstDefArg) (errs atmo.Errors) {
	me.Orig = orig

	var isconstexpr bool
	switch v := orig.NameOrConstVal.(type) {
	case *atmolang.AstIdent:
		if isconstexpr = true; !(v.IsTag || v.Val == atmo.Syn_Undef || ( /*AstIdentVar*/ v.Val[0] == '_' && len(v.Val) > 1 && v.Val[1] != '_')) {
			if ustr.IsRepeat(v.Val, '_') {
				if isconstexpr, me.AstIdentBase.Val, me.AstIdentBase.Orig = false, ustr.Int(len(v.Val))+"_"+ctx.nextPrefix(), v; len(v.Val) > 1 {
					errs.AddNaming(&v.Tokens[0], "invalid def arg name")
				}
			} else if cxn, ok1 := errs.AddVia(ctx.newAstExprFromIdent(v)).(*AstIdentName); ok1 {
				isconstexpr, me.AstIdentBase = false, cxn.AstIdentBase
			}
		}
	case *atmolang.AstExprLitFloat, *atmolang.AstExprLitUint, *atmolang.AstExprLitRune, *atmolang.AstExprLitStr:
		isconstexpr = true
	}

	if isconstexpr && orig.Affix != nil {
		errs.AddSyn(&orig.Affix.Toks()[0], "def-arg affix illegal where the arg is itself a constant value")
	}
	if orig.Affix != nil {
		ctx.addCoercion(me, errs.AddVia(ctx.newAstExprFrom(orig.Affix)).(IAstExpr))
	}
	if isconstexpr {
		me.AstIdentBase.Val = ctx.nextPrefix() + orig.NameOrConstVal.Toks()[0].Meta.Orig
		appl := B.Appl1(B.IdentName("§"), ctx.ensureAstAtomFor(errs.AddVia(ctx.newAstExprFrom(orig.NameOrConstVal)).(IAstExpr)))
		appl.AstExprBase.Orig = orig.NameOrConstVal
		ctx.addCoercion(me, appl)
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
