package atmocorefn

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func (me *AstDef) initFrom(orig *atmolang.AstDef) {
	me.state.genNamePrefs = []string{orig.Name.Val}
	me.state.wrapBody = func(expr IAstExpr) IAstExpr { return expr }
	me.Orig = orig
	me.Errs.Add(me.initName())
	me.Errs.Add(me.initArgs())
}

func (me *AstDef) initName() (errs atmo.Errors) {
	tok := &me.Orig.Name.Tokens[0]
	if me.Name, errs = me.newAstIdentFrom(&me.Orig.Name); len(errs) == 0 {
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

func (me *AstDef) initArgs() (errs atmo.Errors) {
	me.Args = make([]AstDefArg, len(me.Orig.Args))
	for i := range me.Orig.Args {
		errs.Add(me.Args[i].initFrom(me, &me.Orig.Args[i], i))
	}
	return
}

func (me *AstDef) newAstIdentFrom(orig *atmolang.AstIdent) (ident IAstIdent, errs atmo.Errors) {
	if orig.IsTag || ustr.BeginsUpper(orig.Val) {
		var tag AstIdentTag
		ident, errs = &tag, tag.initFrom(me, orig)

	} else if orig.IsOpish {
		if orig.Val == "()" {
			var empar AstIdentEmptyParens
			ident = &empar
		} else {
			var op AstIdentOp
			ident, errs = &op, op.initFrom(me, orig)
		}

	} else if orig.Val[0] != '_' {
		var name AstIdentName
		ident, errs = &name, name.initFrom(me, orig)

	} else if ustr.IsRepeat(orig.Val) {
		var unsco AstIdentUnderscores
		ident, errs = &unsco, unsco.initFrom(me, orig)

	} else if orig.Val[1] != '_' {
		var idvar AstIdentVar
		ident, errs = &idvar, idvar.initFrom(me, orig)

	} else {
		errs.AddFrom(atmo.ErrCatNaming, &orig.Tokens[0], "invalid identifier: begins with multiple underscores")
	}
	return
}

func (me *AstIdentBase) initFrom(ctx *AstDef, from *atmolang.AstIdent) (errs atmo.Errors) {
	me.Val = from.Val
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
	case *atmolang.AstExprLitFloat:
		var lit AstLitFloat
		lit.initFrom(ctx, v)
		constexpr = &lit
	case *atmolang.AstExprLitUint:
		var lit AstLitUint
		lit.initFrom(ctx, v)
		constexpr = &lit
	case *atmolang.AstExprLitRune:
		var lit AstLitRune
		lit.initFrom(ctx, v)
		constexpr = &lit
	case *atmolang.AstExprLitStr:
		var lit AstLitStr
		lit.initFrom(ctx, v)
		constexpr = &lit
	default:
		panic(v)
	}

	if constexpr != nil {
		me.AstIdentName.Val = "~arg~" + ustr.Int(argIdx)
	}
	return
}

func (me *AstLitBase) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.Orig = orig
}

func (me *AstLitFloat) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.BaseTokens().Tokens[0].Float
}

func (me *AstLitUint) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.BaseTokens().Tokens[0].Uint
}

func (me *AstLitRune) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.BaseTokens().Tokens[0].Rune()
}

func (me *AstLitStr) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.BaseTokens().Tokens[0].Str
}
