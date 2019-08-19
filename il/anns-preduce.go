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

func (me *PValPrimConst) Rewritten(rewrite func(IPreduced) IPreduced) IPreduced {
	return rewrite(me)
}
func (me *PValPrimConst) String() string {
	return "eqPrim(" + fmt.Sprintf("%v", me.ConstVal) + ")"
}

func (me *PValEqType) Errs() Errors { return me.Of.Errs() }
func (me *PValEqType) Rewritten(rewrite func(IPreduced) IPreduced) IPreduced {
	if of := rewrite(me.Of).(*PVal); of != me.Of {
		this := *me
		this.Of = of
		return rewrite(&this)
	}
	return rewrite(me)
}
func (me *PValEqType) String() string {
	return "eqType(" + me.Of.PValFactBase.String() + ")"
}

func (me *PValEqVal) Errs() Errors { return me.To.Errs() }
func (me *PValEqVal) Rewritten(rewrite func(IPreduced) IPreduced) IPreduced {
	if to := rewrite(me.To).(*PVal); to != me.To {
		this := *me
		this.To = to
		return rewrite(&this)
	}
	return rewrite(me)
}
func (me *PValEqVal) String() string {
	return "eqTo(" + me.To.PValFactBase.String() + ")"
}

func (me *PValFn) Errs() Errors { return append(me.Arg.Errs(), me.Ret.Errs()...) }
func (me *PValFn) Rewritten(rewrite func(IPreduced) IPreduced) IPreduced {
	parg, pret := rewrite(&me.Arg).(*PVal), rewrite(&me.Ret).(*PVal)
	if parg != &me.Arg || pret != &me.Ret {
		this := *me
		this.Arg, this.Ret = *parg, *pret
		return rewrite(&this)
	}
	return rewrite(me)
}
func (me *PValFn) String() string {
	return "fn(" + me.Arg.String() + "->" + me.Ret.String() + ")"
}

func (me *PValErr) Errs() Errors { return Errors{me.Error} }
func (me *PValErr) Rewritten(rewrite func(IPreduced) IPreduced) IPreduced {
	return rewrite(me)
}
func (me *PValErr) String() string {
	return "err(" + me.Error.Error() + ")"
}

func (me *PValAbyss) Rewritten(rewrite func(IPreduced) IPreduced) IPreduced {
	return rewrite(me)
}
func (me *PValAbyss) String() string {
	return "abyss(" + me.PValFactBase.String() + ")"
}

func (me *PValLink) Errs() Errors { return me.To.Errs() }
func (me *PValLink) Rewritten(rewrite func(IPreduced) IPreduced) IPreduced {
	if to := rewrite(me.To).(*PVal); to != me.To {
		this := *me
		this.To = to
		return rewrite(&this)
	}
	return rewrite(me)
}
func (me *PValLink) String() string {
	return "link(" + me.To.Self().String() + ")"
}

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

func (me *PVal) FnEnsure(loc IrRef) (ret *PValFn, fromLoc bool) {
	ret = me.Fn()
	if fromLoc = (ret == nil); fromLoc {
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
		errs = append(errs, f.Errs()...)
	}
	return
}

func (me *PVal) FromAppl(fn *PValFn, curApplArg *PVal) {
	this := fn.Ret.Rewritten(func(pred IPreduced) IPreduced {
		if pred == &fn.Arg {
			return curApplArg
		}
		return pred
	}).(*PVal)
	*me = *this
}

func (me *PVal) Rewritten(rewrite func(IPreduced) IPreduced) IPreduced {
	var diff bool
	facts := make([]IPreduced, len(me.Facts))
	for i, f := range me.Facts {
		if facts[i] = rewrite(f); facts[i] != f {
			diff = true
		}
	}
	if diff {
		this := *me
		this.Facts = facts
		return rewrite(&this)
	}
	return rewrite(me)
}

func (me *PVal) String() string {
	buf := ustd.BytesWriter{Data: make([]byte, 0, len(me.Facts)*16)}
	buf.WriteString("[|")
	buf.WriteString(me.PValFactBase.String())
	buf.WriteString("| ")
	for i, f := range me.Facts {
		buf.WriteString(f.String())
		if i != (len(me.Facts) - 1) {
			buf.WriteString(" AND ")
		}
	}
	buf.WriteString(" ]")
	return buf.String()
}
