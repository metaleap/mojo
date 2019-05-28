package atmosess

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang/irfun"
)

type astDefRef struct {
	*atmolang_irfun.AstDefTop
	kit string
}

func (me *Ctx) kitsRepopulateAstNamesInScope() (errs atmo.Errors) {
	kitrepops := make(map[*Kit]atmo.Empty, len(me.Kits.all))

	{ // FIRST: namesInScopeOwn
		for _, kit := range me.Kits.all {
			if len(kit.state.defsGoneIdsNames) > 0 || len(kit.state.defsBornIdsNames) > 0 {
				kit.Errs.Stage1BadNames = kit.Errs.Stage1BadNames[0:0]
				kitrepops[kit], kit.lookups.namesInScopeAll, kit.lookups.namesInScopeOwn =
					atmo.Exists, nil, make(atmolang_irfun.AnnNamesInScope, len(kit.topLevelDefs))
				for _, tld := range kit.topLevelDefs {
					kit.lookups.namesInScopeOwn.Add(&kit.Errs.Stage1BadNames, tld.Name.Val, tld)
				}
			}
		}
	}
	{ // NEXT: namesInScopeExt (utilizing above namesInScopeOwn adds hence separate Ctx.Kits.all loop)
		for _, kit := range me.Kits.all {
			if len(kit.Imports) > 0 {
				var totaldefscount int
				var anychangesinkimps bool
				kimps := me.Kits.all.Where(func(k *Kit) (iskimp bool) {
					if iskimp = ustr.In(k.ImpPath, kit.Imports...); iskimp {
						totaldefscount, anychangesinkimps = totaldefscount+len(k.topLevelDefs), anychangesinkimps || len(k.state.defsGoneIdsNames) > 0 || len(k.state.defsBornIdsNames) > 0
					}
					return
				})
				if anychangesinkimps {
					if _, ok := kitrepops[kit]; !ok {
						kit.Errs.Stage1BadNames = kit.Errs.Stage1BadNames[0:0]
					}
					kitrepops[kit], kit.lookups.namesInScopeAll, kit.lookups.namesInScopeExt =
						atmo.Exists, nil, make(atmolang_irfun.AnnNamesInScope, totaldefscount)
					for _, kimp := range kimps {
						for k, v := range kimp.lookups.namesInScopeOwn {
							nodes := make([]atmolang_irfun.IAstNode, len(v))
							for i, n := range v {
								nodes[i] = astDefRef{kit: kimp.ImpPath,
									AstDefTop: n.(*atmolang_irfun.AstDefTop) /* ok to panic here bc should-never-happen-else-its-a-bug */}
							}
							kit.lookups.namesInScopeExt.Add(&kit.Errs.Stage1BadNames, k, nodes...)
						}
					}
				}
			}
		}
	}

	for kit := range kitrepops {
		kit.lookups.namesInScopeAll = make(atmolang_irfun.AnnNamesInScope, len(kit.lookups.namesInScopeExt)+len(kit.lookups.namesInScopeOwn))
		for k, v := range kit.lookups.namesInScopeOwn {
			nodes := make([]atmolang_irfun.IAstNode, len(v))
			copy(nodes, v)
			kit.lookups.namesInScopeAll[k] = nodes
		}
		for k, v := range kit.lookups.namesInScopeExt {
			kit.lookups.namesInScopeAll[k] = append(kit.lookups.namesInScopeAll[k], v...)
		}
		for _, tld := range kit.topLevelDefs {
			kit.Errs.Stage1BadNames.Add(kit.lookups.namesInScopeAll.RepopulateAstDefsAndIdentsFor(&tld.AstDef))
		}
		errs.Add(kit.Errs.Stage1BadNames)
	}

	return
}
