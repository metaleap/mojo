# atmo
--
    import "github.com/metaleap/atmo"


## Usage

```go
const (
	EnvVarKitsDirs = "ATMO_KITS_DIRS"
	NameAutoKit    = "Std"
	SrcFileExt     = ".at"

	KnownIdentDecl   = ":="
	KnownIdentCoerce = "§"
	KnownIdentOpOr   = "or"
	KnownIdentUndef  = "÷0"
	KnownIdentEq     = "=="
)
```

```go
var (
	// ∈ aka "exists" for any maps-that-are-sets
	Є       = Exist{}
	Options struct {
		// sorted errors, kits, source files, defs etc:
		// should be enabled for consistency in user-facing tools such as REPLs or language servers.
		// could remain off for mere script runners, transpilers etc.
		Sorts bool
	}
)
```

#### func  ErrFauxPos

```go
func ErrFauxPos(maybeSrcFilePath string) (pos *udevlex.Pos)
```

#### func  SortMaybe

```go
func SortMaybe(s sort.Interface)
```

#### type Error

```go
type Error struct {
}
```


#### func  ErrAtPos

```go
func ErrAtPos(cat ErrorCategory, code int, pos *udevlex.Pos, length int, msg string) (err *Error)
```

#### func  ErrAtTok

```go
func ErrAtTok(cat ErrorCategory, code int, tok *udevlex.Token, msg string) *Error
```

#### func  ErrAtToks

```go
func ErrAtToks(cat ErrorCategory, code int, toks udevlex.Tokens, msg string) *Error
```

#### func  ErrBug

```go
func ErrBug(code int, toks udevlex.Tokens, msg string) *Error
```

#### func  ErrFrom

```go
func ErrFrom(cat ErrorCategory, code int, maybeSrcFilePath string, err error) *Error
```

#### func  ErrLex

```go
func ErrLex(code int, pos *udevlex.Pos, msg string) *Error
```

#### func  ErrNaming

```go
func ErrNaming(code int, tok *udevlex.Token, msg string) *Error
```

#### func  ErrPreduce

```go
func ErrPreduce(code int, toks udevlex.Tokens, msg string) *Error
```

#### func  ErrRef

```go
func ErrRef(err *Error) *Error
```

#### func  ErrSess

```go
func ErrSess(code int, maybePath string, msg string) *Error
```

#### func  ErrSyn

```go
func ErrSyn(code int, tok *udevlex.Token, msg string) *Error
```

#### func  ErrTodo

```go
func ErrTodo(code int, toks udevlex.Tokens, msg string) *Error
```

#### func  ErrUnreach

```go
func ErrUnreach(code int, toks udevlex.Tokens, msg string) *Error
```

#### func (*Error) Cat

```go
func (me *Error) Cat() ErrorCategory
```

#### func (*Error) Code

```go
func (me *Error) Code() int
```

#### func (*Error) CodeAndCat

```go
func (me *Error) CodeAndCat() string
```

#### func (*Error) Error

```go
func (me *Error) Error() string
```

#### func (*Error) IsRef

```go
func (me *Error) IsRef() bool
```

#### func (*Error) Len

```go
func (me *Error) Len() int
```

#### func (*Error) Msg

```go
func (me *Error) Msg() string
```

#### func (*Error) Pos

```go
func (me *Error) Pos() *udevlex.Pos
```

#### func (*Error) UpdatePosOffsets

```go
func (me *Error) UpdatePosOffsets(offsets IErrPosOffsets)
```

#### type ErrorCategory

```go
type ErrorCategory int
```


```go
const (
	ErrCatBug ErrorCategory
	ErrCatTodo
	ErrCatLexing
	ErrCatParsing
	ErrCatNaming
	ErrCatPreduce
	ErrCatSess
	ErrCatUnreachable
)
```

#### func (ErrorCategory) String

```go
func (me ErrorCategory) String() string
```

#### type Errors

```go
type Errors []*Error
```


#### func (*Errors) Add

```go
func (me *Errors) Add(errs ...*Error) (anyAdded bool)
```

#### func (*Errors) AddAt

```go
func (me *Errors) AddAt(cat ErrorCategory, code int, pos *udevlex.Pos, length int, msg string) *Error
```

#### func (*Errors) AddBug

```go
func (me *Errors) AddBug(code int, toks udevlex.Tokens, msg string) *Error
```

#### func (*Errors) AddFromTok

```go
func (me *Errors) AddFromTok(cat ErrorCategory, code int, tok *udevlex.Token, msg string) *Error
```

#### func (*Errors) AddFromToks

```go
func (me *Errors) AddFromToks(cat ErrorCategory, code int, toks udevlex.Tokens, msg string) *Error
```

#### func (*Errors) AddLex

```go
func (me *Errors) AddLex(code int, pos *udevlex.Pos, msg string) *Error
```

#### func (*Errors) AddNaming

```go
func (me *Errors) AddNaming(code int, tok *udevlex.Token, msg string) *Error
```

#### func (*Errors) AddPreduce

```go
func (me *Errors) AddPreduce(code int, toks udevlex.Tokens, msg string) *Error
```

#### func (*Errors) AddSess

```go
func (me *Errors) AddSess(code int, maybePath string, msg string) *Error
```

#### func (*Errors) AddSyn

```go
func (me *Errors) AddSyn(code int, toks udevlex.Tokens, msg string) *Error
```

#### func (*Errors) AddTodo

```go
func (me *Errors) AddTodo(code int, toks udevlex.Tokens, msg string) *Error
```

#### func (*Errors) AddUnreach

```go
func (me *Errors) AddUnreach(code int, toks udevlex.Tokens, msg string) *Error
```

#### func (*Errors) AddVia

```go
func (me *Errors) AddVia(v interface{}, errs Errors) interface{}
```

#### func (Errors) Len

```go
func (me Errors) Len() int
```
Len implements `sort.Interface`.

#### func (Errors) Less

```go
func (me Errors) Less(i int, j int) bool
```
Less implements `sort.Interface`.

#### func (Errors) Refs

```go
func (me Errors) Refs() (refs Errors)
```

#### func (Errors) Strings

```go
func (me Errors) Strings() (s []string)
```

#### func (Errors) Swap

```go
func (me Errors) Swap(i int, j int)
```
Swap implements `sort.Interface`.

#### func (Errors) UpdatePosOffsets

```go
func (me Errors) UpdatePosOffsets(offsets IErrPosOffsets)
```

#### type Exist

```go
type Exist struct{}
```


#### type IErrPosOffsets

```go
type IErrPosOffsets interface {
	PosOffsetLine() int
	PosOffsetByte() int
}
```


#### type StringKeys

```go
type StringKeys map[string]Exist
```


#### func (StringKeys) Exists

```go
func (me StringKeys) Exists(s string) (ok bool)
```

#### func (StringKeys) SortedBy

```go
func (me StringKeys) SortedBy(isLessThan func(string, string) bool) (sorted []string)
```

#### func (StringKeys) String

```go
func (me StringKeys) String() (s string)
```
