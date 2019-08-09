# atmoil
--
    import "github.com/metaleap/atmo/il"

Package `atmo/il` implements the intermediate-language representation that a
lexed-and-parsed `atmo/ast` is transformed into as a next step. Whereas the
lex-and-parse phase in the `atmo/ast` package ("stage 0") only cared about
syntax, a few initial (more-)semantic validations occur at the very next "stage
1" (AST-to-IL), such as eg. def-name and arg-name validations, nonsensical
placeholders et al. The following "stage 2" (in `atmo/session`) then does
prerequisite initial names-analyses and it too both operates chiefly on
`atmo/il` node types plus utilizes its auxiliary types provided for this.

`atmo/ast` transforms and/or desugars into `atmo/il` such that only idents,
atomic literals, unary calls, lambdas, and nullary defs remain in the IL. Case
expressions desugar into combinations of calls to basic `Std`-built-in funcs
such as `true`, `false`, `or`, `==` etc. Underscore placeholders obtain meaning
(or err), usually going from "de-facto lambdas" towards actual lambdas. Every
"let" def turns into an application-of-a-lambda. Def-name and def-arg affixes
are repositioned as wrapping around the def's body. N-ary defs are
unary-then-nullary-fied via inner lambdas, n-ary calls via nested sub-calls.
Other than these transforms, no reductions, rewritings or removals occur in this
(AST-to-IL) "stage 1".

## Usage

```go
const (
	ErrFromAst_DefNameInvalidIdent
	ErrFromAst_DefArgNameMultipleUnderscores
	ErrFromAst_UnhandledStandaloneUnderscores
)
```

```go
const (
	ErrNames_ShadowingNotAllowed
	ErrNames_IdentRefersToMalformedDef
)
```

#### func  DbgPrintToStderr

```go
func DbgPrintToStderr(node IIrNode)
```

#### func  DbgPrintToString

```go
func DbgPrintToString(node IIrNode) string
```

#### func  ExprFrom

```go
func ExprFrom(orig IAstExpr) (IIrExpr, Errors)
```

#### type AnnNamesInScope

```go
type AnnNamesInScope map[string][]IIrNode
```

AnnNamesInScope contains per-name all nodes known-in-scope that declare that
name; every `IIrNode` is one of `*atmoil.IrDef`, `*atmoil.IrArg`,
`atmosess.IrDefRef`

#### func (AnnNamesInScope) Add

```go
func (me AnnNamesInScope) Add(name string, nodes ...IIrNode)
```

#### func (AnnNamesInScope) RepopulateDefsAndIdentsFor

```go
func (me AnnNamesInScope) RepopulateDefsAndIdentsFor(tld *IrDef, node IIrNode, currentlyErroneousButKnownGlobalsNames StringKeys, nodeAncestors ...IIrNode) (errs Errors)
```

#### type IIrExpr

```go
type IIrExpr interface {
	IIrNode
	IsAtomic() bool
	// contains filtered or unexported methods
}
```


#### type IIrNode

```go
type IIrNode interface {
	Print() IAstNode
	Origin() IAstNode

	EquivTo(sameTypedNode IIrNode, ignoreNames bool) bool

	IsDef() *IrDef
	IsExt() bool
	RefersTo(string) bool
	// contains filtered or unexported methods
}
```


#### type IPreduced

```go
type IPreduced interface {
	IsErrOrAbyss() bool
	Self() *Preduced
	SummaryCompact() string
}
```


#### type IrAppl

```go
type IrAppl struct {
	IrExprBase
	Callee  IIrExpr
	CallArg IIrExpr
}
```


#### func (*IrAppl) EquivTo

```go
func (me *IrAppl) EquivTo(node IIrNode, ignoreNames bool) bool
```

#### func (*IrAppl) IsDef

```go
func (*IrAppl) IsDef() *IrDef
```

#### func (*IrAppl) IsExt

```go
func (*IrAppl) IsExt() bool
```

#### func (*IrAppl) Origin

```go
func (me *IrAppl) Origin() IAstNode
```

#### func (*IrAppl) Print

```go
func (me *IrAppl) Print() IAstNode
```

#### func (*IrAppl) RefersTo

```go
func (me *IrAppl) RefersTo(name string) bool
```

#### type IrArg

```go
type IrArg struct {
	IrIdentDecl
}
```


#### func (*IrArg) EquivTo

```go
func (me *IrArg) EquivTo(node IIrNode, ignoreNames bool) bool
```

#### func (*IrArg) IsDef

```go
func (*IrArg) IsDef() *IrDef
```

