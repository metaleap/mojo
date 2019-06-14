package atmosess

import (
	"github.com/metaleap/atmo"
)

func (me *Ctx) refreshDefsFacts(defIdsBorn map[string]*Kit, defIdsGone map[string]*Kit, defIdsDependantsOfNamesOfChange map[string]*Kit) (freshFactsErrs atmo.Errors) {
	done := make(map[string]bool, len(defIdsBorn)+len(defIdsDependantsOfNamesOfChange))
	for defid, kit := range defIdsBorn {
		me.refreshDefFacts(kit, defid, done)
	}
	for defid, kit := range defIdsDependantsOfNamesOfChange {
		me.refreshDefFacts(kit, defid, done)
	}

	return
}

func (me *Ctx) refreshDefFacts(kit *Kit, defId string, done map[string]bool) {
	if isdone, isdoing := done[defId]; isdone {
		return
	} else if isdoing {
		panic("TODO: handle circular dependencies aka recursion!")
	}
	done[defId] = false

	def := kit.lookups.tlDefsByID[defId]
	def.Facts().Facts = nil

	done[defId] = true
}
