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

func (me AnnNamesInScope) copyAndAdd(tld *IrDef, addArg *IrAbs, errs *Errors) (namesInScopeCopy AnnNamesInScope) {
	argname := addArg.Arg.Val
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
	case *IrAbs:
		errs.Add(me.copyAndAdd(tld, n, &errs).RepopulateDefsAndIdentsFor(tld, n.Body, currentlyErroneousButKnownGlobalsNames, append(nodeAncestors, n)...)...)
	case *IrAppl:
		errs.Add(me.RepopulateDefsAndIdentsFor(tld, n.Callee, currentlyErroneousButKnownGlobalsNames, append(nodeAncestors, n)...)...)
		errs.Add(me.RepopulateDefsAndIdentsFor(tld, n.CallArg, currentlyErroneousButKnownGlobalsNames, append(nodeAncestors, n)...)...)
	case *IrIdentName:
		if _, existsunparsed := currentlyErroneousButKnownGlobalsNames[n.Val]; existsunparsed {
			errs.AddUnreach(ErrNames_IdentRefersToMalformedDef, tld.AstOrigToks(n), "`"+n.Val+"` found but with syntax errors")
		} else {
			n.Anns.Candidates = me[n.Val]
			if len(n.Anns.Candidates) == 1 {
				if abs, isabs := n.Anns.Candidates[0].(*IrAbs); isabs {
					if n.Anns.AbsIdx = abs.Anns.AbsIdx; n.Anns.AbsIdx < 0 {
						n.Anns.AbsIdx = 0
					}

					n.Anns.ArgIdx = 0
					var found bool
					for i := len(nodeAncestors) - 1; i != 0; i-- {
						if lam, is := nodeAncestors[i].(*IrAbs); is {
							n.Anns.ArgIdx++
							if found = (abs == lam); found {
								break
							}
						}
					}
				}
			}
		}
	}
	return
}