#### func (*IrArg) IsExt

```go
func (*IrArg) IsExt() bool
```

#### func (*IrArg) Origin

```go
func (me *IrArg) Origin() IAstNode
```

#### func (*IrArg) Print

```go
func (me *IrArg) Print() IAstNode
```

#### type IrBuild

```go
type IrBuild struct{}
```


```go
var (
	BuildIr IrBuild
)
```

#### func (IrBuild) Appl1

```go
func (IrBuild) Appl1(callee IIrExpr, callArg IIrExpr) *IrAppl
```

#### func (IrBuild) ApplN

```go
func (IrBuild) ApplN(ctx *ctxIrFromAst, callee IIrExpr, callArgs ...IIrExpr) (appl *IrAppl)
```

#### func (IrBuild) IdentName

```go
func (IrBuild) IdentName(name string) *IrIdentName
```

#### func (IrBuild) IdentNameCopy

```go
func (IrBuild) IdentNameCopy(identBase *IrIdentBase) *IrIdentName
```

#### func (IrBuild) IdentTag

```go
func (IrBuild) IdentTag(name string) *IrLitTag
```

#### func (IrBuild) Undef

```go
func (IrBuild) Undef() *IrNonValue
```

#### type IrDef

```go
type IrDef struct {
	Name IrIdentDecl
	Body IIrExpr

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
}
```


#### func (*IrDef) EquivTo

```go
func (me *IrDef) EquivTo(node IIrNode, ignoreNames bool) bool
```

#### func (*IrDef) Errors

```go
func (me *IrDef) Errors() (errs Errors)
```

#### func (*IrDef) FindAll

```go
func (me *IrDef) FindAll(where func(IIrNode) bool) (matches [][]IIrNode)
```

#### func (*IrDef) FindAny

```go
func (me *IrDef) FindAny(where func(IIrNode) bool) (firstMatchWithAncestorsPrepended []IIrNode)
```

#### func (*IrDef) FindByOrig

```go
func (me *IrDef) FindByOrig(orig IAstNode) []IIrNode
```

#### func (*IrDef) FindDescendants

```go
func (me *IrDef) FindDescendants(traverseIntoMatchesToo bool, max int, pred func(IIrNode) bool) (paths [][]IIrNode)
```

#### func (*IrDef) HasAnyOf

```go
func (me *IrDef) HasAnyOf(nodes ...IIrNode) bool
```

#### func (*IrDef) HasErrors

```go
func (me *IrDef) HasErrors() bool
```

#### func (*IrDef) HasIdentDecl

```go
func (me *IrDef) HasIdentDecl(name string) bool
```

#### func (*IrDef) IsDef

```go
func (me *IrDef) IsDef() *IrDef
```

#### func (*IrDef) IsExt

```go
func (*IrDef) IsExt() bool
```

#### func (*IrDef) IsLam

```go
func (me *IrDef) IsLam() (ifSo *IrLam)
```

#### func (*IrDef) NamesInScopeAt

```go
func (me *IrDef) NamesInScopeAt(descendantNodeInQuestion IIrNode, knownGlobalsInScope AnnNamesInScope, excludeInternalIdents bool) (namesInScope AnnNamesInScope)
```

#### func (*IrDef) OrigDef

```go
func (me *IrDef) OrigDef() (origDef *AstDef)
```

#### func (*IrDef) OrigToks

```go
func (me *IrDef) OrigToks(node IIrNode) (toks udevlex.Tokens)
```

#### func (*IrDef) Origin

```go
func (me *IrDef) Origin() IAstNode
```

#### func (*IrDef) Print

```go
func (me *IrDef) Print() IAstNode
```

#### func (*IrDef) RefersTo

```go
func (me *IrDef) RefersTo(name string) (refersTo bool)
```

#### func (*IrDef) RefersToOrDefines

```go
func (me *IrDef) RefersToOrDefines(name string) (relatesTo bool)
```

#### func (*IrDef) RefsTo

```go
func (me *IrDef) RefsTo(name string) (refs []IIrExpr)
```

#### func (*IrDef) Walk

```go
func (me *IrDef) Walk(whetherToKeepTraversing func(curNodeAncestors []IIrNode, curNode IIrNode, curNodeDescendantsThatWillBeTraversedIfReturningTrue ...IIrNode) bool)
```

#### type IrDefs

```go
type IrDefs []IrDef
```


#### type IrExprAtomBase

```go
type IrExprAtomBase struct {
	IrExprBase
}
```


#### func (*IrExprAtomBase) IsAtomic

