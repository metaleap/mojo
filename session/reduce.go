package atmosess

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang/irfun"
)

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

func (me *Ctx) reReduceAffectedIRsIfAnyKitsReloaded() {
	if me.state.someKitsReloaded {
		me.state.someKitsReloaded = false

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
			if _, rc := me.reduceIfNotAlready(kit, defid); rc.Err != nil {
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
		me.reduce(kit, def, red, rc)
	}
	return
}

func (me *Ctx) reduce(kit *Kit, defTop *atmolang_irfun.AstDefTop, red *defReduced, rc *defReducedCase) {
	rc.Result, rc.Ret.Desc, rc.Err = me.reduceExpr(kit, defTop.Body, &defTop.AstDef)
}

func (me *Ctx) reduceExpr(kit *Kit, body atmolang_irfun.IAstExpr, ancestors ...atmolang_irfun.IAstNode) (result interface{}, retDesc iDefReducedCaseVal, err *atmo.Error) {
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
		handledef := func(def *atmolang_irfun.AstDef) {
			if def.Arg == nil {
				result, retDesc, err = me.reduceExpr(kit, def.Body, expr)
			} else {
				panic("def with arg not yet impl")
			}
		}

		var found bool
		namesinscope := append(make([]string, 0, len(kit.lookups.allNames)+8), kit.lookups.allNames...)
		if def := expr.LetDef(expr.Val); def != nil { // thanks stupid golang interface-nilness double semantics...
			found = true
			handledef(def)
		} else {
			namesinscope = append(namesinscope, expr.Names()...)
		}
		if !found {
			for i := len(ancestors) - 1; i >= 0; i-- {
				if ald, _ := ancestors[i].(atmolang_irfun.IAstExprWithLetDefs); ald != nil {
					if def := ald.LetDef(expr.Val); def != nil {
						found = true
						handledef(def)
					} else {
						namesinscope = append(namesinscope, ald.Names()...)
					}
				} else if def, _ := ancestors[i].(*atmolang_irfun.AstDef); def != nil {
					if def.Name.Val == expr.Val {
						found = true
						handledef(def)
					} else if def.Arg != nil && def.Arg.Val == expr.Val {
						found = true
					} else {
						namesinscope = append(namesinscope, def.Name.Val)
					}
				}
				if found {
					break
				}
			}
		}
		if !found {
			if defids := kit.lookups.tlDefIDsByName[expr.Val]; len(defids) == 1 {
				if def := kit.lookups.tlDefsByID[defids[0]]; def != nil {
					found = true
					handledef(&def.AstDef)
				}
			} else if len(defids) > 1 {
				panic(expr.Val)
			}
		}
		if !found {
			namesinscope = ustr.Similes(expr.Val, namesinscope...)
			err = atmo.ErrNaming(&expr.Orig.Tokens[0], "unknown: "+expr.Val+ustr.If(len(namesinscope) == 0, "", " (did you mean `"+ustr.Join(namesinscope, "` or `")+"`?)"))
		}
	}
	return
}
