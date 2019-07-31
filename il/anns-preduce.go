package atmoil

import (
	"strconv"
)

func (me *Preduced) Self() *Preduced { return me }

func (me *PFunc) SummaryCompact() string { return "->" }

func (me *PPrimAtomicConstFloat) SummaryCompact() string {
	return strconv.FormatFloat(me.Val, 'f', -1, 64)
}

func (me *PPrimAtomicConstUint) SummaryCompact() string { return strconv.FormatUint(me.Val, 10) }

func (me *PPrimAtomicConstTag) SummaryCompact() string { return me.Val }

func (me *PAbyss) SummaryCompact() string { return "ABYSS" }
