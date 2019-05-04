package atmolang_irfun

var (
	B                 AstBuilder
	builderSingletons struct {
		identTagTrue  *AstIdentTag
		identTagFalse *AstIdentTag
	}
)

type AstBuilder struct{}

func (AstBuilder) Appl(callee IAstExpr, arg IAstExpr) *AstAppl {
	if !callee.IsAtomic() {
		panic(callee)
	}
	if !arg.IsAtomic() {
		panic(arg)
	}
	return &AstAppl{AtomicCallee: callee, AtomicArg: arg}
}

func (AstBuilder) Appls(ctx *ctxAstInit, callee IAstExpr, args ...IAstExpr) (appl *AstAppl) {
	if !callee.IsAtomic() {
		panic(callee)
	}
	for i := range args {
		if !args[i].IsAtomic() {
			panic(args[i])
		}
		if i == 0 {
			appl = &AstAppl{AtomicCallee: callee, AtomicArg: args[i]}
		} else {
			appl = &AstAppl{AtomicCallee: ctx.ensureAstAtomFor(appl), AtomicArg: args[i]}
		}
	}
	return
}

func (AstBuilder) Case(ifThis IAstExpr, thenThat IAstExpr) *AstCases {
	return &AstCases{Ifs: []IAstExpr{ifThis}, Thens: []IAstExpr{thenThat}}
}

func (AstBuilder) IdentName(name string) *AstIdentName {
	return &AstIdentName{AstIdentBase{Val: name}, AstExprLetBase{}}
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
