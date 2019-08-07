package atmoil

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
)

func (me AnnNamesInScope) Add(name string, nodes ...IIrNode) {
	me[name] = append(me[name], nodes...)
}

func (me AnnNamesInScope) copyAndAdd(tld *IrDefTop, add interface{}, errs *atmo.Errors) (namesInScopeCopy AnnNamesInScope) {
	var addarg *IrArg
	var adddefs IrDefs
	var namestoadd []string
	var numerrs int
	switch addwhat := add.(type) {
	case *IrArg:
		addarg, namestoadd = addwhat, []string{addwhat.IrIdentBase.Val}
		if cands := me[namestoadd[0]]; len(cands) > 0 {
			for _, cand := range cands {
				if !cand.IsExt() {
					me.errNameWouldShadow(tld, errs, addarg, namestoadd[0])
					numerrs++
					break
				}
			}
		}
	case IrDefs:
		adddefs, namestoadd = addwhat, make([]string, len(addwhat))
		for i := range adddefs {
			namestoadd[i] = adddefs[i].Name.Val
			if cands := me[namestoadd[i]]; len(cands) > 0 {
				for _, cand := range cands {
					if cd := cand.IsDef(); cd == nil ||
						((!cd.IsExt()) && (cd.IsLam() == nil || adddefs[i].IsLam() == nil)) {
						me.errNameWouldShadow(tld, errs, &adddefs[i], namestoadd[i])
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
	for name, nodes := range me {
		if !ustr.In(name, namestoadd...) {
			namesInScopeCopy[name] = nodes // safe to keep existing slice as-is
		} else {
			namesInScopeCopy.Add(name, nodes...) // effectively copy existing slice
		}
	}
	// add new names:
	if addarg != nil {
		k, v := addarg.IrIdentBase.Val, addarg
		namesInScopeCopy.Add(k, v)
	} else {
		for i := range adddefs {
			if namestoadd[i] != "" {
				k, v := adddefs[i].Name.Val, &adddefs[i]
				namesInScopeCopy.Add(k, v)
			}
		}
	}
	return
}

func (me AnnNamesInScope) RepopulateDefsAndIdentsFor(tld *IrDefTop, node IIrNode, currentlyErroneousButKnownGlobalsNames map[string]int) (errs atmo.Errors) {
	inscope := me
	switch n := node.(type) {
	case *IrDef:
		if lam := n.IsLam(); lam != nil {
			inscope = inscope.copyAndAdd(tld, &lam.Arg, &errs)
		}
		errs.Add(inscope.RepopulateDefsAndIdentsFor(tld, n.Body, currentlyErroneousButKnownGlobalsNames)...)
	case *IrAppl:
		errs.Add(inscope.RepopulateDefsAndIdentsFor(tld, n.Callee, currentlyErroneousButKnownGlobalsNames)...)
		errs.Add(inscope.RepopulateDefsAndIdentsFor(tld, n.CallArg, currentlyErroneousButKnownGlobalsNames)...)
	case *IrIdentName:
		if existsunparsed := currentlyErroneousButKnownGlobalsNames[n.Val]; existsunparsed > 0 {
			errs.AddUnreach(ErrNames_IdentRefersToMalformedDef, tld.OrigToks(n), "`"+n.Val+"` found but with syntax errors")
		} else {
			n.Anns.Candidates = inscope[n.Val]
		}
	}
	return
}

func (AnnNamesInScope) errNameWouldShadow(maybeTld *IrDefTop, errs *atmo.Errors, node IIrNode, name string) {
	toks := node.origToks()
	if def := node.IsDef(); def != nil {
		if t := def.Name.origToks(); len(t) > 0 {
			toks = t
		}
	}
	if len(toks) == 0 && maybeTld != nil {
		toks = maybeTld.OrigToks(node)
	}
	errs.AddNaming(ErrNames_ShadowingNotAllowed, toks.First1(), "name `"+name+"` already defined (rename required)")
}
