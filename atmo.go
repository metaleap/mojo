package atmo

import (
	"sort"
	"strconv"
)

type Exist struct{}
type StringKeys map[string]Exist
type StringCounts map[string]int

func (me StringKeys) Exists(s string) (ok bool) {
	if me != nil {
		_, ok = me[s]
	}
	return
}

type sorter struct {
	s    []string
	less func(string, string) bool
}

func (me *sorter) Swap(i int, j int)      { me.s[i], me.s[j] = me.s[j], me.s[i] }
func (me *sorter) Len() int               { return len(me.s) }
func (me *sorter) Less(i int, j int) bool { return me.less(me.s[i], me.s[j]) }

func (me StringKeys) SortedBy(isLessThan func(string, string) bool) (sorted []string) {
	i, sorted := 0, make([]string, len(me))
	for k := range me {
		sorted[i] = k
		i++
	}
	sort.Sort(&sorter{s: sorted, less: isLessThan})
	return
}

func (me StringKeys) String() (s string) {
	s = "{"
	for k := range me {
		s += strconv.Quote(k) + ","
	}
	return s + "}"
}

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
	// ∈ aka "exists"
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
