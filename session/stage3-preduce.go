package atmosess

import (
	"github.com/go-leap/dev/lex"
	. "github.com/metaleap/atmo"
	. "github.com/metaleap/atmo/il"
)

func (me *ctxPreducing) def() IrDefRef {
	return IrDefRef{Kit: me.curNode.owningKit, IrDef: me.curNode.owningTopDef}
}

func (me *ctxPreducing) ref(node IIrNode) IrRef {
	return IrRef{Node: node, Def: me.def()}
}

func (me *ctxPreducing) toks(n IIrNode) udevlex.Tokens {
	return me.curNode.owningTopDef.AstOrigToks(n)
}

func (me *Ctx) rePreduceTopLevelDefs(defIds map[*IrDef]*Kit) (freshErrs Errors) {
	for def := range defIds {
		def.Ann.Preduced, def.Errs.Stage3Preduce = nil, nil
	}
	ctxpred := ctxPreducing{curSessCtx: me}
	for def, kit := range defIds {
		ctxpred.curNode.owningKit, ctxpred.curNode.owningTopDef, ctxpred.curAbs = kit, def, make(map[*IrAbs]*PVal, 32)
		pred := ctxpred.preduce(def) // does set def.Ann.Preduced, and it must happen there not here
		def.Errs.Stage3Preduce = pred.Errs()
		freshErrs.Add(def.Errs.Stage3Preduce...)
	}
	return
}

func (me *Ctx) Preduce(nodeOwningKit *Kit, maybeNodeOwningTopDef *IrDef, node IIrNode) *PVal {
	ctxpreduce := &ctxPreducing{curSessCtx: me, curAbs: make(map[*IrAbs]*PVal, 8)}
	ctxpreduce.curNode.owningKit, ctxpreduce.curNode.owningTopDef = nodeOwningKit, maybeNodeOwningTopDef
	return ctxpreduce.preduce(node)
}

func (me *ctxPreducing) preduce(node IIrNode) (ret *PVal) {
	var newval PVal
	ref := me.ref(node)

	switch this := node.(type) {

	case IrDefRef:
		prevkit := me.curNode.owningKit
		me.curNode.owningKit = this.Kit
		ret = me.preduce(this.IrDef)
		me.curNode.owningKit = prevkit

	case *IrDef:
		if this.Ann.Preduced == nil && this.Errs.Stage3Preduce == nil { // only actively preduce if not already there --- both set to nil preparatorily in rePreduceTopLevelDefs
			this.Errs.Stage3Preduce = make(Errors, 0, 0) // not nil anymore now
			if this.HasErrors() {
				this.Ann.Preduced = newval.AddAbyss(ref)
			} else {
				prevtopdef := me.curNode.owningTopDef
				me.curNode.owningTopDef = this
				this.Ann.Preduced = me.preduce(this.Body)
				me.curNode.owningTopDef = prevtopdef
			}
		}
		ret = this.Ann.Preduced

	case *IrLitFloat:
		ret = newval.AddPrimConst(ref, this.Val)

	case *IrLitUint:
		ret = newval.AddPrimConst(ref, this.Val)

	case *IrLitTag:
		ret = newval.AddPrimConst(ref, this.Val)

	case *IrNonValue:
		ret = newval.AddAbyss(ref)

	case *IrIdentName:
		switch len(this.Ann.Candidates) {
		case 0:
			ret = newval.AddErr(ref, ErrNaming(ErrNames_NotFound, me.toks(this).First1(), "not in scope or not defined: `"+this.Name+"`"))
		case 1:
			ret = me.preduce(this.Ann.Candidates[0])
			ret.AddUsed(ref)
		default:
			ret = newval.AddErr(ref, ErrNaming(ErrNames_Ambiguous, me.toks(this).First1(), "ambiguous: `"+this.Name+"`"))
		}

	case *IrArg:
		pafn := me.preduce(me.curNode.owningTopDef.ArgOwnerAbs(this))
		ret = &(pafn.Fn().Arg)

	case *IrAbs:
		ret = me.curAbs[this]
		if ret == nil {
			ret = &newval
			me.curAbs[this] = ret

			rfn := newval.FnAdd(ref)
			rfn.Ret.Add(me.preduce(this.Body))
		}

	case *IrAppl:
		pcallee := me.preduce(this.Callee)
		parg := me.preduce(this.CallArg)

		islocal := pcallee.Loc.Def == me.curNode.owningTopDef
		if islocal {
		}
		pcfn := pcallee.FnEnsure(ref)

		parg.AddLink(ref, &pcfn.Arg)

		println("CALLEE", pcallee.String, "ARG", parg.String())
		newval.FromAppl(pcfn, parg)
		ret = &newval
	default:
		panic(this)
	}
	return
}
