package atmosess

import (
	"fmt"
	"strconv"

	"github.com/go-leap/str"
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
		return "()"
	} else if len(me) == 1 {
		return me[0].String()
	}
	s := "("
	for i, fact := range me {
		if i > 0 {
			s += " & "
		}
		s += fact.String()
	}
	return s + ")"
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

func (me *Ctx) substantiateKitsDefsFactsAsNeeded() {
	reSubstFirst, reSubstNext := make(map[string]*Kit, 8), make(map[string]*Kit, 16)

	namesofchange := make(map[string]bool, 4)
	for _, kit := range me.Kits.all {
		if len(kit.state.defsGoneIDsNames) > 0 {
			for defid, defname := range kit.state.defsGoneIDsNames {
				if me.state.kitsReprocessing.ever {
					namesofchange[defname] = true
				}
				if dins := kit.defsFacts[defname]; dins != nil {
					dins.overloadDrop(defid)
				}
			}
		}
		if len(kit.state.defsBornIDsNames) > 0 {
			for defid, defname := range kit.state.defsBornIDsNames {
				if reSubstFirst[defid] = kit; me.state.kitsReprocessing.ever {
					namesofchange[defname] = true
				}
				if dans := kit.defsFacts[defname]; dans != nil && dans.overloadById(defid) != nil {
					panic(defid) // to see if this ever occurs
				}
			}
		}
		kit.state.defsGoneIDsNames, kit.state.defsBornIDsNames = nil, nil
	}
	if me.Kits.all.collectReferencers(namesofchange, reSubstNext, true); len(reSubstNext) > 0 {
		println("DEPS of " + fmt.Sprintf("%#v", namesofchange) + ":")
		for defid, kit := range reSubstNext {
			println("\t", kit.ImpPath, defid, kit.lookups.tlDefsByID[defid].Name.Val)
			if tld := kit.lookups.tlDefsByID[defid]; tld != nil {
				if dnf := kit.defsFacts[tld.Name.Val]; dnf != nil {
					dnf.overloadDrop(defid)
				}
			}
		}
	}
	var errs []error
	for defid, kit := range reSubstFirst {
		if errors := me.substantiateKitTopLevelDefFacts(kit, defid, true).Errs(); len(errors) > 0 {
			for _, e := range errors {
				if !e.IsRef() {
					errs = append(errs, e)
				}
			}
		}
	}
	for defid, kit := range reSubstNext {
		if errors := me.substantiateKitTopLevelDefFacts(kit, defid, false).Errs(); len(errors) > 0 {
			for _, e := range errors {
				if !e.IsRef() {
					errs = append(errs, e)
				}
			}
		}
	}
	me.onErrs(errs, nil)
}

func (me *Ctx) substantiateKitTopLevelDefFacts(kit *Kit, defId string, forceResubst bool) (dol *defIdFacts) {
	if kit == nil {
		return nil
	}
	def := kit.lookups.tlDefsByID[defId]
	if def == nil {
		return nil
	}
	facts := kit.defsFacts[def.Name.Val]
	if facts == nil {
		facts = &defNameFacts{}
		kit.defsFacts[def.Name.Val] = facts
	}
	if forceResubst {
		facts.overloadDrop(defId)
	}
	if dol = facts.overloadById(defId); dol == nil {
		dol = &defIdFacts{def: def, cache: make(map[*atmolang_irfun.AstDef]*valFacts)}
		facts.overloads = append(facts.overloads, dol)
		var finish defValFinisher
		if dol.valFacts, finish = me.substantiateFactsForDef(kit, dol, &def.AstDef); finish != nil {
			finish(kit, dol, &def.AstDef)
		}
	}
	return
}

func (me *Ctx) substantiateFactsForDef(kit *Kit, tld *defIdFacts, fullArgsScope ...*atmolang_irfun.AstDef) (ret valFacts, finish defValFinisher) {
	if def := fullArgsScope[len(fullArgsScope)-1]; def.Arg == nil {
		ret = me.substantiateFactsForExpr(kit, def.Body, tld, fullArgsScope...)
	} else {
		vfc := &valFactCallable{arg: valFactCallableArg{orig: def.Arg}}
		ret, finish = valFacts{vfc}, func(k *Kit, t *defIdFacts, d *atmolang_irfun.AstDef) {
			vfc.ret = me.substantiateFactsForExpr(k, d.Body, t, fullArgsScope...)
		}
	}
	return
}

func (me *Ctx) substantiateFactsForExpr(kit *Kit, astExpr atmolang_irfun.IAstExpr, tld *defIdFacts, fullArgsScope ...*atmolang_irfun.AstDef) (findings valFacts) {
	if xld, _ := astExpr.(atmolang_irfun.IAstExprWithLetDefs); xld != nil {
		if lds := xld.LetDefs(); len(lds) > 0 {
			nu := make([]*atmolang_irfun.AstDef, 0, len(lds))
			for i := range lds {
				ld := &lds[i]
				if _, ok := tld.cache[ld]; !ok {
					tld.cache[ld] = &valFacts{}
					nu = append(nu, ld)
				}
			}
			for _, ld := range nu {
				deffacts, finish := me.substantiateFactsForDef(kit, tld, append(fullArgsScope, ld)...)
				if tld.cache[ld].add(deffacts...); finish != nil {
					finish(kit, tld, ld)
				}
			}
		}
	}
	switch expr := astExpr.(type) {
	case *atmolang_irfun.AstLitFloat:
		findings.add(&valFactPrim{orig: expr, value: expr.Val})
	case *atmolang_irfun.AstLitRune:
		findings.add(&valFactPrim{orig: expr, value: expr.Val})
	case *atmolang_irfun.AstLitStr:
		findings.add(&valFactPrim{orig: expr, value: expr.Val})
	case *atmolang_irfun.AstLitUint:
		findings.add(&valFactPrim{orig: expr, value: expr.Val})
	case *atmolang_irfun.AstIdentTag:
		findings.add(&valFactPrim{orig: expr, value: expr.Val})
	case *atmolang_irfun.AstIdentName:
		findings = me.substantiateFactsForExprIdentName(kit, expr, tld, fullArgsScope...)
	case *atmolang_irfun.AstAppl:
	default:
		panic(expr)
	}
	return
}

