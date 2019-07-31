package atmolang

const (
	_ = 1100 + iota
	ErrLexing_IndentationInconsistent
	ErrLexing_IoFileOpenFailure
	ErrLexing_IoFileReadFailure
	ErrLexing_Tokenization
)
const (
	_ = 1200 + iota
	ErrParsing_DefBodyMissing
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
	ErrParsing_IdentExpected
)
const (
	_ = 1300 + iota
	ErrDesugaring_BranchMalformed_CaseResultMissing
)
