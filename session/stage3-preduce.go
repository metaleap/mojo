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

func (me *Ctx) rePreduceTopLevelDefs(defIds map[*atmoil.IrDefTop]*Kit) (freshErrs atmo.Errors) {
	for def := range defIds {
		def.Anns.Preduced, def.Errs.Stage3Preduce = nil, nil
	}
	for def, kit := range defIds {
		_ = me.Preduce(kit, def, def)
	}
	return
}

func (me *Ctx) Preduce(nodeOwningKit *Kit, maybeNodeOwningTopDef *atmoil.IrDefTop, node atmoil.INode) atmoil.IPreduced {
	ctxpreduce := &ctxPreduce{curSessCtx: me, inFlight: make(map[atmoil.INode]atmo.Exist, 32)}
	ctxpreduce.curNode.owningKit, ctxpreduce.curNode.owningTopDef = nodeOwningKit, maybeNodeOwningTopDef
	return ctxpreduce.preduceIlNode(node)
}

func (me *ctxPreduce) preduceIlNode(node atmoil.INode) (ret atmoil.IPreduced) {
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
				curtopdef := me.curNode.owningTopDef
				me.curNode.owningTopDef = this
				this.Anns.Preduced = me.preduceIlNode(&this.IrDef)
				me.curNode.owningTopDef = curtopdef
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

		default:
			panic(this)
		}
	}
	delete(me.inFlight, node)
	return
}
