package molang

import (
	lex "github.com/go-leap/dev/lex"
)

// IExpr constructors

func Id(name string) *ExprIdent                { return &ExprIdent{Name: name} }
func Op(name string, lone bool) *ExprIdent     { return &ExprIdent{Name: name, OpLike: true, OpLone: lone} }
func Lf(lit float64) *ExprLitFloat             { return &ExprLitFloat{Lit: lit} }
func Lu(lit uint64, origBase int) *ExprLitUInt { return &ExprLitUInt{Lit: lit, Base: origBase} }
func Lr(lit rune) *ExprLitRune                 { return &ExprLitRune{Lit: lit} }
func Lt(lit string) *ExprLitText               { return &ExprLitText{Lit: lit} }

// other ISyn constructors

func Ap(callee IExpr, arg IExpr) *ExprAppl     { return &ExprAppl{Callee: callee, Arg: arg} }
func Ab(args []string, body IExpr) *ExprLambda { return &ExprLambda{Args: args, Body: body} }
func Nu(tag string, arity uint64) *ExprCtor    { return &ExprCtor{Tag: tag, Arity: int(arity)} }

type ISyn interface {
	init(lex.Tokens)
	isCore() bool
	isExpr() bool
	isLit() bool
	Pos() *lex.TokenMeta
	Toks() lex.Tokens
}

type IExpr interface {
	ISyn
}

type syn struct {
	toks lex.Tokens
	// root   *SynMod
	// parent ISyn
}

func (this *syn) init(toks lex.Tokens) { this.toks = toks }

func (*syn) isCore() bool { return true }

func (*syn) isExpr() bool { return false }

func (*syn) isLit() bool { return false }

func (this *syn) Pos() *lex.TokenMeta { return &this.toks[0].Meta }

func (this *syn) Toks() lex.Tokens { return this.toks }

type expr struct{ syn }

func (*expr) isExpr() bool { return true }

type exprLit struct{ expr }

func (*exprLit) isLit() bool { return true }

type SynDef struct {
	syn
	Name     string
	Args     []string
	Body     IExpr
	TopLevel bool
}

type SynCaseAlt struct {
	syn
	Tag   string
	Binds []string
	Body  IExpr
}

type ExprCtor struct {
	expr
	Tag   string
	Arity int
}

type ExprAppl struct {
	expr
	Callee IExpr
	Arg    IExpr
}

type ExprLambda struct {
	expr
	Args []string
	Body IExpr
}

type ExprLetIn struct {
	expr
	Rec  bool
	Defs []*SynDef
	Body IExpr
}

type ExprCaseOf struct {
	expr
	Scrut IExpr
	Alts  []*SynCaseAlt
}

type ExprIdent struct {
	expr
	Name   string
	OpLike bool
	OpLone bool
}

type ExprLitFloat struct {
	exprLit
	Lit float64
}

type ExprLitUInt struct {
	exprLit
	Base int
	Lit  uint64
}

type ExprLitRune struct {
	exprLit
	Lit rune
}

type ExprLitText struct {
	exprLit
	Lit string
}
