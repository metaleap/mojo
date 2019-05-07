package atmosess

func (me *Ctx) renewAndRevalidateAffectedIRsIfAnyKitsReloaded() {
	if me.state.someKitsReloaded {
		me.state.someKitsReloaded = false
	}
}
