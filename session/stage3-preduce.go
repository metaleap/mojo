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
		ctxpred.curNode.owningKit, ctxpred.curNode.owningTopDef = kit, def
		_ = ctxpred.preduce(def) // does set def.Anns.Preduced, and it must happen there not here
	}
	return
}

func (me *Ctx) Preduce(nodeOwningKit *Kit, maybeNodeOwningTopDef *IrDef, node IIrNode) IPreduced {
	ctxpreduce := &ctxPreducing{curSessCtx: me}
	ctxpreduce.curNode.owningKit, ctxpreduce.curNode.owningTopDef = nodeOwningKit, maybeNodeOwningTopDef
	return ctxpreduce.preduce(node, maybeNodeOwningTopDef.AncestorsOf(node)...)
}

func (me *ctxPreducing) preduce(node IIrNode, nodeAncestors ...IIrNode) (ret IPreduced) {
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
				prevtopdef := me.curNode.owningTopDef
				me.curNode.owningTopDef = this
				{
					this.Anns.Preduced = &PEnv{}
				}
				me.curNode.owningTopDef = prevtopdef
			}
		}
		ret = this.Anns.Preduced

	case *IrLitFloat:
		ret = (&PVal{}).AddPrimConst(append(nodeAncestors, this), this.Val)

	case *IrLitUint:
		ret = (&PVal{}).AddPrimConst(append(nodeAncestors, this), this.Val)

	case *IrLitTag:
		ret = (&PVal{}).AddPrimConst(append(nodeAncestors, this), this.Val)

	case *IrNonValue:

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

	case *IrArg:

	case *IrAppl:

	default:
		panic(this)
	}
	return
}
