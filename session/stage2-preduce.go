package atmosess

import (
	"strconv"

	"github.com/go-leap/str"
	"github.com/metaleap/atmo/il"
)

type IPreduced interface {
	self() *preducedBase
	SummaryCompact() string
}

type preducedBase struct {
	origNode atmoil.INode
}

func (me *preducedBase) self() *preducedBase { return me }

type PPrimAtomicConstFloat struct {
	preducedBase
	Val float64
}

func (me *PPrimAtomicConstFloat) SummaryCompact() string {
	return strconv.FormatFloat(me.Val, 'f', -1, 64)
}

type PPrimAtomicConstUint struct {
	preducedBase
	Val uint64
}

func (me *PPrimAtomicConstUint) SummaryCompact() string { return strconv.FormatUint(me.Val, 10) }

type PPrimAtomicConstTag struct {
	preducedBase
	Val string
}

func (me *PPrimAtomicConstTag) SummaryCompact() string { return me.Val }

type PFailure struct {
	preducedBase

	ErrMsg string
}

func (me *PFailure) SummaryCompact() string { return me.ErrMsg }

type PCallable struct {
	preducedBase
}

func (me *PCallable) SummaryCompact() string { return "->" }

type ctxPreduce struct {
	owner          *Ctx
	cachedByTldIds map[string]IPreduced
}

func (me *Ctx) PreduceExpr(kit *Kit, maybeTopDefId string, expr atmoil.IExpr) IPreduced {
	var maybetld *atmoil.IrDefTop
	if maybeTopDefId != "" {
		maybetld = kit.lookups.tlDefsByID[maybeTopDefId]
	}
	return me.state.preduce.preduceIlNode(kit, maybetld, expr)
}

func (me *ctxPreduce) preduceIlNode(kit *Kit, maybeTld *atmoil.IrDefTop, node atmoil.INode) (ret IPreduced) {
	switch this := node.(type) {
	case *atmoil.IrLitFloat:
		ret = &PPrimAtomicConstFloat{Val: this.Val}
	case *atmoil.IrLitUint:
		ret = &PPrimAtomicConstUint{Val: this.Val}
	case *atmoil.IrIdentTag:
		ret = &PPrimAtomicConstTag{Val: this.Val}
	case *atmoil.IrIdentName:
		cands := this.Anns.Candidates
		if len(cands) == 0 {
			ret = &PFailure{ErrMsg: "not defined or not in scope: `" + this.Val + "`"}
		} else if len(cands) == 1 {
			ret = me.preduceIlNode(kit, maybeTld, cands[0])
		} else {
			ret = &PFailure{ErrMsg: "ambiguous: `" + this.Val + "` (" + ustr.Int(len(cands)) + " such names in scope)"}
		}
	case *atmoil.IrDefTop:
		ret = me.preduceIlNode(kit, this, &this.IrDef)
	case *atmoil.IrDef:
		ret = me.preduceIlNode(kit, maybeTld, this.Body)
	case IrDefRef:
		ret = me.preduceIlNode(this.Kit, this.IrDefTop, this.IrDefTop)
	// case *atmoil.IrAppl:

	default:
		panic(this)
	}
	ret.self().origNode = node
	return
}
