package atmoil

const (
	_ = 2100 + iota
	ErrFromAst_DefNameInvalidIdent
	ErrFromAst_DefArgNameMultipleUnderscores
	ErrFromAst_UnhandledStandaloneUnderscores
)
const (
	_ = 2200 + iota
	ErrNames_IdentRefersToMalformedDef
)
