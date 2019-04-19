package atmocorefn

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type Def struct {
	Orig     *atmolang.AstDef
	TopLevel *atmolang.AstFileTopLevelChunk
	Name     string
	Errs     atmo.Errors
}

func (me *Def) populate() {
	me.populateName()
	me.populateArgs()
}

func (me *Def) populateName() {
	if me.Name = me.Orig.Name.Val; me.TopLevel != nil && me.TopLevel.Ast.DefIsUnexported {
		me.Name = me.Name[1:]
	}
	tok := &me.Orig.Name.Tokens[0]
	if ustr.In(me.Name, langReservedOps...) {
		me.Errs.AddTok(atmo.ErrCatNaming, tok, "reserved token not permissible as def name: `"+me.Name+"`")
	}
	if me.Orig.Name.IsTag || ((!me.Orig.Name.IsOpish) && !ustr.BeginsLower(me.Name)) {
		me.Errs.AddTok(atmo.ErrCatNaming, tok, "def names must be lower-case (or operator tokens)")
	}
	if me.Orig.Name.Affix != "" {
		me.Errs.AddSyn(tok, "affixes are for def args, not for def names: drop `:"+me.Orig.Name.Affix+"`")
	}
}

func (me *Def) populateArgs() {
}

func (me *Def) populateMeta() {
}

func (me *Def) populateBody() {
}
