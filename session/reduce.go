package atmosess

import (
	"fmt"

	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang/irfun"
)

type astNodeExt struct {
	atmolang_irfun.IAstNode
	kit string
}

type defReduced struct {
	overloads []*defOverload
}

func (me *defReduced) overloadByID(id string) *defOverload {
	for _, rc := range me.overloads {
		if rc.ID == id {
			return rc
		}
	}
	return nil
}

func (me *defReduced) dropOverload(id string) {
	for i, rc := range me.overloads {
		if rc.ID == id {
			me.overloads = append(me.overloads[:i], me.overloads[i+1:]...)
			break
		}
	}
}

type defOverload struct {
	ID     string
	Err    *atmo.Error
	Ret    defOverloadRet
	Result interface{}
}

type defOverloadRet struct {
	Desc iDefReducedValDesc
}

type defReducedValDescs []iDefReducedValDesc

func (me defReducedValDescs) String() (s string) {
	for i, vd := range me {
		if i > 0 {
			s += " AND "
		}
		s += vd.String()
	}
	return
}

type iDefReducedValDesc interface {
	// Satisfies(iDefReducedValDesc) bool
	String() string
}

type defReducedValPrimType int

const (
	_ defReducedValPrimType = iota
	rcRetTypeAtomLitFloat
	rcRetTypeAtomLitRune
	rcRetTypeAtomLitStr
	rcRetTypeAtomLitUint
	rcRetTypeAtomIdentTag
)

type defReducedValPrim struct {
	primType defReducedValPrimType
	val      atmolang_irfun.IAstExpr
}

func (me *defReducedValPrim) String() string {
	return "always exactly prim-type " + ustr.Int(int(me.primType)) + " and value " + atmolang_irfun.DbgPrintToString(me.val)
}

type defReducedValArgRef struct {
	name            string
	alwaysSatisfies defReducedValDescs
}

func (me *defReducedValArgRef) String() string {
	return "value of arg `" + me.name + "`, which always satisfies " + me.alwaysSatisfies.String()
}

type defReducedValDefRef struct {
	arg  *defReducedValArgRef
	body iDefReducedValDesc
}

func (me *defReducedValDefRef) String() string {
	return "callable that takes (" + me.arg.String() + ") and returns (" + me.body.String() + ")"
}

func (me *Ctx) reprocessAffectedIRsIfAnyKitsReloaded() {
	if me.state.someKitsNeedReprocessing {
		me.state.someKitsNeedReprocessing = false

		me.kitsRepopulateIdentNamesInScope()
		needsReReducing := make(map[string]*Kit, 32)

		for _, kit := range me.Kits.all {
			if len(kit.state.defsGoneIDsNames) > 0 {
				for defid, defname := range kit.state.defsGoneIDsNames {
					if red := kit.defsReduced[defname]; red != nil {
						red.dropOverload(defid)
					}
				}
			}
			if len(kit.state.defsNew) > 0 {
				for _, defid := range kit.state.defsNew {
					if tldef := kit.lookups.tlDefsByID[defid]; tldef != nil && len(tldef.Errors) == 0 {
						if red := kit.defsReduced[tldef.Name.Val]; red != nil {
							red.dropOverload(defid)
						}
						needsReReducing[defid] = kit
					}
				}
			}
			kit.state.defsGoneIDsNames, kit.state.defsNew = nil, nil
		}
		var errs []error
		for defid, kit := range needsReReducing {
			me.state.reduceArgs = map[string]iDefReducedValDesc{}
			if _, rc := me.reduceIfNotAlready(kit, defid); rc.Err != nil && !rc.Err.IsRef() {
				errs = append(errs, rc.Err)
			}
		}
		me.onErrs(errs, nil)
	}
}

func (me *Ctx) reduceIfNotAlready(kit *Kit, defId string) (red *defReduced, rc *defOverload) {
	def := kit.lookups.tlDefsByID[defId]
	if red = kit.defsReduced[def.Name.Val]; red == nil {
		red = &defReduced{}
		kit.defsReduced[def.Name.Val] = red
	}
	if rc = red.overloadByID(defId); rc == nil {
		rc = &defOverload{ID: defId}
		red.overloads = append(red.overloads, rc)
		rc.Result, rc.Ret.Desc, rc.Err = me.reduceExpr(kit, def.Body)
	}
	return
}

