package atmosess

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
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
	if 1 > 0 {
		return
	}
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

	case *atmoil.IrLitTag:
		ret = &atmoil.PPrimAtomicConstTag{Val: this.Val}

	case *atmoil.IrIdentName:
		println(ustr.Times("\t", me.dbgIndent)+"INTO_NAME", this.Val)
		me.dbgIndent++
		if len(this.Anns.Candidates) == 0 {
			ret = &atmoil.PErr{Err: atmo.ErrNaming(4321, me.toks(this).First1(), "notInScope")}
		} else if len(this.Anns.Candidates) == 1 {
			ret = me.preduce(this.Anns.Candidates[0])
		} else {
			ret = &atmoil.PErr{Err: atmo.ErrNaming(1234, me.toks(this).First1(), "ambiguous")}
		}
		me.dbgIndent--
		println(ustr.Times("\t", me.dbgIndent)+"DONE_NAME", ret.SummaryCompact())

	case IrDefRef:
		curkit := me.curNode.owningKit
		me.curNode.owningKit = this.Kit
		ret = me.preduce(this.IrDefTop)
		me.curNode.owningKit = curkit

	case *atmoil.IrDefTop:
		if 1 > 0 || (this.Anns.Preduced == nil && this.Errs.Stage3Preduce == nil) { // only actively preduce if not already there --- both set to nil preparatorily in rePreduceTopLevelDefs
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
		println(ustr.Times("\t", me.dbgIndent)+"INTO_DEF", ustr.ReplB(atmoil.DbgPrintToString(this), '\n', ' '))
		me.dbgIndent++
		if this.Arg == nil {
			ret = me.preduce(this.Body)
		} else {
			ret = &atmoil.PCallable{Arg: &atmoil.PHole{Def: this}, Ret: &atmoil.PHole{Def: this}}
		}
		me.dbgIndent--
		println(ustr.Times("\t", me.dbgIndent)+"DONE_DEF", ret.SummaryCompact())

	case *atmoil.IrArg:
		println(ustr.Times("\t", me.dbgIndent)+"INTO_ARG", this.Val)
		me.dbgIndent++

		if ret == nil {
			ret = &atmoil.PErr{Err: atmo.ErrPreduce(4567, me.toks(this), "argNotSet: "+this.Val)}
		}
		me.dbgIndent--
		println(ustr.Times("\t", me.dbgIndent)+"DONE_ARG", ret.SummaryCompact())

	case *atmoil.IrAppl:
		println(ustr.Times("\t", me.dbgIndent)+"INTO_APPL", atmoil.DbgPrintToString(this))
		me.dbgIndent++
		if callee := me.preduce(this.Callee); callee.IsErrOrAbyss() {
			ret = callee
		} else {
			if callable, iscallable := callee.(*atmoil.PCallable); !iscallable {
				ret = &atmoil.PErr{Err: atmo.ErrPreduce(2345, me.toks(this.Callee), "notCallable: "+callee.SummaryCompact())}
			} else {
				ret = me.preduce(callable.Arg.Def.Body)
				if callable, iscallable = ret.(*atmoil.PCallable); iscallable {
				}
			}
		}
		me.dbgIndent--
		println(ustr.Times("\t", me.dbgIndent)+"DONE_APPL", ret.SummaryCompact())

	case *atmoil.IrLam:
		ret = &atmoil.PAbyss{}

	default:
		panic(this)
	}
	return
}
