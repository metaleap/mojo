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
	case LLInstrComment:
		llEmitInstrComment(&ll)
	case LLInstrBrTo:
		llEmitInstrBrTo(&ll)
	case LLInstrBrIf:
		llEmitInstrBrIf(&ll)
	case LLInstrRet:
		llEmitInstrRet(&ll)
	case LLInstrUnreachable:
		llEmitInstrUnreachable()
	case LLInstrLet:
		llEmitInstrLet(&ll)
	case LLInstrSwitch:
		llEmitInstrSwitch(&ll)
	case LLInstrConvert:
		llEmitInstrConvert(&ll)
	case LLTypeVoid:
		llEmitTypeVoid()
	case LLTypeInt:
		llEmitTypeInt(&ll)
	case LLTypePtr:
		llEmitTypePtr(&ll)
	case LLTypeArr:
		llEmitTypeArr(&ll)
	case LLTypeStruct:
		llEmitTypeStruct(&ll)
	case LLTypeFunc:
		llEmitTypeFunc(&ll)
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
	case LLInstrAlloca:
		llEmitInstrAlloca(&ll)
	case LLInstrLoad:
		llEmitInstrLoad(&ll)
	case LLInstrCall:
		llEmitInstrCall(&ll)
	case LLInstrBinOp:
		llEmitInstrBinOp(&ll)
	case LLInstrCmpI:
		llEmitInstrCmpI(&ll)
	case LLInstrPhi:
		llEmitInstrPhi(&ll)
	case LLInstrGep:
		llEmitInstrGep(&ll)
	case Any:
		switch ll_inner := ll.(type) {
		case nil:
			unreachable()
		default:
			llEmit(ll_inner)
		}
	case nil:
		unreachable()
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
	write(Str("\n"))
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
	if ll_func.external {
		write(Str("declare "))
	} else {
		write(Str("\ndefine "))
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
	if !ll_func.external {
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
	for i := range ll_block.instrs {
		write(Str("  "))
		llEmit(ll_block.instrs[i])
		write(Str("\n"))
	}
}

func llEmitTypeVoid() {
	write(Str("void"))
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

func llEmitTypeFunc(ll_type_fun *LLTypeFunc) {
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
	_, is_void := ll_expr_typed.ty.(LLTypeVoid)
	if !is_void {
		write(Str(" "))
		_, is_instr := ll_expr_typed.expr.(LLInstr)
		if is_instr {
			write(Str("("))
		}
		llEmit(ll_expr_typed.expr)
		if is_instr {
			write(Str(")"))
		}
	}
}

func llEmitInstrComment(ll_instr_comment *LLInstrComment) {
	write(Str("; "))
	write(ll_instr_comment.comment_text)
}

func llEmitInstrBrIf(ll_instr_br_if *LLInstrBrIf) {
	write(Str("br i1 "))
	llEmit(ll_instr_br_if.cond)
	write(Str(", label %"))
	write(ll_instr_br_if.block_name_if_true)
	write(Str(", label %"))
	write(ll_instr_br_if.block_name_if_false)
}

func llEmitInstrBrTo(ll_instr_br_to *LLInstrBrTo) {
	write(Str("br label %"))
	write(ll_instr_br_to.block_name)
}

func llEmitInstrRet(ll_instr_ret *LLInstrRet) {
	write(Str("ret "))
	llEmit(ll_instr_ret.expr)
}

func llEmitInstrUnreachable() {
	write(Str("unreachable"))
}

func llEmitInstrLet(ll_instr_let *LLInstrLet) {
	write(Str("%"))
	write(ll_instr_let.name)
	write(Str(" = "))
	llEmit(ll_instr_let.instr)
}

func llEmitInstrSwitch(ll_instr_switch *LLInstrSwitch) {
	write(Str("switch "))
	llEmit(ll_instr_switch.comparee)
	write(Str(", label %"))
	write(ll_instr_switch.default_block_name)
	write(Str(" ["))
	for i := range ll_instr_switch.cases {
		llEmit(ll_instr_switch.cases[i].expr)
		write(Str(", label %"))
		write(ll_instr_switch.cases[i].block_name)
		if i > 0 {
			write(Str(",\n    "))
		}
	}
	write(Str("]"))
}

func llEmitInstrAlloca(ll_instr_alloca *LLInstrAlloca) {
	write(Str("alloca "))
	llEmit(ll_instr_alloca.ty)
	write(Str(", "))
	llEmit(ll_instr_alloca.num_elems)
}

func llEmitInstrLoad(ll_instr_load *LLInstrLoad) {
	write(Str("load "))
	llEmit(ll_instr_load.ty)
	write(Str(", "))
	llEmit(ll_instr_load.expr)
}

func llEmitInstrCall(ll_instr_call *LLInstrCall) {
	write(Str("call "))
	llEmit(ll_instr_call.ty)
	write(Str(" "))
	llEmit(ll_instr_call.callee)
	write(Str("("))
	for i := range ll_instr_call.args {
		if i > 0 {
			write(Str(", "))
		}
		llEmit(ll_instr_call.args[i])
	}
	write(Str(")"))
}

func llEmitInstrConvert(ll_instr_convert *LLInstrConvert) {
	switch ll_instr_convert.convert_kind {
	case ll_convert_int_to_ptr:
		write(Str("inttoptr "))
	case ll_convert_ptr_to_int:
		write(Str("ptrtoint "))
	case ll_convert_trunc:
		write(Str("trunc "))
	default:
		panic(ll_instr_convert.convert_kind)
	}
	llEmit(ll_instr_convert.expr)
	write(Str(" to "))
	llEmit(ll_instr_convert.ty)
}

func llEmitInstrBinOp(ll_instr_bin_op *LLInstrBinOp) {
	var op_kind string
	switch ll_instr_bin_op.op_kind {
	case ll_bin_op_add:
		op_kind = "add"
	case ll_bin_op_udiv:
		op_kind = "udiv"
	default:
		fail(ll_instr_bin_op.op_kind)
	}

	write(Str(op_kind))
	write(Str(" "))
	llEmit(ll_instr_bin_op.ty)
	write(Str(" "))
	llEmit(ll_instr_bin_op.lhs)
	write(Str(", "))
	llEmit(ll_instr_bin_op.rhs)
}

func llEmitInstrCmpI(ll_instr_cmp_i *LLInstrCmpI) {
	var cmp_kind string
	switch ll_instr_cmp_i.cmp_kind {
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
		fail(ll_instr_cmp_i.cmp_kind)
	}

	write(Str("icmp "))
	write(Str(cmp_kind))
	write(Str(" "))
	llEmit(ll_instr_cmp_i.ty)
	write(Str(" "))
	llEmit(ll_instr_cmp_i.lhs)
	write(Str(", "))
	llEmit(ll_instr_cmp_i.rhs)
}

func llEmitInstrPhi(ll_instr_phi *LLInstrPhi) {
	write(Str("phi "))
	llEmit(ll_instr_phi.ty)
	for i := range ll_instr_phi.predecessors {
		if i > 0 {
			write(Str(","))
		}
		write(Str(" ["))
		llEmit(ll_instr_phi.predecessors[i].expr)
		write(Str(", %"))
		write(ll_instr_phi.predecessors[i].block_name)
		write(Str("]"))
	}
}

func llEmitInstrGep(ll_instr_gep *LLInstrGep) {
	write(Str("getelementptr "))
	llEmit(ll_instr_gep.ty)
	write(Str(", "))
	llEmit(ll_instr_gep.base_ptr)
	for i := range ll_instr_gep.indices {
		write(Str(", "))
		llEmit(ll_instr_gep.indices[i])
	}
}
