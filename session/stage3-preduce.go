package atmosess

import (
	"github.com/go-leap/dev/lex"
	. "github.com/metaleap/atmo"
	. "github.com/metaleap/atmo/il"
)

func (me *ctxPreducing) toks(n IIrNode) udevlex.Tokens {
	return me.curNode.owningTopDef.AstOrigToks(n)
}

func (me *Ctx) rePreduceTopLevelDefs(defIds map[*IrDef]*Kit) (freshErrs Errors) {
	for def := range defIds {
		def.Anns.Preduced, def.Errs.Stage3Preduce = nil, nil
	}
	ctxpred := ctxPreducing{curSessCtx: me}
	for def, kit := range defIds {
		ctxpred.curNode.owningKit, ctxpred.curNode.owningTopDef, ctxpred.curAbs = kit, def, make(map[*IrAbs]IPreduced, 32)
		_ = ctxpred.preduce(def) // does set def.Anns.Preduced, and it must happen there not here
	}
	return
}

func (me *Ctx) Preduce(nodeOwningKit *Kit, maybeNodeOwningTopDef *IrDef, node IIrNode) IPreduced {
	ctxpreduce := &ctxPreducing{curSessCtx: me, curAbs: make(map[*IrAbs]IPreduced, 8)}
	ctxpreduce.curNode.owningKit, ctxpreduce.curNode.owningTopDef = nodeOwningKit, maybeNodeOwningTopDef
	return ctxpreduce.preduce(node, maybeNodeOwningTopDef.AncestorsOf(node)...)
}

func (me *ctxPreducing) preduce(node IIrNode, nodeAncestors ...IIrNode) (ret IPreduced) {
	var newval PVal

	switch this := node.(type) {

	case IrDefRef:
		prevkit := me.curNode.owningKit
		me.curNode.owningKit = this.Kit
		ret = me.preduce(this.IrDef)
		me.curNode.owningKit = prevkit

	case *IrDef:
		if this.Anns.Preduced == nil && this.Errs.Stage3Preduce == nil { // only actively preduce if not already there --- both set to nil preparatorily in rePreduceTopLevelDefs
			this.Errs.Stage3Preduce = make(Errors, 0, 0) // not nil anymore now
			if !this.HasErrors() {
				prevtopdef, prevenv := me.curNode.owningTopDef, me.curEnv
				me.curNode.owningTopDef, me.curEnv = this, nil
				{
					this.Anns.Preduced = me.preduce(this.Body)
				}
				me.curNode.owningTopDef, me.curEnv = prevtopdef, prevenv
			}
		}
		ret = this.Anns.Preduced

	case *IrLitFloat:
		ret = newval.AddPrimConst(append(nodeAncestors, this), this.Val)

	case *IrLitUint:
		ret = newval.AddPrimConst(append(nodeAncestors, this), this.Val)

	case *IrLitTag:
		ret = newval.AddPrimConst(append(nodeAncestors, this), this.Val)

	case *IrNonValue:
		ret = &newval
		switch {
		case this.OneOf.TempStrLit:
			ret = newval.AddErr(append(nodeAncestors, this), ErrTodo(0, me.toks(this), "str-lits"))
		case this.OneOf.Undefined:
			ret = newval.AddAbyss(append(nodeAncestors, this))
		}

	case *IrIdentName:
		switch len(this.Anns.Candidates) {
		case 1:
			if abs, isabs := this.Anns.Candidates[0].(*IrAbs); isabs {
				ret = me.preduce(&abs.Arg)
			} else {
				ret = me.preduce(this.Anns.Candidates[0])
			}
		}

	case *IrAbs:
		ret = me.curAbs[this]
		if ret == nil {
			ret = &newval
			me.curAbs[this] = ret
			nodeancestors := append(nodeAncestors, this)
			rfn := newval.EnsureFn(nodeancestors, true)
			rfn.Ret.Add(me.preduce(this.Body, nodeancestors...))
		}

	case *IrAppl:

	default:
		panic(this)
	}
	return
}
