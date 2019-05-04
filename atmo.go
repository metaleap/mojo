package atmo

import (
	"sort"
)

const (
	EnvVarKitsDirs = "ATMO_KITS_DIRS"
	NameAutoKit    = "omni"
	SrcFileExt     = ".at"
)

var Options struct {
	// sorted errors, kits, source files, defs etc:
	// should be enabled for consistency in user-facing tools such as REPLs or language servers.
	// could remain off for mere script runners, transpilers etc.
	Sorts bool
}

func SortMaybe(s sort.Interface) {
	if Options.Sorts {
		sort.Sort(s)
	}
}
