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
		def.Ann.Preduced, def.Errs.Stage3Preduce = nil, nil
	}
	ctxpred := ctxPreducing{curSessCtx: me}
	for def, kit := range defIds {
		ctxpred.curNode.owningKit, ctxpred.curNode.owningTopDef, ctxpred.curAbs = kit, def, make(map[*IrAbs]IPreduced, 32)
		_ = ctxpred.preduce(def) // does set def.Ann.Preduced, and it must happen there not here
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
		if this.Ann.Preduced == nil && this.Errs.Stage3Preduce == nil { // only actively preduce if not already there --- both set to nil preparatorily in rePreduceTopLevelDefs
			this.Errs.Stage3Preduce = make(Errors, 0, 0) // not nil anymore now
			if this.HasErrors() {
				this.Ann.Preduced = newval.AddAbyss(append(nodeAncestors, this))
			} else {
				prevtopdef, prevenv := me.curNode.owningTopDef, me.curEnv
				me.curNode.owningTopDef, me.curEnv = this, nil
				{
					this.Ann.Preduced = me.preduce(this.Body)
				}
				me.curNode.owningTopDef, me.curEnv = prevtopdef, prevenv
			}
		}
		ret = this.Ann.Preduced

	case *IrLitFloat:
		ret = newval.AddPrimConst(append(nodeAncestors, this), this.Val)

	case *IrLitUint:
		ret = newval.AddPrimConst(append(nodeAncestors, this), this.Val)

	case *IrLitTag:
		ret = newval.AddPrimConst(append(nodeAncestors, this), this.Val)

	case *IrNonValue:
		ret = newval.AddAbyss(append(nodeAncestors, this))

	case *IrIdentName:
		switch len(this.Ann.Candidates) {
		case 0:
			ret = newval.AddErr(append(nodeAncestors, this), ErrNaming(ErrNames_NotFound, me.toks(this).First1(), "not in scope or not defined: `"+this.Name+"`"))
		case 1:
			if arg, isarg := this.Ann.Candidates[0].(*IrArg); isarg {
				ret = me.preduce(me.curNode.owningTopDef.ArgOwnerAbs(arg))
			} else {
				ret = me.preduce(this.Ann.Candidates[0])
			}
		default:
			ret = newval.AddErr(append(nodeAncestors, this), ErrNaming(ErrNames_Ambiguous, me.toks(this).First1(), "ambiguous: `"+this.Name+"`"))
		}

	case *IrAbs:
		ret = me.curAbs[this]
		if ret == nil {
			ret = &newval
			me.curAbs[this] = ret
			nodeancestors := append(nodeAncestors, this)

			rfn := newval.EnsureFn(nodeancestors)
			rfn.Ret.Add(me.preduce(this.Body, nodeancestors...))
		}

	case *IrAppl:
		ret = newval.AddAbyss(append(nodeAncestors, this))

	default:
		panic(this)
	}
	return
}
