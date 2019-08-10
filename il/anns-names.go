package atmoil

import (
	. "github.com/metaleap/atmo"
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

func (me AnnNamesInScope) copyAndAdd(tld *IrDef, addArg *IrArg, errs *Errors) (namesInScopeCopy AnnNamesInScope) {
	argname := addArg.IrIdentBase.Val
	namesInScopeCopy = make(AnnNamesInScope, len(me)+1)
	for name, nodes := range me {
		if name != argname {
			namesInScopeCopy[name] = nodes
		}
	}
	namesInScopeCopy[argname] = []IIrNode{addArg}
	return
}

func (me AnnNamesInScope) RepopulateDefsAndIdentsFor(tld *IrDef, node IIrNode, currentlyErroneousButKnownGlobalsNames StringKeys, nodeAncestors ...IIrNode) (errs Errors) {
	switch n := node.(type) {
	case *IrDef:
		errs.Add(me.RepopulateDefsAndIdentsFor(tld, n.Body, currentlyErroneousButKnownGlobalsNames, append(nodeAncestors, n)...)...)
	case *IrLam:
		errs.Add(me.copyAndAdd(tld, &n.Arg, &errs).RepopulateDefsAndIdentsFor(tld, n.Body, currentlyErroneousButKnownGlobalsNames, append(nodeAncestors, n)...)...)
	case *IrAppl:
		errs.Add(me.RepopulateDefsAndIdentsFor(tld, n.Callee, currentlyErroneousButKnownGlobalsNames, append(nodeAncestors, n)...)...)
		errs.Add(me.RepopulateDefsAndIdentsFor(tld, n.CallArg, currentlyErroneousButKnownGlobalsNames, append(nodeAncestors, n)...)...)
	case *IrIdentName:
		if _, existsunparsed := currentlyErroneousButKnownGlobalsNames[n.Val]; existsunparsed {
			errs.AddUnreach(ErrNames_IdentRefersToMalformedDef, tld.AstOrigToks(n), "`"+n.Val+"` found but with syntax errors")
		} else {
			n.Anns.Candidates = me[n.Val]
			if len(n.Anns.Candidates) == 1 {
				if arg, isarg := n.Anns.Candidates[0].(*IrArg); isarg {
					n.Anns.ArgIdx = arg.Anns.LamIdx
					break
				}
			}
		}
	}
	return
}
