package atmoil

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
)

// Add does not validate, merely appends as a convenience short-hand notation.
// Outside-package callers (ie. `atmosess` pkg) only use it for adding names of
// (imported or kit-owned) top-level defs which cannot be rejected (unlike locals / args).
func (me AnnNamesInScope) Add(name string, nodes ...INode) {
	me[name] = append(me[name], nodes...)
}

func (me AnnNamesInScope) copyAndAdd(tld *IrDefTop, add interface{}, errs *atmo.Errors) (namesInScopeCopy AnnNamesInScope) {
	var addarg *IrDefArg
	var adddefs IrDefs
	var namestoadd []string
	var numerrs int
	switch addwhat := add.(type) {
	case *IrDefArg:
		addarg, namestoadd = addwhat, []string{addwhat.IrIdentBase.Val}
		if cands := me[namestoadd[0]]; len(cands) > 0 {
			me.errDuplName(tld, errs, addarg, namestoadd[0])
			numerrs++
		}
	case IrDefs:
		adddefs, namestoadd = addwhat, make([]string, len(addwhat))
		for i := range adddefs {
			namestoadd[i] = adddefs[i].Name.Val
			if cands := me[namestoadd[i]]; len(cands) > 0 {
				for _, c := range cands {
					if c.IsDef() == nil {
						me.errDuplName(tld, errs, &adddefs[i], namestoadd[i])
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

func (me AnnNamesInScope) RepopulateDefsAndIdentsFor(tld *IrDefTop, node INode, currentlyErroneousButKnownGlobalsNames map[string]int) (errs atmo.Errors) {
	inscope := me
	if let := node.Let(); let != nil {
		if len(let.Defs) > 0 {
			inscope = inscope.copyAndAdd(tld, let.Defs, &errs)
			for i := range let.Defs {
				errs.Add(inscope.RepopulateDefsAndIdentsFor(tld, &let.Defs[i], currentlyErroneousButKnownGlobalsNames)...)
			}
		}
		let.Anns.NamesInScope = inscope
	}
	switch n := node.(type) {
	case *IrDef:
		if n.Arg != nil {
			inscope = inscope.copyAndAdd(tld, n.Arg, &errs)
		}
		errs.Add(inscope.RepopulateDefsAndIdentsFor(tld, n.Body, currentlyErroneousButKnownGlobalsNames)...)
	case *IrAppl:
		errs.Add(inscope.RepopulateDefsAndIdentsFor(tld, n.AtomicCallee, currentlyErroneousButKnownGlobalsNames)...)
		errs.Add(inscope.RepopulateDefsAndIdentsFor(tld, n.AtomicArg, currentlyErroneousButKnownGlobalsNames)...)
	case *IrIdentName:
		if existsunparsed := currentlyErroneousButKnownGlobalsNames[n.Val]; existsunparsed > 0 {
			errs.AddUnreach(ErrNames_IdentRefersToMalformedDef, tld.OrigToks(n), "`"+n.Val+"` found but with syntax errors")
		} else if n.Anns.Candidates = inscope[n.Val]; len(n.Anns.Candidates) == 0 {
			me.errUnknownName(tld, &errs, n)
		}
	}
	return
}

func (AnnNamesInScope) errDuplName(maybeTld *IrDefTop, errs *atmo.Errors, node INode, name string) {
	toks := node.origToks()
	if def := node.IsDef(); def != nil {
		if t := def.Name.origToks(); len(t) > 0 {
			toks = t
		}
	}
	if len(toks) == 0 && maybeTld != nil {
		toks = maybeTld.OrigToks(node)
	}
	errs.AddNaming(ErrNames_ShadowingNotAllowed, toks.First1(), "name `"+name+"` already in scope (rename required)")
}

func (AnnNamesInScope) errUnknownName(tld *IrDefTop, errs *atmo.Errors, n *IrIdentName) {
	errs.AddNaming(ErrNames_NotKnownInCurScope, tld.OrigToks(n).First1(), "name `"+n.Val+"` not in scope (possible typo or missing import?)")
}
