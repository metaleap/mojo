package main

func llEmit(ll_something Any) {
	switch ll := ll_something.(type) {
	case *LLModule:
		llEmitModule(ll)
	case *LLGlobal:
		llEmitGlobal(ll)
	case *LLFunc:
		llEmitFunc(ll)
	case *LLBasicBlock:
		llEmitBlock(ll)
	case LLStmtBr:
		llEmitStmtBr(&ll)
	case LLStmtRet:
		llEmitStmtRet(&ll)
	case LLStmtLet:
		llEmitStmtLet(&ll)
	case LLStmtSwitch:
		llEmitStmtSwitch(&ll)
	case LLTypeInt:
		llEmitTypeInt(&ll)
	case LLTypePtr:
		llEmitTypePtr(&ll)
	case LLTypeArr:
		llEmitTypeArr(&ll)
	case LLTypeStruct:
		llEmitTypeStruct(&ll)
	case LLTypeFun:
		llEmitTypeFun(&ll)
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

func llEmitBlock(ll_block *LLBasicBlock) {
	write(ll_block.name)
	write(Str(":\n"))
	for i := range ll_block.stmts {
		write(Str("  "))
		llEmit(&ll_block.stmts[i])
		write(Str("\n"))
	}
}

func llEmitStmtBr(ll_stmt_br *LLStmtBr) {
	write(Str("br label %"))
	write(ll_stmt_br.block_name)
}

func llEmitStmtRet(ll_stmt_ret *LLStmtRet) {
	write(Str("ret "))
	llEmit(ll_stmt_ret.expr)
}

func llEmitStmtLet(ll_stmt_let *LLStmtLet) {
	write(Str("%"))
	write(ll_stmt_let.name)
	write(Str(" = "))
	llEmit(ll_stmt_let.expr)
}

func llEmitStmtSwitch(ll_stmt_switch *LLStmtSwitch) {
	write(Str("switch"))
	llEmit(ll_stmt_switch.comparee)
	write(Str(", label %"))
	llEmit(ll_stmt_switch.default_block_name)
	write(Str(" ["))
	for i := range ll_stmt_switch.cases {
		llEmit(ll_stmt_switch.cases[i].expr)
		write(Str(", label %"))
		write(ll_stmt_switch.cases[i].block_name)
		if i > 0 {
			write(Str(",\n    "))
		}
	}
	write(Str("]"))
}

func llEmitTypeInt(ll_type_int *LLTypeInt) {
	write(Str("i"))
	write(uintToStr(uint64(ll_type_int.bit_width), 10, 1, nil))
}

func llEmitTypePtr(ll_type_ptr *LLTypePtr) {
	llEmit(ll_type_ptr.ty)
	write(Str("*"))
}

func llEmitTypeArr(ll_type_arr *LLTypeArr) {
	write(Str("["))
	write(uintToStr(uint64(ll_type_arr.size), 10, 1, nil))
	write(Str(" x "))
	llEmit(ll_type_arr.ty)
	write(Str("]"))
}

func llEmitTypeStruct(ll_type_struct *LLTypeStruct) {
	write(Str("{"))
	for i := range ll_type_struct.fields {
		if i > 0 {
			write(Str(", "))
		}
		llEmit(ll_type_struct.fields[i])
	}
	write(Str("}"))
}

func llEmitTypeFun(ll_type_fun *LLTypeFun) {
	llEmit(ll_type_fun.ty)
	write(Str("("))
	for i := range ll_type_fun.params {
		if i > 0 {
			write(Str(", "))
		}
		llEmit(ll_type_fun.params[i])
	}
	write(Str(")"))
}
