package atmoil

const (
	_                           = iota
	ErrInit_DefNameInvalidIdent = iota + 2100
	ErrInit_DefNameReserved
	ErrInit_DefArgNameUnderscores
	ErrInit_LeftoverUnderscores
	ErrInit_IdentRefersToMalformedDef
)
const (
	_                            = iota
	ErrNames_ShadowingNotAllowed = iota + 2200
	ErrNames_UndefinedOrUnimported
)
