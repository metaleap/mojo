package atmosess

// import (
// 	"github.com/metaleap/atmo/lang/irfun"
// )

func (me *Ctx) reReduceAffectedIRsIfAnyKitsReloaded() {
	if me.state.someKitsReloaded {
		me.state.someKitsReloaded = false

		for _, kit := range me.Kits.all {
			if len(kit.state.defsNew) > 0 {
				for _, defid := range kit.state.defsNew {
					if tldef := kit.lookups.tlDefsByID[defid]; tldef != nil && len(tldef.Errors) == 0 {

					}
				}
			}
		}
	}
}
