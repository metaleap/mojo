# atem
--
    import "github.com/metaleap/atmo/old/atem"

_atem_ is both a minimal and low-level interpreted functional programming
language IR (_intermediate representation_, ie. not to be hand-written) and its
reference interpreter implementation (in lib form). It prioritizes staying
low-LoC enough to be able to port it over to any other current or future lang /
tech stack swiftly and trivially, over other concerns, by design. The choice of
a JSON code format is likewise motivated by the stated "no-brainer, low-effort
portability" objective.

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

For "real-world production runtime artifact" purposes, _atem_ isn't intended;
rather, transpilers and / or compilers to 3rd-party mature and widely enjoyed
interpreters / bytecode VMs or intermediate ASM targets like LLVM-IR would be
the envisioned generally-preferable direction anyway, except such trans-/
compilers must naturally be done in _atmo_ as well, so _atem_ is way to get from
nowhere to _there_, and to be _able_ (not required) to replicate this original
bootstrapping on any sort of tech base at any time whenever necessary.

The initial inspiration / iteration for _atem_ was the elegantly minimalist
[SAPL](https://github.com/metaleap/go-machines/tree/master/sapl) approach
presented by Jansen / Koopman / Plasmeijer, but unlike the above-linked
"by-the-paper" implementation, _atem_ diverges even in its initial form in
various aspects and will continue to evolve various details in tandem with the
birthing of _atmo_.

Many of SAPL's design essentials still apply for now: all funcs are top-level
(no lambdas or other locals), as such support 0 - n args (rather than all-unary
as in plain lambda-calculus-representing source languages). There are no names:
global funcs and, inside them, their args are referred to by integer indices.
Thus most expression types are atomic `int`s: arg-refs, func-refs, and plain
integral numbers. The only non-atomic expression type is `ExprCall`, made of
`Callee` and `Args` (plus an `IsClosure` flag). Divergences from SAPL: our calls
are n-ary not unary; our func-refs, if negative, denote a binary
primitive-instruction op-code such as addition, multiply, equality-testing etc.
that is handled natively by the interpreter; our func-refs don't carry around
their number-of-args, instead they're looked up together with the `Body` in the
`Prog` via the indicated index. Finally, the lazy-ish evaluator approach has
been replaced with a "mostly eager-ish" interpretation approach. Meaning: the
callee is evaluated down to a callable first, then args to a call that are
marked (in source) as unused are not evaluated, the others are evaluated before
final consumption.

## Usage

```go
var OpPrtDst = os.Stderr.Write
```
OpPrtDst is the output sink for all `OpPrt` primitive instructions. Must never
be `nil` during any `Prog`s that do potentially invoke `OpPrt`.

#### func  Eq

```go
func Eq(expr Expr, cmp Expr) bool
```
Eq is the implementation of the `OpEq` prim-op instruction code.

#### func  ListOfExprsToString

```go
func ListOfExprsToString(expr Expr) string
```
ListOfExprsToString is a wrapper around the combined usage of `ListOfExprs` and
`ListToBytes` to extract the List-closure-encoded `string` of an `Eval` result,
if it is one. Otherwise, `expr.JsonSrc()` is returned for convenience.

#### func  ListToBytes

```go
func ListToBytes(maybeNumList []Expr) (retNumListAsBytes []byte)
```
ListToBytes examines the given `[]Expr`, as normally obtained via `ListOfExprs`
and accumulates a `[]byte` slice as long as all elements in said list are
`ExprNumInt` values in the range 0 - 255. If the input is `nil`, so will be
`retNumListAsBytes`. If the input has a `len` of zero, so will
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
string during `Eval` (via `ExprCall`s of `StdFuncCons` and `StdFuncNil`).

#### func  ListOfExprs

```go
func ListOfExprs(expr Expr) (ret []Expr)
```
ListOfExprs dissects the given `expr` into an `[]Expr` slice only if it is a
closure resulting from `StdFuncCons` / `StdFuncNil` usage during `Eval`. The
individual element `Expr`s are not themselves scrutinized however. The `ret` is
`return`ed as `nil` if `expr` isn't a product of `StdFuncCons` / `StdFuncNil`
usage; yet a non-`nil`, zero-`len` `ret` will result from a mere `StdFuncNil`
construction, aka. "empty linked-list value" `Expr`.

The result of `ListOfExprs` can be passed to `ListToBytes` to extract the
`string` value represented by `expr`, if any.

#### func  ListsFrom

```go
func ListsFrom(strs []string) (ret Expr)
```
ListsFrom creates from `strs` linked-lists via `ListFrom`, and returns a
linked-list of those.

#### type ExprArgRef

```go
type ExprArgRef int
```


#### func (ExprArgRef) JsonSrc

```go
func (me ExprArgRef) JsonSrc() string
```
JsonSrc implements the `Expr` interface.

#### type ExprCall

```go
type ExprCall struct {
	Callee    Expr
	Args      []Expr
	IsClosure int // determined at load time, not in input source: if `> 0` (indicating number of missing args), callee is an `ExprFuncRef` and all args are `ExprNumInt` or `ExprFuncRef` or further such `ExprCall`s with `.IsClosure > 0`
}
```


#### func (*ExprCall) JsonSrc

```go
func (me *ExprCall) JsonSrc() string
```
JsonSrc implements the `Expr` interface.

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
JsonSrc implements the `Expr` interface.

#### type ExprNumInt

```go
type ExprNumInt int
```


#### func (ExprNumInt) JsonSrc

```go
func (me ExprNumInt) JsonSrc() string
```
JsonSrc implements the `Expr` interface.

#### type FuncDef

```go
type FuncDef struct {
	// Args holds this `FuncDef`'s arguments: each `int` denotes how often the `Body`
	// references this arg, the arg's "identity" however is just its index in `Args`
	Args []int
	Body Expr
	Meta []string // ignored and not used in this lib: but still loaded from JSON and (re)emitted by `FuncDef.JsonSrc()`
}
```


#### func (*FuncDef) JsonSrc

```go
func (me *FuncDef) JsonSrc(dropFuncDefMetas bool) string
```
JsonSrc emits the re-`LoadFromJson`able representation of this `FuncDef`.

#### type OpCode

```go
type OpCode int
```

OpCode denotes a "primitive instruction", eg. one that is hardcoded in the
interpreter and invoked when encountering a call to a negative `ExprFuncRef`
supplied with two operand arguments.

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
	// Less-than test between 2 `ExprNumInt`s, result is `StdFuncTrue` or `StdFuncFalse`
	OpLt OpCode = -7
	// Greater-than test between 2 `ExprNumInt`s, result is `StdFuncTrue` or `StdFuncFalse`
	OpGt OpCode = -8
	// Writes both `Expr`s (the first one a string-ish `StdFuncCons`tructed linked-list of `ExprNumInt`s) to `OpPrtDst`, result is the right-hand-side `Expr` of the 2 input `Expr` operands
	OpPrt OpCode = -42
	// Evaluates the 2nd `Expr` with respect to the 1st. If the 1st is `StdFuncNil`, the 2nd encodes any expression to be evaluated in the context of the current `Prog`, else in the context of the `Prog` encoded by the 1st. Encoding is via `StdFuncNil` / `StdFuncCons` lists arranged just like the JSON format.
	OpEval OpCode = -4242
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
string parseable into an integer, and `ExprCall` is a variable length (greater
than 1) array of any of those possibilities. A `panic` occurs on any sort of
error encountered from the input `src`.

A note on `ExprCall`s, their `Args` orderings are on-load reversed from those
being read in or emitted back out via `JsonSrc()`. Args in the JSON format are
ordered in a common intuitive manner: `[callee, arg1, arg2, arg3]`, but an
`ExprCall` created from this will have an `Args` slice of `[arg3, arg2, arg1]`
throughout its lifetime. Still, its `JsonSrc()` emits the original ordering. If
the callee is another `ExprCall`, expect a JSON source notation of eg.
`[[callee, x, y, z], a, b, c]` to turn into a single `ExprCall` with `Args` of
[c, b, a, z, y, x], it would be re-emitted as `[callee, x, y, z, a, b, c]`.
`ExprCall.Args` and `FuncDef.Args` orderings are consistent in the JSON source
code format (when loading or emitting), but not at run time.

A note on `ExprArgRef`s: these take different forms in the JSON format and at
runtime. In the former, two intuitive-to-emit styles are supported: if positive
they denote 0-based indexing such that 0 refers to the `FuncDef`'s first arg, 1
to the second, 2 to the third etc; if negative, they're read with -1 referring
to the `FuncDef`'s last arg, -2 to the one-before-last, -3 to the
one-before-one-before-last etc. Both styles at load time are translated into a
form expected at run time, where 0 turns into -2, 1 into -3, 2 into -4 etc, for
marginally speedier call-stack accesses in the interpreter.
`ExprArgRef.JsonSrc()` will restore the 0-based indexing form, however.

#### func (Prog) Eval

```go
func (me Prog) Eval(expr Expr, big bool) Expr
```
Eval reduces `expr` to an `ExprNumInt`, an `ExprFuncRef` or a closure value (an
`*ExprCall` with `.IsClosure > 0`, see field description there), the latter can
be tested for linked-list-ness and extracted via `ListOfExprs`.

The evaluator is akin to a tree-walking interpreter of the input `Prog` but
given the nature of the `atem` intermediate-representation language, that
amounts to a sort of register machine. A call stack is kept so that `Eval` never
needs to recursively call itself. Any stack entry beyond the "root" / "base" one
(that at first holds `expr` and at the end the final result value) represents a
call: it at first holds both said call's callee and its args. The former is
evaluated first (only down to a "callable": `ExprFuncRef` or closure), next then
only those args are evaluated that are actually needed. Finally, the
"callable"'s body (or prim-op) is evaluated further, consuming those
freshly-obtained arg values while producing the call's result value. (If in a
call not enough args are supplied to the callee, the result is a closure that
does keep its fully-evaluated args around for later completion.)

The `big` arg fine-tunes how much call-stack memory to pre-allocate at once
beforehand. If `true`, this will be to the tune of ~2 MB, else under 10 KB. Put
simply, `true` is for full-program running, `false` is for smallish "drive-by" /
"side-car" expression evaluation attempts in the context of a given `Prog` such
as in REPLs, optimizers, compilers or similar tooling.

#### func (Prog) JsonSrc

```go
func (me Prog) JsonSrc(dropFuncDefMetas bool) string
```
JsonSrc emits the re-`LoadFromJson`able representation of this `Prog`.
