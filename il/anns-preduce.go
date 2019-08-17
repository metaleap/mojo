package atmoil

import (
	"fmt"
	"github.com/go-leap/std"
	. "github.com/metaleap/atmo"
)

func (me *PValFactBase) Errs() Errors        { return nil }
func (me *PValFactBase) Self() *PValFactBase { return me }
func (me *PValFactBase) String() string {
	return me.Loc.Def.IsDef().AstOrigToks(me.Loc.Node).String()
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

func (me *PValLink) String() string { return "link(" + me.To.Self().String() + ")" }

func (me *PVal) AddAbyss(loc IrRef) *PVal {
	fact := PValAbyss{}
	fact.Loc, me.Facts = loc, append(me.Facts, &fact)
	return me
}

func (me *PVal) AddLink(loc IrRef, to *PVal) *PVal {
	fact := PValLink{To: to}
	fact.Loc, me.Facts = loc, append(me.Facts, &fact)
	return me
}

func (me *PVal) AddUsed(loc IrRef) *PVal {
	fact := PValUsed{}
	fact.Loc, me.Facts = loc, append(me.Facts, &fact)
	return me
}

func (me *PVal) AddErr(loc IrRef, err *Error) *PVal {
	fact := PValErr{Error: err}
	fact.Loc, me.Facts = loc, append(me.Facts, &fact)
	return me
}

func (me *PVal) Fn() *PValFn {
	for _, f := range me.Facts {
		if fn, is := f.(*PValFn); is {
			return fn
		}
	}
	return nil
}

func (me *PVal) FnAdd(loc IrRef) *PValFn {
	fact, abs := PValFn{}, loc.Node.(*IrAbs)
	fact.Arg.Loc.Def, fact.Arg.Loc.Node = loc.Def, &abs.Arg
	fact.Ret.Loc.Def, fact.Ret.Loc.Node = loc.Def, abs.Body
	fact.Loc, me.Facts = loc, append(me.Facts, &fact)
	return &fact
}

func (me *PVal) FnEnsure(loc IrRef) (ret *PValFn) {
	if ret = me.Fn(); ret == nil {
		fact, appl := &PValFn{}, loc.Node.(*IrAppl)
		fact.Arg.Loc.Def, fact.Arg.Loc.Node = loc.Def, appl.CallArg
		fact.Ret.Loc.Def, fact.Ret.Loc.Node = loc.Def, appl.Callee
		fact.Loc, me.Facts, ret = loc, append(me.Facts, fact), fact
	}
	return
}

func (me *PVal) AddPrimConst(loc IrRef, constVal interface{}) *PVal {
	fact := PValPrimConst{ConstVal: constVal}
	fact.Loc, me.Facts = loc, append(me.Facts, &fact)
	return me
}

func (me *PVal) Add(oneOrMultipleFacts IPreduced) *PVal {
	switch f := oneOrMultipleFacts.(type) {
	case *PVal:
		me.Facts = append(me.Facts, f.Facts...)
	default:
		me.Facts = append(me.Facts, f)
	}
	return me
}

func (me *PVal) Errs() (errs Errors) {
	for _, f := range me.Facts {
		errs.Add(f.Errs()...)
	}
	return
}

func (me *PVal) FromAppl(fn *PValFn, curApplArg *PVal) {
}

func (me *PVal) String() string {
	buf := ustd.BytesWriter{Data: make([]byte, 0, len(me.Facts)*16)}
	buf.WriteString("[ ")
	for i, f := range me.Facts {
		buf.WriteString(f.String())
		if i != (len(me.Facts) - 1) {
			buf.WriteString(" AND ")
		}
	}
	buf.WriteString(" ]")
	return buf.String()
}
