package atmosess

import (
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang/irfun"
)

func (me *Ctx) substantiateKitsDefsFactsAsNeeded() (errs []error) {
	reSubstFirst, reSubstNext := make(map[string]*Kit, 8), make(map[string]*Kit, 16)

	namesofchange := make(atmo.StringsUnorderedButUnique, 4)
	for _, kit := range me.Kits.all {
		for defid, defname := range kit.state.defsGoneIdsNames {
			namesofchange[defname] = atmo.Exists
			if defnamefacts := kit.defsFacts[defname]; defnamefacts != nil {
				defnamefacts.overloadDrop(defid)
			}
		}
		for defid, defname := range kit.state.defsBornIdsNames {
			reSubstFirst[defid] = kit
			namesofchange[defname] = atmo.Exists
			if defnamefacts := kit.defsFacts[defname]; defnamefacts != nil && defnamefacts.overloadById(defid) != nil {
				panic(defid) // tells us we have a bug in our housekeeping
			}
		}
		kit.state.defsGoneIdsNames, kit.state.defsBornIdsNames = nil, nil
	}
	me.Kits.all.collectReferencers(namesofchange, reSubstNext, true)
	for defid, kit := range reSubstFirst {
		errs = append(errs, me.substantiateKitTopLevelDefFacts(kit, defid, true).errsNonRef()...)
	}
	for defid, kit := range reSubstNext {
		errs = append(errs, me.substantiateKitTopLevelDefFacts(kit, defid, reSubstFirst[defid] != kit).errsNonRef()...)
	}
	return
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
	// switch candidates := expr.Anns.ResolvesTo; len(candidates) {

	// case 0: // uncomplicated fail: name unknown / not found in scope
	// 	i, namesinscope := 0, make([]string, len(expr.NamesInScope))
	// 	for k := range expr.NamesInScope {
	// 		i, namesinscope[i] = i+1, k
	// 	}
	// 	namesinscope = ustr.Similes(expr.Val, namesinscope...)
	// 	findings.errs(true).AddNaming(expr.Orig.Toks().First(nil), "unknown: `"+expr.Val+ustr.If(len(namesinscope) == 0, "`", "` (did you mean `"+ustr.Join(namesinscope, "` or `")+"`?)"))

	// case 1: // uncomplicated best-case: 1 name-match exactly
	// 	switch cand := candidates[0].(type) {
	// 	case *atmolang_irfun.AstDef:
	// 		if dol := tld.cache[cand]; dol != nil {
	// 			findings.add(&valFactRef{valFacts: dol})
	// 		} else {
	// 			panic(cand)
	// 		}
	// 	case *atmolang_irfun.AstDefTop:
	// 		if dol := me.substantiateKitTopLevelDefFacts(kit, cand.Id, false); dol != nil {
	// 			findings.add(&valFactRef{valFacts: &dol.valFacts})
	// 		} else {
	// 			panic(cand.Id)
	// 		}
	// 	case astDefRef:
	// 		if dol := me.substantiateKitTopLevelDefFacts(me.Kits.all.ByImpPath(cand.kit), cand.Id, false); dol != nil {
	// 			findings.add(&valFactRef{valFacts: &dol.valFacts})
	// 		} else {
	// 			panic(cand.kit + "@" + cand.Id)
	// 		}
	// 	case *atmolang_irfun.AstDefArg:
	// 		var argdef *atmolang_irfun.AstDef
	// 		istopleveldefarg := (tld.def.Arg == cand)
	// 		if istopleveldefarg {
	// 			argdef = &tld.def.AstDef
	// 		} else {
	// 			for i := len(fullArgsScope) - 1; i >= 0; i, argdef = i-1, nil {
	// 				if argdef = fullArgsScope[i]; argdef.Arg == cand {
	// 					break
	// 				}
	// 			}
	// 		}
	// 		if argdef == nil {
	// 			panic(cand)
	// 		} else if istopleveldefarg {
	// 			findings.add(&valFactArgRef{tld.callable(false)})
	// 		} else {
	// 			findings.add(&valFactArgRef{tld.cache[argdef].callable(false)})
	// 		}
	// 	default:
	// 		panic(cand)
	// 	}

	// default: // >1 candidates name-match our ‹curExprIdent›
	// 	var argless []atmolang_irfun.IAstNode // verify all are argful to qualify
	// 	for _, cand := range candidates {
	// 		if !cand.IsDefWithArg() {
	// 			argless = append(argless, cand)
	// 		}
	// 	}
	// 	if len(argless) == 0 { // fine: but still-todo
	// 		findings.errs(true).AddTodo(expr.Orig.Toks().First(nil), "resolve overloads")

	// 	} else { // fail: multiple ‹curExprIdent›-named arg-less defs and/or args-in-scope shadow each other and/or same-named known argful defs, that's illegal
	// 		errmsg := "ambiguous: `" + expr.Val + "`, competing candidates from "
	// 		{ // BEGIN: lame boilerplate to hint in err-msg which deps are involved in duplicate name
	// 			locs, exts, candkits := false, false, make(map[string]int, len(argless))
	// 			for _, cand := range argless {
	// 				if nx, ok := cand.(astDefRef); ok && nx.AstDefTop != nil && nx.kit != "" {
	// 					exts, candkits[nx.kit] = true, 1+candkits[nx.kit]
	// 				} else {
	// 					locs, candkits[kit.ImpPath] = true, 1+candkits[kit.ImpPath]
	// 				}
	// 			}
	// 			for k, v := range candkits {
	// 				errmsg += k + " (" + ustr.Int(v) + "×) ─── "
	// 			}
	// 			if locs {
	// 				if errmsg += "rename culprits in " + kit.ImpPath; exts {
	// 					errmsg += " and/or "
	// 				}
	// 			}
	// 			if exts {
	// 				errmsg += "qualify which import was meant"
	// 			}
	// 		} // END: lame boilerplate
	// 		findings.errs(true).AddNaming(expr.Orig.Toks().First(nil), errmsg)
	// 	}
	// }
	return
}
