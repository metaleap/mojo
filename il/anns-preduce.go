package atmoil

import (
	"strconv"
)

func (me *Preduced) IsErrOrAbyss() bool { return false }
func (me *Preduced) Self() *Preduced    { return me }

func (me *PFunc) SummaryCompact() string {
	return me.Arg.SummaryCompact() + "->" + me.Ret.SummaryCompact()
}

func (me *PPrimAtomicConstFloat) SummaryCompact() string {
	return strconv.FormatFloat(me.Val, 'f', -1, 64)
}

func (me *PPrimAtomicConstUint) SummaryCompact() string { return strconv.FormatUint(me.Val, 10) }

func (me *PPrimAtomicConstTag) SummaryCompact() string { return me.Val }

func (me *PAbyss) SummaryCompact() string { return "ABYSS" }
func (me *PAbyss) IsErrOrAbyss() bool     { return true }

func (me *PHole) SummaryCompact() string { return "HOLE" }

func (me *PErr) SummaryCompact() string { return me.Err.Error() }
func (me *PErr) IsErrOrAbyss() bool     { return true }
