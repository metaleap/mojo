package atmoil

import (
	"strconv"
)

type IPreduced interface {
	Self() *Preduced
	SummaryCompact() string
}

// Preduced is embedded in all `IPreduced` implementers.
type Preduced struct {
	OrigNodes []INode
}

func (me *Preduced) Self() *Preduced { return me }

type PFunc struct {
	Preduced
	Cases []struct {
		Arg IPreduced
		Ret IPreduced
	}
}

func (me *PFunc) SummaryCompact() string { return "->" }

type PPrimAtomicConstFloat struct {
	Preduced
	Val float64
}

func (me *PPrimAtomicConstFloat) SummaryCompact() string {
	return strconv.FormatFloat(me.Val, 'f', -1, 64)
}

type PPrimAtomicConstUint struct {
	Preduced
	Val uint64
}

func (me *PPrimAtomicConstUint) SummaryCompact() string { return strconv.FormatUint(me.Val, 10) }

type PPrimAtomicConstTag struct {
	Preduced
	Val string
}

func (me *PPrimAtomicConstTag) SummaryCompact() string { return me.Val }
