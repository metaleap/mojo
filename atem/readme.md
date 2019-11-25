# atem
--
    import "github.com/metaleap/atmo/atem"


## Usage

```go
var OpPrtDst io.Writer = os.Stderr
```

#### func  Bytes

```go
func Bytes(maybeNumList []Expr) (retNumListAsBytes []byte)
```

#### type Expr

```go
type Expr interface{ String() string }
```


#### type ExprArgRef

```go
type ExprArgRef int
```


#### func (ExprArgRef) String

```go
func (me ExprArgRef) String() string
```

#### type ExprCall

```go
type ExprCall struct {
	Callee Expr
	Arg    Expr
}
```


#### func (ExprCall) String

```go
func (me ExprCall) String() string
```

#### type ExprFuncRef

```go
type ExprFuncRef int
```


```go
const (
	StdFuncId    ExprFuncRef = 0
	StdFuncTrue  ExprFuncRef = 1
	StdFuncFalse ExprFuncRef = 2
	StdFuncNil   ExprFuncRef = 3
	StdFuncCons  ExprFuncRef = 4
)
```

#### func (ExprFuncRef) String

```go
func (me ExprFuncRef) String() string
```

#### type ExprNumInt

```go
type ExprNumInt int
```


#### func (ExprNumInt) String

```go
func (me ExprNumInt) String() string
```

#### type FuncDef

```go
type FuncDef struct {
	Args []int
	Body Expr
}
```


#### func (*FuncDef) String

```go
func (me *FuncDef) String() string
```

#### type OpCode

```go
type OpCode int
```


```go
const (
	OpAdd OpCode = -1
	OpSub OpCode = -2
	OpMul OpCode = -3
	OpDiv OpCode = -4
	OpMod OpCode = -5
	OpEq  OpCode = -6
	OpLt  OpCode = -7
	OpGt  OpCode = -8
	OpPrt OpCode = -42
)
```

#### type Prog

```go
type Prog []FuncDef
```


#### func  LoadFromJson

```go
func LoadFromJson(src []byte) Prog
```

#### func (Prog) Eval

```go
func (me Prog) Eval(expr Expr, stack []Expr) Expr
```

#### func (Prog) ExprList

```go
func (me Prog) ExprList(expr Expr) (ret []Expr)
```

#### func (Prog) ExprString

```go
func (me Prog) ExprString(expr Expr) string
```

#### func (Prog) String

```go
func (me Prog) String() string
```
