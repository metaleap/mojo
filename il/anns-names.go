package atmoil

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
)

type AnnNamesInScope map[string][]IAstNode

func (me AnnNamesInScope) Add(maybeTld *AstDefTop, errs *atmo.Errors, k string, v ...IAstNode) {
	if errs == nil && maybeTld != nil {
		errs = &maybeTld.Errs.Stage1BadNames
	}
	var nonargdef IAstNode
	for i, n := range v {
		if !n.IsDefWithArg() {
			if nonargdef = n; errs != nil && len(v) > 1 {
				var ic int
				if i == 0 {
					ic = 1
				}
				me.errDuplName(maybeTld, errs, n, k, ic, v...)
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
				me.errDuplName(maybeTld, errs, nonargdef, k, 0, existing...)
			} else {
				for i, n := range existing {
					if !n.IsDefWithArg() {
						me.errDuplName(maybeTld, errs, v[0], k, i, existing...)
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

func (me AnnNamesInScope) copyAndAdd(tld *AstDefTop, add interface{}, errs *atmo.Errors) (namesInScopeCopy AnnNamesInScope) {
	var addarg *AstDefArg
	var adddefs AstDefs
	var namestoadd []string
	var numerrs int
	switch addwhat := add.(type) {
	case *AstDefArg:
		addarg, namestoadd = addwhat, []string{addwhat.AstIdentBase.Val}
		if cands := me[namestoadd[0]]; len(cands) > 0 {
			me.errDuplName(tld, errs, addarg, namestoadd[0], 0, cands[0])
			numerrs++
		}
	case AstDefs:
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
	for k, v := range me {
		if !ustr.In(k, namestoadd...) {
			namesInScopeCopy[k] = v // safe to keep existing slice as-is
		} else {
			namesInScopeCopy.add(k, v...) // effectively copy existing slice
		}
	}
	// add new names:
	if addarg != nil {
		k, v := addarg.AstIdentBase.Val, addarg
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

func (me AnnNamesInScope) RepopulateAstDefsAndIdentsFor(tld *AstDefTop, node IAstNode, currentlyErroneousButKnownGlobalsNames atmo.StringKeys) (errs atmo.Errors) {
	inscope := me
	if let := node.Let(); let != nil {
		if len(let.Defs) > 0 {
			inscope = inscope.copyAndAdd(tld, let.Defs, &errs)
			for i := range let.Defs {
				errs.Add(inscope.RepopulateAstDefsAndIdentsFor(tld, &let.Defs[i], currentlyErroneousButKnownGlobalsNames))
			}
		}
		let.Anns.NamesInScope = inscope
	}
	switch n := node.(type) {
	case *AstDef:
		if n.Arg != nil {
			inscope = inscope.copyAndAdd(tld, n.Arg, &errs)
		}
		errs.Add(inscope.RepopulateAstDefsAndIdentsFor(tld, n.Body, currentlyErroneousButKnownGlobalsNames))
	case *AstAppl:
		errs.Add(inscope.RepopulateAstDefsAndIdentsFor(tld, n.AtomicCallee, currentlyErroneousButKnownGlobalsNames))
		errs.Add(inscope.RepopulateAstDefsAndIdentsFor(tld, n.AtomicArg, currentlyErroneousButKnownGlobalsNames))
	case *AstIdentName:
		if n.Anns.Candidates = inscope[n.Val]; len(n.Anns.Candidates) == 0 {
			_, existsthough := currentlyErroneousButKnownGlobalsNames[n.Val]
			me.errUnknownName(tld, &errs, n, existsthough)
		}
	}
	return
}

func (AnnNamesInScope) errDuplName(maybeTld *AstDefTop, errs *atmo.Errors, n IAstNode, name string, cIdx int, cands ...IAstNode) {
	toks := n.origToks()
	if len(toks) == 0 && maybeTld != nil {
		toks = maybeTld.OrigToks(n)
	}
	errs.AddNaming(toks.First(nil), "nullary name `"+name+"` already in scope (rename required)")
}

func (AnnNamesInScope) errUnknownName(tld *AstDefTop, errs *atmo.Errors, n *AstIdentName, currentlyErroneousButKnown bool) {
	tok := tld.OrigToks(n).First(nil)
	if currentlyErroneousButKnown {
		errs.AddUnreach(tok, "name `"+n.Val+"` has syntax errors", len(tok.Meta.Orig))
	} else {
		errs.AddNaming(tok, "name `"+n.Val+"` not in scope (possible typo or missing import?)")
	}
}
