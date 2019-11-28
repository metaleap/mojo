// atem is a minimal and low-level interpretable functional intermediate language
// and interpreter. It prioritizes staying low-LoC enough to be able to rewrite it
// on any other tech stack any time, over other concerns. At the time of writing,
// the core "parsing" / loading in this Go-based implementation is ~40 LoCs (the
// choice of a JSON code format is similarly motivated by the goal to allow for
// swift re-implementations in any contemporary or future lang / tech stack), the
// core interpretation / eval'ing parts around ~55 LoCs, basic AST node types and
// their `String()` implementations around ~40 LoCs, and helpers for injecting or
// extracting "lists of ints" (strings at the end of the day) into / from run
// time another ~40 LoCs. All counts approximate and net, excluding comments.
//
// This focus doesn't make for the most efficient interpreter in the world, but
// that isn't the objective for atem. The goal is to provide the bootstrapping
// layer for **atmo**. An initial compiler from atmo to atem is being coded in my
// [toy Lambda Calculus dialect](https://github.com/metaleap/go-machines/tree/master/toylam)
// and then again in (the initial iteration of) atmo itself. The atem interpreter
// will also suffice / go a long way for REPL purposes and later on abstract /
// symbolic interpretation / partial evaluation for the experimental
// type-checking approaches envisioned to be explored within the impending
// ongoing evolution of atmo once its initial incarnation is birthed.
//
// For "ultimate real-world runtime artifact" purposes, atem isn't intended;
// rather, transpilers and / or compilers to 3rd-party mature and widely enjoyed
// interpreters / bytecode VMs or intermediate ASM targets such as LLVM would
// be the envisioned generally-preferable direction anyway, except such trans-/
// compilers want to naturally be done in atmo as well, so atem is way to get from
// nowhere to _there_, and to be _able_ (not forced) to replicate this original
// bootstrapping on any sort of tech base at any time whenever necessary.
//
// The initial inspiration (and iteration) for atem was the elegant and minimalist
// [SAPL](https://github.com/metaleap/go-machines/tree/master/sapl) approach
// presented by Jansen / Koopman / Plasmeijer, but unlike the above-linked
// by-the-paper implementation, atem diverges even in its initial form in
// various aspects and will continue to evolve various details in tandem with
// the birthing of atmo.
//
// SAPL's basics still apply for now: all funcs are top-level (no lambdas or
// other locals), as such support 0 - n args (rather than all-unary). There
// are no names: global funcs and, inside them, their args are referred to by
// integer indices. Thus most expressions are atomic: arg-refs, func-refs,
// and plain integers. The only non-atomic expression is call / application:
// it is composed of two sub-expressions, the callee and the arg. Divergences:
// our func-refs, if negative, denote a binary primitive-instruction op-code
// such as `ADD` etc. that is handled natively by the interpreter. Unlike SAPL,
// our func-refs don't carry around their number-of-args, instead they're looked
// up in the `Prog`. For calls / applications, likely will move from the current
// unary style to n-ary for efficiency reasons, without breaking partial-application
// of course, or degrading our overall LoCs aims unduly.
package atem

import (
	"strconv"
)

// The few standard func defs the interpreter needs to know of as a minimum, and
// their inviolably hereby-prescribed standard indexes within a `Prog`. Every atem
// code generator must emit implementations for them all, and placed correctly.
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

type (
	Prog    []FuncDef
	FuncDef struct {
		// Args holds this `FuncDef`'s arguments: each `int` denotes how often the
		// `Body` references this arg (although the interpreter only cares about
		// 0 or greater), the arg's "identity" however is just its index in `Args`
		Args          []int
		Body          Expr
		OrigNameMaybe string
	}
	Expr interface {
		// String emits the re-`LoadFromJson`able representation of this `Expr`.
		String() string
	}
	ExprNumInt  int
	ExprArgRef  int
	ExprFuncRef int
	ExprAppl    struct {
		Callee Expr
		Arg    Expr
	}
)

// String emits the re-`LoadFromJson`able representation of this `ExprNumInt`.
func (me ExprNumInt) String() string { return strconv.Itoa(int(me)) }

// String emits the re-`LoadFromJson`able representation of this `ExprArgRef`.
func (me ExprArgRef) String() string { return "\"" + strconv.Itoa(int(me)) + "\"" }

// String emits the re-`LoadFromJson`able representation of this `ExprFuncRef`.
func (me ExprFuncRef) String() string { return "[" + strconv.Itoa(int(me)) + "]" }

// String emits the re-`LoadFromJson`able representation of this `ExprAppl`.
func (me ExprAppl) String() string { return "[" + me.Callee.String() + ", " + me.Arg.String() + "]" }

// String emits the re-`LoadFromJson`able representation of this `FuncDef`.
func (me *FuncDef) String() string {
	outjson := "[ ["
	for i, a := range me.Args {
		if i > 0 {
			outjson += ","
		}
		outjson += strconv.Itoa(a)
	}
	return outjson + "],\n\t\t" + me.Body.String() + " ]"
}

// String emits the re-`LoadFromJson`able representation of this `Prog`.
func (me Prog) String() string {
	outjson := "[ "
	for i, def := range me {
		if i > 0 {
			outjson += ", "
		}
		outjson += def.String() + "\n"
	}
	return outjson + "]\n"
}
