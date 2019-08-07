// Package `atmo/il` implements the intermediate-language representation that
// a lexed-and-parsed `atmo/ast` is transformed into as a next step. Whereas
// the lex-and-parse phase in the `atmo/ast` package ("stage 0") only cared
// about syntax, a few initial (more-)semantic validations occur at the very
// next "stage 1" (AST-to-IL), such as eg. def-name and arg-name validations,
// nonsensical placeholders et al. The following "stage 2" (in `atmo/session`)
// then does prerequisite initial names-analyses and it too both operates chiefly
// on `atmo/il` node types plus utilizes its auxiliary types provided for this.
//
// `atmo/ast` transforms and/or desugars into `atmo/il` such that only idents,
// atomic literals, unary calls, and nullary or unary defs remain in the IL.
// Case expressions desugar into combinations of calls to basic `Std`-built-in
// funcs such as `true`, `false`, `or`, `==` etc. Underscore placeholders obtain
// meaning (or err), usually going from "de-facto lambdas" towards added inner
// named-local defs ("let"s). (Both ident exprs and call exprs can own any
// number of named local defs, whether from AST or dynamically generated.)
// Def-name and def-arg affixes are repositioned as wrapping around the def's
// body. N-ary defs are unary-fied via added inner named-locals, n-ary calls
// via nested sub-calls. ~~All callees and call-args are ensured to be atomic
// via added inner named-locals if and as needed.~~ Other than these transforms,
// no reductions, rewritings or removals occur in this (AST-to-IL) "stage 1".
package atmoil

import (
	"github.com/go-leap/dev/lex"
	. "github.com/metaleap/atmo"
	. "github.com/metaleap/atmo/ast"
)

type ctxIrFromAst struct {
	curTopLevelDef  *IrDefTop
	defsScope       *IrDefs
	defArgs         map[*IrDef]*IrArg
	coerceCallables map[IIrNode]IIrExpr
	counter         struct {
		val   byte
		times int
	}
}

type IrDefs []IrDef

type IrTopDefs []*IrDefTop

type IIrNode interface {
	Print() IAstNode
	Origin() IAstNode
	origToks() udevlex.Tokens
	EquivTo(IIrNode) bool
	findByOrig(IIrNode, IAstNode) []IIrNode
	IsDef() *IrDef
	IsExt() bool
	Let() *IrExprLetBase
	RefersTo(string) bool
	refsTo(string) []IIrExpr
	walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) bool
}

type IIrExpr interface {
	IIrNode
	IsAtomic() bool
	exprBase() *IrExprBase
}

type irNodeBase struct {
	// some `IIrExpr`s' own `Orig` fields or `INode.Origin()` implementations might
	// point to (on-the-fly dynamically computed in-memory) desugared nodes, this
	// one always points to the "real origin" node (might be identical or not)
	Orig IAstNode
}

type IrLam struct {
	IrExprBase
	Arg  IrArg
	Body IIrExpr
}

type IrDef struct {
	irNodeBase

	Name IrIdentDecl
	Body IIrExpr
}

type IrDefTop struct {
	IrDef

	Id           string
	OrigTopChunk *AstFileChunk
	Anns         struct {
		Preduced IPreduced
	}
	Errs struct {
		Stage1AstToIr  Errors
		Stage2BadNames Errors
		Stage3Preduce  Errors
	}

	refersTo map[string]bool
}

type IrArg struct {
	IrIdentDecl
}

type IrExprBase struct {
	irNodeBase
}

type IrExprAtomBase struct {
	IrExprBase
}

type irLitBase struct {
	IrExprAtomBase
}

type IrLitUint struct {
	irLitBase
	Val uint64
}

type IrLitFloat struct {
	irLitBase
	Val float64
}

type IrLitTag struct {
	irLitBase
	Val string
}

type IrExprLetBase struct {
	Defs      IrDefs
	letOrig   *AstExprLet
	letPrefix string

	Anns struct {
		// like `IrIdentName.Anns.Candidates`, contains the following `INode` types:
		// *atmoil.IrDef, *atmoil.IrArg, *atmoil.IrDefTop, atmosess.IrDefRef
		NamesInScope AnnNamesInScope
	}
}

type IrNonValue struct {
	IrExprAtomBase
	OneOf struct {
		LeftoverPlaceholder bool
		Undefined           bool
		TempStrLit          bool
	}
}

type IrIdentBase struct {
	IrExprAtomBase
	Val string
}

type IrIdentDecl struct {
	IrIdentBase
}

type IrIdentName struct {
	IrIdentBase
	IrExprLetBase

	Anns struct {
		// like `IrExprLetBase.Anns.NamesInScope`, contains the following `IIrNode` types:
		// *atmoil.IrDef, *atmoil.IrArg, *atmoil.IrDefTop, atmosess.IrDefRef
		Candidates []IIrNode
	}
}

type IrAppl struct {
	IrExprBase
	IrExprLetBase
	Callee  IIrExpr
	CallArg IIrExpr
}

type AnnNamesInScope map[string][]IIrNode

type IrBuild struct{}
type IPreduced interface {
	IsErrOrAbyss() bool
	Self() *Preduced
	SummaryCompact() string
}

// Preduced is embedded in all `IPreduced` implementers.
type Preduced struct {
}

type PPrimAtomicConstUint struct {
	Preduced
	Val uint64
}

type PPrimAtomicConstFloat struct {
	Preduced
	Val float64
}

type PPrimAtomicConstTag struct {
	Preduced
	Val string
}

type PErr struct {
	Preduced
	Err *Error
}

type PAbyss struct {
	Preduced
}

type PHole struct {
	Preduced
	Def *IrDef
}

type PCallable struct {
	Preduced
	Arg *PHole
	Ret *PHole
}