func (me *Ctx) substantiateFactsForExprAppl(kit *Kit, expr *atmolang_irfun.AstAppl, tld *defIdFacts, fullArgsScope ...*atmolang_irfun.AstDef) (findings valFacts) {
	cfacts := me.substantiateFactsForExpr(kit, expr.AtomicCallee, tld, fullArgsScope...)
	if cerrs := cfacts.errs(false); cerrs != nil && len(cerrs.Errors) > 0 {
		cfacts.errs(true).Add(cerrs.Errors.Refs())
	}
	afacts := me.substantiateFactsForExpr(kit, expr.AtomicArg, tld, fullArgsScope...)
	if aerrs := afacts.errs(false); aerrs != nil && len(aerrs.Errors) > 0 {
		afacts.errs(true).Add(aerrs.Errors.Refs())
	}
	callable := cfacts.callable(false)
	if callable == nil {
		findings.errs(true).AddSubst(expr.AtomicCallee.OrigToks().First(nil), "not callable: "+cfacts.String())
	} else {
	}
	return
}

func (me *Ctx) substantiateFactsForExprIdentName(kit *Kit, expr *atmolang_irfun.AstIdentName, tld *defIdFacts, fullArgsScope ...*atmolang_irfun.AstDef) (findings valFacts) {
	switch candidates := expr.NamesInScope[expr.Val]; len(candidates) {

	case 0: // uncomplicated fail: name unknown / not found in scope
		i, namesinscope := 0, make([]string, len(expr.NamesInScope))
		for k := range expr.NamesInScope {
			i, namesinscope[i] = i+1, k
		}
		namesinscope = ustr.Similes(expr.Val, namesinscope...)
		findings.errs(true).AddNaming(&expr.Orig.Tokens[0], "unknown: `"+expr.Val+ustr.If(len(namesinscope) == 0, "`", "` (did you mean `"+ustr.Join(namesinscope, "` or `")+"`?)"))

	case 1: // uncomplicated best-case: 1 name-match exactly
		switch cand := candidates[0].(type) {
		case *atmolang_irfun.AstDef:
			if dol := tld.cache[cand]; dol != nil {
				findings.add(&valFactRef{valFacts: dol})
			} else {
				panic(cand)
			}
		case *atmolang_irfun.AstDefTop:
			if dol := me.substantiateKitTopLevelDefFacts(kit, cand.Id, false); dol != nil {
				findings.add(&valFactRef{valFacts: &dol.valFacts})
			} else {
				panic(cand.Id)
			}
		case astDefRef:
			if dol := me.substantiateKitTopLevelDefFacts(me.Kits.all.ByImpPath(cand.kit), cand.Id, false); dol != nil {
				findings.add(&valFactRef{valFacts: &dol.valFacts})
			} else {
				panic(cand.kit + "@" + cand.Id)
			}
		case *atmolang_irfun.AstDefArg:
			var argdef *atmolang_irfun.AstDef
			istopleveldefarg := (tld.def.Arg == cand)
			if istopleveldefarg {
				argdef = &tld.def.AstDef
			} else {
				for i := len(fullArgsScope) - 1; i >= 0; i, argdef = i-1, nil {
					if argdef = fullArgsScope[i]; argdef.Arg == cand {
						break
					}
				}
			}
			if argdef == nil {
				panic(cand)
			} else if istopleveldefarg {
				findings.add(&valFactArgRef{tld.callable(false)})
			} else {
				findings.add(&valFactArgRef{tld.cache[argdef].callable(false)})
			}
		default:
			panic(cand)
		}

	default: // >1 candidates name-match our ‹curExprIdent›
		var argless []atmolang_irfun.IAstNode // verify all are argful to qualify
		for _, cand := range candidates {
			if !cand.IsDefWithArg() {
				argless = append(argless, cand)
			}
		}
		if len(argless) == 0 { // fine: but still-todo
			findings.errs(true).AddTodo(&expr.Orig.Tokens[0], "resolve overloads")

		} else { // fail: multiple ‹curExprIdent›-named arg-less defs and/or args-in-scope shadow each other and/or same-named known argful defs, that's illegal
			errmsg := "ambiguous: `" + expr.Val + "`, competing candidates from "
			{ // BEGIN: lame boilerplate to hint in err-msg which deps are involved in duplicate name
				locs, exts, candkits := false, false, make(map[string]int, len(argless))
				for _, cand := range argless {
					if nx, ok := cand.(astDefRef); ok && nx.AstDefTop != nil && nx.kit != "" {
						exts, candkits[nx.kit] = true, 1+candkits[nx.kit]
					} else {
						locs, candkits[kit.ImpPath] = true, 1+candkits[kit.ImpPath]
					}
				}
				for k, v := range candkits {
					errmsg += k + " (" + ustr.Int(v) + "×) ─── "
				}
				if locs {
					if errmsg += "rename culprits in " + kit.ImpPath; exts {
						errmsg += " and/or "
					}
				}
				if exts {
					errmsg += "qualify which import was meant"
				}
			} // END: lame boilerplate
			findings.errs(true).AddNaming(&expr.Orig.Tokens[0], errmsg)
		}
	}
	return
}
