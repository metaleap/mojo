package atmocorefn

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func newAstIdentFrom(orig *atmolang.AstIdent) (ident IAstIdent, errs atmo.Errors) {
	if orig.IsTag || ustr.BeginsUpper(orig.Val) {
		var tag AstIdentTag
		ident, errs = &tag, tag.initFrom(orig)

	} else if orig.IsOpish {
		if orig.Val == "()" {
			var empar AstIdentEmptyParens
			ident = &empar
		} else {
			var op AstIdentOp
			ident, errs = &op, op.initFrom(orig)
		}

	} else if orig.Val[0] != '_' {
		var name AstIdentName
		ident, errs = &name, name.initFrom(orig)

	} else if ustr.IsRepeat(orig.Val) {
		var unsco AstIdentUnderscores
		ident, errs = &unsco, unsco.initFrom(orig)

	} else if orig.Val[1] != '_' {
		var idvar AstIdentVar
		ident, errs = &idvar, idvar.initFrom(orig)

	} else {
		errs.AddFrom(atmo.ErrCatNaming, &orig.Tokens[0], "invalid identifier: begins with multiple underscores")
	}
	return
}

func (me *AstIdentBase) initFrom(from *atmolang.AstIdent) (errs atmo.Errors) {
	me.Val = from.Val
	return
}

func (me *AstDef) initFrom(orig *atmolang.AstDef) {
	me.Orig = orig
	me.Errs.Add(me.initName())
	me.Errs.Add(me.initArgs())
}

func (me *AstDef) initName() (errs atmo.Errors) {
	tok := &me.Orig.Name.Tokens[0]
	if me.Name, errs = newAstIdentFrom(&me.Orig.Name); len(errs) == 0 {
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
		errs.Add(me.Args[i].initFrom(&me.Orig.Args[i]))
	}
	return
}

func (me *AstDefArg) initFrom(orig *atmolang.AstDefArg) (errs atmo.Errors) {
	return
}

func (me *AstLitBase) initFrom(orig atmolang.IAstExprAtomic) {
	me.Orig = orig
}

func (me *AstLitFloat) initFrom(orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(orig)
	me.Val = orig.BaseTokens().Tokens[0].Float
}

func (me *AstLitUint) initFrom(orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(orig)
	me.Val = orig.BaseTokens().Tokens[0].Uint
}

func (me *AstLitRune) initFrom(orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(orig)
	me.Val = orig.BaseTokens().Tokens[0].Rune()
}

func (me *AstLitStr) initFrom(orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(orig)
	me.Val = orig.BaseTokens().Tokens[0].Str
}