```go
func (me *IrExprAtomBase) IsAtomic() bool
```

#### func (*IrExprAtomBase) IsDef

```go
func (*IrExprAtomBase) IsDef() *IrDef
```

#### func (*IrExprAtomBase) IsExt

```go
func (*IrExprAtomBase) IsExt() bool
```

#### func (*IrExprAtomBase) Origin

```go
func (me *IrExprAtomBase) Origin() IAstNode
```

#### func (*IrExprAtomBase) RefersTo

```go
func (me *IrExprAtomBase) RefersTo(string) bool
```

#### type IrExprBase

```go
type IrExprBase struct {
}
```


#### func (*IrExprBase) IsAtomic

```go
func (*IrExprBase) IsAtomic() bool
```

#### func (*IrExprBase) IsDef

```go
func (*IrExprBase) IsDef() *IrDef
```

#### func (*IrExprBase) IsExt

```go
func (*IrExprBase) IsExt() bool
```

#### func (*IrExprBase) Origin

```go
func (me *IrExprBase) Origin() IAstNode
```

#### type IrIdentBase

```go
type IrIdentBase struct {
	IrExprAtomBase
	Val string
}
```


#### func (*IrIdentBase) IsDef

```go
func (*IrIdentBase) IsDef() *IrDef
```

#### func (*IrIdentBase) IsExt

```go
func (*IrIdentBase) IsExt() bool
```

#### func (*IrIdentBase) Origin

```go
func (me *IrIdentBase) Origin() IAstNode
```

#### func (*IrIdentBase) Print

```go
func (me *IrIdentBase) Print() IAstNode
```

#### type IrIdentDecl

```go
type IrIdentDecl struct {
	IrIdentBase
}
```


#### func (*IrIdentDecl) EquivTo

```go
func (me *IrIdentDecl) EquivTo(node IIrNode, ignoreNames bool) bool
```

#### func (*IrIdentDecl) IsDef

```go
func (*IrIdentDecl) IsDef() *IrDef
```

#### func (*IrIdentDecl) IsExt

```go
func (*IrIdentDecl) IsExt() bool
```

#### func (*IrIdentDecl) Origin

```go
func (me *IrIdentDecl) Origin() IAstNode
```

#### type IrIdentName

```go
type IrIdentName struct {
	IrIdentBase

	Anns struct {
		// ArgIdx is 0 if not pointing to an `*IrArg`, else the De Bruijn index
		ArgIdx int

		// *atmoil.IrDef, *atmoil.IrArg, atmosess.IrDefRef
		Candidates []IIrNode
	}
}
```


#### func (*IrIdentName) EquivTo

```go
func (me *IrIdentName) EquivTo(node IIrNode, ignoreNames bool) bool
```

#### func (*IrIdentName) IsArgRef

```go
func (me *IrIdentName) IsArgRef(maybeSpecificArg *IrArg) bool
```

#### func (*IrIdentName) IsDef

```go
func (*IrIdentName) IsDef() *IrDef
```

#### func (*IrIdentName) IsExt

```go
func (*IrIdentName) IsExt() bool
```

#### func (*IrIdentName) Origin

```go
func (me *IrIdentName) Origin() IAstNode
```

#### func (*IrIdentName) RefersTo

```go
func (me *IrIdentName) RefersTo(name string) bool
```

#### func (*IrIdentName) ResolvesTo

```go
func (me *IrIdentName) ResolvesTo(n IIrNode) bool
```

#### type IrLam

```go
type IrLam struct {
	IrExprBase
	Arg  IrArg
	Body IIrExpr
}
```


#### func (*IrLam) EquivTo

```go
func (me *IrLam) EquivTo(node IIrNode, ignoreNames bool) bool
```

#### func (*IrLam) IsDef

```go
func (*IrLam) IsDef() *IrDef
```

#### func (*IrLam) IsExt

```go
func (*IrLam) IsExt() bool
```

#### func (*IrLam) Origin

```go
func (me *IrLam) Origin() IAstNode
```

#### func (*IrLam) Print

```go
func (me *IrLam) Print() IAstNode
```

#### func (*IrLam) RefersTo

```go
func (me *IrLam) RefersTo(name string) bool
```

#### type IrLitFloat

```go
type IrLitFloat struct {
	Val float64
}
```


#### func (*IrLitFloat) EquivTo

```go
func (me *IrLitFloat) EquivTo(node IIrNode, ignoreNames bool) bool
```

#### func (*IrLitFloat) Print

```go
func (me *IrLitFloat) Print() IAstNode
```

#### type IrLitTag

