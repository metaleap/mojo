# atmoil
--
    import "github.com/metaleap/atmo/il"

Package `atmoil` implements the intermediate-language representation that a
lexed-and-parsed `atmolang` AST is transformed into as a next step. Whereas the
lex-and-parse phase in the `atmolang` package ("stage 0") only cared about
syntax, a few initial (more) semantic validations occur at the very next "stage
1" (AST-to-IL), such as eg. def-name and arg-name validations, nonsensical
placeholders et al. The following "stage 2" (in `atmosess`) then does
prerequisite initial names-analyses and it too both operates chiefly on `atmoil`
node types plus utilizes its auxiliary types provided for this.

`atmolang` transforms and/or desugars into `atmoil` such that only idents,
atomic literals, unary calls, and nullary or unary defs remain in the IL. Case
expressions desugar into combinations of calls to basic `Std`-built-in funcs
such as `true`, `false`, `or`, `==` etc. Underscore placeholders obtain meaning
(or err), usually going from "de-facto lambdas" towards added inner named-local
defs ("let"s). (Both ident exprs and call exprs can own any number of named
local defs, whether from AST or dynamically generated.) Def-name and def-arg
affixes are repositioned as wrapping around the def's body. N-ary defs are
unary-fied via added inner named-locals, n-ary calls via nested sub-calls. ~~All
callees and call-args are ensured to be atomic via added inner named-locals if
and as needed.~~ Other than these transforms, no reductions, rewritings or
removals occur in this (AST-to-IL) "stage 1".

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
func DbgPrintToStderr(node INode)
```

#### func  DbgPrintToString

```go
func DbgPrintToString(node INode) string
```

#### type AnnNamesInScope

```go
type AnnNamesInScope map[string][]INode
```


#### func (AnnNamesInScope) Add

```go
func (me AnnNamesInScope) Add(name string, nodes ...INode)
```

#### func (AnnNamesInScope) RepopulateDefsAndIdentsFor

```go
func (me AnnNamesInScope) RepopulateDefsAndIdentsFor(tld *IrDefTop, node INode, currentlyErroneousButKnownGlobalsNames map[string]int) (errs atmo.Errors)
```

#### type Builder

```go
type Builder struct{}
```


```go
var (
	Build Builder
)
```

#### func (Builder) Appl1

```go
func (Builder) Appl1(callee IExpr, callArg IExpr) *IrAppl
```

#### func (Builder) ApplN

```go
func (Builder) ApplN(ctx *ctxIrFromAst, callee IExpr, callArgs ...IExpr) (appl *IrAppl)
```

#### func (Builder) IdentName

```go
func (Builder) IdentName(name string) *IrIdentName
```

#### func (Builder) IdentNameCopy

```go
func (Builder) IdentNameCopy(identBase *IrIdentBase) *IrIdentName
```

#### func (Builder) IdentTag

```go
func (Builder) IdentTag(name string) *IrLitTag
```

#### func (Builder) Undef

```go
func (Builder) Undef() *IrNonValue
```

#### type IExpr

```go
type IExpr interface {
	INode
	IsAtomic() bool
	// contains filtered or unexported methods
}
```


#### func  ExprFrom

```go
func ExprFrom(orig atmolang.IAstExpr) (IExpr, atmo.Errors)
```

#### type INode

```go
type INode interface {
	Print() atmolang.IAstNode
	Origin() atmolang.IAstNode

	EquivTo(INode) bool

	IsDef() *IrDef
	IsExt() bool
	Let() *IrExprLetBase
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
	IrExprLetBase
	Orig    *atmolang.AstExprAppl
	Callee  IExpr
	CallArg IExpr
}
```


#### func (*IrAppl) EquivTo

```go
func (me *IrAppl) EquivTo(node INode) bool
```

#### func (*IrAppl) IsDef

```go
func (*IrAppl) IsDef() *IrDef
```

#### func (*IrAppl) IsExt

```go
func (*IrAppl) IsExt() bool
```

#### func (*IrAppl) Let

```go
func (me *IrAppl) Let() *IrExprLetBase
```

#### func (*IrAppl) Origin

```go
func (me *IrAppl) Origin() atmolang.IAstNode
```

#### func (*IrAppl) Print

```go
func (me *IrAppl) Print() atmolang.IAstNode
```

#### func (*IrAppl) RefersTo

```go
func (me *IrAppl) RefersTo(name string) bool
```

#### type IrArg

```go
type IrArg struct {
	IrIdentDecl
	Orig *atmolang.AstDefArg
}
```


#### func (*IrArg) EquivTo

```go
func (me *IrArg) EquivTo(node INode) bool
```

#### func (*IrArg) IsDef

```go
func (*IrArg) IsDef() *IrDef
```

#### func (*IrArg) IsExt

```go
func (*IrArg) IsExt() bool
```

#### func (*IrArg) Let

```go
func (*IrArg) Let() *IrExprLetBase
```

#### func (*IrArg) Origin

```go
func (me *IrArg) Origin() atmolang.IAstNode
```

#### func (*IrArg) Print

```go
func (me *IrArg) Print() atmolang.IAstNode
```

#### type IrDef

```go
type IrDef struct {
	OrigDef *atmolang.AstDef

	Name IrIdentDecl
	Arg  *IrArg
	Body IExpr
}
```


#### func (*IrDef) EquivTo

```go
func (me *IrDef) EquivTo(node INode) bool
```

#### func (*IrDef) HasArgRefsOtherThan

```go
func (me *IrDef) HasArgRefsOtherThan(argNamesHave atmo.StringKeys) (argNamesIncomplete bool)
```

#### func (*IrDef) IsDef

```go
func (me *IrDef) IsDef() *IrDef
```

#### func (*IrDef) IsExt

```go
func (*IrDef) IsExt() bool
```

#### func (*IrDef) Let

```go
func (*IrDef) Let() *IrExprLetBase
```

#### func (*IrDef) Origin

```go
func (me *IrDef) Origin() atmolang.IAstNode
```

#### func (*IrDef) Print

```go
func (me *IrDef) Print() atmolang.IAstNode
```

#### func (*IrDef) RefersTo

```go
func (me *IrDef) RefersTo(name string) bool
```

#### type IrDefTop

```go
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
}
```


#### func (*IrDefTop) Errors

```go
func (me *IrDefTop) Errors() (errs atmo.Errors)
```

#### func (*IrDefTop) FindAll

```go
func (me *IrDefTop) FindAll(where func(INode) bool) (matches [][]INode)
```

#### func (*IrDefTop) FindAny

```go
func (me *IrDefTop) FindAny(where func(INode) bool) (firstMatch []INode)
```

#### func (*IrDefTop) FindByOrig

```go
func (me *IrDefTop) FindByOrig(orig atmolang.IAstNode) []INode
```

#### func (*IrDefTop) FindDescendants

```go
func (me *IrDefTop) FindDescendants(traverseIntoMatchesToo bool, max int, pred func(INode) bool) (paths [][]INode)
```

#### func (*IrDefTop) ForAllLocalDefs

```go
func (me *IrDefTop) ForAllLocalDefs(onLocalDef func(*IrDef) (done bool))
```

#### func (*IrDefTop) HasAnyOf

```go
func (me *IrDefTop) HasAnyOf(nodes ...INode) bool
```

#### func (*IrDefTop) HasDef

```go
func (me *IrDefTop) HasDef(maybeDef INode) (defIfLocal *IrDef)
```

#### func (*IrDefTop) HasErrors

```go
func (me *IrDefTop) HasErrors() bool
```

#### func (*IrDefTop) HasIdentDecl

```go
func (me *IrDefTop) HasIdentDecl(name string) bool
```

#### func (*IrDefTop) IsExt

```go
func (*IrDefTop) IsExt() bool
```

#### func (*IrDefTop) Let

```go
func (*IrDefTop) Let() *IrExprLetBase
```

#### func (*IrDefTop) OrigToks

```go
func (me *IrDefTop) OrigToks(node INode) (toks udevlex.Tokens)
```

#### func (*IrDefTop) Print

```go
func (me *IrDefTop) Print() atmolang.IAstNode
```

#### func (*IrDefTop) RefersTo

```go
func (me *IrDefTop) RefersTo(name string) (refersTo bool)
```

#### func (*IrDefTop) RefersToOrDefines

```go
func (me *IrDefTop) RefersToOrDefines(name string) (relatesTo bool)
```

#### func (*IrDefTop) RefsTo

```go
func (me *IrDefTop) RefsTo(name string) (refs []IExpr)
```

#### func (*IrDefTop) Walk

```go
func (me *IrDefTop) Walk(shouldTraverse func(curNodeAncestors []INode, curNode INode, curNodeDescendantsThatWillBeTraversedIfReturningTrue ...INode) bool)
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

