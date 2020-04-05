package main

func llEmit(ll_something Any) {
	switch ll := ll_something.(type) {
	case *LLModule:
		llEmitModule(ll)
	case *LLGlobal:
		llEmitGlobal(ll)
	case *LLFunc:
		llEmitFunc(ll)
	case LLType, LLExpr, LLStmt:
		switch ll_other := ll.(type) {
		default:
			llEmit(ll_other)
		}
	default:
		fail(ll)
	}
}

func llEmitModule(ll_mod *LLModule) {
	write(Str("\ntarget datalayout = \""))
	write(ll_mod.target_datalayout)
	write(Str("\"\ntarget triple = \""))
	write(ll_mod.target_triple)
	write(Str("\"\n\n"))
	for i := range ll_mod.globals {
		llEmit(&ll_mod.globals[i])
		write(Str("\n"))
	}
	write(Str("\n\n"))
	for i := range ll_mod.funcs {
		llEmit(&ll_mod.funcs[i])
		write(Str("\n"))
	}
}

func llEmitGlobal(ll_global *LLGlobal) {
	write(Str("@"))
	write(ll_global.name)
	if ll_global.constant {
		write(Str(" = constant "))
	} else if ll_global.external {
		write(Str(" = external global "))
	} else {
		write(Str(" = global "))
	}
	llEmit(ll_global.ty)
	if ll_global.initializer != nil {
		write(Str(" "))
		llEmit(ll_global.initializer)
	}
}

func llEmitFunc(ll_func *LLFunc) {
	if len(ll_func.basic_blocks) == 0 {
		write(Str("declare "))
	} else {
		write(Str("\n\ndefine "))
	}
	llEmit(ll_func.ty)
	write(Str(" @"))
	write(ll_func.name)
	write(Str("("))
	for i := range ll_func.params {
		param := &ll_func.params[i]
		if i > 0 {
			write(Str(", "))
		}
		llEmit(param.ty)
		if len(param.name) != 0 {
			write(Str(" %"))
			write(param.name)
		}
	}
	write(Str(")"))
	if len(ll_func.basic_blocks) != 0 {
		write(Str(" {\n"))
		for i := range ll_func.basic_blocks {
			llEmit(&ll_func.basic_blocks[i])
		}
		write(Str("}\n"))
	}
}
