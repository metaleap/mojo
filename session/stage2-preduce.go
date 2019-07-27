package atmosess

import (
	"strconv"

	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
)

type IPreduced interface {
	self() *preducedBase
	SummaryCompact() string
}

type preducedBase struct {
	origNodePath []atmoil.INode
}

func (me *preducedBase) self() *preducedBase { return me }

type PConstValAtomic struct {
	preducedBase
	LitVal interface{}
}

func (me *PConstValAtomic) SummaryCompact() string {
	switch lv := me.LitVal.(type) {
	case float64:
		return strconv.FormatFloat(lv, 'f', -1, 64)
	case uint64:
		return strconv.FormatUint(lv, 10)
	case string: // tag
		return lv
	default:
		panic(lv)
	}
}

type PConstValCompound struct {
	preducedBase

	// later will be more principled as compound will encompass lists-of-any, tuples, relations / maps / records etc.
	TmpVal string
}

func (me *PConstValCompound) SummaryCompact() string {
	return strconv.Quote(me.TmpVal)
}

type PCallable struct {
	preducedBase
}

type ctxPreduce struct {
	owner *Ctx
}

func (me *Ctx) PreduceExpr(kit *Kit, expr atmoil.IExpr) (IPreduced, atmo.Errors) {
	return me.state.preduce.preduceIlNode(kit, expr)
}

func (me *ctxPreduce) preduceIlNode(kit *Kit, nodePath ...atmoil.INode) (ret IPreduced, errs atmo.Errors) {
	switch n := nodePath[len(nodePath)-1].(type) {
	case *atmoil.IrLitFloat:
		ret = &PConstValAtomic{LitVal: n.Val}
	case *atmoil.IrLitUint:
		ret = &PConstValAtomic{LitVal: n.Val}
	case *atmoil.IrIdentTag:
		ret = &PConstValAtomic{LitVal: n.Val}
	case *atmoil.IrLitStr:
		ret = &PConstValCompound{TmpVal: n.Val}
	default:
		panic(n)
	}
	ret.self().origNodePath = nodePath
	return
}
