# atem
--
    import "github.com/metaleap/atmo/atem"

atem is a minimal and low-level interpretable functional intermediate language
and interpreter. It prioritizes staying low-LoC enough to be able to rewrite it
on any other tech stack any time, over other concerns. At the time of writing,
the core "parsing" / loading in this Go-based implementation is ~40 LoCs (the
choice of a JSON code format is similarly motivated by the goal to allow for
swift re-implementations in any contemporary or future lang / tech stack), the
core interpretation / eval'ing parts around ~55 LoCs, basic AST node types and
their `String()` implementations around ~40 LoCs, and helpers for injecting or
extracting "lists of ints" (strings at the end of the day) into / from run time
another ~40 LoCs. All counts approximate and net, excluding comments.

This focus doesn't make for the most efficient interpreter in the world, but
that isn't the objective for atem. The goal is to provide the bootstrapping
layer for **atmo**. An initial compiler from atmo to atem is being coded in my
[toy Lambda Calculus
dialect](https://github.com/metaleap/go-machines/tree/master/toylam) and then
again in (the initial iteration of) atmo itself. The atem interpreter will also
suffice / go a long way for REPL purposes and later on abstract / symbolic
interpretation / partial evaluation for the experimental type-checking
approaches envisioned to be explored within the impending ongoing evolution of
atmo once its initial incarnation is birthed.

For "ultimate real-world runtime artifact" purposes, atem isn't intended;
rather, transpilers and / or compilers to 3rd-party mature and widely enjoyed
interpreters / bytecode VMs or intermediate ASM targets such as LLVM would be
the envisioned generally-preferable direction anyway, except such trans-/
compilers want to naturally be done in atmo as well, so atem is way to get from
nowhere to _there_, and to be _able_ (not forced) to replicate this original
bootstrapping on any sort of tech base at any time whenever necessary.

The initial inspiration (and iteration) for atem was the elegant and minimalist
[SAPL](https://github.com/metaleap/go-machines/tree/master/sapl) approach
presented by Jansen / Koopman / Plasmeijer, but unlike the above-linked
by-the-paper implementation, atem diverges even in its initial form in various
aspects and will continue to evolve various details in tandem with the birthing
of atmo.

SAPL's basics still apply for now: all funcs are top-level (no lambdas or other
locals), as such support 0 - n args (rather than all-unary). There are no names:
global funcs and, inside them, their args are referred to by integer indices.
Thus most expressions are atomic: arg-refs, func-refs, and plain integers. The
only non-atomic expression is call / application: it is composed of two
sub-expressions, the callee and the arg. Divergences: our func-refs, if
negative, denote a binary primitive-instruction op-code such as `ADD` etc. that
is handled natively by the interpreter. Unlike SAPL, our func-refs don't carry
around their number-of-args, instead they're looked up in the `Prog`. For calls
/ applications, likely will move from the current unary style to n-ary for
efficiency reasons, without breaking partial-application of course, or degrading
our overall LoCs aims unduly.

## Usage

```go
var OpPrtDst io.Writer = os.Stderr
```

#### func  ListToBytes

```go
func ListToBytes(maybeNumList []Expr) (retNumListAsBytes []byte)
```
ListToBytes examines the given `[]Expr`, as normally obtained via
`Prog.ListOfExprs` and accumulates a `[]byte` slice as long as all elements in
said list are `ExprNumInt` values in the range 0 - 255. If the input is `nil`,
so will be `retNumListAsBytes`. If the input has a `len` of zero, so will
`retNumListAsBytes`. If any of the input `Expr`s isn't an in-range `ExprNumInt`,
then too will `retNumListAsBytes` be `nil`.

#### type Expr

```go
type Expr interface {
	// String emits the re-`LoadFromJson`able representation of this `Expr`.
	String() string
}
```


#### type ExprAppl

```go
type ExprAppl struct {
	Callee Expr
	Arg    Expr
}
```


#### func (ExprAppl) String

```go
func (me ExprAppl) String() string
```
String emits the re-`LoadFromJson`able representation of this `ExprAppl`.

#### type ExprArgRef

```go
type ExprArgRef int
```


#### func (ExprArgRef) String

```go
func (me ExprArgRef) String() string
```
String emits the re-`LoadFromJson`able representation of this `ExprArgRef`.

#### type ExprFuncRef

```go
type ExprFuncRef int
```


```go
const (
	// I combinator aka identity function
	StdFuncId ExprFuncRef = 0
	// K combinator aka konst aka boolish of true
	StdFuncTrue ExprFuncRef = 1
	// boolish of false
	StdFuncFalse ExprFuncRef = 2
	// end of linked-list
	StdFuncNil ExprFuncRef = 3
	// next link in linked-list
	StdFuncCons ExprFuncRef = 4
)
```
The few standard func defs the interpreter needs to know of as a minimum, and
their inviolably hereby-prescribed standard indexes within a `Prog`. Every atem
code generator must emit implementations for them all, and placed correctly.

#### func (ExprFuncRef) String

```go
func (me ExprFuncRef) String() string
```
String emits the re-`LoadFromJson`able representation of this `ExprFuncRef`.

#### type ExprNumInt

```go
type ExprNumInt int
```


#### func (ExprNumInt) String

```go
func (me ExprNumInt) String() string
```
String emits the re-`LoadFromJson`able representation of this `ExprNumInt`.

#### type FuncDef

```go
type FuncDef struct {
	// Args holds this `FuncDef`'s arguments: each `int` denotes how often the
	// `Body` references this arg (although the interpreter only cares about
	// 0 or greater), the arg's "identity" however is just its index in `Args`
	Args          []int
	Body          Expr
	OrigNameMaybe string
}
```


#### func (*FuncDef) String

```go
func (me *FuncDef) String() string
```
String emits the re-`LoadFromJson`able representation of this `FuncDef`.

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
LoadFromJson parses and decodes a JSON `src` into an atem `Prog`. The format is
expected to be: `[ func, func, ... , func ]` where `func` means: ` [ args, body
]` where `args` is a numbers array and `body` is the reverse of each concrete
`Expr` implementer's `String` method implementation, meaning: `ExprNumInt` is a
JSON number, `ExprFuncRef` is a length-1 numbers array, `ExprArgRef` is a JSON
string parseable into an integer, and `ExprAppl` is a variable length (greater
than 1) array of any of those possibilities. A `panic` occurs on any sort of
error encountered from the input `src`.

#### func (Prog) Eval

```go
func (me Prog) Eval(expr Expr, stack []Expr) Expr
```
Eval operates thusly:

- encountering an `ExprAppl`, its `Arg` is `append`ed to the `stack` and its
`Callee` is then `Eval`'d;

- encountering an `ExprFuncRef`, the `stack` is checked for having the proper
minimum required `len` with regard to the referenced `FuncDef`'s number of
`Args`. If okay, the pertinent number of args is taken (and removed) from the
`stack` and the referenced `FuncDef`'s `Body`, rewritten with all inner
`ExprArgRef`s (including those inside `ExprAppl`s) resolved to the `stack`
entries, is `Eval`'d (with the appropriately reduced `stack`);

- encountering any other `Expr` type, it is merely returned.

Corner cases for the `ExprFuncRef` situation: if the `stack` has too small a
`len`, an `ExprAppl` representing the partial application is returned; if the
`ExprFuncRef` is negative and thus referring to a primitive-instruction
`OpCode`, the expected minimum required `len` for the `stack` is 2 and if this
is met, the primitive instruction is carried out. For "boolish" prim ops, namely
`OpEq`, `OpGt`, `OpLt` the resulting `StdFuncFalse` or `StdFuncTrue` is directly
applied (as described above) with the remainder of the `stack`. For "integer"
prim ops (`OpAdd`, `OpSub` etc.) the 2 operands are forced into `ExprNumInt`s
and an `ExprNumInt` result will be returned. For `OpPrt`, the side-effect write
of both operands to `OpPrtDst` is performed and the second operand is then
returned. Other / unknown op-codes `panic` with a `[3]Expr` of first the
`ExprFuncRef` followed by both its operands.

#### func (Prog) ListOfExprs

```go
func (me Prog) ListOfExprs(expr Expr) (ret []Expr)
```
ListOfExprs dissects the given `expr` into an `[]Expr` slice only if it is a
closure resulting from `StdFuncCons` / `StdFuncNil` usage during `Eval`. The
individual element `Expr`s are not themselves scrutinized however. The `ret` is
`return`ed as `nil` if `expr` isn't a product of `StdFuncCons` / `StdFuncNil`
usage; yet a non-`nil`, zero-`len` `ret` will result from a mere `StdFuncNil`
construction, aka. "empty linked-list value" `Expr`.

#### func (Prog) ListOfExprsToString

```go
func (me Prog) ListOfExprsToString(expr Expr) string
```
ListOfExprsToString is a wrapper around the combined usage of `Prog.ListOfExprs`
and `ListToBytes` to extract the List-closure-encoded `string` of an `Eval`
result.

#### func (Prog) String

```go
func (me Prog) String() string
```
String emits the re-`LoadFromJson`able representation of this `Prog`.
