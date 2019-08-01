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
	memoizecap := len(defIds) * 16
	ctxpred := ctxPreducing{curSessCtx: me, inFlight: make(map[atmoil.INode]atmo.Exist, 128), memoized: make(map[atmoil.INode]atmoil.IPreduced, memoizecap)}
	for def, kit := range defIds {
		ctxpred.curNode.owningKit, ctxpred.curNode.owningTopDef = kit, def
		_ = ctxpred.preduce(def)
	}
	if len(ctxpred.memoized) > memoizecap {
		println("toolil:", memoizecap, "vs.", len(ctxpred.memoized), "for", len(defIds))
	}
	return
}

func (me *Ctx) Preduce(nodeOwningKit *Kit, maybeNodeOwningTopDef *atmoil.IrDefTop, node atmoil.INode) atmoil.IPreduced {
	ctxpreduce := &ctxPreducing{curSessCtx: me, inFlight: make(map[atmoil.INode]atmo.Exist, 32), memoized: make(map[atmoil.INode]atmoil.IPreduced, 64)}
	ctxpreduce.curNode.owningKit, ctxpreduce.curNode.owningTopDef = nodeOwningKit, maybeNodeOwningTopDef
	return ctxpreduce.preduce(node)
}

func (me *ctxPreducing) preduce(node atmoil.INode) (ret atmoil.IPreduced) {
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
			if len(this.Anns.Candidates) == 0 {
				ret = &atmoil.PErr{Err: atmo.ErrNaming(4321, me.toks(this).First1(), "notInScope")}
			} else if len(this.Anns.Candidates) == 1 {
				ret = me.preduce(this.Anns.Candidates[0])
			} else {
				ret = &atmoil.PErr{Err: atmo.ErrNaming(1234, me.toks(this).First1(), "ambiguous")}
			}

		case IrDefRef:
			curkit := me.curNode.owningKit
			me.curNode.owningKit = this.Kit
			ret = me.preduce(this.IrDefTop)
			me.curNode.owningKit = curkit

		case *atmoil.IrDefTop:
			if this.Anns.Preduced == nil && this.Errs.Stage3Preduce == nil { // only actively preduce if not already there --- both set to nil preparatorily in rePreduceTopLevelDefs
				this.Errs.Stage3Preduce = make(atmo.Errors, 0, 0) // not nil anymore now
				if this.HasErrors() {
					this.Anns.Preduced = &atmoil.PErr{Err: this.Errors()[0]}
				} else {
					curtopdef := me.curNode.owningTopDef
					me.curNode.owningTopDef = this
					this.Anns.Preduced = me.preduce(&this.IrDef)
					me.curNode.owningTopDef = curtopdef
				}
			}
			ret = this.Anns.Preduced

		case *atmoil.IrDef:
			if this.Arg == nil {
				ret = me.preduce(this.Body)
			} else {
				ret = &atmoil.PCallable{Arg: &atmoil.PHole{Def: this}, Ret: &atmoil.PHole{Def: this}}
			}

		case *atmoil.IrDefArg:
			if curargval, ok := me.callArgs[this.Val]; !ok {
				ret = &atmoil.PErr{Err: atmo.ErrPreduce(4567, me.toks(this), "argNotSet: "+this.Val)}
			} else {
				ret = me.preduce(curargval)
			}

		case *atmoil.IrAppl:
			isoutermostappl := (me.callArgs == nil)
			if isoutermostappl {
				me.callArgs = make(map[string]atmoil.IExpr, 2)
			}
			callee := me.preduce(this.AtomicCallee)
			if callable, _ := callee.(*atmoil.PCallable); callable == nil {
				ret = &atmoil.PErr{Err: atmo.ErrPreduce(2345, me.toks(this.AtomicCallee), "notCallable: "+callee.SummaryCompact())}
			} else {
				arg := callable.Arg.Def.Arg
				argname := arg.Val
				if _, argexists := me.callArgs[argname]; argexists {
					ret = &atmoil.PErr{Err: atmo.ErrPreduce(3456, me.toks(this), "argExists: "+argname)}
				} else {
					me.callArgs[argname] = this.AtomicArg
					ret = me.preduce(callable.Ret.Def.Body)
				}
			}
			if isoutermostappl {
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
