package atmosess

import (
	"github.com/metaleap/atmo/lang/irfun"
)

func (me *Ctx) renewAndRevalidateAffectedIRsIfAnyKitsReloaded() {
	if me.state.someKitsReloaded {
		me.state.someKitsReloaded = false

		for _, kit := range me.Kits.all {
			if len(kit.state.defsNew) > 0 {
				for _, defid := range kit.state.defsNew {
					me.validateNames(kit.lookups.tlDefsByID[defid])
				}
			}
		}
	}
}

func (me *Ctx) validateNames(tlDef *atmolang_irfun.AstDefTop) {
	// namesinscope := make(map[string]bool, 32)
}
