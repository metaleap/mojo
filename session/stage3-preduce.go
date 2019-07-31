package atmosess

import (
	"github.com/go-leap/dev/lex"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
)

func (me *ctxPreduce) toks(n atmoil.INode) (toks udevlex.Tokens) {
	if tld := me.curNode.owningTopDef; tld != nil {
		toks = tld.OrigToks(n)
	}
	if orig := n.Origin(); orig != nil {
		toks = orig.Toks()
	}
	return
}

func (me *Ctx) Preduce(kit *Kit, node atmoil.INode) (atmoil.IPreduced, atmo.Errors) {
	ctxpreduce := &ctxPreduce{curSessCtx: me, inFlight: make(map[atmoil.INode]atmo.Exist, 32)}
	ctxpreduce.curNode.owningKit = kit
	ctxpreduce.curNode.owningTopDef, _ = node.(*atmoil.IrDefTop)
	return ctxpreduce.preduceIlNode(node)
}

func (me *ctxPreduce) preduceIlNode(node atmoil.INode) (ret atmoil.IPreduced, freshErrs atmo.Errors) {
	if _, rec := me.inFlight[node]; rec {
		ret = &atmoil.PAbyss{}
	} else {
		me.inFlight[node] = atmo.Є
		switch this := node.(type) {
		case *atmoil.IrLitFloat:
			ret = &atmoil.PPrimAtomicConstFloat{Val: this.Val}
		case *atmoil.IrLitUint:
			ret = &atmoil.PPrimAtomicConstUint{Val: this.Val}
		case *atmoil.IrIdentTag:
			ret = &atmoil.PPrimAtomicConstTag{Val: this.Val}
		case *atmoil.IrIdentName:
			cands := this.Anns.Candidates
			if len(cands) == 0 {
				freshErrs.AddPreduce(4321, me.toks(this), "notInScope")
			} else if len(cands) == 1 {
				ret, freshErrs = me.preduceIlNode(cands[0])
			} else {
				freshErrs.AddPreduce(1234, me.toks(this), "ambiguous")
			}
		case IrDefRef:
			curkit := me.curNode.owningKit
			me.curNode.owningKit = this.Kit
			ret, freshErrs = me.preduceIlNode(this.IrDefTop)
			me.curNode.owningKit = curkit
		case *atmoil.IrDefTop:
			curtopdef := me.curNode.owningTopDef
			me.curNode.owningTopDef = this
			ret, freshErrs = me.preduceIlNode(&this.IrDef)
			me.curNode.owningTopDef = curtopdef
		case *atmoil.IrDef:
			ret, freshErrs = me.preduceIlNode(this.Body)
		case *atmoil.IrDefArg:
		case *atmoil.IrAppl:
		default:
			panic(this)
		}
	}
	if ret != nil {
		ret.Self().OrigNodes = append(ret.Self().OrigNodes, node)
	}
	delete(me.inFlight, node)
	return
}
