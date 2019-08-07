package atmoil

var (
	BuildIr IrBuild
)

func (IrBuild) Appl1(callee IIrExpr, callArg IIrExpr) *IrAppl {
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

func (IrBuild) ApplN(ctx *ctxIrFromAst, callee IIrExpr, callArgs ...IIrExpr) (appl *IrAppl) {
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

func (IrBuild) IdentName(name string) *IrIdentName {
	return &IrIdentName{IrIdentBase: IrIdentBase{Val: name}}
}

func (IrBuild) IdentNameCopy(identBase *IrIdentBase) *IrIdentName {
	return &IrIdentName{IrIdentBase: *identBase}
}

func (IrBuild) IdentTag(name string) *IrLitTag {
	return &IrLitTag{Val: name}
}

func (IrBuild) Undef() *IrNonValue {
	var node IrNonValue
	node.OneOf.Undefined = true
	return &node
}
