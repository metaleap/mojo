# atemlang
--
    import "github.com/metaleap/atem/lang"


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
	Comments []AstComment
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
	Conds []IAstExpr
	Body  IAstExpr
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


#### type AstDef

```go
type AstDef struct {
	AstBaseTokens
	Name       AstIdent
	Args       []AstIdent
	Meta       []IAstExpr
	Body       IAstExpr
	IsTopLevel bool
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


#### type AstExprAtomBase

```go
type AstExprAtomBase struct {
	AstExprBase
	AstBaseComments
}
```


#### type AstExprBase

```go
type AstExprBase struct {
	AstBaseTokens
}
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

#### type AstExprLet

```go
type AstExprLet struct {
	AstExprBase
	Defs []AstDef
	Body IAstExpr
}
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


#### type AstExprLitRune

```go
type AstExprLitRune struct {
	AstExprLitBase
	Val rune
}
```


#### type AstExprLitStr

```go
type AstExprLitStr struct {
	AstExprLitBase
	Val string
}
```


#### type AstExprLitUint

```go
type AstExprLitUint struct {
	AstExprLitBase
	Val uint64
}
```


#### type AstFile

```go
type AstFile struct {
	TopLevel []AstFileTopLevelChunk

	LastLoad struct {
		Src  []byte
		Time int64
	}

	Options struct {
		ApplStyle ApplStyle
	}
	SrcFilePath string
}
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
func (me *AstFile) Print(pf IPrintFormatter) []byte
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


#### type AstFiles

```go
type AstFiles []AstFile
```


#### func (AstFiles) Contains

```go
func (me AstFiles) Contains(srcFilePath string) bool
```

#### func (*AstFiles) RemoveAt

```go
func (me *AstFiles) RemoveAt(i int)
```

#### type AstIdent

```go
type AstIdent struct {
	AstExprAtomBase
	Val     string
	Affix   string
	IsOpish bool
	IsTag   bool
}
```


#### type AstTopLevel

```go
type AstTopLevel struct {
	AstBaseTokens
	AstBaseComments
	Def             *AstDef
	DefIsUnexported bool
}
```


#### type CtxPrint

```go
type CtxPrint struct {
	IPrintFormatter
	File           *AstFile
	CurTopLevel    *AstTopLevel
	CurIndentLevel int

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

#### type IAstExpr

```go
type IAstExpr interface {
	IAstNode
}
```


#### type IAstNode

```go
type IAstNode interface {
	// contains filtered or unexported methods
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
