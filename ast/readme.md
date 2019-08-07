# atmoast
--
    import "github.com/metaleap/atmo/ast"

Package `atmo/ast` offers AST node structures and supporting auxiliary types and
funcs, plus implements the lexing and parsing into such ASTs. It has no notion
of kits or imports, and little-to-no semantic prepossession, being chiefly
concerned with syntactical analysis (eg. it does not care if a call is made to a
number or other known-non-callable, etc.).

## Usage

```go
const (
	ErrLexing_IndentationInconsistent
	ErrLexing_IoFileOpenFailure
	ErrLexing_IoFileReadFailure
	ErrLexing_Tokenization
)
```

```go
const (
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
```

```go
const (
	ErrDesugaring_BranchMalformed_CaseResultMissing
)
```

#### func  LexAndGuess

```go
func LexAndGuess(fauxSrcFileNameForErrs string, src []byte) (guessIsDef bool, guessIsExpr bool, lexedToks udevlex.Tokens, err *Error)
```

#### func  LexAndParseDefOrExpr

```go
func LexAndParseDefOrExpr(def bool, toks udevlex.Tokens) (ret IAstNode, err *Error)
```

#### func  PrintTo

```go
func PrintTo(curTopLevel *AstDef, node IAstNode, out io.Writer, prominentForDebugPurposes bool, applStyle ApplStyle)
```

#### func  PrintToStderr

```go
func PrintToStderr(node IAstNode)
```

#### type ApplStyle

```go
type ApplStyle int
```


```go
const (
	APPLSTYLE_SVO ApplStyle = iota
	APPLSTYLE_VSO
	APPLSTYLE_SOV
)
```

#### type AstBaseComments

```go
type AstBaseComments struct {
}
```


#### func (*AstBaseComments) Comments

```go
func (me *AstBaseComments) Comments() *astBaseComments
```

#### type AstBaseExpr

```go
type AstBaseExpr struct {
	AstBaseTokens
	AstBaseComments
}
```


#### func (*AstBaseExpr) Desugared

```go
func (*AstBaseExpr) Desugared(func() string) (IAstExpr, Errors)
```

#### func (*AstBaseExpr) IsAtomic

```go
func (*AstBaseExpr) IsAtomic() bool
```

#### type AstBaseExprAtom

```go
type AstBaseExprAtom struct {
	AstBaseExpr
}
```


#### func (*AstBaseExprAtom) IsAtomic

```go
func (*AstBaseExprAtom) IsAtomic() bool
```

#### type AstBaseExprAtomLit

```go
type AstBaseExprAtomLit struct {
	AstBaseExprAtom
}
```


#### type AstBaseTokens

```go
type AstBaseTokens struct {
	Tokens udevlex.Tokens
}
```


#### func (*AstBaseTokens) Toks

```go
func (me *AstBaseTokens) Toks() udevlex.Tokens
```

#### type AstBuild

```go
type AstBuild struct{}
```


```go
var (
	BuildAst AstBuild
)
```

#### func (AstBuild) Appl

```go
func (AstBuild) Appl(callee IAstExpr, args ...IAstExpr) *AstExprAppl
```

#### func (AstBuild) Arg

```go
func (AstBuild) Arg(nameOrConstVal IAstExprAtomic, affix IAstExpr) *AstDefArg
```

#### func (AstBuild) Cases

```go
func (AstBuild) Cases(scrutinee IAstExpr, alts ...AstCase) *AstExprCases
```

#### func (AstBuild) Def

```go
func (AstBuild) Def(name string, body IAstExpr, argNames ...string) *AstDef
```

#### func (AstBuild) Ident

```go
func (AstBuild) Ident(val string) *AstIdent
```

#### func (AstBuild) Let

```go
func (AstBuild) Let(body IAstExpr, defs ...AstDef) *AstExprLet
```

#### func (AstBuild) LitFloat

```go
func (AstBuild) LitFloat(val float64) *AstExprLitFloat
```

#### func (AstBuild) LitRune

```go
func (AstBuild) LitRune(val int32) *AstExprLitUint
```

#### func (AstBuild) LitStr

```go
func (AstBuild) LitStr(val string) *AstExprLitStr
```

#### func (AstBuild) LitUint

```go
func (AstBuild) LitUint(val uint64) *AstExprLitUint
```

#### func (AstBuild) Tag

```go
func (AstBuild) Tag(val string) (ret *AstIdent)
```

#### type AstCase

```go
type AstCase struct {
	AstBaseTokens
	Conds []IAstExpr
	Body  IAstExpr
}
```


