package atmosess

import (
	"github.com/go-leap/dev/lex"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
)

func (me *ctxPreducing) toks(n atmoil.INode) (toks udevlex.Tokens) {
	if tld := me.curNode.owningTopDef; tld != nil {
		toks = tld.OrigToks(n)
	}
	if orig := n.Origin(); orig != nil {
		toks = orig.Toks()
	}
	return
}

func (me *Ctx) rePreduceTopLevelDefs(defIds map[*atmoil.IrDefTop]*Kit) (freshErrs atmo.Errors) {
	for def := range defIds {
		def.Anns.Preduced, def.Errs.Stage3Preduce = nil, nil
	}
	memoizecap := len(defIds) * 5
	ctxpred := ctxPreducing{curSessCtx: me, inFlight: make(map[atmoil.INode]atmo.Exist, 16), memoized: make(map[atmoil.INode]atmoil.IPreduced, memoizecap)}
	for def, kit := range defIds {
		ctxpred.curNode.owningKit, ctxpred.curNode.owningTopDef = kit, def
		_ = ctxpred.preduceIlNode(def)
	}
	if len(ctxpred.memoized) > memoizecap {
		panic(len(ctxpred.memoized)) // until Std-libs have emerged and grown realistic
	}
	return
}

func (me *Ctx) Preduce(nodeOwningKit *Kit, maybeNodeOwningTopDef *atmoil.IrDefTop, node atmoil.INode) atmoil.IPreduced {
	ctxpreduce := &ctxPreducing{curSessCtx: me, inFlight: make(map[atmoil.INode]atmo.Exist, 8), memoized: make(map[atmoil.INode]atmoil.IPreduced, 64)}
	ctxpreduce.curNode.owningKit, ctxpreduce.curNode.owningTopDef = nodeOwningKit, maybeNodeOwningTopDef
	return ctxpreduce.preduceIlNode(node)
}

var ctxPreducingTmpMax int

func (me *ctxPreducing) preduceIlNode(node atmoil.INode) (ret atmoil.IPreduced) {
	if len(me.inFlight) > ctxPreducingTmpMax {
		ctxPreducingTmpMax = len(me.inFlight)
		println("INFLIGHT", ctxPreducingTmpMax)
	}
	if _, rec := me.inFlight[node]; rec {
		ret = &atmoil.PAbyss{}
	} else if ret, _ = me.memoized[node]; ret != nil {
		return
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
				ret = &atmoil.PErr{Err: atmo.ErrNaming(4321, me.toks(this).First1(), "notInScope")}
			} else if len(cands) == 1 {
				ret = me.preduceIlNode(cands[0])
			} else {
				ret = &atmoil.PErr{Err: atmo.ErrNaming(1234, me.toks(this).First1(), "ambiguous")}
			}

		case IrDefRef:
			curkit := me.curNode.owningKit
			me.curNode.owningKit = this.Kit
			ret = me.preduceIlNode(this.IrDefTop)
			me.curNode.owningKit = curkit

		case *atmoil.IrDefTop:
			if this.Anns.Preduced == nil && this.Errs.Stage3Preduce == nil { // only actively preduce if not already there --- both set to nil preparatorily in rePreduceTopLevelDefs
				this.Errs.Stage3Preduce = make(atmo.Errors, 0, 0) // not nil anymore now
				if this.HasErrors() {
					this.Anns.Preduced = &atmoil.PErr{Err: this.Errors()[0]}
				} else {
					curtopdef := me.curNode.owningTopDef
					me.curNode.owningTopDef = this
					this.Anns.Preduced = me.preduceIlNode(&this.IrDef)
					me.curNode.owningTopDef = curtopdef
				}
			}
			ret = this.Anns.Preduced

		case *atmoil.IrDef:
			ret = me.preduceIlNode(this.Body)
			if this.Arg != nil && ret != nil {
				ret = &atmoil.PCallable{Arg: me.preduceIlNode(this.Arg), Ret: ret}
			}

		case *atmoil.IrDefArg:
			ret = &atmoil.PHole{}

		case *atmoil.IrAppl:
			isouter := (me.callArgs == nil)
			if isouter {
				me.callArgs = make(map[string]interface{})
			}
			callee := me.preduceIlNode(this.AtomicCallee)
			if callable, _ := callee.(*atmoil.PCallable); callable == nil {
				ret = &atmoil.PErr{Err: atmo.ErrPreduce(2345, me.toks(this.AtomicCallee), "notCallable")}
			} else {

			}
			if isouter {
				me.callArgs = nil
			}

		default:
			panic(this)
		}
	}
	me.memoized[node] = ret
	delete(me.inFlight, node)
	return
}
