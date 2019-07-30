package atmo

import (
	"sort"
	"strconv"

	"github.com/go-leap/str"
)

const (
	EnvVarKitsDirs = "ATMO_KITS_DIRS"
	NameAutoKit    = "Std"
	SrcFileExt     = ".at"

	KnownIdentDecl   = ":="
	KnownIdentCoerce = "§"
	KnownIdentOpOr   = "or"
	KnownIdentUndef  = "÷0"
	KnownIdentEq     = "=="
)

var (
	// ∈ aka "exists" for any maps-that-are-sets
	Є       = Exist{}
	Options struct {
		// sorted errors, kits, source files, defs etc:
		// should be enabled for consistency in user-facing tools such as REPLs or language servers.
		// could remain off for mere script runners, transpilers etc.
		Sorts bool
	}
)

func SortMaybe(s sort.Interface) {
	if Options.Sorts && s != nil {
		sort.Sort(s)
	}
}

func (me StringKeys) Exists(s string) (ok bool) {
	if me != nil {
		_, ok = me[s]
	}
	return
}

func (me StringKeys) SortedBy(isLessThan func(string, string) bool) (sorted []string) {
	sorted = make([]string, len(me))
	var i int
	for k := range me {
		sorted[i] = k
		i++
	}
	sort.Sort(ustr.Sortable(sorted, isLessThan))
	return
}

func (me StringKeys) String() (s string) {
	s = "{"
	for k := range me {
		s += strconv.Quote(k) + ","
	}
	return s + "}"
}