#### type AstComment

```go
type AstComment struct {
	AstBaseTokens
	Val           string
	IsLineComment bool
}
```


#### type AstComments

```go
type AstComments []AstComment
```


#### type AstDef

```go
type AstDef struct {
	AstBaseTokens
	Name       AstIdent
	NameAffix  IAstExpr
	Args       []AstDefArg
	Meta       []IAstExpr
	Body       IAstExpr
	IsTopLevel bool
}
```


#### func (*AstDef) Print

```go
func (me *AstDef) Print(ctxp *CtxPrint)
```

#### func (*AstDef) ToUnary

```go
func (me *AstDef) ToUnary() *AstDef
```

#### type AstDefArg

```go
type AstDefArg struct {
	AstBaseTokens
	NameOrConstVal IAstExpr
	Affix          IAstExpr
}
```


#### type AstExprAppl

```go
type AstExprAppl struct {
	AstBaseExpr
	Callee IAstExpr
	Args   []IAstExpr
}
```


#### func (*AstExprAppl) CalleeAndArgsOrdered

```go
func (me *AstExprAppl) CalleeAndArgsOrdered(applStyle ApplStyle) (ret []IAstExpr)
```

#### func (*AstExprAppl) ClaspishByTokens

```go
func (me *AstExprAppl) ClaspishByTokens() (claspish bool)
```

#### func (*AstExprAppl) Desugared

```go
func (me *AstExprAppl) Desugared(prefix func() string) (IAstExpr, Errors)
```

#### func (*AstExprAppl) ToUnary

```go
func (me *AstExprAppl) ToUnary() (unary *AstExprAppl)
```

#### type AstExprCases

```go
type AstExprCases struct {
	AstBaseExpr
	Scrutinee IAstExpr
	Alts      []AstCase
}
```


#### func (*AstExprCases) Default

```go
func (me *AstExprCases) Default() *AstCase
```

#### func (*AstExprCases) Desugared

```go
func (me *AstExprCases) Desugared(prefix func() string) (expr IAstExpr, errs Errors)
```

#### type AstExprLet

```go
type AstExprLet struct {
	AstBaseExpr
	Defs []AstDef
	Body IAstExpr
}
```


#### func (*AstExprLet) Desugared

```go
func (me *AstExprLet) Desugared(prefix func() string) (expr IAstExpr, errs Errors)
```

#### type AstExprLitFloat

```go
type AstExprLitFloat struct {
	AstBaseExprAtomLit
	Val float64
}
```


#### func (*AstExprLitFloat) String

```go
func (me *AstExprLitFloat) String() string
```

#### type AstExprLitStr

```go
type AstExprLitStr struct {
	AstBaseExprAtomLit
	Val string
}
```


#### func (*AstExprLitStr) String

```go
func (me *AstExprLitStr) String() string
```

#### type AstExprLitUint

```go
type AstExprLitUint struct {
	AstBaseExprAtomLit
	Val uint64
}
```


#### func (*AstExprLitUint) FromRune

```go
func (me *AstExprLitUint) FromRune() bool
```

#### func (*AstExprLitUint) String

```go
func (me *AstExprLitUint) String() string
```

#### type AstFile

```go
type AstFile struct {
	TopLevel []AstFileChunk

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
}
```


#### func (*AstFile) CountNetLinesOfCode

```go
func (me *AstFile) CountNetLinesOfCode(onlyCountErrless bool) (sloc int)
```

#### func (*AstFile) CountTopLevelDefs

```go
func (me *AstFile) CountTopLevelDefs(onlyCountErrless bool) (total int, unexported int)
```

#### func (*AstFile) Errors

```go
func (me *AstFile) Errors() Errors
```

#### func (*AstFile) HasDefs

```go
func (me *AstFile) HasDefs(name string, includeUnparsed bool) bool
```

#### func (*AstFile) HasErrors

```go
func (me *AstFile) HasErrors() (r bool)
```

#### func (*AstFile) LexAndParseFile

```go
func (me *AstFile) LexAndParseFile(onlyIfModifiedSinceLastLoad bool, stdinIfNoSrcFile bool, noChangesDetected *bool) (freshErrs Errors)
```

#### func (*AstFile) LexAndParseSrc

```go
func (me *AstFile) LexAndParseSrc(r io.Reader, noChangesDetected *bool) (freshErrs Errors)
```

#### func (*AstFile) Print

```go
func (me *AstFile) Print(fmt IPrintFmt) []byte
```

#### func (*AstFile) TopLevelChunkAt

