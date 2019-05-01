package atmolang

import (
	"github.com/go-leap/str"
)

var (
	B                 AstBuilder
	builderSingletons struct {
		identTrue  *AstIdent
		identFalse *AstIdent
	}
)

type AstBuilder struct{}

func (AstBuilder) Ident(val string) *AstIdent {
	isnotopish := len(val) == 0 || val[0] == '_' || ustr.BeginsLetter(val)
	return &AstIdent{Val: val, IsTag: ustr.BeginsUpper(val), IsOpish: !isnotopish}
}

func (AstBuilder) IdentTrue() *AstIdent {
	if builderSingletons.identTrue == nil {
		builderSingletons.identTrue = B.Ident("True")
	}
	return builderSingletons.identTrue
}

func (AstBuilder) IdentFalse() *AstIdent {
	if builderSingletons.identFalse == nil {
		builderSingletons.identFalse = B.Ident("False")
	}
	return builderSingletons.identFalse
}

func (AstBuilder) LitFloat(val float64) *AstExprLitFloat {
	return &AstExprLitFloat{Val: val}
}

func (AstBuilder) LitUint(val uint64) *AstExprLitUint {
	return &AstExprLitUint{Val: val}
}

func (AstBuilder) LitRune(val rune) *AstExprLitRune {
	return &AstExprLitRune{Val: val}
}

func (AstBuilder) LitStr(val string) *AstExprLitStr {
	return &AstExprLitStr{Val: val}
}

func (AstBuilder) Let(body IAstExpr, defs ...AstDef) *AstExprLet {
	return &AstExprLet{Body: body, Defs: defs}
}

func (AstBuilder) Appl(callee IAstExpr, args ...IAstExpr) *AstExprAppl {
	return &AstExprAppl{Callee: callee, Args: args}
}

func (AstBuilder) Def(name string, body IAstExpr, argNames ...string) *AstDef {
	def := AstDef{Body: body, Args: make([]AstDefArg, len(argNames))}
	def.Name.Val = name
	for i := range argNames {
		def.Args[i].NameOrConstVal = &AstIdent{Val: argNames[i]}
	}
	return &def
}

func (AstBuilder) Arg(nameOrConstVal IAstExprAtomic, affix IAstExpr) *AstDefArg {
	return &AstDefArg{NameOrConstVal: nameOrConstVal, Affix: affix}
}

func (AstBuilder) Cases(scrutinee IAstExpr, alts ...AstCase) *AstExprCases {
	defaultindex := -1
	for i := range alts {
		if len(alts[i].Conds) == 0 {
			defaultindex = i
			break
		}
	}
	return &AstExprCases{defaultIndex: defaultindex, Scrutinee: scrutinee, Alts: alts}
}
