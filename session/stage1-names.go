package atmosess

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang/irfun"
)

type AstDefRef struct {
	*atmolang_irfun.AstDefTop
	KitImpPath string
}

func (me *Ctx) kitsRepopulateAstNamesInScope() (namesOfChange atmo.StringsUnorderedButUnique, defIdsBorn map[string]*Kit, errs atmo.Errors) {
	kitrepops := make(map[*Kit]atmo.Exist, len(me.Kits.All))
	defIdsBorn, namesOfChange = make(map[string]*Kit, 2), make(atmo.StringsUnorderedButUnique, 4)

	{ // FIRST: namesInScopeOwn
		for _, kit := range me.Kits.All {
			if kit.WasEverToBeLoaded {
				if len(kit.state.defsGoneIdsNames) > 0 || len(kit.state.defsBornIdsNames) > 0 {
					for /*defid*/ _, defname := range kit.state.defsGoneIdsNames {
						namesOfChange[defname] = atmo.Є
						// if defnamefacts := kit.defsFacts[defname]; defnamefacts != nil {
						// 	defnamefacts.overloadDrop(defid)
						// }
					}
					for defid, defname := range kit.state.defsBornIdsNames {
						defIdsBorn[defid] = kit
						namesOfChange[defname] = atmo.Є
						// if defnamefacts := kit.defsFacts[defname]; defnamefacts != nil && defnamefacts.overloadById(defid) != nil {
						// 	panic(defid) // tells us we have a bug in our housekeeping
						// }
					}

					kitrepops[kit], kit.lookups.namesInScopeAll, kit.lookups.namesInScopeOwn =
						atmo.Є, nil, make(atmolang_irfun.AnnNamesInScope, len(kit.topLevelDefs))
					for _, tld := range kit.topLevelDefs {
						tld.Errs.Stage1BadNames = nil
						kit.lookups.namesInScopeOwn.Add(tld, &tld.Errs.Stage1BadNames, tld.Name.Val, tld)
					}
				}
			}
		}
	}
	{ // NEXT: namesInScopeExt (utilizing above namesInScopeOwn adds hence separate Ctx.Kits.all loop)
		for _, kit := range me.Kits.All {
			if kit.WasEverToBeLoaded {
				if len(kit.Imports) > 0 {
					var totaldefscount int
					var anychangesinkimps bool
					kimps := me.Kits.All.Where(func(k *Kit) (iskimp bool) {
						if iskimp = ustr.In(k.ImpPath, kit.Imports...); iskimp {
							totaldefscount, anychangesinkimps = totaldefscount+len(k.topLevelDefs), anychangesinkimps || len(k.state.defsGoneIdsNames) > 0 || len(k.state.defsBornIdsNames) > 0
						}
						return
					})
					if anychangesinkimps || len(kit.lookups.namesInScopeExt) == 0 {
						if _, alreadymarked := kitrepops[kit]; !alreadymarked {
							for _, tld := range kit.topLevelDefs {
								tld.Errs.Stage1BadNames = nil
							}
						}
						kitrepops[kit], kit.lookups.namesInScopeAll, kit.lookups.namesInScopeExt =
							atmo.Є, nil, make(atmolang_irfun.AnnNamesInScope, totaldefscount)
						for _, kimp := range kimps {
							for k, v := range kimp.lookups.namesInScopeOwn {
								nodes := make([]atmolang_irfun.IAstNode, len(v))
								for i, n := range v {
									nodes[i] = AstDefRef{KitImpPath: kimp.ImpPath,
										AstDefTop: n.(*atmolang_irfun.AstDefTop) /* ok to panic here bc should-never-happen-else-its-a-bug */}
								}
								kit.lookups.namesInScopeExt.Add(nil, nil, k, nodes...)
							}
						}
					}
				}
			}
		}
	}

	for kit := range kitrepops {
		kit.state.defsBornIdsNames, kit.state.defsGoneIdsNames = nil, nil
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
			tld.Errs.Stage1BadNames.Add(kit.lookups.namesInScopeAll.RepopulateAstDefsAndIdentsFor(tld, &tld.AstDef))
			errs.Add(tld.Errs.Stage1BadNames)
		}
	}
	return
}
