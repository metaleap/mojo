package atmoil

import (
	"github.com/go-leap/str"
	"strconv"
)

func (me *Preduced) Self() *Preduced { return me }

func (me *PCallable) SummaryCompact() string {
	if me.Arg.Def != nil {
		return ustr.ReplB(DbgPrintToString(me.Arg.Def), '\n', ' ')
	}
	return me.Arg.SummaryCompact() + "->" + me.Ret.SummaryCompact()
}

func (me *PClosure) SummaryCompact() (s string) {
	s = "{"
	for k, v := range me.ArgsEnv {
		s += k.Val + "=(" + DbgPrintToString(v) + "),"
	}
	return s + "}" + me.Def.SummaryCompact()
}

func (me *PPrimAtomicConstFloat) SummaryCompact() string {
	return strconv.FormatFloat(me.Val, 'f', -1, 64)
}

func (me *PPrimAtomicConstUint) SummaryCompact() string { return strconv.FormatUint(me.Val, 10) }

func (me *PPrimAtomicConstTag) SummaryCompact() string { return me.Val }

func (me *PAbyss) SummaryCompact() string { return "ABYSS" }

func (me *PHole) SummaryCompact() string { return "HOLE" }

func (me *PErr) SummaryCompact() string { return me.Err.Error() }