```go
type IrLitTag struct {
	Val string
}
```


#### func (*IrLitTag) EquivTo

```go
func (me *IrLitTag) EquivTo(node IIrNode, ignoreNames bool) bool
```

#### func (*IrLitTag) Print

```go
func (me *IrLitTag) Print() IAstNode
```

#### type IrLitUint

```go
type IrLitUint struct {
	Val uint64
}
```


#### func (*IrLitUint) EquivTo

```go
func (me *IrLitUint) EquivTo(node IIrNode, ignoreNames bool) bool
```

#### func (*IrLitUint) Print

```go
func (me *IrLitUint) Print() IAstNode
```

#### type IrNonValue

```go
type IrNonValue struct {
	IrExprAtomBase
	OneOf struct {
		LeftoverPlaceholder bool
		Undefined           bool
		TempStrLit          bool
	}
}
```


#### func (*IrNonValue) EquivTo

```go
func (me *IrNonValue) EquivTo(node IIrNode, ignoreNames bool) bool
```

#### func (*IrNonValue) IsDef

```go
func (*IrNonValue) IsDef() *IrDef
```

#### func (*IrNonValue) IsExt

```go
func (*IrNonValue) IsExt() bool
```

#### func (*IrNonValue) Origin

```go
func (me *IrNonValue) Origin() IAstNode
```

#### func (*IrNonValue) Print

```go
func (me *IrNonValue) Print() IAstNode
```

#### type IrTopDefs

```go
type IrTopDefs []*IrDef
```


#### func (IrTopDefs) ByName

```go
func (me IrTopDefs) ByName(name string, onlyFor *AstFile) (defs []*IrDef)
```

#### func (IrTopDefs) IndexByID

```go
func (me IrTopDefs) IndexByID(id string) int
```

#### func (IrTopDefs) Len

```go
func (me IrTopDefs) Len() int
```

#### func (IrTopDefs) Less

```go
func (me IrTopDefs) Less(i int, j int) bool
```

#### func (*IrTopDefs) ReInitFrom

```go
func (me *IrTopDefs) ReInitFrom(kitSrcFiles AstFiles) (droppedTopLevelDefIdsAndNames map[string]string, newTopLevelDefIdsAndNames map[string]string, freshErrs Errors)
```

#### func (IrTopDefs) Swap

```go
func (me IrTopDefs) Swap(i int, j int)
```

#### type PAbyss

```go
type PAbyss struct {
	Preduced
}
```


#### func (*PAbyss) IsErrOrAbyss

```go
func (me *PAbyss) IsErrOrAbyss() bool
```

#### func (*PAbyss) SummaryCompact

```go
func (me *PAbyss) SummaryCompact() string
```

#### type PCallable

```go
type PCallable struct {
	Preduced
	Arg *PHole
	Ret *PHole
}
```


#### func (*PCallable) SummaryCompact

```go
func (me *PCallable) SummaryCompact() string
```

#### type PErr

```go
type PErr struct {
	Preduced
	Err *Error
}
```


#### func (*PErr) IsErrOrAbyss

```go
func (me *PErr) IsErrOrAbyss() bool
```

#### func (*PErr) SummaryCompact

```go
func (me *PErr) SummaryCompact() string
```

#### type PHole

```go
type PHole struct {
	Preduced
	Def *IrDef
}
```


#### func (*PHole) SummaryCompact

```go
func (me *PHole) SummaryCompact() string
```

#### type PPrimAtomicConstFloat

```go
type PPrimAtomicConstFloat struct {
	Preduced
	Val float64
}
```


#### func (*PPrimAtomicConstFloat) SummaryCompact

```go
func (me *PPrimAtomicConstFloat) SummaryCompact() string
```

#### type PPrimAtomicConstTag

```go
type PPrimAtomicConstTag struct {
	Preduced
	Val string
}
```


#### func (*PPrimAtomicConstTag) SummaryCompact

```go
func (me *PPrimAtomicConstTag) SummaryCompact() string
```

#### type PPrimAtomicConstUint

```go
type PPrimAtomicConstUint struct {
	Preduced
	Val uint64
}
```


#### func (*PPrimAtomicConstUint) SummaryCompact

```go
func (me *PPrimAtomicConstUint) SummaryCompact() string
```

#### type Preduced

```go
type Preduced struct {
}
```

Preduced is embedded in all `IPreduced` implementers.

#### func (*Preduced) IsErrOrAbyss

```go
func (me *Preduced) IsErrOrAbyss() bool
```

#### func (*Preduced) Self

```go
func (me *Preduced) Self() *Preduced
```
