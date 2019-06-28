package atmolang

import (
	"github.com/go-leap/str"
)

var (
	Build Builder
)

type Builder struct{}

func (Builder) Ident(val string) *AstIdent {
	isnotopish := val[0] == '_' || ustr.BeginsLetter(val)
	return &AstIdent{Val: val, IsTag: ustr.BeginsUpper(val), IsOpish: !isnotopish}
}

func (Builder) LitFloat(val float64) *AstExprLitFloat {
	return &AstExprLitFloat{Val: val}
}

func (Builder) LitUint(val uint64) *AstExprLitUint {
	return &AstExprLitUint{Val: val}
}

func (Builder) LitRune(val int32) *AstExprLitUint {
	return &AstExprLitUint{Val: uint64(val)}
}

func (Builder) LitStr(val string) *AstExprLitStr {
	return &AstExprLitStr{Val: val}
}

func (Builder) Let(body IAstExpr, defs ...AstDef) *AstExprLet {
	return &AstExprLet{Body: body, Defs: defs}
}

func (Builder) Appl(callee IAstExpr, args ...IAstExpr) *AstExprAppl {
	return &AstExprAppl{Callee: callee, Args: args}
}

func (Builder) Def(name string, body IAstExpr, argNames ...string) *AstDef {
	def := AstDef{Body: body, Args: make([]AstDefArg, len(argNames))}
	def.Name.Val = name
	for i := range argNames {
		def.Args[i].NameOrConstVal = &AstIdent{Val: argNames[i]}
	}
	return &def
}

func (Builder) Arg(nameOrConstVal IAstExprAtomic, affix IAstExpr) *AstDefArg {
	return &AstDefArg{NameOrConstVal: nameOrConstVal, Affix: affix}
}

func (Builder) Cases(scrutinee IAstExpr, alts ...AstCase) *AstExprCases {
	defaultindex := -1
	for i := range alts {
		if len(alts[i].Conds) == 0 {
			defaultindex = i
			break
		}
	}
	return &AstExprCases{defaultIndex: defaultindex, Scrutinee: scrutinee, Alts: alts}
}
