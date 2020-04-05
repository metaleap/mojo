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
	case LLExprIdentLocal:
		llEmitExprIdentLocal(ll)
	case LLExprIdentGlobal:
		llEmitExprIdentGlobal(ll)
	case LLExprLitInt:
		llEmitExprLitInt(ll)
	case LLExprLitStr:
		llEmitExprLitStr(ll)
	case LLExprTyped:
		llEmitExprTyped(&ll)
	case LLExprAlloca:
		llEmitExprAlloca(&ll)
	case LLExprLoad:
		llEmitExprLoad(&ll)
	case LLExprCall:
		llEmitExprCall(&ll)
	case LLExprBinOp:
		llEmitExprBinOp(&ll)
	case LLExprCmpI:
		llEmitExprCmpI(&ll)
	case LLExprPhi:
		llEmitExprPhi(&ll)
	case LLExprGep:
		llEmitExprGep(&ll)
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

func llEmitExprIdentLocal(ll_expr_ident_local LLExprIdentLocal) {
	write(Str("%"))
	write(ll_expr_ident_local)
}

func llEmitExprIdentGlobal(ll_expr_ident_global LLExprIdentGlobal) {
	write(Str("@"))
	write(ll_expr_ident_global)
}

func llEmitExprLitInt(ll_expr_lit_int LLExprLitInt) {
	write(uintToStr(uint64(ll_expr_lit_int), 10, 1, nil))
}

func llEmitExprLitStr(ll_expr_lit_str LLExprLitStr) {
	write(Str("c\""))
	for i := range ll_expr_lit_str {
		if char := ll_expr_lit_str[i]; char >= 32 && char < 127 {
			write(ll_expr_lit_str[i : i+1])
		} else {
			write(Str("\\"))
			write(uintToStr(uint64(char), 16, 2, nil))
		}
	}
	write(Str("\""))
}

func llEmitExprTyped(ll_expr_typed *LLExprTyped) {
	llEmit(ll_expr_typed.ty)
	write(Str(" "))
	llEmit(ll_expr_typed.expr)
}

func llEmitExprAlloca(ll_expr_alloca *LLExprAlloca) {
	write(Str("alloca "))
	llEmit(ll_expr_alloca.ty)
	write(Str(", "))
	llEmit(ll_expr_alloca.num_elems)
}

func llEmitExprLoad(ll_expr_load *LLExprLoad) {
	write(Str("load "))
	llEmit(ll_expr_load.ty)
	write(Str(", "))
	llEmit(ll_expr_load.ty)
	write(Str("* "))
	llEmit(ll_expr_load.expr)
}

func llEmitExprCall(ll_expr_call *LLExprCall) {
	write(Str("call "))
	llEmit(ll_expr_call.callee)
	write(Str("("))
	for i := range ll_expr_call.args {
		if i > 0 {
			write(Str(", "))
		}
		llEmit(ll_expr_call.args[i])
	}
	write(Str(")"))
}

func llEmitExprBinOp(ll_expr_bin_op *LLExprBinOp) {
	var op_kind string
	switch ll_expr_bin_op.op_kind {
	case ll_bin_op_add:
		op_kind = "add"
	default:
		fail(ll_expr_bin_op.op_kind)
	}

	write(Str(op_kind))
	write(Str(" "))
	llEmit(ll_expr_bin_op.ty)
	write(Str(" "))
	llEmit(ll_expr_bin_op.lhs)
	write(Str(", "))
	llEmit(ll_expr_bin_op.rhs)
}

func llEmitExprCmpI(ll_expr_cmp_i *LLExprCmpI) {
	var cmp_kind string
	switch ll_expr_cmp_i.cmp_kind {
	case ll_cmp_i_eq:
		cmp_kind = "eq"
	case ll_cmp_i_ne:
		cmp_kind = "ne"
	case ll_cmp_i_ugt:
		cmp_kind = "ugt"
	case ll_cmp_i_uge:
		cmp_kind = "uge"
	case ll_cmp_i_ult:
		cmp_kind = "ult"
	case ll_cmp_i_ule:
		cmp_kind = "ule"
	case ll_cmp_i_sgt:
		cmp_kind = "sgt"
	case ll_cmp_i_sge:
		cmp_kind = "sge"
	case ll_cmp_i_slt:
		cmp_kind = "slt"
	case ll_cmp_i_sle:
		cmp_kind = "sle"
	default:
		fail(ll_expr_cmp_i.cmp_kind)
	}

	write(Str("icmp "))
	write(Str(cmp_kind))
	write(Str(" "))
	llEmit(ll_expr_cmp_i.ty)
	write(Str(" "))
	llEmit(ll_expr_cmp_i.lhs)
	write(Str(", "))
	llEmit(ll_expr_cmp_i.rhs)
}

func llEmitExprPhi(ll_expr_phi *LLExprPhi) {
	write(Str("phi "))
	llEmit(ll_expr_phi.ty)
	for i := range ll_expr_phi.predecessors {
		if i > 0 {
			write(Str(","))
		}
		write(Str(" ["))
		llEmit(ll_expr_phi.predecessors[i].expr)
		write(Str(", %"))
		write(ll_expr_phi.predecessors[i].block_name)
		write(Str("]"))
	}
}

func llEmitExprGep(ll_expr_gep *LLExprGep) {
	write(Str("getelementptr "))
	llEmit(ll_expr_gep.ty)
	write(Str(", "))
	llEmit(ll_expr_gep.base_ptr)
	for i := range ll_expr_gep.indices {
		write(Str(", "))
		llEmit(ll_expr_gep.indices[i])
	}
}
