package atmocorefn

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func (me *AstDef) initFrom(orig *atmolang.AstDef) {
	me.Orig = orig
	me.initName()
}

func (me *AstDef) initName() {
	name, tok := me.Orig.Name.Val, &me.Orig.Name.Tokens[0]
	if me.TopLevel.Ast.DefIsUnexported {
		name = name[1:]
	}

	if name == "" || ustr.In(name, langReservedOps...) {
		me.Errs.AddTok(atmo.ErrCatNaming, tok, "reserved token not permissible as def name: `"+me.Orig.Name.Val+"`")
	} else if me.Orig.Name.IsTag || ((!me.Orig.Name.IsOpish) && !ustr.BeginsLower(name)) {
		me.Errs.AddTok(atmo.ErrCatNaming, tok, "def names must be lower-case (or operator tokens)")
	}

}

func (me *AstIdentBase) initFrom(from *atmolang.AstIdent) (errs atmo.Errors) {
	me.Name = from.Val
	return
}

func (me *AstIdentBase) FromOrig() atmolang.IAstNode { return me.Orig }
