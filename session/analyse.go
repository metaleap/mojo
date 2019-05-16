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

type defInferences struct {
	overloads []*defOverload
}

func (me *defInferences) overloadByID(id string) *defOverload {
	for _, rc := range me.overloads {
		if rc.ID == id {
			return rc
		}
	}
	return nil
}

func (me *defInferences) dropOverload(id string) {
	for i, rc := range me.overloads {
		if rc.ID == id {
			me.overloads = append(me.overloads[:i], me.overloads[i+1:]...)
			break
		}
	}
}

type defOverload struct {
	ID  string
	Err *atmo.Error
	Ret defOverloadRet
}

type defOverloadRet struct {
	Desc iDefValDesc
}

type defValDescs []iDefValDesc

func (me defValDescs) String() (s string) {
	for i, vd := range me {
		if i > 0 {
			s += " AND "
		}
		s += vd.String()
	}
	return
}

type iDefValDesc interface {
	String() string
}

type defValPrimType int

const (
	_ defValPrimType = iota
	rcRetTypeAtomLitFloat
	rcRetTypeAtomLitRune
	rcRetTypeAtomLitStr
	rcRetTypeAtomLitUint
	rcRetTypeAtomIdentTag
)

type defValPrim struct {
	primType defValPrimType
	val      atmolang_irfun.IAstExpr
}

func (me *defValPrim) String() string {
	return "always exactly prim-type " + ustr.Int(int(me.primType)) + " and value " + atmolang_irfun.DbgPrintToString(me.val)
}

type defValArgRef struct {
	name string
	// alwaysSatisfies defValDescs
}

func (me *defValArgRef) String() string {
	return "value of arg `" + me.name + "`" //, which always satisfies " + me.alwaysSatisfies.String()
}

type defValCallable struct {
	arg  *defValArgRef
	body iDefValDesc
}

func (me *defValCallable) String() string {
	return "callable that takes (" + me.arg.String() + ") and returns (" + me.body.String() + ")"
}

func (me *Ctx) reprocessAffectedIRsIfAnyKitsReloaded() {
	if me.state.someKitsNeedReprocessing {
		me.state.someKitsNeedReprocessing = false

		me.kitsRepopulateIdentNamesInScope()
		needsReInferring := make(map[string]*Kit, 32)

		for _, kit := range me.Kits.all {
			if len(kit.state.defsGoneIDsNames) > 0 {
				for defid, defname := range kit.state.defsGoneIDsNames {
					if dins := kit.defsInferences[defname]; dins != nil {
						dins.dropOverload(defid)
					}
				}
			}
			if len(kit.state.defsNew) > 0 {
				for _, defid := range kit.state.defsNew {
					if tldef := kit.lookups.tlDefsByID[defid]; tldef != nil && len(tldef.Errors) == 0 {
						if dans := kit.defsInferences[tldef.Name.Val]; dans != nil {
							dans.dropOverload(defid)
						}
						needsReInferring[defid] = kit
					}
				}
			}
			kit.state.defsGoneIDsNames, kit.state.defsNew = nil, nil
		}
		var errs []error
		for defid, kit := range needsReInferring {
			if _, rc := me.inferIfNotAlready(kit, defid); rc.Err != nil && !rc.Err.IsRef() {
				errs = append(errs, rc.Err)
			}
		}
		me.onErrs(errs, nil)
	}
}

func (me *Ctx) inferIfNotAlready(kit *Kit, defId string) (rda *defInferences, rdo *defOverload) {
	def := kit.lookups.tlDefsByID[defId]
	if rda = kit.defsInferences[def.Name.Val]; rda == nil {
		rda = &defInferences{}
		kit.defsInferences[def.Name.Val] = rda
	}
	if rdo = rda.overloadByID(defId); rdo == nil {
		rdo = &defOverload{ID: defId}
		rda.overloads = append(rda.overloads, rdo)
		rdo.Ret.Desc, rdo.Err = me.inferExpr(kit, def.Body)
	}
	return
}

func (me *Ctx) inferExpr(kit *Kit, body atmolang_irfun.IAstExpr) (retDesc iDefValDesc, err *atmo.Error) {
	switch expr := body.(type) {
	case *atmolang_irfun.AstLitFloat:
		retDesc = &defValPrim{val: expr, primType: rcRetTypeAtomLitFloat}
	case *atmolang_irfun.AstLitRune:
		retDesc = &defValPrim{val: expr, primType: rcRetTypeAtomLitRune}
	case *atmolang_irfun.AstLitStr:
		retDesc = &defValPrim{val: expr, primType: rcRetTypeAtomLitStr}
	case *atmolang_irfun.AstLitUint:
		retDesc = &defValPrim{val: expr, primType: rcRetTypeAtomLitUint}
	case *atmolang_irfun.AstIdentTag:
		retDesc = &defValPrim{val: expr, primType: rcRetTypeAtomIdentTag}
	case *atmolang_irfun.AstIdentName:
		retDesc, err = me.inferExprIdentName(kit, expr)
	case *atmolang_irfun.AstAppl:
		retDesc, err = me.inferExprAppl(kit, expr)
	default:
		err = atmo.ErrTodo(expr.Origin().Toks().First(nil), fmt.Sprintf("%T", expr))
	}
	return
}

func (me *Ctx) inferExprAppl(kit *Kit, expr *atmolang_irfun.AstAppl) (retDesc iDefValDesc, err *atmo.Error) {
	if callee, cerr := me.inferExpr(kit, expr.AtomicCallee); cerr != nil {
		err = atmo.ErrRef(cerr)
	} else if arg, aerr := me.inferExpr(kit, expr.AtomicArg); aerr != nil {
		err = atmo.ErrRef(aerr)
	} else if calleedef, _ := callee.(*defValCallable); calleedef == nil {
		err = atmo.ErrInfer(expr.AtomicCallee.Origin().Toks().First(nil), "not callable")
	} else {
		if arg == nil {
		}
		// if result, retDesc, err = me.reduceExpr(kit, calleedef.body); err != nil {
		// 	err = atmo.ErrRef(err)
		// }
	}
	return
}

func (me *Ctx) inferExprIdentName(kit *Kit, expr *atmolang_irfun.AstIdentName) (retDesc iDefValDesc, err *atmo.Error) {
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
			retDesc = &defValArgRef{name: maybearg.AstIdentName.Val}
		} else {
			def, _ := candidates[0].(*atmolang_irfun.AstDef)
			if def == nil {
				def = &candidates[0].(*atmolang_irfun.AstDefTop).AstDef
			}
			if def.Arg != nil {
				rvdbody, e := me.inferExpr(kit, def.Body)
				if e != nil {
					err = atmo.ErrRef(e)
				} else {
					retDesc = &defValCallable{arg: &defValArgRef{name: def.Arg.AstIdentName.Val}, body: rvdbody}
				}
			} else if retDesc, err = me.inferExpr(kit, def.Body); err != nil {
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
