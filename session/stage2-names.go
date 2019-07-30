package atmosess

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
)

type IrDefRef struct {
	*atmoil.IrDefTop
	Kit *Kit
}

func (me *Ctx) kitsRepopulateNamesInScope() (namesOfChange atmo.StringKeys, defIdsBorn map[string]*Kit, defIdsGone map[string]*Kit, errs atmo.Errors) {
	kitrepops := make(map[*Kit]map[string]int, len(me.Kits.All))
	defIdsBorn, defIdsGone, namesOfChange = make(map[string]*Kit, 2), make(map[string]*Kit, 2), make(atmo.StringKeys, 4)

	{ // FIRST: namesInScopeOwn
		for _, kit := range me.Kits.All {
			if kit.WasEverToBeLoaded {
				if len(kit.state.defsGoneIdsNames) > 0 || len(kit.state.defsBornIdsNames) > 0 {
					for defid, defname := range kit.state.defsGoneIdsNames {
						defIdsGone[defid] = kit
						namesOfChange[defname] = atmo.Є
					}
					for defid, defname := range kit.state.defsBornIdsNames {
						defIdsBorn[defid] = kit
						namesOfChange[defname] = atmo.Є
					}

					kitrepops[kit], kit.lookups.namesInScopeAll, kit.lookups.namesInScopeOwn =
						map[string]int{}, nil, make(atmoil.AnnNamesInScope, len(kit.topLevelDefs))
					for _, tld := range kit.topLevelDefs {
						tld.Errs.Stage2BadNames = nil
						if tld.Name.Val != "" {
							kit.lookups.namesInScopeOwn.Add(tld.Name.Val, tld)
						}
					}
				}
			}
		}
	}
	{ // NEXT: namesInScopeExt (utilizing above namesInScopeOwn adds hence separate Ctx.Kits.all loop)
		for _, kit := range me.Kits.All {
			if kit.WasEverToBeLoaded {
				if kitimports := kit.Imports(); len(kitimports) > 0 {
					var totaldefscount int
					var anychangesinkimps bool
					kimps := me.Kits.All.Where(func(k *Kit) (iskimp bool) {
						if iskimp = ustr.In(k.ImpPath, kitimports...); iskimp {
							totaldefscount, anychangesinkimps = totaldefscount+len(k.topLevelDefs), anychangesinkimps || len(k.state.defsGoneIdsNames) > 0 || len(k.state.defsBornIdsNames) > 0
						}
						return
					})
					if anychangesinkimps || len(kit.lookups.namesInScopeExt) == 0 {
						if _, alreadymarked := kitrepops[kit]; !alreadymarked {
							kitrepops[kit] = map[string]int{}
							for _, tld := range kit.topLevelDefs {
								tld.Errs.Stage2BadNames = nil
							}
						}
						kit.lookups.namesInScopeAll, kit.lookups.namesInScopeExt =
							nil, make(atmoil.AnnNamesInScope, totaldefscount)
						for _, kimp := range kimps {
							for name, nodesown := range kimp.lookups.namesInScopeOwn {
								nodes := make([]atmoil.INode, 0, len(nodesown))
								for _, n := range nodesown {
									deftop := n.(*atmoil.IrDefTop) /* ok to panic here bc should-never-happen-else-its-a-bug */
									if !deftop.OrigTopChunk.Ast.Def.IsUnexported {
										nodes = append(nodes, IrDefRef{Kit: kimp,
											IrDefTop: deftop})
									}
								}
								kit.lookups.namesInScopeExt.Add(name, nodes...)
							}
						}
					}
				}
			}
		}
	}

	for kit, badglobalsnames := range kitrepops {
		kit.state.defsBornIdsNames, kit.state.defsGoneIdsNames = nil, nil
		kit.lookups.namesInScopeAll = make(atmoil.AnnNamesInScope, len(kit.lookups.namesInScopeExt)+len(kit.lookups.namesInScopeOwn))
		for k, v := range kit.lookups.namesInScopeOwn {
			nodes := make([]atmoil.INode, len(v))
			copy(nodes, v)
			kit.lookups.namesInScopeAll[k] = nodes
		}
		for k, v := range kit.lookups.namesInScopeExt {
			kit.lookups.namesInScopeAll.Add(k, v...)
		}
		me.kitGatherAllUnparsedGlobalsNames(kit, badglobalsnames)
		for _, tld := range kit.topLevelDefs {
			tld.Errs.Stage2BadNames.Add(kit.lookups.namesInScopeAll.RepopulateDefsAndIdentsFor(tld, &tld.IrDef, badglobalsnames)...)
			errs.Add(tld.Errs.Stage2BadNames...)
		}
	}
	return
}
