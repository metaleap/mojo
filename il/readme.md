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
	AstOrig() IAstNode
	Print() IAstNode
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


#### type IrAbs

```go
type IrAbs struct {
	IrExprBase
	Arg  IrArg
	Body IIrExpr

	Anns struct {
		AbsIdx int
	}
}
```


#### func (*IrAbs) AstOrig

```go
func (me *IrAbs) AstOrig() IAstNode
```

#### func (*IrAbs) EquivTo

```go
func (me *IrAbs) EquivTo(node IIrNode, ignoreNames bool) bool
```

#### func (*IrAbs) IsDef

```go
func (*IrAbs) IsDef() *IrDef
```

#### func (*IrAbs) IsExt

```go
func (*IrAbs) IsExt() bool
```

#### func (*IrAbs) Print

```go
func (me *IrAbs) Print() IAstNode
```

#### func (*IrAbs) RefersTo

```go
func (me *IrAbs) RefersTo(name string) bool
```

#### type IrAppl

```go
type IrAppl struct {
	IrExprBase
	Callee  IIrExpr
	CallArg IIrExpr
}
```


#### func (*IrAppl) AstOrig

```go
func (me *IrAppl) AstOrig() IAstNode
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


#### func (*IrArg) AstOrig

```go
func (me *IrArg) AstOrig() IAstNode
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

#### func (IrBuild) LitTag

```go
func (IrBuild) LitTag(name string) *IrLitTag
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


#### func (*IrDef) AncestorsAndChildrenOf

```go
func (me *IrDef) AncestorsAndChildrenOf(node IIrNode) (nodeAncestors []IIrNode, nodeChildren []IIrNode)
```

#### func (*IrDef) AncestorsOf

```go
func (me *IrDef) AncestorsOf(node IIrNode) (nodeAncestors []IIrNode)
```

#### func (*IrDef) AstOrig

```go
func (me *IrDef) AstOrig() IAstNode
```

#### func (*IrDef) AstOrigToks

```go
func (me *IrDef) AstOrigToks(node IIrNode) (toks udevlex.Tokens)
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

#### func (*IrDef) NamesInScopeAt

```go
func (me *IrDef) NamesInScopeAt(descendantNodeInQuestion IIrNode, knownGlobalsInScope AnnNamesInScope, excludeInternalIdents bool) (namesInScope AnnNamesInScope)
```

#### func (*IrDef) OrigDef

```go
func (me *IrDef) OrigDef() (origDef *AstDef)
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
func (me *IrDef) Walk(whetherToKeepTraversing func(curNodeAncestors []IIrNode, curNode IIrNode, curNodeChildrenThatWillBeTraversedIfReturningTrue ...IIrNode) bool)
```

#### type IrDefs

```go
type IrDefs []*IrDef
```


#### func (IrDefs) ByName

```go
func (me IrDefs) ByName(name string, onlyFor *AstFile) (defs []*IrDef)
```

#### func (IrDefs) IndexByID

```go
func (me IrDefs) IndexByID(id string) int
```

#### func (IrDefs) Len

```go
func (me IrDefs) Len() int
```

#### func (IrDefs) Less

```go
func (me IrDefs) Less(i int, j int) bool
```

#### func (*IrDefs) ReInitFrom

```go
func (me *IrDefs) ReInitFrom(kitSrcFiles AstFiles) (droppedTopLevelDefIdsAndNames map[string]string, newTopLevelDefIdsAndNames map[string]string, freshErrs Errors)
```

#### func (IrDefs) Swap

```go
func (me IrDefs) Swap(i int, j int)
```

#### type IrExprAtomBase

```go
type IrExprAtomBase struct {
	IrExprBase
}
```


#### func (*IrExprAtomBase) AstOrig

```go
func (me *IrExprAtomBase) AstOrig() IAstNode
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

#### func (*IrExprAtomBase) RefersTo

```go
func (me *IrExprAtomBase) RefersTo(string) bool
```

#### type IrExprBase

```go
type IrExprBase struct {
}
```


#### func (*IrExprBase) AstOrig

```go
func (me *IrExprBase) AstOrig() IAstNode
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

#### type IrIdentBase

```go
type IrIdentBase struct {
	IrExprAtomBase
	Val string
}
```


#### func (*IrIdentBase) AstOrig

```go
func (me *IrIdentBase) AstOrig() IAstNode
```

#### func (*IrIdentBase) IsDef

```go
func (*IrIdentBase) IsDef() *IrDef
```

#### func (*IrIdentBase) IsExt

```go
func (*IrIdentBase) IsExt() bool
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


#### func (*IrIdentDecl) AstOrig

```go
func (me *IrIdentDecl) AstOrig() IAstNode
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

#### type IrIdentName

```go
type IrIdentName struct {
	IrIdentBase

	Anns struct {
		AbsIdx     int
		ArgIdx     int
		Candidates []IIrNode
	}
}
```


#### func (*IrIdentName) AstOrig

```go
func (me *IrIdentName) AstOrig() IAstNode
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

#### func (*IrIdentName) RefersTo

```go
func (me *IrIdentName) RefersTo(name string) bool
```

#### func (*IrIdentName) ResolvesTo

```go
func (me *IrIdentName) ResolvesTo(n IIrNode) bool
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


#### func (*IrNonValue) AstOrig

```go
func (me *IrNonValue) AstOrig() IAstNode
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

#### func (*IrNonValue) Print

```go
func (me *IrNonValue) Print() IAstNode
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

#### type PFunc

```go
type PFunc struct {
	Preduced
	Orig *IrAbs
	Arg  *PMeta
	Ret  *PMeta
}
```


#### func (*PFunc) SummaryCompact

```go
func (me *PFunc) SummaryCompact() string
```

#### type PMeta

```go
type PMeta struct {
	Preduced
}
```


#### func (*PMeta) SummaryCompact

```go
func (me *PMeta) SummaryCompact() string
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
