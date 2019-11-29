# atem
--
    import "github.com/metaleap/atmo/atem"

_atem_ is both a minimal and low-level interpreted functional intermediate
language and its reference interpreter. It prioritizes staying low-LoC enough to
be able to port it to any other tech stack swiftly, over other concerns. At the
time of writing, the "parsing" / loading in this Go-based implementation is ~42
LoCs (the choice of a JSON code format too is motivated by the goal to allow for
swift re-implementations in any contemporary or future lang / tech stack), the
interpreting / eval'ing parts around ~55 LoCs, AST node type formulations and
their `JsonSrc()` / `ToJson()` implementations around ~45 LoCs, and helpers for
forcing "`Eval` result list-closures" into actual `[]int` or `[]byte` slices or
`string`s, another ~40 LoCs. All counts approximate and net (excluding comments,
blank lines etc).

This focus doesn't make for the most efficient interpreter in the world, but
that isn't the objective for _atem_. The goal is to provide the bootstrapping
basis for **atmo**. An initial compiler from _atmo_ to _atem_ is being coded in
my [toy Lambda Calculus
dialect](https://github.com/metaleap/go-machines/tree/master/toylam) and then
again in (the initial iteration of) _atmo_ itself. The _atem_ interpreter will
also suffice / go a long way for REPL purposes and later on abstract / symbolic
interpretation / partial evaluation for the experimental type-checking
approaches envisioned to be explored within the impending ongoing evolution of
_atmo_ once its initial incarnation is birthed.

For "ultimate real-world runtime artifact" purposes, _atem_ isn't intended;
rather, transpilers and / or compilers to 3rd-party mature and widely enjoyed
interpreters / bytecode VMs or intermediate ASM targets like LLVM-IR would be
the envisioned generally-preferable direction anyway, except such trans-/
compilers must naturally be done in _atmo_ as well, so _atem_ is way to get from
nowhere to _there_, and to be _able_ (not forced) to replicate this original
bootstrapping on any sort of tech base at any time whenever necessary.

The initial inspiration / iteration for _atem_ was the elegantly minimalist
[SAPL](https://github.com/metaleap/go-machines/tree/master/sapl) approach
presented by Jansen / Koopman / Plasmeijer, but unlike the above-linked
"by-the-paper" implementation, _atem_ diverges even in its initial form in
various aspects and will continue to evolve various details in tandem with the
birthing of _atmo_.

SAPL's basics still apply for now: all funcs are top-level (no lambdas or other
locals), as such support 0 - n args (rather than all-unary). There are no names:
global funcs and, inside them, their args are referred to by integer indices.
Thus most expressions are atomic: arg-refs, func-refs, and plain integers. The
only non-atomic expression is call / application: it is composed of two
sub-expressions, the callee and the arg. Divergences: our func-refs, if
negative, denote a binary primitive-instruction op-code such as addition,
multiply, equality-testing etc. that is handled natively by the interpreter.
Unlike SAPL, our func-refs don't carry around their number-of-args, instead
they're looked up in the `Prog`. For applications / calls, likely will move from
the current unary style to n-ary, if feasible without breaking
partial-application or degrading our overall LoCs aims.

## Usage

```go
var OpPrtDst io.Writer = os.Stderr
```
OpPrtDst is the output destination for all `OpPrt` primitive instructions. Must
never be `nil` during any `Prog`s that do potentially invoke `OpPrt`.

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
	// JsonSrc emits the re-`LoadFromJson`able representation of this `Expr`.
	JsonSrc() string
}
```


#### func  ListFrom

```go
func ListFrom(str []byte) (ret Expr)
```
ListFrom converts the specified byte string to a linked-list representing a text
string during `Eval` (via `ExprAppl`s of `StdFuncCons` and `StdFuncNil`).

#### func  ListsFrom

```go
func ListsFrom(strs []string) (ret Expr)
```
ListsFrom creates from `strs` linked-lists via `ListFrom`, and returns a
linked-list of those.

#### type ExprAppl

```go
type ExprAppl struct {
	Callee Expr
	Arg    Expr
}
```


#### func (ExprAppl) JsonSrc

```go
func (me ExprAppl) JsonSrc() string
```
JsonSrc emits the re-`LoadFromJson`able representation of this `ExprAppl`.

#### type ExprArgRef

```go
type ExprArgRef int
```


#### func (ExprArgRef) JsonSrc

```go
func (me ExprArgRef) JsonSrc() string
```
JsonSrc emits a non-re-`LoadFromJson`able representation of this `ExprArgRef`.

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
	// K I aka. boolish of false
	StdFuncFalse ExprFuncRef = 2
	// end of linked-list
	StdFuncNil ExprFuncRef = 3
	// next link in linked-list
	StdFuncCons ExprFuncRef = 4
)
```
The few standard func defs the interpreter needs to know of as a minimum, and
their inviolably hereby-decreed standard indices within a `Prog`. Every atem
code generator must emit implementations for them all, and placed correctly.

