// Package `atmoil` implements the intermediate-language representation that a
// lexed-and-parsed `atmolang` AST is transformed into as a next step. Whereas
// the lex-and-parse phase in the `atmolang` package ("stage 0") only cared
// about syntax, a few initial (more) semantic validations occur at the very
// next "stage 1" (AST-to-IL), such as eg. def-name and arg-name validations,
// nonsensical placeholders et al. The following "stage 2" (in `atmosess`) then
// does prerequisite initial names-analyses and it too both operates chiefly
// on `atmoil` node types plus utilizes its auxiliary types provided for this.
//
// `atmolang` transforms and/or desugars into `atmoil` such that only idents,
// atomic literals, unary calls, and nullary or unary defs remain in the IL.
// Case expressions desugar into combinations of calls to basic `Std`-built-in
// funcs such as `true`, `false`, `or`, `==` etc. Underscore placeholders obtain
// meaning (or err), usually going from "de-facto lambdas" towards added inner
// named-local defs ("let"s). (Both ident exprs and call exprs can own any
// number of named local defs, whether from AST or dynamically generated.)
// Def-name and def-arg affixes are repositioned as wrapping around the def's
// body. N-ary defs are unary-fied via added inner named-locals, n-ary calls
// via nested sub-calls. All callees and call-args are ensured to be atomic
// via added inner named-locals if and as needed. Other than these transforms,
// no reductions, rewritings or removals occur in this (AST-to-IL) "stage 1".
package atmoil

import (
	"github.com/go-leap/dev/lex"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type ctxIrFromAst struct {
	curTopLevelDef  *IrDefTop
	defsScope       *IrDefs
	coerceCallables map[INode]IExpr
	counter         struct {
		val   byte
		times int
	}
}

type IrDefs []IrDef

type IrTopDefs []*IrDefTop

type INode interface {
	Print() atmolang.IAstNode
	Origin() atmolang.IAstNode
	origToks() udevlex.Tokens
	EquivTo(INode) bool
	findByOrig(INode, atmolang.IAstNode) []INode
	IsDef() *IrDef
	Let() *IrExprLetBase
	RefersTo(string) bool
	refsTo(string) []IExpr
	walk(ancestors []INode, self INode, on func([]INode, INode, ...INode) bool)
}

type IExpr interface {
	INode
	IsAtomic() bool
	exprBase() *IrExprBase
}

type irNodeBase struct {
}

type IrDef struct {
	irNodeBase
	OrigDef *atmolang.AstDef

	Name IrIdentDecl
	Arg  *IrDefArg
	Body IExpr
}

type IrDefTop struct {
	IrDef

	Id           string
	OrigTopChunk *atmolang.SrcTopChunk
	Anns         struct {
		Preduced IPreduced
	}
	Errs struct {
		Stage1AstToIr  atmo.Errors
		Stage2BadNames atmo.Errors
		Stage3Preduce  atmo.Errors
	}

	refersTo map[string]bool
}

type IrDefArg struct {
	IrIdentDecl
	Orig     *atmolang.AstDefArg
	ownerDef *IrDef
}

type IrExprBase struct {
	irNodeBase

	// some `IIrExpr`s' own `Orig` fields or `INode.Origin()` implementations might
	// point to (on-the-fly dynamically computed in-memory) desugared nodes, this
	// one always points to the "real origin" node (might be identical or not)
	Orig atmolang.IAstExpr
}

type IrExprAtomBase struct {
	IrExprBase
}

type irLitBase struct {
	IrExprAtomBase
}

type IrLitStr struct {
	irLitBase
	Val string
}

type IrLitUint struct {
	irLitBase
	Val uint64
}

type IrLitFloat struct {
	irLitBase
	Val float64
}

type IrExprLetBase struct {
	Defs      IrDefs
	letOrig   *atmolang.AstExprLet
	letPrefix string

	Anns struct {
		// like `IrIdentName.Anns.Candidates`, contains the following `INode` types:
		// *atmoil.IrDef, *atmoil.IrDefArg, *atmoil.IrDefTop, atmosess.IrDefRef
		NamesInScope AnnNamesInScope
	}
}

type IrIdentBase struct {
	IrExprAtomBase
	Val string
}

type IrNonValue struct {
	IrExprAtomBase
	OneOf struct {
		LeftoverPlaceholder bool
		Undefined           bool
	}
}

type IrIdentTag struct {
	IrIdentBase
}

type IrIdentDecl struct {
	IrIdentBase
}

type IrIdentName struct {
	IrIdentBase
	IrExprLetBase

	Anns struct {
		// like `IrExprLetBase.Anns.NamesInScope`, contains the following `IIrNode` types:
		// *atmoil.IrDef, *atmoil.IrDefArg, *atmoil.IrDefTop, atmosess.IrDefRef
		Candidates []INode
	}
}

type IrAppl struct {
	IrExprBase
	IrExprLetBase
	Orig         *atmolang.AstExprAppl
	AtomicCallee IExpr
	AtomicArg    IExpr
}

type AnnNamesInScope map[string][]INode

type IPreduced interface {
	Self() *Preduced
	SummaryCompact() string
}

// Preduced is embedded in all `IPreduced` implementers.
type Preduced struct {
}

type PCallable struct {
	Preduced
	Arg *PHole
	Ret *PHole
}

type PCallables struct {
	Preduced
	Cases []PCallable
}

type PErr struct {
	Preduced
	Err *atmo.Error
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

type PAbyss struct {
	Preduced
}

type PHole struct {
	Preduced
	Def *IrDef
}

type Builder struct{}
