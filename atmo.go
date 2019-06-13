package atmo

import (
	"sort"
	"strconv"
)

type Exist struct{}
type StringKeys map[string]Exist

func (me StringKeys) String() (s string) {
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

	KnownIdentI      = "it"
	KnownIdentK      = "ever"
	KnownIdentCoerce = "§"
	KnownIdentOpAnd  = "&&"
	KnownIdentOpOr   = "||"
	KnownIdentOpNeg  = "~"
	KnownIdentUndef  = "÷0"
	KnownIdentIf     = "if"
	KnownIdentEq     = "=="
	KnownIdentNEq    = "/="
	KnownIdentDot    = "."
	KnownIdentColon  = ":"
)

var (
	KnownIdents = StringKeys{KnownIdentI: Є, KnownIdentK: Є, KnownIdentCoerce: Є, KnownIdentOpAnd: Є, KnownIdentOpOr: Є, KnownIdentOpNeg: Є, KnownIdentUndef: Є, KnownIdentIf: Є, KnownIdentEq: Є, KnownIdentNEq: Є}

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
