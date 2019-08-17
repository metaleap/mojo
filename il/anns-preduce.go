package atmoil

import (
	"fmt"
	"github.com/go-leap/std"
	. "github.com/metaleap/atmo"
)

func (me *PValFactBase) Errs() Errors        { return nil }
func (me *PValFactBase) Self() *PValFactBase { return me }
func (me *PValFactBase) String() string {
	return me.From.Def.IsDef().AstOrigToks(me.From.Node).Pos().String()
}

func (me *PValUsed) String() string { return "used(" + me.PValFactBase.String() + ")" }

func (me *PValPrimConst) String() string { return "eqPrim(" + fmt.Sprintf("%v", me.ConstVal) + ")" }

func (me *PValEqType) String() string { return "eqType(" + me.Of.PValFactBase.String() + ")" }

func (me *PValEqVal) String() string { return "eqTo(" + me.To.PValFactBase.String() + ")" }

func (me *PValNever) String() string { return "never(" + me.Never.Self().String() + ")" }

func (me *PValFn) String() string { return "fn(" + me.Arg.String() + "->" + me.Ret.String() + ")" }

func (me *PValErr) Errs() Errors   { return Errors{me.Error} }
func (me *PValErr) String() string { return "err(" + me.Error.Error() + ")" }

func (me *PValAbyss) String() string { return "abyss(" + me.PValFactBase.String() + ")" }

func (me *PVal) AddAbyss(from IrRef) *PVal {
	fact := PValAbyss{}
	fact.From, me.Facts = from, append(me.Facts, &fact)
	return me
}

func (me *PVal) AddErr(from IrRef, err *Error) *PVal {
	fact := PValErr{Error: err}
	fact.From, me.Facts = from, append(me.Facts, &fact)
	return me
}

func (me *PVal) EnsureFn(from IrRef) *PValFn {
	for _, f := range me.Facts {
		if fn, is := f.(*PValFn); is {
			return fn
		}
	}

	fact, abs := PValFn{}, from.Node.(*IrAbs)
	fact.Arg.From.Def, fact.Arg.From.Node = from.Def, &abs.Arg
	fact.Ret.From.Def, fact.Ret.From.Node = from.Def, abs.Body
	fact.From, me.Facts = from, append(me.Facts, &fact)
	return &fact
}

func (me *PVal) AddPrimConst(from IrRef, constVal interface{}) *PVal {
	fact := PValPrimConst{ConstVal: constVal}
	fact.From, me.Facts = from, append(me.Facts, &fact)
	return me
}

func (me *PVal) Add(oneOrMultipleFacts IPreduced) *PVal {
	switch f := oneOrMultipleFacts.(type) {
	case *PVal:
		me.Facts = append(me.Facts, f.Facts...)
	case *PEnv:
		me.Facts = append(me.Facts, f.Facts...)
	default:
		me.Facts = append(me.Facts, f)
	}
	return me
}

func (me *PVal) Errs() (errs Errors) {
	for _, f := range me.Facts {
		if e, _ := f.(*PValErr); e != nil {
			errs.Add(e.Error)
		}
	}
	return
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

func (me *PEnv) Errs() (errs Errors) {
	if errs = me.PVal.Errs(); me.Link != nil {
		errs.Add(me.Link.Errs()...)
	}
	return
}

func (me *PEnv) Flatten() {
	if me.Link != nil {
		me.Link.Flatten()
		me.Facts = append(me.Facts, me.Link.Facts...)
		me.Link = nil
	}
}
