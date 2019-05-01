package atmocorefn

var (
	B                 AstBuilder
	builderSingletons struct {
		identTagTrue  *AstIdentTag
		identTagFalse *AstIdentTag
	}
)

type AstBuilder struct{}

func (AstBuilder) Appl(callee IAstIdent, arg IAstExprAtomic) *AstAppl {
	return &AstAppl{Callee: callee, Arg: arg}
}

func (AstBuilder) Appls(ctx *AstDef, callee IAstIdent, args ...IAstExprAtomic) (appl *AstAppl) {
	for i := range args {
		if i == 0 {
			appl = &AstAppl{Callee: callee, Arg: args[i]}
		} else {
			appl = &AstAppl{Callee: ctx.ensureAstAtomFor(appl, true).(IAstIdent), Arg: args[i]}
		}
	}
	return
}

func (AstBuilder) Case(ifThis IAstExpr, thenThat IAstExpr) *AstCases {
	return &AstCases{Ifs: [][]IAstExpr{{ifThis}}, Thens: []IAstExpr{thenThat}}
}

func (AstBuilder) IdentName(name string) *AstIdentName {
	return &AstIdentName{AstIdentBase{Val: name}}
}
func (AstBuilder) IdentEmptyParens() *AstIdentEmptyParens {
	return &AstIdentEmptyParens{AstIdentBase{Val: "()"}}
}

func (AstBuilder) IdentVar(name string) *AstIdentVar {
	return &AstIdentVar{AstIdentBase{Val: name}}
}

func (AstBuilder) IdentTag(name string) *AstIdentTag {
	return &AstIdentTag{AstIdentBase{Val: name}}
}

func (AstBuilder) IdentTagTrue() *AstIdentTag {
	if builderSingletons.identTagTrue == nil {
		builderSingletons.identTagTrue = B.IdentTag("True")
	}
	return builderSingletons.identTagTrue
}

func (AstBuilder) IdentTagFalse() *AstIdentTag {
	if builderSingletons.identTagFalse == nil {
		builderSingletons.identTagFalse = B.IdentTag("False")
	}
	return builderSingletons.identTagFalse
}

func (AstBuilder) DefArgs(names ...string) (r []AstDefArg) {
	r = make([]AstDefArg, len(names))
	for i := range names {
		r[i].AstIdentName.Val = names[i]
	}
	return
}
