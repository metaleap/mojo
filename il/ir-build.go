package atmoil

var (
	Build Builder
)

type Builder struct{}

func (Builder) Appl1(atomicCallee IExpr, atomicArg IExpr) *IrAppl {
	if !atomicCallee.IsAtomic() {
		panic(atomicCallee)
	}
	if !atomicArg.IsAtomic() {
		panic(atomicArg)
	}
	return &IrAppl{AtomicCallee: atomicCallee, AtomicArg: atomicArg}
}

func (Builder) ApplN(ctx *ctxIrInit, atomicCallee IExpr, atomicArgs ...IExpr) (appl *IrAppl) {
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

func (Builder) IdentNameCopy(identBase *IrIdentBase) *IrIdentName {
	return &IrIdentName{IrIdentBase: *identBase}
}

func (Builder) IdentTag(name string) *IrIdentTag {
	return &IrIdentTag{IrIdentBase{Val: name}}
}

func (Builder) Undef() *IrSpecial {
	var node IrSpecial
	node.OneOf.Undefined = true
	return &node
}
