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
	Cases []*defReducedCase
}

func (me *defReduced) caseByID(id string) *defReducedCase {
	for _, rc := range me.Cases {
		if rc.ID == id {
			return rc
		}
	}
	return nil
}

func (me *defReduced) dropCase(id string) {
	for i, rc := range me.Cases {
		if rc.ID == id {
			me.Cases = append(me.Cases[:i], me.Cases[i+1:]...)
			break
		}
	}
}

type defReducedCase struct {
	ID     string
	Err    *atmo.Error
	Args   []defReducedCaseArg
	Ret    defReducedCaseRet
	Result interface{}
}

type defReducedCaseArg struct {
	Name string
	Desc iDefReducedCaseVal
}

type defReducedCaseRet struct {
	Desc iDefReducedCaseVal
}

type iDefReducedCaseVal interface {
	String() string
}

type defReducedCasePrimType int

const (
	_ defReducedCasePrimType = iota
	rcRetTypeAtomLitFloat
	rcRetTypeAtomLitRune
	rcRetTypeAtomLitStr
	rcRetTypeAtomLitUint
	rcRetTypeAtomIdentTag
	// rcRetTypeFunc
)

type defReducedCaseValPrim struct {
	primType defReducedCasePrimType
	val      atmolang_irfun.IAstExpr
}

func (me *defReducedCaseValPrim) String() string {
	return "(always exactly prim-type " + ustr.Int(int(me.primType)) + " and value " + atmolang_irfun.DbgPrintToString(me.val) + ")"
}

type defReducedCaseValArg struct {
	name string
}

type defReducedCaseValAnd struct {
	vals []iDefReducedCaseVal
}

type defReducedCaseValOr struct {
	vals []iDefReducedCaseVal
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
						red.dropCase(defid)
					}
				}
			}
			if len(kit.state.defsNew) > 0 {
				for _, defid := range kit.state.defsNew {
					if tldef := kit.lookups.tlDefsByID[defid]; tldef != nil && len(tldef.Errors) == 0 {
						if red := kit.defsReduced[tldef.Name.Val]; red != nil {
							red.dropCase(defid)
						}
						needsReReducing[defid] = kit
					}
				}
			}
			kit.state.defsGoneIDsNames, kit.state.defsNew = nil, nil
		}
		var errs []error
		for defid, kit := range needsReReducing {
			if _, rc := me.reduceIfNotAlready(kit, defid); rc.Err != nil && !rc.Err.IsRef() {
				errs = append(errs, rc.Err)
			}
		}
		me.onErrs(errs, nil)
	}
}

func (me *Ctx) reduceIfNotAlready(kit *Kit, defId string) (red *defReduced, rc *defReducedCase) {
	def := kit.lookups.tlDefsByID[defId]
	if red = kit.defsReduced[def.Name.Val]; red == nil {
		red = &defReduced{}
		kit.defsReduced[def.Name.Val] = red
	}
	if rc = red.caseByID(defId); rc == nil {
		rc = &defReducedCase{ID: defId}
		red.Cases = append(red.Cases, rc)
		rc.Result, rc.Ret.Desc, rc.Err = me.reduceExpr(kit, def.Body)
	}
	return
}

func (me *Ctx) reduceExpr(kit *Kit, body atmolang_irfun.IAstExpr) (result interface{}, retDesc iDefReducedCaseVal, err *atmo.Error) {
	switch expr := body.(type) {
	case *atmolang_irfun.AstLitFloat:
		result, retDesc = expr, &defReducedCaseValPrim{val: expr, primType: rcRetTypeAtomLitFloat}
	case *atmolang_irfun.AstLitRune:
		result, retDesc = expr, &defReducedCaseValPrim{val: expr, primType: rcRetTypeAtomLitRune}
	case *atmolang_irfun.AstLitStr:
		result, retDesc = expr, &defReducedCaseValPrim{val: expr, primType: rcRetTypeAtomLitStr}
	case *atmolang_irfun.AstLitUint:
		result, retDesc = expr, &defReducedCaseValPrim{val: expr, primType: rcRetTypeAtomLitUint}
	case *atmolang_irfun.AstIdentTag:
		result, retDesc = expr, &defReducedCaseValPrim{val: expr, primType: rcRetTypeAtomIdentTag}
	case *atmolang_irfun.AstIdentName:
		result, retDesc, err = me.reduceExprIdentName(kit, expr)
	default:
		err = atmo.ErrTodo(expr.Origin().Toks().First(nil), fmt.Sprintf("%T", expr))
	}
	return
}

func (me *Ctx) reduceExprIdentName(kit *Kit, expr *atmolang_irfun.AstIdentName) (result interface{}, retDesc iDefReducedCaseVal, err *atmo.Error) {
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
			err = atmo.ErrTodo(&expr.Orig.Tokens[0], "resolve args")
		} else {
			def, _ := candidates[0].(*atmolang_irfun.AstDef)
			if def == nil {
				def = &candidates[0].(*atmolang_irfun.AstDefTop).AstDef
			}
			if def.Arg != nil {
				err = atmo.ErrTodo(&expr.Orig.Tokens[0], "resolve to func")
			} else if result, retDesc, err = me.reduceExpr(kit, def.Body); err != nil && !err.IsRef() {
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