```go
func (me *AstFile) TopLevelChunkAt(pos0ByteOffset int) *AstFileChunk
```

#### type AstFileChunk

```go
type AstFileChunk struct {
	Src     []byte
	SrcFile *AstFile

	Ast AstTopLevel
}
```


#### func (*AstFileChunk) At

```go
func (me *AstFileChunk) At(byte0PosOffsetInSrcFile int) []IAstNode
```

#### func (*AstFileChunk) Encloses

```go
func (me *AstFileChunk) Encloses(byte0PosOffsetInSrcFile int) bool
```

#### func (*AstFileChunk) Errs

```go
func (me *AstFileChunk) Errs() Errors
```

#### func (*AstFileChunk) HasErrors

```go
func (me *AstFileChunk) HasErrors() bool
```

#### func (*AstFileChunk) Id

```go
func (me *AstFileChunk) Id() string
```

#### func (*AstFileChunk) PosOffsetByte

```go
func (me *AstFileChunk) PosOffsetByte() int
```
PosOffsetByte implements `atmo.IErrPosOffsets`.

#### func (*AstFileChunk) PosOffsetLine

```go
func (me *AstFileChunk) PosOffsetLine() int
```
PosOffsetLine implements `atmo.IErrPosOffsets`.

#### func (*AstFileChunk) Print

```go
func (me *AstFileChunk) Print(p *CtxPrint)
```

#### type AstFiles

```go
type AstFiles []*AstFile
```


#### func (AstFiles) ByFilePath

```go
func (me AstFiles) ByFilePath(srcFilePath string) *AstFile
```

#### func (AstFiles) Index

```go
func (me AstFiles) Index(srcFilePath string) int
```

#### func (AstFiles) Len

```go
func (me AstFiles) Len() int
```

#### func (AstFiles) Less

```go
func (me AstFiles) Less(i int, j int) bool
```

#### func (*AstFiles) RemoveAt

```go
func (me *AstFiles) RemoveAt(idx int)
```

#### func (AstFiles) Swap

```go
func (me AstFiles) Swap(i int, j int)
```

#### func (AstFiles) TopLevelChunkByDefId

```go
func (me AstFiles) TopLevelChunkByDefId(defId string) *AstFileChunk
```

#### type AstIdent

```go
type AstIdent struct {
	AstBaseExprAtom
	Val     string
	IsOpish bool
	IsTag   bool
}
```


#### func (*AstIdent) IsName

```go
func (me *AstIdent) IsName(opishOk bool) bool
```

#### func (*AstIdent) IsPlaceholder

```go
func (me *AstIdent) IsPlaceholder() bool
```

#### func (*AstIdent) IsVar

```go
func (me *AstIdent) IsVar() bool
```

#### func (*AstIdent) String

```go
func (me *AstIdent) String() string
```

#### type AstTopLevel

```go
type AstTopLevel struct {
	AstBaseTokens
	AstBaseComments
	Def struct {
		Orig         *AstDef
		NameIfErr    string
		IsUnexported bool
	}
}
```


#### type CtxPrint

```go
type CtxPrint struct {
	Fmt            IPrintFmt
	ApplStyle      ApplStyle
	NoComments     bool
	CurTopLevel    *AstDef
	CurIndentLevel int
	OneIndentLevel string

	ustd.BytesWriter
}
```


#### func (*CtxPrint) Print

```go
func (me *CtxPrint) Print(node IAstNode) *CtxPrint
```

#### func (*CtxPrint) WriteLineBreaksThenIndent

```go
func (me *CtxPrint) WriteLineBreaksThenIndent(numLines int)
```

#### type IAstComments

```go
type IAstComments interface {
	Comments() *astBaseComments
}
```


#### type IAstExpr

```go
type IAstExpr interface {
	IAstNode
	IAstComments
	IsAtomic() bool
	Desugared(func() string) (IAstExpr, Errors)
}
```


#### type IAstExprAtomic

```go
type IAstExprAtomic interface {
	IAstExpr
	String() string
}
```


#### type IAstNode

```go
type IAstNode interface {
	Toks() udevlex.Tokens
	// contains filtered or unexported methods
}
```


#### type IPrintFmt

```go
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
```

IPrintFmt is fully implemented by `PrintFormatterMinimal`, for custom formatters
it'll be best to embed this and then override specifics.

#### type PrintFmtMinimal

```go
type PrintFmtMinimal struct{ *CtxPrint }
```

PrintFmtMinimal implements `IPrintFmt`.

#### func (*PrintFmtMinimal) OnComment