#### func (ExprFuncRef) JsonSrc

```go
func (me ExprFuncRef) JsonSrc() string
```
JsonSrc emits the re-`LoadFromJson`able representation of this `ExprFuncRef`.

#### type ExprNumInt

```go
type ExprNumInt int
```


#### func (ExprNumInt) JsonSrc

```go
func (me ExprNumInt) JsonSrc() string
```
JsonSrc emits the re-`LoadFromJson`able representation of this `ExprNumInt`.

#### type FuncDef

```go
type FuncDef struct {
	// Args holds this `FuncDef`'s arguments: each `int` denotes how often the
	// `Body` references this arg (note that the interpreter does not currently
	// use this info), the arg's "identity" however is just its index in `Args`
	Args          []int
	Body          Expr
	OrigNameMaybe string
}
```


#### func (*FuncDef) ToJson

```go
func (me *FuncDef) ToJson() string
```
ToJson emits the re-`LoadFromJson`able representation of this `FuncDef`. (It's
not called `JsonSrc` in order to make clear that `FuncDef` is not an `Expr`
implementer.)

#### type OpCode

```go
type OpCode int
```

OpCode denotes a "primitive instruction", eg. one that is hardcoded in the
interpreter and invoked when encountering a call to a negative `ExprFuncRef`
with at least 2 operands on the current `Eval` stack. All `OpCode`-denoted
primitive instructions consume always exactly 2 operands from said stack.

```go
const (
	// Addition of 2 `ExprNumInt`s, result 1 `ExprNumInt`
	OpAdd OpCode = -1
	// Subtraction of 2 `ExprNumInt`s, result 1 `ExprNumInt`
	OpSub OpCode = -2
	// Multiplication of 2 `ExprNumInt`s, result 1 `ExprNumInt`
	OpMul OpCode = -3
	// Division of 2 `ExprNumInt`s, result 1 `ExprNumInt`
	OpDiv OpCode = -4
	// Modulo of 2 `ExprNumInt`s, result 1 `ExprNumInt`
	OpMod OpCode = -5
	// Equality test between 2 `Expr`s, result is `StdFuncTrue` or `StdFuncFalse`
	OpEq OpCode = -6
	// Less-than test between 2 `Expr`s, result is `StdFuncTrue` or `StdFuncFalse`
	OpLt OpCode = -7
	// Greater-than test between 2 `Expr`s, result is `StdFuncTrue` or `StdFuncFalse`
	OpGt OpCode = -8
	// Writes both `Expr`s (the first one a string-ish `StdFuncCons`tructed linked-list of `ExprNumInt`s) to `OpPrtDst`, result is the right-hand-side `Expr` of the 2 input `Expr` operands
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
`Expr` implementer's `JsonSrc` method implementation, meaning: `ExprNumInt` is a
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
is met, the primitive instruction is carried out, its `Expr` result then being
`Eval`'d with the reduced-by-2 `stack`. Unknown op-codes `panic` with a
`[3]Expr` of first the `ExprFuncRef` followed by both its operands.

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
result, if it is one. Otherwise, `expr.JsonSrc()` is returned for convenience.

#### func (Prog) ToJson

```go
func (me Prog) ToJson() string
```
ToJson emits the re-`LoadFromJson`able representation of this `Prog`. (It's not
called `JsonSrc` in order to make clear that `Prog` is not an `Expr`
implementer.)
