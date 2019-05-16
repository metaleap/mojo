package atmosess

import (
	"strconv"

	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang/irfun"
)

type IValFact interface {
	String() string
}

type ValFacts []IValFact

func (me ValFacts) String() string {
	s := "("
	for i, fact := range me {
		if i > 0 {
			s += " & "
		}
		s += fact.String()
	}
	return s + ")"
}

type valFactPrimLit struct {
	value interface{}
}

func (me valFactPrimLit) String() string {
	switch v := me.value.(type) {
	case *atmolang_irfun.AstLitFloat:
		return strconv.FormatFloat(v.Val, 'G', -1, 64)
	case *atmolang_irfun.AstLitUint:
		return strconv.FormatUint(v.Val, 10)
	case *atmolang_irfun.AstLitStr:
		return strconv.Quote(v.Val)
	case *atmolang_irfun.AstLitRune:
		return strconv.QuoteRune(v.Val)
	default:
		panic(v)
	}
}

type valFactPrimTag struct {
	value string
}

func (me valFactPrimTag) String() string { return me.value }

type valFactCallable struct {
	arg ValFacts
	ret ValFacts
}

func (me valFactCallable) String() string { return me.arg.String() + " -> " + me.ret.String() }

type defNameFacts struct {
	overloads []*defIdFacts
}

func (me *defNameFacts) overloadByID(id string) *defIdFacts {
	for _, rc := range me.overloads {
		if rc.id == id {
			return rc
		}
	}
	return nil
}

func (me *defNameFacts) dropOverload(id string) {
	for i, rc := range me.overloads {
		if rc.id == id {
			me.overloads = append(me.overloads[:i], me.overloads[i+1:]...)
			break
		}
	}
}

type defIdFacts struct {
	id    string
	errs  atmo.Errors
	facts ValFacts
}

func (me *Ctx) reprocessAffectedIRsIfAnyKitsReloaded() {
	if me.state.someKitsNeedReprocessing {
		me.state.someKitsNeedReprocessing = false

		me.kitsRepopulateIdentNamesInScope()
		needsReSubst := make(map[string]*Kit, 32)

		for _, kit := range me.Kits.all {
			if len(kit.state.defsGoneIDsNames) > 0 {
				for defid, defname := range kit.state.defsGoneIDsNames {
					if dins := kit.defsFacts[defname]; dins != nil {
						dins.dropOverload(defid)
					}
				}
			}
			if len(kit.state.defsNew) > 0 {
				for _, defid := range kit.state.defsNew {
					if tldef := kit.lookups.tlDefsByID[defid]; tldef != nil && len(tldef.Errors) == 0 {
						if dans := kit.defsFacts[tldef.Name.Val]; dans != nil {
							dans.dropOverload(defid)
						}
						needsReSubst[defid] = kit
					}
				}
			}
			kit.state.defsGoneIDsNames, kit.state.defsNew = nil, nil
		}
		var errs []error
		for defid, kit := range needsReSubst {
			if rc := me.substantiateFactsIfNotAlready(kit, defid); len(rc.errs) > 0 {
				for i := range rc.errs {
					if e := &rc.errs[i]; !e.IsRef() {
						errs = append(errs, e)
					}
				}
			}
		}
		me.onErrs(errs, nil)
	}
}

func (me *Ctx) substantiateFactsIfNotAlready(kit *Kit, defId string) (dol *defIdFacts) {
	def := kit.lookups.tlDefsByID[defId]
	facts := kit.defsFacts[def.Name.Val]
	if facts == nil {
		facts = &defNameFacts{}
		kit.defsFacts[def.Name.Val] = facts
	}
	if dol = facts.overloadByID(defId); dol == nil {
		dol = &defIdFacts{id: defId}
		facts.overloads = append(facts.overloads, dol)
		// `def` is an `AstDefTop` and as such known to have no `Arg`, ever, so just using its `Body` only is fine:
		dol.facts, dol.errs = me.substantiateFactsForExpr(kit, dol, def.Body)
	}
	return
}

func (me *Ctx) substantiateFactsForExpr(kit *Kit, dol *defIdFacts, expr atmolang_irfun.IAstExpr) (facts ValFacts, errs atmo.Errors) {
	return
}
