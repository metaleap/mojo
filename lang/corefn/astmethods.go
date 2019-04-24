package atmocorefn

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func (me *AstIdentBase) initFrom(from *atmolang.AstIdent) (errs atmo.Errors) {
	me.Val = from.Val
	return
}

func (me *AstIdentBase) Origin() atmolang.IAstNode { return me.Orig }

func (me *AstDef) initFrom(orig *atmolang.AstDef) {
	me.Orig = orig
	me.Errs.Add(me.initName())
	me.Errs.Add(me.initArgs())
}

func (me *AstDef) Origin() atmolang.IAstNode { return me.Orig }

func (me *AstDef) initName() (errs atmo.Errors) {
	tok := &me.Orig.Name.Tokens[0]
	if me.Name, errs = newIdentFrom(&me.Orig.Name); len(errs) == 0 {
		switch name := me.Name.(type) {
		case *AstIdentName:
			// all ok
		case *AstIdentOp:
			if name.Val == "" || ustr.In(name.Val, langReservedOps...) {
				errs.AddFrom(atmo.ErrCatNaming, tok, "reserved token not permissible as def name: `"+tok.Meta.Orig+"`")
			}
		case *AstIdentTag:
			errs.AddFrom(atmo.ErrCatNaming, tok, "not a valid def name: `"+name.Val+"` is upper-case (this is reserved for tags)")
		case *AstIdentVar:
			errs.AddFrom(atmo.ErrCatNaming, tok, "not a valid def name: `"+tok.Meta.Orig+"` (begins with multiple underscores)")
		default:
			errs.AddFrom(atmo.ErrCatNaming, tok, "not a valid def name: `"+tok.Meta.Orig+"`")
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
	// me.Orig = orig
	// tok := &orig.Tokens[0]
	// switch /*a :=*/ orig.NameOrConstVal.(type) {

	// default:
	// 	errs.AddSyn(tok, "not a valid def arg: `"+tok.Meta.Orig+"`")
	// }
	return
}

func (me *AstLitBase) Origin() atmolang.IAstNode {
	return me.Orig
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

func (me *AstIdentUnderscores) Num() int { return len(me.Val) }
