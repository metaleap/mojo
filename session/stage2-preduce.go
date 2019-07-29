package atmosess

import (
	"github.com/go-leap/dev/lex"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
)

type ctxPreduce struct {
	owner          *Ctx
	cachedByTldIds map[string]atmoil.IPreduced
	curNodeCtx     struct {
		kit    *Kit
		topDef *atmoil.IrDefTop
	}
}

func (me *ctxPreduce) toks(n atmoil.INode) udevlex.Tokens { return me.curNodeCtx.topDef.OrigToks(n) }

func (me *Ctx) PreduceExpr(kit *Kit, expr atmoil.IExpr) (atmoil.IPreduced, atmo.Errors) {
	me.state.preduce.curNodeCtx.kit = kit
	return me.state.preduce.preduceIlNode(expr)
}

func (me *ctxPreduce) preduceIlNode(node atmoil.INode) (ret atmoil.IPreduced, freshErrs atmo.Errors) {
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
		curkit := me.curNodeCtx.kit
		me.curNodeCtx.kit = this.Kit
		ret, freshErrs = me.preduceIlNode(this.IrDefTop)
		me.curNodeCtx.kit = curkit
	case *atmoil.IrDefTop:
		pred, exists := me.cachedByTldIds[this.Id]
		if !exists {
			me.cachedByTldIds[this.Id] = nil

			curtopdef := me.curNodeCtx.topDef
			me.curNodeCtx.topDef = this
			pred, this.Errs.Stage2Preduce = me.preduceIlNode(&this.IrDef)
			me.curNodeCtx.topDef = curtopdef

			me.cachedByTldIds[this.Id] = pred
		}

		ret, freshErrs = pred, this.Errs.Stage2Preduce
	case *atmoil.IrDef:
		ret, freshErrs = me.preduceIlNode(this.Body)
	// case *atmoil.IrAppl:

	default:
		panic(this)
	}
	if ret != nil {
		ret.Self().OrigNodes = append(ret.Self().OrigNodes, node)
	}
	return
}