```go
func (me *PrintFmtMinimal) OnComment(leads IAstNode, trails IAstNode, node *AstComment)
```

#### func (*PrintFmtMinimal) OnDef

```go
func (me *PrintFmtMinimal) OnDef(_ *AstTopLevel, node *AstDef)
```

#### func (*PrintFmtMinimal) OnDefArg

```go
func (me *PrintFmtMinimal) OnDefArg(_ *AstDef, argIdx int, node *AstDefArg)
```

#### func (*PrintFmtMinimal) OnDefBody

```go
func (me *PrintFmtMinimal) OnDefBody(def *AstDef, node IAstExpr)
```

#### func (*PrintFmtMinimal) OnDefMeta

```go
func (me *PrintFmtMinimal) OnDefMeta(_ *AstDef, _ int, node IAstExpr)
```

#### func (*PrintFmtMinimal) OnDefName

```go
func (me *PrintFmtMinimal) OnDefName(_ *AstDef, node *AstIdent)
```

#### func (*PrintFmtMinimal) OnExprApplArg

```go
func (me *PrintFmtMinimal) OnExprApplArg(_ bool, appl *AstExprAppl, argIdx int, node IAstExpr)
```

#### func (*PrintFmtMinimal) OnExprApplName

```go
func (me *PrintFmtMinimal) OnExprApplName(_ bool, _ *AstExprAppl, node IAstExpr)
```

#### func (*PrintFmtMinimal) OnExprCasesBody

```go
func (me *PrintFmtMinimal) OnExprCasesBody(_ *AstCase, node IAstExpr)
```

#### func (*PrintFmtMinimal) OnExprCasesCond

```go
func (me *PrintFmtMinimal) OnExprCasesCond(_ *AstCase, _ int, node IAstExpr)
```

#### func (*PrintFmtMinimal) OnExprCasesScrutinee

```go
func (me *PrintFmtMinimal) OnExprCasesScrutinee(_ bool, _ *AstExprCases, node IAstExpr)
```

#### func (*PrintFmtMinimal) OnExprLetBody

```go
func (me *PrintFmtMinimal) OnExprLetBody(_ *AstExprLet, node IAstExpr)
```

#### func (*PrintFmtMinimal) OnExprLetDef

```go
func (me *PrintFmtMinimal) OnExprLetDef(_ *AstExprLet, _ int, node *AstDef)
```

#### func (*PrintFmtMinimal) OnTopLevelChunk

```go
func (me *PrintFmtMinimal) OnTopLevelChunk(tlc *AstFileChunk, node *AstTopLevel)
```

#### func (*PrintFmtMinimal) PrintInParensIf

```go
func (me *PrintFmtMinimal) PrintInParensIf(node IAstNode, ifCases bool, ifNotAtomicOrClaspish bool)
```

#### func (*PrintFmtMinimal) SetCtxPrint

```go
func (me *PrintFmtMinimal) SetCtxPrint(ctxPrint *CtxPrint)
```

#### type PrintFmtPretty

```go
type PrintFmtPretty struct{ PrintFmtMinimal }
```

PrintFmtPretty implements `IPrintFmt`.

#### func (*PrintFmtPretty) OnComment

```go
func (me *PrintFmtPretty) OnComment(leads IAstNode, trails IAstNode, node *AstComment)
```

#### func (*PrintFmtPretty) OnDefBody

```go
func (me *PrintFmtPretty) OnDefBody(def *AstDef, node IAstExpr)
```

#### func (*PrintFmtPretty) OnExprCasesBody

```go
func (me *PrintFmtPretty) OnExprCasesBody(_ *AstCase, node IAstExpr)
```

#### func (*PrintFmtPretty) OnExprCasesCond

```go
func (me *PrintFmtPretty) OnExprCasesCond(_ *AstCase, _ int, node IAstExpr)
```

#### func (*PrintFmtPretty) OnExprCasesScrutinee

```go
func (me *PrintFmtPretty) OnExprCasesScrutinee(_ bool, _ *AstExprCases, node IAstExpr)
```

#### func (*PrintFmtPretty) OnExprLetBody

```go
func (me *PrintFmtPretty) OnExprLetBody(_ *AstExprLet, node IAstExpr)
```

#### func (*PrintFmtPretty) OnExprLetDef

```go
func (me *PrintFmtPretty) OnExprLetDef(let *AstExprLet, idx int, node *AstDef)
```

#### func (*PrintFmtPretty) OnTopLevelChunk

```go
func (me *PrintFmtPretty) OnTopLevelChunk(tlc *AstFileChunk, node *AstTopLevel)
```
