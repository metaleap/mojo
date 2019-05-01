package atmocorefn

var (
	Builder           AstBuilder
	builderSingletons struct {
		identTagTrue  *AstIdentTag
		identTagFalse *AstIdentTag
	}
)

type AstBuilder struct{}

func (AstBuilder) Appl(callee IAstIdent, arg IAstExprAtomic) *AstAppl {
	return &AstAppl{Callee: callee, Arg: arg}
}

func (AstBuilder) Appls(ctx *AstDef, callee IAstIdent, args ...IAstExprAtomic) *AstAppl {
	var appl AstAppl
	if len(args) == 2 {
		// 0
		appl.Callee, appl.Arg = callee, args[0]
		// 1
		oldappl := appl
		appl = AstAppl{Callee: ctx.ensureAstAtomFor(&oldappl, true).(IAstIdent), Arg: args[1]}
	} else {
		panic("OY")
		for i := range args {
			if i == 0 {
				appl.Callee, appl.Arg = callee, args[i]
			} else {
				appl = AstAppl{Callee: ctx.ensureAstAtomFor(&appl, true).(IAstIdent), Arg: args[i]}
			}
		}
	}
	return &appl
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

func (AstBuilder) IdentOp(name string) *AstIdentOp {
	return &AstIdentOp{AstIdentBase{Val: name}}
}

func (AstBuilder) IdentVar(name string) *AstIdentVar {
	return &AstIdentVar{AstIdentBase{Val: name}}
}

func (AstBuilder) IdentTag(name string) *AstIdentTag {
	return &AstIdentTag{AstIdentBase{Val: name}}
}

func (AstBuilder) IdentTagTrue() *AstIdentTag {
	if builderSingletons.identTagTrue == nil {
		builderSingletons.identTagTrue = Builder.IdentTag("True")
	}
	return builderSingletons.identTagTrue
}

func (AstBuilder) IdentTagFalse() *AstIdentTag {
	if builderSingletons.identTagFalse == nil {
		builderSingletons.identTagFalse = Builder.IdentTag("False")
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