func (me *Ctx) reduceExpr(kit *Kit, body atmolang_irfun.IAstExpr) (result interface{}, retDesc iDefReducedValDesc, err *atmo.Error) {
	switch expr := body.(type) {
	case *atmolang_irfun.AstLitFloat:
		result, retDesc = expr, &defReducedValPrim{val: expr, primType: rcRetTypeAtomLitFloat}
	case *atmolang_irfun.AstLitRune:
		result, retDesc = expr, &defReducedValPrim{val: expr, primType: rcRetTypeAtomLitRune}
	case *atmolang_irfun.AstLitStr:
		result, retDesc = expr, &defReducedValPrim{val: expr, primType: rcRetTypeAtomLitStr}
	case *atmolang_irfun.AstLitUint:
		result, retDesc = expr, &defReducedValPrim{val: expr, primType: rcRetTypeAtomLitUint}
	case *atmolang_irfun.AstIdentTag:
		result, retDesc = expr, &defReducedValPrim{val: expr, primType: rcRetTypeAtomIdentTag}
	case *atmolang_irfun.AstIdentName:
		result, retDesc, err = me.reduceExprIdentName(kit, expr)
	case *atmolang_irfun.AstAppl:
		result, retDesc, err = me.reduceExprAppl(kit, expr)
	default:
		err = atmo.ErrTodo(expr.Origin().Toks().First(nil), fmt.Sprintf("%T", expr))
	}
	return
}

func (me *Ctx) reduceExprAppl(kit *Kit, expr *atmolang_irfun.AstAppl) (result interface{}, retDesc iDefReducedValDesc, err *atmo.Error) {
	if _, callee, cerr := me.reduceExpr(kit, expr.AtomicCallee); cerr != nil {
		err = atmo.ErrRef(cerr)
	} else if _, arg, aerr := me.reduceExpr(kit, expr.AtomicArg); aerr != nil {
		err = atmo.ErrRef(aerr)
	} else if calleedef, _ := callee.(*defReducedValDefRef); calleedef == nil {
		err = atmo.ErrReducing(expr.AtomicCallee.Origin().Toks().First(nil), "not callable")
	} else {
		me.state.reduceArgs[calleedef.arg.name] = arg
		panic(fmt.Sprintf("%T", calleedef.body))
		// if result, retDesc, err = me.reduceExpr(kit, calleedef.body); err != nil {
		// 	err = atmo.ErrRef(err)
		// }
	}
	return
}

func (me *Ctx) reduceExprIdentName(kit *Kit, expr *atmolang_irfun.AstIdentName) (result interface{}, retDesc iDefReducedValDesc, err *atmo.Error) {
	candidates := expr.NamesInScope[expr.Val]
	switch len(candidates) {
	case 0:
		idx, namesinscope := 0, make([]string, len(expr.NamesInScope))
		for k := range expr.NamesInScope {
			idx, namesinscope[idx] = idx+1, k
		}
		namesinscope = ustr.Similes(expr.Val, namesinscope...)
		err = atmo.ErrNaming(&expr.Orig.Tokens[0], "unknown: `"+expr.Val+ustr.If(len(namesinscope) == 0, "`", "` (did you mean `"+ustr.Join(namesinscope, "` or `")+"`?)"))
	case 1:
		if maybearg, _ := candidates[0].(*atmolang_irfun.AstDefArg); maybearg != nil {
			retDesc = &defReducedValArgRef{name: maybearg.AstIdentName.Val}
		} else {
			def, _ := candidates[0].(*atmolang_irfun.AstDef)
			if def == nil {
				def = &candidates[0].(*atmolang_irfun.AstDefTop).AstDef
			}
			if def.Arg != nil {
				_, rvdbody, e := me.reduceExpr(kit, def.Body)
				if e != nil {
					err = atmo.ErrRef(e)
				} else {
					retDesc = &defReducedValDefRef{arg: &defReducedValArgRef{name: def.Arg.AstIdentName.Val}, body: rvdbody}
				}
			} else if result, retDesc, err = me.reduceExpr(kit, def.Body); err != nil {
				err = atmo.ErrRef(err)
			}
		}
	default:
		var nonarity []atmolang_irfun.IAstNode
		for _, cand := range candidates {
			if !cand.IsDefWithArg() {
				nonarity = append(nonarity, cand)
			}
		}
		if len(nonarity) > 0 {
			locs, exts, candkits := false, false, make(map[string]int, len(nonarity))
			for _, cand := range nonarity {
				if nx, ok := cand.(astNodeExt); ok && nx.IAstNode != nil && nx.kit != "" {
					exts, candkits[nx.kit] = true, 1+candkits[nx.kit]
				} else {
					locs, candkits[kit.ImpPath] = true, 1+candkits[kit.ImpPath]
				}
			}
			var str string
			for k, v := range candkits {
				str += k + " (" + ustr.Int(v) + "×) ─── "
			}
			if locs {
				str += "rename culprits in " + kit.ImpPath
			}
			if locs && exts {
				str += " and/or "
			}
			if exts {
				str += "qualify which import was meant"
			}
			err = atmo.ErrNaming(&expr.Orig.Tokens[0], "ambiguous: `"+expr.Val+"`, competing candidates from "+str)
		}
		if err == nil {
			err = atmo.ErrTodo(&expr.Orig.Tokens[0], "resolve overloads")
		}
	}
	return
}
