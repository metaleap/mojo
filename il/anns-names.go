package atmoil

import (
	"github.com/metaleap/atmo"
)

func (me AnnNamesInScope) Add(name string, nodes ...IIrNode) {
	me[name] = append(me[name], nodes...)
}

func (me AnnNamesInScope) copy() (fullCopy AnnNamesInScope) {
	fullCopy = make(AnnNamesInScope, len(me))
	for k, v := range me {
		fullCopy[k] = v
	}
	return
}

func (me AnnNamesInScope) copyAndAdd(tld *IrDef, addArg *IrArg, errs *atmo.Errors) (namesInScopeCopy AnnNamesInScope) {
	argname := addArg.IrIdentBase.Val
	if cands := me[argname]; len(cands) > 0 {
		for _, cand := range cands {
			if !cand.IsExt() {
				me.errNameWouldShadow(tld, errs, addArg, argname)
				return me
			}
		}
	}
	namesInScopeCopy = make(AnnNamesInScope, len(me)+1)
	// copy old names:
	for name, nodes := range me {
		if name != argname {
			namesInScopeCopy[name] = nodes // safe to keep existing slice as-is
		} else {
			namesInScopeCopy.Add(name, nodes...) // effectively copy existing slice
		}
	}
	namesInScopeCopy.Add(argname, addArg)
	return
}

func (me AnnNamesInScope) RepopulateDefsAndIdentsFor(tld *IrDef, node IIrNode, currentlyErroneousButKnownGlobalsNames map[string]int) (errs atmo.Errors) {
	switch n := node.(type) {
	case *IrDef:
		errs.Add(me.RepopulateDefsAndIdentsFor(tld, n.Body, currentlyErroneousButKnownGlobalsNames)...)
	case *IrLam:
		errs.Add(me.copyAndAdd(tld, &n.Arg, &errs).RepopulateDefsAndIdentsFor(tld, n.Body, currentlyErroneousButKnownGlobalsNames)...)
	case *IrAppl:
		errs.Add(me.RepopulateDefsAndIdentsFor(tld, n.Callee, currentlyErroneousButKnownGlobalsNames)...)
		errs.Add(me.RepopulateDefsAndIdentsFor(tld, n.CallArg, currentlyErroneousButKnownGlobalsNames)...)
	case *IrIdentName:
		if existsunparsed := currentlyErroneousButKnownGlobalsNames[n.Val]; existsunparsed > 0 {
			errs.AddUnreach(ErrNames_IdentRefersToMalformedDef, tld.OrigToks(n), "`"+n.Val+"` found but with syntax errors")
		} else {
			n.Anns.Candidates = me[n.Val]
		}
	}
	return
}

func (AnnNamesInScope) errNameWouldShadow(maybeTld *IrDef, errs *atmo.Errors, node *IrArg, name string) {
	toks := node.origToks()
	if len(toks) == 0 && maybeTld != nil {
		toks = maybeTld.OrigToks(node)
	}
	errs.AddNaming(ErrNames_ShadowingNotAllowed, toks.First1(), "name `"+name+"` already defined (rename required)")
}
