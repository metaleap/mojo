package atmosess

import (
	"github.com/metaleap/atmo"
)

func (me *Ctx) refreshDefsFacts(defIdsBorn map[string]*Kit, defIdsGone map[string]*Kit, defIdsDependantsOfNamesOfChange map[string]*Kit) (freshErrs []error) {
	done := make(atmo.StringKeys, len(defIdsBorn)+len(defIdsDependantsOfNamesOfChange))
	for defid, kit := range defIdsBorn {
		me.refreshDefFacts(kit, defid, done)
	}
	for defid, kit := range defIdsDependantsOfNamesOfChange {
		me.refreshDefFacts(kit, defid, done)
	}
	return
}

func (me *Ctx) refreshDefFacts(kit *Kit, defId string, done atmo.StringKeys) {
	done[defId] = atmo.Є
}
