package atmoil

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
)

type AnnNamesInScope map[string][]INode

func (me AnnNamesInScope) Add(maybeTld *IrDefTop, errs *atmo.Errors, name string, nodes ...INode) {
	if errs == nil && maybeTld != nil {
		errs = &maybeTld.Errs.Stage1BadNames
	}
	var nonargdef INode
	for i, node := range nodes {
		if !node.IsDefWithArg() {
			if nonargdef = node; errs != nil && len(nodes) > 1 {
				var ic int
				if i == 0 {
					ic = 1
				}
				me.errDuplName(maybeTld, errs, node, name, ic, nodes...)
			} else {
				break
			}
		}
	}
	if existing := me[name]; len(existing) == 0 {
		me[name] = nodes
	} else {
		if errs != nil {
			if nonargdef != nil {
				me.errDuplName(maybeTld, errs, nonargdef, name, 0, existing...)
			} else {
				for i, n := range existing {
					if !n.IsDefWithArg() {
						me.errDuplName(maybeTld, errs, nodes[0], name, i, existing...)
					}
				}
			}
		}
		me[name] = append(existing, nodes...)
	}
}

func (me AnnNamesInScope) add(name string, nodes ...INode) {
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
			me.errDuplName(tld, errs, addarg, namestoadd[0], 0, cands[0])
			numerrs++
		}
	case IrDefs:
		adddefs, namestoadd = addwhat, make([]string, len(addwhat))
		for i := range adddefs {
			namestoadd[i] = adddefs[i].Name.Val
			if cands := me[namestoadd[i]]; len(cands) > 0 {
				for ic, c := range cands {
					if !(c.IsDefWithArg() && adddefs[i].Arg != nil) {
						me.errDuplName(tld, errs, &adddefs[i], namestoadd[i], ic, cands...)
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
			namesInScopeCopy.add(name, nodes...) // effectively copy existing slice
		}
	}
	// add new names:
	if addarg != nil {
		k, v := addarg.IrIdentBase.Val, addarg
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

func (me AnnNamesInScope) RepopulateDefsAndIdentsFor(tld *IrDefTop, node INode, currentlyErroneousButKnownGlobalsNames atmo.StringCounts) (errs atmo.Errors) {
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
			errs.AddUnreach(ErrInit_IdentRefersToMalformedDef, tld.OrigToks(n), "syntax errors in "+ustr.Plu(existsunparsed, "def")+" named `"+n.Val+"`")
		} else if n.Anns.Candidates = inscope[n.Val]; len(n.Anns.Candidates) == 0 {
			me.errUnknownName(tld, &errs, n)
		}
	}
	return
}

func (AnnNamesInScope) errDuplName(maybeTld *IrDefTop, errs *atmo.Errors, node INode, name string, cIdx int, cands ...INode) {
	if name == "" {
		panic("WOT")
	}
	toks := node.origToks()
	if len(toks) == 0 && maybeTld != nil {
		toks = maybeTld.OrigToks(node)
	}
	errs.AddNaming(ErrNames_ShadowingNotAllowed, toks.First1(), "nullary name `"+name+"` already in scope (rename required)")
}

func (AnnNamesInScope) errUnknownName(tld *IrDefTop, errs *atmo.Errors, n *IrIdentName) {
	errs.AddNaming(ErrNames_UndefinedOrUnimported, tld.OrigToks(n).First1(), "name `"+n.Val+"` not in scope (possible typo or missing import?)")
}
