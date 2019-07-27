package atmolang

const (
	_                                 = iota
	ErrLexing_IndentationInconsistent = iota + 1100
	ErrLexing_Other
)
const (
	_                         = iota
	ErrParsing_DefBodyMissing = iota + 1200
	ErrParsing_DefMissing
	ErrParsing_DefHeaderMissing
	ErrParsing_DefHeaderMalformed
	ErrParsing_DefNameAffixMalformed
	ErrParsing_DefNameMalformed
	ErrParsing_DefArgAffixMalformed
	ErrParsing_TokenUnexpected_Separator
	ErrParsing_TokenUnexpected_DefDecl
	ErrParsing_TokenUnexpected_Underscores
	ErrParsing_ExpressionMissing_Accum
	ErrParsing_ExpressionMissing_Case
	ErrParsing_CaseEmpty
	ErrParsing_CaseNoPair
	ErrParsing_CaseNoResult
	ErrParsing_CaseSecondDefault
	ErrParsing_CaseDisjNoResult
	ErrParsing_CommasConsecutive
	ErrParsing_CommasMixDefsAndExprs
	ErrParsing_BracketUnclosed
	ErrParsing_BracketUnopened
)
const (
	_                                               = iota
	ErrDesugaring_BranchMalformed_CaseResultMissing = iota + 1300
)
