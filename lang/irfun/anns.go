package atmolang_irfun

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
)

type AnnNamesInScope map[string][]IAstNode

func (me AnnNamesInScope) Add(errs *atmo.Errors, k string, v ...IAstNode) {
	var nonargdef IAstNode
	for i, n := range v {
		if !n.IsDefWithArg() {
			if nonargdef = n; errs != nil && len(v) > 1 {
				var ic int
				if i == 0 {
					ic = 1
				}
				me.errDuplName(errs, n, k, ic, v...)
			} else {
				break
			}
		}
	}
	if existing := me[k]; len(existing) == 0 {
		me[k] = v
	} else {
		if errs != nil {
			if nonargdef != nil {
				me.errDuplName(errs, nonargdef, k, 0, existing...)
			} else {
				for i, n := range existing {
					if !n.IsDefWithArg() {
						me.errDuplName(errs, v[0], k, i, existing...)
					}
				}
			}
		}
		me[k] = append(existing, v...)
	}
}

func (me AnnNamesInScope) add(k string, v ...IAstNode) {
	me[k] = append(me[k], v...)
}

func (me AnnNamesInScope) copyAndAdd(add interface{}, errs *atmo.Errors) (namesInScopeCopy AnnNamesInScope) {
	var addarg *AstDefArg
	var adddefs AstDefs
	var namestoadd []string
	var numerrs int
	switch addwhat := add.(type) {
	case *AstDefArg:
		addarg, namestoadd = addwhat, []string{addwhat.AstIdentName.Val}
		if cands := me[namestoadd[0]]; len(cands) > 0 {
			me.errDuplName(errs, addarg, namestoadd[0], 0, cands[0])
			numerrs++
		}
	case AstDefs:
		adddefs, namestoadd = addwhat, make([]string, len(addwhat))
		for i := range adddefs {
			namestoadd[i] = adddefs[i].Name.Val
			if cands := me[namestoadd[i]]; len(cands) > 0 {
				for ic, c := range cands {
					if !(c.IsDefWithArg() && adddefs[i].Arg != nil) {
						me.errDuplName(errs, &adddefs[i], namestoadd[i], ic, cands...)
						numerrs, namestoadd[i] = numerrs+1, ""
						break
					}
				}
			}
		}
	default:
		panic(addwhat)
	}
	if numerrs == len(namestoadd) {
		return me
	}

	namesInScopeCopy = make(AnnNamesInScope, len(me)+len(namestoadd))
	// copy old names:
	for k, v := range me {
		if !ustr.In(k, namestoadd...) {
			namesInScopeCopy[k] = v // safe to keep existing slice as-is
		} else {
			namesInScopeCopy.add(k, v...) // effectively copy existing slice
		}
	}
	// add new names:
	if addarg != nil {
		k, v := addarg.AstIdentName.Val, addarg
		namesInScopeCopy.add(k, v)
	} else {
		for i := range adddefs {
			if namestoadd[i] != "" {
				k, v := adddefs[i].Name.Val, &adddefs[i]
				namesInScopeCopy.add(k, v)
			}
		}
	}
	return
}

func (me AnnNamesInScope) RepopulateAstDefsAndIdentsFor(node IAstNode) (errs atmo.Errors) {
	inscope := me
	if ldx, _ := node.(IAstExprWithLetDefs); ldx != nil {
		if lds := ldx.LetDefs(); len(lds) > 0 {
			inscope = inscope.copyAndAdd(lds, &errs)
			for i := range lds {
				errs.Add(inscope.RepopulateAstDefsAndIdentsFor(&lds[i]))
			}
		}
		ldx.astExprLetBase().Anns.NamesInScope = inscope
	}
	switch n := node.(type) {
	case *AstDef:
		if n.Arg != nil {
			inscope = inscope.copyAndAdd(n.Arg, &errs)
		}
		errs.Add(inscope.RepopulateAstDefsAndIdentsFor(n.Body))
	case *AstAppl:
		errs.Add(inscope.RepopulateAstDefsAndIdentsFor(n.AtomicCallee))
		errs.Add(inscope.RepopulateAstDefsAndIdentsFor(n.AtomicArg))
	case *AstIdentName:
		n.Anns.ResolvesTo = inscope[n.Val]
		if tok := n.OrigToks().First(nil); tok != nil {
			println(len(n.Anns.ResolvesTo), tok.Meta.Position.String())
		} else {
			println(len(n.Anns.ResolvesTo), n.Val)
		}
	}
	return
}

func (AnnNamesInScope) errDuplName(errs *atmo.Errors, n IAstNode, name string, cIdx int, cands ...IAstNode) {
	cand := cands[cIdx]
	ctoks := cand.OrigToks()
	if len(ctoks) == 0 && len(cands) > 1 {
		for _, c := range cands {
			if ctoks = c.OrigToks(); len(ctoks) > 0 {
				cand = c
				break
			}
		}
	}
	errs.AddNaming(n.OrigToks().First(nil), "name `"+name+"` already defined in "+ctoks.First(nil).Meta.Position.String())
}
