package atmoil

var (
	Build Builder
)

func (Builder) Appl1(callee IExpr, callArg IExpr) *IrAppl {
	if requireAtomicCalleeAndCallArg {
		if !callee.IsAtomic() {
			panic(callee)
		}
		if !callArg.IsAtomic() {
			panic(callArg)
		}
	}
	return &IrAppl{Callee: callee, CallArg: callArg}
}

func (Builder) ApplN(ctx *ctxIrFromAst, callee IExpr, callArgs ...IExpr) (appl *IrAppl) {
	if requireAtomicCalleeAndCallArg && !callee.IsAtomic() {
		panic(callee)
	}
	for i := range callArgs {
		if requireAtomicCalleeAndCallArg && !callArgs[i].IsAtomic() {
			panic(i)
		}
		if i == 0 {
			appl = &IrAppl{Callee: callee, CallArg: callArgs[i]}
		} else {
			appl = &IrAppl{Callee: ctx.ensureAtomic(appl), CallArg: callArgs[i]}
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

func (Builder) IdentTag(name string) *IrLitTag {
	return &IrLitTag{Val: name}
}

func (Builder) Undef() *IrNonValue {
	var node IrNonValue
	node.OneOf.Undefined = true
	return &node
}
