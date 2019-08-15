package atmoil

import (
	"fmt"
	"github.com/go-leap/std"
)

func (me *PValFactBase) From() (*IrDef, IIrNode) { return me.from[0].(*IrDef), me.from[len(me.from)-1] }
func (me *PValFactBase) IsEnv() *PEnv            { return nil }
func (me *PValFactBase) Self() *PValFactBase     { return me }
func (me *PValFactBase) String() string {
	def, node := me.From()
	return def.AstOrigToks(node).Pos().String()
}

func (me *PValUsed) String() string { return "used(" + me.PValFactBase.String() + ")" }

func (me *PValPrimConst) String() string { return "eqPrim(" + fmt.Sprintf("%v", me.ConstVal) + ")" }

func (me *PValEqType) String() string { return "eqType(" + me.Of.PValFactBase.String() + ")" }

func (me *PValEqVal) String() string { return "eqTo(" + me.To.PValFactBase.String() + ")" }

func (me *PValNever) String() string { return "never(" + me.Never.Self().String() + ")" }

func (me *PValFn) String() string { return "fn(" + me.Arg.String() + "->" + me.Ret.String() + ")" }

func (me *PVal) AddPrimConst(fromNode []IIrNode, constVal interface{}) *PVal {
	fact := PValPrimConst{ConstVal: constVal}
	fact.from, me.Facts = fromNode, append(me.Facts, &fact)
	return me
}

func (me *PVal) String() string {
	buf := ustd.BytesWriter{Data: make([]byte, 0, len(me.Facts)*16)}
	buf.WriteByte('{')
	for i, f := range me.Facts {
		buf.WriteString(f.String())
		if i != (len(me.Facts) - 1) {
			buf.WriteString(" AND ")
		}
	}
	buf.WriteByte('}')
	return buf.String()
}

func (me *PEnv) IsEnv() *PEnv { return me }
func (me *PEnv) Flatten() {
	if me.Link != nil {
		me.Link.Flatten()
		me.Facts = append(me.Facts, me.Link.Facts...)
		me.Link = nil
	}
}
