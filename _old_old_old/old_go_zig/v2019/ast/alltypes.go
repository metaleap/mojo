// Package `atmo/ast` offers AST node structures and supporting auxiliary
// types and funcs, plus implements the lexing and parsing into such ASTs.
// It has no notion of kits or imports, and little-to-no semantic prepossession,
// being chiefly concerned with syntactical analysis (eg. it does not care
// if a call is made to a number or other known-non-callable, etc.).
package atmoast

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/std"
	. "github.com/metaleap/atmo/old/v2019"
)

type ApplStyle int

const (
	APPLSTYLE_SVO ApplStyle = iota
	APPLSTYLE_VSO
	APPLSTYLE_SOV
)

type preLexTopLevelChunk struct {
	src                   []byte
	pos                   int
	line                  int
	numLinesTabIndented   int
	numLinesSpaceIndented int
}

type ctxTldParse struct {
	curTopLevel     *AstFileChunk
	curTopDef       *AstDef
	brackets        []byte
	bracketsHalfIdx int
}

type AstFiles []*AstFile

type AstFile struct {
	TopLevel []AstFileChunk
	errs     struct {
		loading *Error
	}
	LastLoad struct {
		Src      []byte
		Time     int64
		FileSize int64
		NumLines int
	}
	Options struct {
		ApplStyle ApplStyle
		TmpAltSrc []byte
	}
	SrcFilePath string

	_toks udevlex.Tokens
	_errs Errors
}

type AstFileChunk struct {
	Src     []byte
	SrcFile *AstFile
	offset  struct {
		Ln int
		B  int
	}
	preLex struct {
		numLinesTabIndented   int
		numLinesSpaceIndented int
	}
	id       [3]uint64
	_id      string
	_errs    Errors
	srcDirty bool
	errs     struct {
		lexing  Errors
		parsing *Error
	}
	Ast AstTopLevel
}

type IAstNode interface {
	print(*CtxPrint)
	at(IAstNode, int) []IAstNode
	Toks() udevlex.Tokens
}

type IAstComments interface {
	Comments() *astBaseComments
}

type IAstExpr interface {
	IAstNode
	IAstComments
	IsAtomic() bool
	Desugared(func() string) (IAstExpr, Errors)
}

type IAstExprAtomic interface {
	IAstExpr
	String() string
}

type AstBaseTokens struct {
	Tokens udevlex.Tokens
}

type astBaseComments = struct {
	Leading  AstComments
	Trailing AstComments
}

type AstBaseComments struct {
	comments astBaseComments
}

type AstTopLevel struct {
	AstBaseTokens
	AstBaseComments
	Def struct {
		Orig         *AstDef
		NameIfErr    string
		IsUnexported bool
	}
}

type AstComments []AstComment

type AstComment struct {
	AstBaseTokens
	Val           string
	IsLineComment bool
}

type AstDef struct {
	AstBaseTokens
	Name       AstIdent
	NameAffix  IAstExpr
	Args       []AstDefArg
	Meta       []IAstExpr
	Body       IAstExpr
	IsTopLevel bool
}

type AstDefArg struct {
	AstBaseTokens
	NameOrConstVal IAstExpr
	Affix          IAstExpr
}

type AstBaseExpr struct {
	AstBaseTokens
	AstBaseComments
}

type AstBaseExprAtom struct {
	AstBaseExpr
}

type AstBaseExprAtomLit struct {
	AstBaseExprAtom
}

type AstExprLitUint struct {
	AstBaseExprAtomLit
	Val uint64
}

type AstExprLitFloat struct {
	AstBaseExprAtomLit
	Val float64
}

type AstExprLitStr struct {
	AstBaseExprAtomLit
	Val string
}

type AstIdent struct {
	AstBaseExprAtom
	Val     string
	IsOpish bool
	IsTag   bool
}

type AstExprAppl struct {
	AstBaseExpr
	Callee IAstExpr
	Args   []IAstExpr
}

type AstExprLet struct {
	AstBaseExpr
	Defs []AstDef
	Body IAstExpr
}

type AstExprCases struct {
	AstBaseExpr
	Scrutinee    IAstExpr
	Alts         []AstCase
	defaultIndex int
}

type AstCase struct {
	AstBaseTokens
	Conds []IAstExpr
	Body  IAstExpr
}

type AstBuild struct{}

// IPrintFmt is fully implemented by `PrintFormatterMinimal`, for custom
// formatters it'll be best to embed this and then override specifics.
type IPrintFmt interface {
	SetCtxPrint(*CtxPrint)
	OnTopLevelChunk(*AstFileChunk, *AstTopLevel)
	OnDef(*AstTopLevel, *AstDef)
	OnDefName(*AstDef, *AstIdent)
	OnDefArg(*AstDef, int, *AstDefArg)
	OnDefMeta(*AstDef, int, IAstExpr)
	OnDefBody(*AstDef, IAstExpr)
	OnExprLetBody(*AstExprLet, IAstExpr)
	OnExprLetDef(*AstExprLet, int, *AstDef)
	OnExprApplName(bool, *AstExprAppl, IAstExpr)
	OnExprApplArg(bool, *AstExprAppl, int, IAstExpr)
	OnExprCasesScrutinee(bool, *AstExprCases, IAstExpr)
	OnExprCasesCond(*AstCase, int, IAstExpr)
	OnExprCasesBody(*AstCase, IAstExpr)
	OnComment(IAstNode, IAstNode, *AstComment)
}

type CtxPrint struct {
	Fmt            IPrintFmt
	ApplStyle      ApplStyle
	NoComments     bool
	CurTopLevel    *AstDef
	CurIndentLevel int
	OneIndentLevel string

	ustd.BytesWriter

	fmtCtxSet bool
}

// PrintFmtMinimal implements `IPrintFmt`.
type PrintFmtMinimal struct{ *CtxPrint }

// PrintFmtPretty implements `IPrintFmt`.
type PrintFmtPretty struct{ PrintFmtMinimal }
