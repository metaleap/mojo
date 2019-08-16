package atmosess

import (
	"github.com/go-leap/str"
	. "github.com/metaleap/atmo"
	. "github.com/metaleap/atmo/il"
)

func (me *Ctx) kitsRepopulateNamesInScope() (namesOfChange StringKeys, defIdsBorn map[string]*Kit, defIdsGone map[string]*Kit, errs Errors) {
	kitrepops := make(map[*Kit]StringKeys, len(me.Kits.All))
	defIdsBorn, defIdsGone, namesOfChange = make(map[string]*Kit, 2), make(map[string]*Kit, 2), make(StringKeys, 4)

	{ // FIRST: namesInScopeOwn
		for _, kit := range me.Kits.All {
			if kit.WasEverToBeLoaded {
				if len(kit.state.defsGoneIdsNames) != 0 || len(kit.state.defsBornIdsNames) != 0 {
					for defid, defname := range kit.state.defsGoneIdsNames {
						defIdsGone[defid] = kit
						namesOfChange[defname] = Є
					}
					for defid, defname := range kit.state.defsBornIdsNames {
						defIdsBorn[defid] = kit
						namesOfChange[defname] = Є
					}

					kitrepops[kit], kit.lookups.namesInScopeAll, kit.lookups.namesInScopeOwn =
						StringKeys{}, nil, make(AnnNamesInScope, len(kit.topLevelDefs))
					for _, tld := range kit.topLevelDefs {
						tld.Errs.Stage2BadNames = nil
						if tld.Ident.Name != "" {
							kit.lookups.namesInScopeOwn.Add(tld.Ident.Name, tld)
						}
					}
				}
			}
		}
	}
	{ // NEXT: namesInScopeExt (utilizing above namesInScopeOwn adds hence separate Ctx.Kits.all loop)
		for _, kit := range me.Kits.All {
			if kit.WasEverToBeLoaded {
				if kitimports := kit.Imports(); len(kitimports) != 0 {
					var totaldefscount int
					var anychangesinkimps bool
					kimps := me.Kits.All.Where(func(k *Kit) (iskimp bool) {
						if iskimp = ustr.In(k.ImpPath, kitimports...); iskimp {
							totaldefscount, anychangesinkimps = totaldefscount+len(k.topLevelDefs), anychangesinkimps || len(k.state.defsGoneIdsNames) != 0 || len(k.state.defsBornIdsNames) != 0
						}
						return
					})
					if anychangesinkimps || len(kit.lookups.namesInScopeExt) == 0 {
						if _, alreadymarked := kitrepops[kit]; !alreadymarked {
							kitrepops[kit] = StringKeys{}
							for _, tld := range kit.topLevelDefs {
								tld.Errs.Stage2BadNames = nil
							}
						}
						kit.lookups.namesInScopeAll, kit.lookups.namesInScopeExt =
							nil, make(AnnNamesInScope, totaldefscount)
						for _, kimp := range kimps {
							for name, nodesown := range kimp.lookups.namesInScopeOwn {
								nodes := make([]IIrNode, 0, len(nodesown))
								for _, n := range nodesown {
									def := n.(*IrDef) /* ok to panic here bc should-never-happen-else-its-a-bug */
									if !def.AstFileChunk.Ast.Def.IsUnexported {
										nodes = append(nodes, IrDefRef{Kit: kimp,
											IrDef: def})
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
		kit.lookups.namesInScopeAll = make(AnnNamesInScope, len(kit.lookups.namesInScopeExt)+len(kit.lookups.namesInScopeOwn))
		for name, nodes := range kit.lookups.namesInScopeOwn {
			kit.lookups.namesInScopeAll.Add(name, nodes...)
		}
		for name, nodes := range kit.lookups.namesInScopeExt {
			kit.lookups.namesInScopeAll.Add(name, nodes...)
		}
		me.kitGatherAllUnparsedGlobalsNames(kit, badglobalsnames)
		for _, tld := range kit.topLevelDefs {
			tld.Errs.Stage2BadNames.Add(kit.lookups.namesInScopeAll.RepopulateDefsAndIdentsFor(tld, tld, badglobalsnames)...)
			errs.Add(tld.Errs.Stage2BadNames...)
		}
	}
	return
}

func (me *Kit) NamesInScope() AnnNamesInScope {
	return me.lookups.namesInScopeAll
}
