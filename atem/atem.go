// _atem_ is both a minimal and low-level interpreted functional programming
// language IR (_intermediate representation_, ie. not to be hand-written) and
// its reference interpreter implementation (in lib form). It prioritizes
// staying low-LoC enough to be able to port it over to any other current or
// future lang / tech stack swiftly and trivially, over other concerns, by
// design. The choice of a JSON code format is likewise motivated by the
// stated "no-brainer, low-effort portability" objective.
//
// This focus doesn't make for the most efficient interpreter in the world, but that
// isn't the objective for _atem_. The goal is to provide the bootstrapping basis
// for **atmo**. An initial compiler from _atmo_ to _atem_ is being coded in my
// [toy Lambda Calculus dialect](https://github.com/metaleap/go-machines/tree/master/toylam)
// and then again in (the initial iteration of) _atmo_ itself. The _atem_
// interpreter will also suffice / go a long way for REPL purposes and later
// on abstract / symbolic interpretation / partial evaluation for the
// experimental type-checking approaches envisioned to be explored within the
// impending ongoing evolution of _atmo_ once its initial incarnation is birthed.
//
// For "real-world production runtime artifact" purposes, _atem_ isn't intended;
// rather, transpilers and / or compilers to 3rd-party mature and widely enjoyed
// interpreters / bytecode VMs or intermediate ASM targets like LLVM-IR would
// be the envisioned generally-preferable direction anyway, except such trans-/
// compilers must naturally be done in _atmo_ as well, so _atem_ is way to get
// from nowhere to _there_, and to be _able_ (not required) to replicate this
// original bootstrapping on any sort of tech base at any time whenever necessary.
//
// The initial inspiration / iteration for _atem_ was the elegantly minimalist
// [SAPL](https://github.com/metaleap/go-machines/tree/master/sapl) approach
// presented by Jansen / Koopman / Plasmeijer, but unlike the above-linked
// "by-the-paper" implementation, _atem_ diverges even in its initial form in
// various aspects and will continue to evolve various details in tandem with
// the birthing of _atmo_.
//
// Many of SAPL's design essentials still apply for now: all funcs are top-level
// (no lambdas or other locals), as such support 0 - n args (rather than
// all-unary as in plain lambda-calculus-representing source languages). There
// are no names: global funcs and, inside them, their args are referred to by
// integer indices. Thus most expression types are atomic `int`s: arg-refs,
// func-refs, and plain integral numbers. The only non-atomic expression type
// is `ExprCall`, made of `Callee` and `Args` (plus an `IsClosure` flag).
// Divergences from SAPL: our calls are n-ary not unary; our func-refs, if
// negative, denote a binary primitive-instruction op-code such as addition,
// multiply, equality-testing etc. that is handled natively by the interpreter;
// our func-refs don't carry around their number-of-args, instead they're
// looked up together with the `Body` in the `Prog` via the indicated index.
// Finally, the lazy-ish evaluator approach has been replaced with a "mostly
// eager-ish" interpretation approach. Meaning: the callee is evaluated down
// to a callable first, then args to a call that are marked (in source) as
// unused are not evaluated, the others are evaluated before final consumption.
package atem

import (
	"os"
	"strconv"
)

// The few standard func defs the interpreter needs to know of as a minimum, and
// their inviolably hereby-decreed standard indices within a `Prog`. Every atem
// code generator must emit implementations for them all, and placed correctly.
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

type (
	Prog    []FuncDef
	FuncDef struct {
		// Args holds this `FuncDef`'s arguments: each `int` denotes how often the `Body`
		// references this arg, the arg's "identity" however is just its index in `Args`
		Args        []int
		Body        Expr
		Meta        []string // ignored and not used in this lib: but still loaded from JSON and (re)emitted by `FuncDef.JsonSrc()`
		selector    int
		allArgsUsed bool
		isMereAlias bool
	}
	Expr interface {
		// JsonSrc emits the re-`LoadFromJson`able representation of this `Expr`.
		JsonSrc() string
	}
	ExprNumInt  int
	ExprArgRef  int
	ExprFuncRef int
	ExprCall    struct {
		Callee    Expr
		Args      []Expr
		IsClosure int // determined at load time, not in input source: if `> 0` (indicating number of missing args), callee is an `ExprFuncRef` and all args are `ExprNumInt` or `ExprFuncRef` or further such `ExprCall`s with `.IsClosure > 0`
	}

	// OpCode denotes a "primitive instruction", eg. one that is hardcoded in
	// the interpreter and invoked when encountering a call to a negative
	// `ExprFuncRef` supplied with two operand arguments.
	OpCode int
)

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
)

// OpPrtDst is the output sink for all `OpPrt` primitive instructions.
// Must never be `nil` during any `Prog`s that do potentially invoke `OpPrt`.
var OpPrtDst = os.Stderr.Write

// JsonSrc implements the `Expr` interface.
func (me ExprNumInt) JsonSrc() string { return strconv.Itoa(int(me)) }

// JsonSrc implements the `Expr` interface.
func (me ExprArgRef) JsonSrc() string { return "\"" + strconv.Itoa(int(-me)-2) + "\"" }

// JsonSrc implements the `Expr` interface.
func (me ExprFuncRef) JsonSrc() string { return "[" + strconv.Itoa(int(me)) + "]" }

// JsonSrc implements the `Expr` interface.
func (me *ExprCall) JsonSrc() string {
	ret := "[" + me.Callee.JsonSrc()
	for i := len(me.Args) - 1; i > -1; i-- {
		argstr := "null" // at runtime only, for dropped args. not valid in our JSON source format
		if me.Args[i] != nil {
			argstr = me.Args[i].JsonSrc()
		}
		ret += ", " + argstr
	}
	ret += "]"
	return ret
}

// JsonSrc emits the re-`LoadFromJson`able representation of this `FuncDef`.
func (me *FuncDef) JsonSrc(dropFuncDefMetas bool) string {
	outjson := "[ ["
	if !dropFuncDefMetas {
		for i, mstr := range me.Meta {
			if i > 0 {
				outjson += ","
			}
			outjson += strconv.Quote(mstr)
		}
	}
	outjson += "], ["
	for i, a := range me.Args {
		if i > 0 {
			outjson += ","
		}
		outjson += strconv.Itoa(a)
	}
	return outjson + "],\n\t\t" + me.Body.JsonSrc() + " ]"
}

// JsonSrc emits the re-`LoadFromJson`able representation of this `Prog`.
func (me Prog) JsonSrc(dropFuncDefMetas bool) string {
	outjson := "[ "
	for i, def := range me {
		if i > 0 {
			outjson += ", "
		}
		outjson += def.JsonSrc(dropFuncDefMetas) + "\n"
	}
	return outjson + "]\n"
}
