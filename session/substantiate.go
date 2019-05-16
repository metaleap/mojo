package atmosess

import (
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
	if len(me) == 1 {
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
func (me *valFactErrs) String() string    { return ustr.Join(me.Errors.Strings(), "\n") }

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

type valFactCallable struct {
	arg valFactCallableArg
	ret valFacts
}

func (me *valFactCallable) Errs() (errs atmo.Errors) {
	errs.Add(me.arg.Errs())
	errs.Add(me.ret.Errs())
	return
}
func (me *valFactCallable) String() string { return me.arg.String() + " -> " + me.ret.String() }

type valFactRef struct {
	*valFacts
}

func (me *valFactRef) Errs() atmo.Errors { return me.valFacts.Errs().Refs() }

type valFactArgRef struct {
	*valFactCallable
}

func (me *valFactArgRef) Errs() atmo.Errors { return me.valFactCallable.arg.Errs().Refs() }
func (me *valFactArgRef) String() string    { return me.valFactCallable.arg.String() }

type defValFinisher func(*Kit, *defIdFacts, *atmolang_irfun.AstDef)

type defNameFacts struct {
	overloads []*defIdFacts
}

func (me *defNameFacts) overloadByID(id string) *defIdFacts {
	for _, rc := range me.overloads {
		if rc.def.ID == id {
			return rc
		}
	}
	return nil
}

func (me *defNameFacts) dropOverload(id string) {
	for i, rc := range me.overloads {
		if rc.def.ID == id {
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
					if tldef := kit.lookups.tlDefsByID[defid]; tldef != nil && len(tldef.Errs) == 0 {
						if dans := kit.defsFacts[tldef.Name.Val]; dans != nil && dans.overloadByID(defid) != nil {
							panic(defid) // to see if this ever occurs
						}
						needsReSubst[defid] = kit
					}
				}
			}
			kit.state.defsGoneIDsNames, kit.state.defsNew = nil, nil
		}
		var errs []error
		for defid, kit := range needsReSubst {
			if errors := me.substantiateFactsIfNotAlready(kit, defid).Errs(); len(errors) > 0 {
				for _, e := range errors {
					if !e.IsRef() {
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
	candidates := expr.NamesInScope[expr.Val]
	switch len(candidates) {
	case 0:
		i, namesinscope := 0, make([]string, len(expr.NamesInScope))
		for k := range expr.NamesInScope {
			i, namesinscope[i] = i+1, k
		}
		namesinscope = ustr.Similes(expr.Val, namesinscope...)
		findings.errs(true).AddNaming(&expr.Orig.Tokens[0], "unknown: `"+expr.Val+ustr.If(len(namesinscope) == 0, "`", "` (did you mean `"+ustr.Join(namesinscope, "` or `")+"`?)"))
	case 1:
		switch cand := candidates[0].(type) {
		case *atmolang_irfun.AstDef:
			findings.add(&valFactRef{valFacts: tld.cache[cand]})
		case *atmolang_irfun.AstDefTop:
			findings.add(&valFactRef{valFacts: &me.substantiateFactsIfNotAlready(kit, cand.ID).valFacts})
		case astNodeExt:
			findings.add(&valFactRef{valFacts: &me.substantiateFactsIfNotAlready(me.Kits.all.ByImpPath(cand.kit), cand.ID).valFacts})
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
	default:
		var argless []atmolang_irfun.IAstNode
		for _, cand := range candidates {
			if !cand.IsDefWithArg() {
				argless = append(argless, cand)
			}
		}
		if len(argless) == 0 {
			findings.errs(true).AddTodo(&expr.Orig.Tokens[0], "resolve overloads")
		} else { // err: multiple arg-less defs and/or args-in-scope shadow each other, that's illegal
			var errmsghints string
			{ // BEGIN: lame boilerplate to hint in err-msg which kits are involved in duplicate name
				locs, exts, candkits := false, false, make(map[string]int, len(argless))
				for _, cand := range argless {
					if nx, ok := cand.(astNodeExt); ok && nx.AstDefTop != nil && nx.kit != "" {
						exts, candkits[nx.kit] = true, 1+candkits[nx.kit]
					} else {
						locs, candkits[kit.ImpPath] = true, 1+candkits[kit.ImpPath]
					}
				}
				for k, v := range candkits {
					errmsghints += k + " (" + ustr.Int(v) + "×) ─── "
				}
				if locs {
					errmsghints += "rename culprits in " + kit.ImpPath
				}
				if locs && exts {
					errmsghints += " and/or "
				}
				if exts {
					errmsghints += "qualify which import was meant"
				}
			} // END: lame boilerplate
			findings.errs(true).AddNaming(&expr.Orig.Tokens[0], "ambiguous: `"+expr.Val+"`, competing candidates from "+errmsghints)
		}
	}
	return
}
