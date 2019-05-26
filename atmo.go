package atmo

import (
	"sort"
	"strconv"
)

type Empty struct{}
type StringsUnorderedButUnique map[string]Empty

func (me StringsUnorderedButUnique) String() (s string) {
	s = "{"
	for k := range me {
		s += strconv.Quote(k) + ","
	}
	return s + "}"
}

const (
	EnvVarKitsDirs = "ATMO_KITS_DIRS"
	NameAutoKit    = "omni"
	SrcFileExt     = ".at"
	Undef          = "()"
)

var (
	Exists  = Empty{}
	Options struct {
		// sorted errors, kits, source files, defs etc:
		// should be enabled for consistency in user-facing tools such as REPLs or language servers.
		// could remain off for mere script runners, transpilers etc.
		Sorts bool
	}
)

func SortMaybe(s sort.Interface) {
	if Options.Sorts {
		sort.Sort(s)
	}
}
