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
	ctxpred := ctxPreducing{curSessCtx: me, curDefs: make(map[*atmoil.IrDef]atmo.Exist, 128)}
	for def, kit := range defIds {
		ctxpred.curNode.owningKit, ctxpred.curNode.owningTopDef = kit, def
		_ = ctxpred.preduce(def)
	}
	return
}

func (me *Ctx) Preduce(nodeOwningKit *Kit, maybeNodeOwningTopDef *atmoil.IrDefTop, node atmoil.INode) atmoil.IPreduced {
	ctxpreduce := &ctxPreducing{curSessCtx: me, curDefs: make(map[*atmoil.IrDef]atmo.Exist, 32)}
	ctxpreduce.curNode.owningKit, ctxpreduce.curNode.owningTopDef = nodeOwningKit, maybeNodeOwningTopDef
	return ctxpreduce.preduce(node)
}

func (me *ctxPreducing) preduce(node atmoil.INode) (ret atmoil.IPreduced) {
	switch this := node.(type) {

	case *atmoil.IrLitFloat:
		ret = &atmoil.PPrimAtomicConstFloat{Val: this.Val}

	case *atmoil.IrLitUint:
		ret = &atmoil.PPrimAtomicConstUint{Val: this.Val}

	case *atmoil.IrIdentTag:
		ret = &atmoil.PPrimAtomicConstTag{Val: this.Val}

	case *atmoil.IrIdentName:
		println("INTO_NAME", this.Val)
		if len(this.Anns.Candidates) == 0 {
			ret = &atmoil.PErr{Err: atmo.ErrNaming(4321, me.toks(this).First1(), "notInScope")}
		} else if len(this.Anns.Candidates) == 1 {
			ret = me.preduce(this.Anns.Candidates[0])
		} else {
			ret = &atmoil.PErr{Err: atmo.ErrNaming(1234, me.toks(this).First1(), "ambiguous")}
		}
		println("DONE_NAME", this.Val)

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
		println("INTO_DEF", this.Name.Val)
		if this.Arg == nil {
			ret = me.preduce(this.Body)
		} else {
			ret = &atmoil.PCallable{Arg: &atmoil.PHole{Def: this}, Ret: &atmoil.PHole{Def: this}}
		}
		println("DONE_DEF", this.Name.Val)

	case *atmoil.IrDefArg:
		println("INTO_ARG", this.Val)
		if curargval, ok := me.callArgs[this]; !ok {
			ret = &atmoil.PErr{Err: atmo.ErrPreduce(4567, me.toks(this), "argNotSet: "+this.Val)}
		} else {
			ret = me.preduce(curargval)
		}
		println("DONE_ARG", this.Val)

	case *atmoil.IrAppl:
		isoutermostappl := (me.callArgs == nil)
		if isoutermostappl {
			me.callArgs = make(map[*atmoil.IrDefArg]atmoil.IExpr, 2)
		}
		callee := me.preduce(this.Callee)
		if callable, _ := callee.(*atmoil.PCallable); callable == nil {
			ret = &atmoil.PErr{Err: atmo.ErrPreduce(2345, me.toks(this.Callee), "notCallable: "+callee.SummaryCompact())}
		} else {
			println("INTO_APPL", callable.Arg.Def.Name.Val)
			arg := callable.Arg.Def.Arg
			println("SET_ARG", arg.Val, atmoil.DbgPrintToString(this.CallArg))
			argprev, hadprev := me.callArgs[arg]
			me.callArgs[arg] = this.CallArg
			ret = me.preduce(callable.Ret.Def.Body)
			println("RET_WAS", ret.SummaryCompact())
			if hadprev {
				println("DEL_ARG", arg.Val, atmoil.DbgPrintToString(this.CallArg))
				me.callArgs[arg] = argprev
			}
			println("DONE_APPL", callable.Arg.Def.Name.Val)
		}
		if isoutermostappl {
			me.callArgs = nil
		}

	default:
		panic(this)
	}
	return
}
