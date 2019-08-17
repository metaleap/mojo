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
// atomic literals, unary calls, lambdas, and nullary defs remain in the IL.
// Case expressions desugar into combinations of calls to basic `Std`-built-in
// funcs such as `true`, `false`, `or`, `==` etc. Underscore placeholders
// obtain meaning (or err), usually going from "de-facto lambdas" towards
// actual lambdas. Every "let" def turns into an application-of-a-lambda.
// Def-name and def-arg affixes are repositioned as wrapping around the def's
// body. N-ary defs are unary-then-nullary-fied via inner lambdas, n-ary calls
// via nested sub-calls. Other than these transforms, no reductions,
// rewritings or removals occur in this (AST-to-IL) "stage 1".
package atmoil

import (
	. "github.com/metaleap/atmo"
	. "github.com/metaleap/atmo/ast"
)

type ctxIrFromAst struct {
	curTopLevelDef  *IrDef
	coerceCallables map[IIrNode]IIrExpr
	counter         struct {
		val   byte
		times int
	}
	absIdx int
	absMax int
}

type IrDefs []*IrDef

type IIrNode interface {
	AstOrig() IAstNode
	Print() IAstNode
	EquivTo(sameTypedNode IIrNode, ignoreNames bool) bool
	findByOrig(IIrNode, IAstNode, func(IIrNode) bool) []IIrNode
	IsDef() *IrDef
	RefersTo(string) bool
	refsTo(string) []IIrExpr
	walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) bool
}

type IIrExpr interface {
	IIrNode
	IsAtomic() bool
	exprBase() *IrExprBase
}

type IIrIdent interface {
	IsInternal() bool
}

type irNodeBase struct {
	Orig IAstNode
}

type IrDef struct {
	irNodeBase

	Ident IrIdentDecl
	Body  IIrExpr

	Id string
	*AstFileChunk
	Ann struct {
		Preduced IPreduced
	}
	Errs struct {
		Stage1AstToIr  Errors
		Stage2BadNames Errors
		Stage3Preduce  Errors
	}

	refersTo map[string]bool
}

type IrAbs struct {
	IrExprBase
	Arg  IrArg
	Body IIrExpr

	Ann struct {
		AbsIdx int
	}
}

type IrArg struct {
	IrIdentDecl
	ownerAbs *IrAbs
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

type IrNonValue struct {
	IrExprAtomBase
	OneOf struct {
		LeftoverPlaceholder bool
		Undefined           bool
	}
}

type IrIdentBase struct {
	IrExprAtomBase
	Name string
}

type IrIdentDecl struct {
	IrIdentBase
}

type IrIdentName struct {
	IrIdentBase

	Ann struct {
		AbsIdx int
		ArgIdx int
		// Candidates may contain either one `*atmoil.IrArg` or any number
		// of `*atmoil.IrDef` or `atmosess.IrDefRef`.
		Candidates []IIrNode
	}
}

type IrAppl struct {
	IrExprBase
	Callee  IIrExpr
	CallArg IIrExpr
}

type IrRef struct {
	Node IIrNode

	// access via `IsDef`, can be *IrDef or atmosess.IrDefRef
	Def IIrNode
}

type IrBuild struct{}

// AnnNamesInScope contains per-name all nodes known-in-scope that declare that name;
// every `IIrNode` is one of `*atmoil.IrDef`, `*atmoil.IrArg`, `atmosess.IrDefRef`.
// (That is for outside consumers, internally it temporarily contains `*atmoil.IrAbs` during `AnnNamesInScope.RepopulateDefsAndIdentsFor` while in flight.)
type AnnNamesInScope map[string][]IIrNode

type IPreduced interface {
	Errs() Errors
	Self() *PValFactBase
	String() string
}

type PValFactBase struct {
	From IrRef
}

type PValUsed struct {
	PValFactBase
}

type PValPrimConst struct {
	PValFactBase
	ConstVal interface{}
}

type PValEqVal struct {
	PValFactBase
	To *PVal
}

type PValEqType struct {
	PValFactBase
	Of *PVal
}

type PValNever struct {
	PValFactBase
	Never IPreduced
}

type PValFn struct {
	PValFactBase
	Arg PVal
	Ret PVal
}

type PValAbyss struct {
	PValFactBase
}

type PValErr struct {
	PValFactBase
	*Error
}

type PVal struct {
	PValFactBase
	Facts []IPreduced
}

type PEnv struct {
	Link *PEnv
	PVal
}