#### func (*IrExprAtomBase) Let

```go
func (*IrExprAtomBase) Let() *IrExprLetBase
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

#### func (*IrExprBase) Let

```go
func (*IrExprBase) Let() *IrExprLetBase
```

#### func (*IrExprBase) Origin

```go
func (me *IrExprBase) Origin() atmolang.IAstNode
```

#### type IrExprLetBase

```go
type IrExprLetBase struct {
	Defs IrDefs

	Anns struct {
		// like `IrIdentName.Anns.Candidates`, contains the following `INode` types:
		// *atmoil.IrDef, *atmoil.IrArg, *atmoil.IrDefTop, atmosess.IrDefRef
		NamesInScope AnnNamesInScope
	}
}
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

#### func (*IrIdentBase) Let

```go
func (*IrIdentBase) Let() *IrExprLetBase
```

#### func (*IrIdentBase) Print

```go
func (me *IrIdentBase) Print() atmolang.IAstNode
```

#### type IrIdentDecl

```go
type IrIdentDecl struct {
	IrIdentBase
}
```


#### func (*IrIdentDecl) EquivTo

```go
func (me *IrIdentDecl) EquivTo(node INode) bool
```

#### func (*IrIdentDecl) IsDef

```go
func (*IrIdentDecl) IsDef() *IrDef
```

