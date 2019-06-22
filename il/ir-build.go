package atmoil

var (
	Build Builder
)

type Builder struct{}

func (Builder) Appl1(atomicCallee IIrExpr, atomicArg IIrExpr) *IrAppl {
	if !atomicCallee.IsAtomic() {
		panic(atomicCallee)
	}
	if !atomicArg.IsAtomic() {
		panic(atomicArg)
	}
	return &IrAppl{AtomicCallee: atomicCallee, AtomicArg: atomicArg}
}

func (Builder) ApplN(ctx *ctxIrInit, atomicCallee IIrExpr, atomicArgs ...IIrExpr) (appl *IrAppl) {
	if !atomicCallee.IsAtomic() {
		panic(atomicCallee)
	}
	for i := range atomicArgs {
		if !atomicArgs[i].IsAtomic() {
			panic(i)
		}
		if i == 0 {
			appl = &IrAppl{AtomicCallee: atomicCallee, AtomicArg: atomicArgs[i]}
		} else {
			appl = &IrAppl{AtomicCallee: ctx.ensureAtomic(appl), AtomicArg: atomicArgs[i]}
		}
	}
	return
}

func (Builder) IdentName(name string) *IrIdentName {
	return &IrIdentName{IrIdentBase: IrIdentBase{Val: name}}
}

func (Builder) IdentTag(name string) *IrIdentTag {
	return &IrIdentTag{IrIdentBase{Val: name}}
}
