# odlang
--
    import "github.com/metaleap/odic/lang"


## Usage

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
	Comments []*AstComment
}
```


#### type AstBaseTokens

```go
type AstBaseTokens struct {
	Tokens udevlex.Tokens
}
```


#### type AstCaseAlt

```go
type AstCaseAlt struct {
	AstBaseTokens
	Conds       []IAstExpr
	Body        IAstExpr
	IsShortForm bool
}
```


#### type AstComment

```go
type AstComment struct {
	AstBaseTokens
	ContentText       string
	IsSelfTerminating bool
}
```


#### type AstDefBase

```go
type AstDefBase struct {
	AstBaseTokens
	Name AstIdent
	Args []AstIdent
	Meta []IAstExpr

	IsDefType bool
}
```


#### func (*AstDefBase) DefBase

```go
func (me *AstDefBase) DefBase() *AstDefBase
```

#### type AstDefFunc

```go
type AstDefFunc struct {
	AstDefBase
	Body IAstExpr
}
```


#### type AstDefType

```go
type AstDefType struct {
	AstDefBase
	Expr IAstTypeExpr
	Tags []AstDefTypeTag
}
```


#### type AstDefTypeTag

```go
type AstDefTypeTag struct {
	Name AstIdent
	Expr IAstTypeExpr
}
```


#### type AstExprAppl

```go
type AstExprAppl struct {
	AstExprBase
	Callee IAstExpr
	Args   []IAstExpr
}
```


#### func (*AstExprAppl) Description

```go
func (me *AstExprAppl) Description() string
```

#### type AstExprAtomBase

```go
type AstExprAtomBase struct {
	AstExprBase
}
```


#### type AstExprBase

```go
type AstExprBase struct {
	AstBaseTokens
}
```


#### func (*AstExprBase) ExprBase

```go
func (me *AstExprBase) ExprBase() *AstExprBase
```

#### func (*AstExprBase) IsOp

```go
func (me *AstExprBase) IsOp(anyOf ...string) bool
```

#### type AstExprCase

```go
type AstExprCase struct {
	AstExprBase
	Scrutinee IAstExpr
	Alts      []AstCaseAlt
}
```


#### func (*AstExprCase) Default

```go
func (me *AstExprCase) Default() *AstCaseAlt
```

#### func (*AstExprCase) Description

```go
func (me *AstExprCase) Description() string
```

#### type AstExprLet

```go
type AstExprLet struct {
	AstExprBase
	Defs []IAstDef
	Body IAstExpr
}
```


#### func (*AstExprLet) Description

```go
func (me *AstExprLet) Description() string
```

#### type AstExprLitBase

```go
type AstExprLitBase struct {
	AstExprAtomBase
}
```


#### type AstExprLitFloat

```go
type AstExprLitFloat struct {
	AstExprLitBase
	Val float64
}
```


#### func (*AstExprLitFloat) Description

```go
func (me *AstExprLitFloat) Description() string
```

#### type AstExprLitRune

```go
type AstExprLitRune struct {
	AstExprLitBase
	Val rune
}
```


#### func (*AstExprLitRune) Description

```go
func (me *AstExprLitRune) Description() string
```

#### type AstExprLitStr

```go
type AstExprLitStr struct {
	AstExprLitBase
	Val string
}
```


#### func (*AstExprLitStr) Description

```go
func (me *AstExprLitStr) Description() string
```

#### type AstExprLitUint

```go
type AstExprLitUint struct {
	AstExprLitBase
	Val uint64
}
```


#### func (*AstExprLitUint) Description

```go
func (me *AstExprLitUint) Description() string
```

#### type AstFile

```go
type AstFile struct {
	TopLevel []AstFileTopLevelChunk

	LastLoad struct {
		Src  []byte
		Time int64
		Size int64
	}

	Options struct {
		ApplStyle ApplStyle
	}
	SrcFilePath string
}
```


#### func (*AstFile) Err

```go
func (me *AstFile) Err() error
```

#### func (*AstFile) Errs

```go
func (me *AstFile) Errs() []error
```

#### func (*AstFile) LexAndParseFile

```go
func (me *AstFile) LexAndParseFile(onlyIfModifiedSinceLastLoad bool, stdinIfNoSrcFilePathSet bool)
```

#### func (*AstFile) LexAndParseSrc

```go
func (me *AstFile) LexAndParseSrc(r io.Reader)
```

#### func (*AstFile) Print

```go
func (me *AstFile) Print(pf IPrintFormatter) (formattedSrc []byte, err error)
```

#### func (*AstFile) Src

```go
func (me *AstFile) Src() []byte
```

#### func (*AstFile) Tokens

```go
func (me *AstFile) Tokens() udevlex.Tokens
```

#### type AstFileTopLevelChunk

```go
type AstFileTopLevelChunk struct {
	Ast AstTopLevel
}
```


#### type AstIdent

```go
type AstIdent struct {
	AstExprAtomBase
}
```


#### func (*AstIdent) BeginsLower

```go
func (me *AstIdent) BeginsLower() bool
```

#### func (*AstIdent) BeginsUpper

```go
func (me *AstIdent) BeginsUpper() bool
```

#### func (*AstIdent) Description

```go
func (me *AstIdent) Description() string
```

#### type AstTopLevel

```go
type AstTopLevel struct {
	AstBaseTokens
	AstBaseComments
	Def IAstDef
}
```


#### type AstTypeExprAppl

```go
type AstTypeExprAppl struct {
	AstTypeExprBase
	Callee IAstExpr
	Args   []IAstExpr
}
```


#### func (*AstTypeExprAppl) Description

```go
func (me *AstTypeExprAppl) Description() string
```

#### type AstTypeExprBase

```go
type AstTypeExprBase struct {
	AstExprBase
	Meta []IAstExpr
}
```


#### func (*AstTypeExprBase) TypeExprBase

```go
func (me *AstTypeExprBase) TypeExprBase() *AstTypeExprBase
```

#### type AstTypeExprIdent

```go
type AstTypeExprIdent struct {
	AstTypeExprBase
}
```


#### func (*AstTypeExprIdent) Description

```go
func (me *AstTypeExprIdent) Description() string
```

#### type AstTypeExprRec

```go
type AstTypeExprRec struct {
	AstTypeExprBase
	Names []AstIdent
	Exprs []IAstTypeExpr
}
```


#### func (*AstTypeExprRec) Description

```go
func (me *AstTypeExprRec) Description() string
```

#### type CtxPrint

```go
type CtxPrint struct {
	IPrintFormatter
	File        *AstFile
	CurTopLevel *AstTopLevel

	ustd.BytesWriter
}
```


#### type Error

```go
type Error struct {
	Msg string
	Pos scanner.Position
	Len int
	Cat ErrorCategory
}
```


#### func (*Error) Error

```go
func (me *Error) Error() (msg string)
```

#### type ErrorCategory

```go
type ErrorCategory int
```


```go
const (
	ErrCatSyntax ErrorCategory
)
```

#### type IAstDef

```go
type IAstDef interface {
	IAstNode
	DefBase() *AstDefBase
	// contains filtered or unexported methods
}
```


#### type IAstExpr

```go
type IAstExpr interface {
	ExprBase() *AstExprBase
	Description() string
}
```


#### type IAstIdent

```go
type IAstIdent interface {
	IAstExpr
	IsOpish() bool
	BeginsUpper() bool
	BeginsLower() bool
}
```


#### type IAstNode

```go
type IAstNode interface {
	// contains filtered or unexported methods
}
```


#### type IAstTypeExpr

```go
type IAstTypeExpr interface {
	IAstExpr
	TypeExprBase() *AstTypeExprBase
}
```


#### type IPrintFormatter

```go
type IPrintFormatter interface {
}
```


#### type PrintFormatterBase

```go
type PrintFormatterBase struct {
}
```


#### type PrintFormatterMinimal

```go
type PrintFormatterMinimal struct {
	PrintFormatterBase
}
```