#### func (*IrIdentDecl) IsExt

```go
func (*IrIdentDecl) IsExt() bool
```

#### func (*IrIdentDecl) Let

```go
func (*IrIdentDecl) Let() *IrExprLetBase
```

#### type IrIdentName

```go
type IrIdentName struct {
	IrIdentBase
	IrExprLetBase

	Anns struct {
		// like `IrExprLetBase.Anns.NamesInScope`, contains the following `IIrNode` types:
		// *atmoil.IrDef, *atmoil.IrArg, *atmoil.IrDefTop, atmosess.IrDefRef
		Candidates []INode
	}
}
```


#### func (*IrIdentName) EquivTo

```go
func (me *IrIdentName) EquivTo(node INode) bool
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

#### func (*IrIdentName) Let

```go
func (me *IrIdentName) Let() *IrExprLetBase
```

#### func (*IrIdentName) Origin

```go
func (me *IrIdentName) Origin() atmolang.IAstNode
```

#### func (*IrIdentName) Print

```go
func (me *IrIdentName) Print() atmolang.IAstNode
```

#### func (*IrIdentName) RefersTo

```go
func (me *IrIdentName) RefersTo(name string) bool
```

#### func (*IrIdentName) ResolvesTo

```go
func (me *IrIdentName) ResolvesTo(n INode) bool
```

#### type IrLam

```go
type IrLam struct {
	IrExprBase
	OrigDef *atmolang.AstDef
	Arg     IrArg
	Body    IExpr
}
```


#### func (*IrLam) EquivTo

```go
func (me *IrLam) EquivTo(node INode) bool
```

#### func (*IrLam) IsDef

```go
func (*IrLam) IsDef() *IrDef
```

#### func (*IrLam) IsExt

```go
func (*IrLam) IsExt() bool
```

#### func (*IrLam) Let

```go
func (*IrLam) Let() *IrExprLetBase
```

#### func (*IrLam) Print

```go
func (me *IrLam) Print() atmolang.IAstNode
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
func (me *IrLitFloat) EquivTo(node INode) bool
```

#### func (*IrLitFloat) Print

```go
func (me *IrLitFloat) Print() atmolang.IAstNode
```

#### type IrLitTag

```go
type IrLitTag struct {
	Val string
}
```


#### func (*IrLitTag) EquivTo

```go
func (me *IrLitTag) EquivTo(node INode) bool
```

#### func (*IrLitTag) Print

```go
func (me *IrLitTag) Print() atmolang.IAstNode
```

#### type IrLitUint

```go
type IrLitUint struct {
	Val uint64
}
```


#### func (*IrLitUint) EquivTo

```go
func (me *IrLitUint) EquivTo(node INode) bool
```

#### func (*IrLitUint) Print

```go
func (me *IrLitUint) Print() atmolang.IAstNode
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
func (me *IrNonValue) EquivTo(node INode) bool
```

#### func (*IrNonValue) IsDef

```go
func (*IrNonValue) IsDef() *IrDef
```

#### func (*IrNonValue) IsExt

```go
func (*IrNonValue) IsExt() bool
```

#### func (*IrNonValue) Let

```go
func (*IrNonValue) Let() *IrExprLetBase
```

#### func (*IrNonValue) Print

```go
func (me *IrNonValue) Print() atmolang.IAstNode
```

#### type IrTopDefs

```go
type IrTopDefs []*IrDefTop
```


#### func (IrTopDefs) ByName

```go
func (me IrTopDefs) ByName(name string, onlyFor *atmolang.AstFile) (defs []*IrDefTop)
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
func (me *IrTopDefs) ReInitFrom(kitSrcFiles atmolang.AstFiles) (droppedTopLevelDefIdsAndNames map[string]string, newTopLevelDefIdsAndNames map[string]string, freshErrs atmo.Errors)
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
	Err *atmo.Error
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
