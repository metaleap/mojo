package atmosess

import (
	"strconv"

	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang/irfun"
)

type IValFact interface {
	Errs() atmo.Errors
	String() string
}

type valFacts []IValFact

func (me *valFacts) add(facts ...IValFact) { *me = append(*me, facts...) }
func (me *valFacts) errs(ensure bool) (vfe *valFactErrs) {
	this := *me
	for i := range this {
		if vfe, _ = this[i].(*valFactErrs); vfe != nil {
			break
		}
	}
	if ensure && vfe == nil {
		vfe = &valFactErrs{}
		*me = append(this, vfe)
	}
	return
}
func (me *valFacts) callable(ensure bool) (vfc *valFactCallable) {
	this := *me
	for i := range this {
		if vfc, _ = this[i].(*valFactCallable); vfc != nil {
			break
		}
	}
	if ensure && vfc == nil {
		vfc = &valFactCallable{}
		*me = append(this, vfc)
	}
	return
}
func (me valFacts) Errs() (errs atmo.Errors) {
	if len(me) == 1 {
		return me[0].Errs()
	}
	for i := range me {
		errs.Add(me[i].Errs())
	}
	return
}
func (me valFacts) String() string {
	if len(me) == 0 {
		return "❲❳"
	} else if len(me) == 1 {
		return me[0].String()
	}
	s := "❲"
	for i, fact := range me {
		if i > 0 {
			s += " & "
		}
		s += fact.String()
	}
	return s + "❳"
}

type ValFacts struct {
	// not used in this package except to return to the outside as a wrapper
	valFacts
	errs atmo.Errors
}

func (me *ValFacts) Errs() atmo.Errors {
	if me.errs == nil {
		me.errs = me.valFacts.Errs()
	}
	return me.errs
}

type valFactErrs struct {
	atmo.Errors
}

func (me *valFactErrs) Errs() atmo.Errors { return me.Errors }
func (me *valFactErrs) String() string    { return "" }

type valFactPrim struct {
	orig  atmolang_irfun.IAstExpr
	value interface{}
}

func (me *valFactPrim) Errs() atmo.Errors { return nil }
func (me *valFactPrim) String() string {
	switch v := me.value.(type) {
	case float64:
		return strconv.FormatFloat(v, 'G', -1, 64)
	case uint64:
		return strconv.FormatUint(v, 10)
	case string:
		if _, ok := me.orig.(*atmolang_irfun.AstIdentTag); ok {
			return v
		}
		return strconv.Quote(v)
	case rune:
		return strconv.QuoteRune(v)
	default:
		panic(v)
	}
}

type valFactCallableArg struct {
	valFacts
	orig *atmolang_irfun.AstDefArg
}

func (me *valFactCallableArg) String() string {
	return "‹" + me.orig.AstIdentName.Val + ": " + me.valFacts.String() + "›"
}

type valFactCallable struct {
	arg valFactCallableArg
	ret valFacts
}

func (me *valFactCallable) Errs() (errs atmo.Errors) {
	errs.Add(me.arg.Errs())
	errs.Add(me.ret.Errs())
	return
}
func (me *valFactCallable) String() string { return me.arg.String() + " » " + me.ret.String() }

type valFactRef struct {
	*valFacts
}

func (me *valFactRef) Errs() atmo.Errors { return me.valFacts.Errs().Refs() }

type valFactArgRef struct {
	*valFactCallable
}

func (me *valFactArgRef) Errs() atmo.Errors { return me.valFactCallable.arg.Errs().Refs() }
func (me *valFactArgRef) String() string    { return me.valFactCallable.arg.orig.AstIdentName.Val }

type defValFinisher func(*Kit, *defIdFacts, *atmolang_irfun.AstDef)

type defNameFacts struct {
	overloads []*defIdFacts // not a map because most (~90+%?) will be of len 1
}

func (me *defNameFacts) overloadById(id string) *defIdFacts {
	for _, dif := range me.overloads {
		if dif.def.Id == id {
			return dif
		}
	}
	return nil
}

func (me *defNameFacts) overloadDrop(id string) {
	for i, dif := range me.overloads {
		if dif.def.Id == id {
			me.overloads = append(me.overloads[:i], me.overloads[i+1:]...)
			break
		}
	}
}

type defIdFacts struct {
	def *atmolang_irfun.AstDefTop
	valFacts
	cache map[*atmolang_irfun.AstDef]*valFacts
}

func (me *defIdFacts) errsNonRef() (errs []error) {
	if errors := me.Errs(); len(errors) > 0 {
		errs = make([]error, 0, len(errors))
		for _, e := range errors {
			if !e.IsRef() {
				errs = append(errs, e)
			}
		}
	}
	return
}
