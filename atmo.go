package atmo

import (
	"math/rand"
	"sort"
	"strconv"
	"time"

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

func init() { rand.Seed(time.Now().UnixNano()) }

func SortMaybe(s sort.Interface) {
	if Options.Sorts && s != nil {
		sort.Sort(s)
	}
}

func StrRand(appendNowNanoToRandStr bool) (rndStr string) {
	if rndStr = strconv.FormatInt(rand.Int63(), 16); appendNowNanoToRandStr {
		rndStr += strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return
}

func (me StringKeys) Exists(s string) (ok bool) {
	if me != nil {
		_, ok = me[s]
	}
	return
}

func (me StringKeys) Sorted(isLessThan func(string, string) bool) (sorted []string) {
	sorted = make([]string, len(me))
	if isLessThan == nil {
		isLessThan = func(s1 string, s2 string) bool { return s1 < s2 }
	}
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
