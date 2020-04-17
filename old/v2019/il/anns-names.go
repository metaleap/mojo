package atmoil

import (
	. "github.com/metaleap/atmo/old/v2019"
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

func (me AnnNamesInScope) copyAndAdd(tld *IrDef, addArg *IrAbs) (namesInScopeCopy AnnNamesInScope) {
	existing := me[addArg.Arg.Name]
	sl := append(make([]IIrNode, 1, 1+len(existing)), existing...)
	sl[0], namesInScopeCopy = addArg, make(AnnNamesInScope, len(me)+1)
	for name, nodes := range me {
		namesInScopeCopy[name] = nodes
	}
	namesInScopeCopy[addArg.Arg.Name] = sl
	return
}

func (me AnnNamesInScope) RepopulateDefsAndIdentsFor(tld *IrDef, node IIrNode, currentlyErroneousButKnownGlobalsNames StringKeys, nodeAncestors ...IIrNode) (errs Errors) {
	switch n := node.(type) {
	case *IrDef:
		errs.Add(me.RepopulateDefsAndIdentsFor(tld, n.Body, currentlyErroneousButKnownGlobalsNames, append(nodeAncestors, n)...)...)
	case *IrAbs:
		errs.Add(me.copyAndAdd(tld, n).RepopulateDefsAndIdentsFor(tld, n.Body, currentlyErroneousButKnownGlobalsNames, append(nodeAncestors, n)...)...)
	case *IrAppl:
		errs.Add(me.RepopulateDefsAndIdentsFor(tld, n.Callee, currentlyErroneousButKnownGlobalsNames, append(nodeAncestors, n)...)...)
		errs.Add(me.RepopulateDefsAndIdentsFor(tld, n.CallArg, currentlyErroneousButKnownGlobalsNames, append(nodeAncestors, n)...)...)
	case *IrIdentName:
		if _, existsunparsed := currentlyErroneousButKnownGlobalsNames[n.Name]; existsunparsed {
			errs.AddUnreach(ErrNames_IdentRefersToMalformedDef, tld.AstOrigToks(n), "`"+n.Name+"` found but with syntax errors")
		} else {
			n.Ann.Candidates = me[n.Name]
			for i, cand := range n.Ann.Candidates {
				if abs, isabs := cand.(*IrAbs); isabs {
					n.Ann.Candidates[i] = &abs.Arg

					if n.Ann.AbsIdx = abs.Ann.AbsIdx; n.Ann.AbsIdx < 0 {
						n.Ann.AbsIdx = 0
					}

					n.Ann.ArgIdx = 0
					var found bool
					for i := len(nodeAncestors) - 1; i != 0; i-- {
						if lam, is := nodeAncestors[i].(*IrAbs); is {
							n.Ann.ArgIdx++
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
