package atmolang_irfun

var (
	B                 AstBuilder
	builderSingletons struct {
		litUndef AstLitUndef
	}
)

type AstBuilder struct{}

func (AstBuilder) Appl1(atomicCallee IAstExpr, atomicArg IAstExpr) *AstAppl {
	if !atomicCallee.IsAtomic() {
		panic(atomicCallee)
	}
	if !atomicArg.IsAtomic() {
		panic(atomicArg)
	}
	return &AstAppl{AtomicCallee: atomicCallee, AtomicArg: atomicArg}
}

func (AstBuilder) ApplN(ctx *ctxAstInit, atomicCallee IAstExpr, atomicArgs ...IAstExpr) (appl *AstAppl) {
	if !atomicCallee.IsAtomic() {
		panic(atomicCallee)
	}
	for i := range atomicArgs {
		if !atomicArgs[i].IsAtomic() {
			panic(atomicArgs[i])
		}
		if i == 0 {
			appl = &AstAppl{AtomicCallee: atomicCallee, AtomicArg: atomicArgs[i]}
		} else {
			appl = &AstAppl{AtomicCallee: ctx.ensureAstAtomFor(appl), AtomicArg: atomicArgs[i]}
		}
	}
	return
}

func (AstBuilder) IdentName(name string) *AstIdentName {
	return &AstIdentName{AstIdentBase{Val: name}, AstExprLetBase{}, nil}
}

func (AstBuilder) IdentVar(name string) *AstIdentVar {
	return &AstIdentVar{AstIdentBase{Val: name}}
}

func (AstBuilder) IdentTag(name string) *AstIdentTag {
	return &AstIdentTag{AstIdentBase{Val: name}}
}

func (AstBuilder) LitUndef() *AstLitUndef { return &builderSingletons.litUndef }
