package atmolang

type AstBuilder struct{}

func (*AstBuilder) Let(body IAstExpr, defs ...AstDef) *AstExprLet {
	return &AstExprLet{Body: body, Defs: defs}
}

func (*AstBuilder) Def(name string, body IAstExpr, argNames ...string) (def AstDef) {
	def.Body, def.Name.Val, def.Args = body, name, make([]AstDefArg, len(argNames))
	for i := range argNames {
		def.Args[i].NameOrConstVal = &AstIdent{Val: argNames[i]}
	}
	return
}

func (*AstBuilder) Arg(nameOrConstVal IAstExprAtomic, affix IAstExpr) AstDefArg {
	return AstDefArg{NameOrConstVal: nameOrConstVal, Affix: affix}
}

func (*AstBuilder) Cases(scrutinee IAstExpr, alts ...AstCase) *AstExprCases {
	defaultindex := -1
	for i := range alts {
		if len(alts[i].Conds) == 0 {
			defaultindex = i
			break
		}
	}
	return &AstExprCases{defaultIndex: defaultindex, Scrutinee: scrutinee, Alts: alts}
}
