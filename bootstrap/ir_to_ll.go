package main

const llmodule_default_target_datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
const llmodule_default_target_triple = "x86_64-unknown-linux-gnu"
const ll_target_word_bit_width = 64

type CtxIrToLL struct {
	ir          *Ir
	ll_mod      *LLModule
	num_funcs   int
	num_globals int
}

func llModuleFrom(ir *Ir, def_name Str) LLModule {
	ret_mod := LLModule{
		target_datalayout: Str(llmodule_default_target_datalayout),
		target_triple:     Str(llmodule_default_target_triple),
		globals:           allocˇLLGlobal(len(ir.defs)),
		funcs:             allocˇLLFunc(len(ir.defs)),
	}
	ret_mod.anns.orig_ir = ir
	ctx := CtxIrToLL{
		ir:          ir,
		ll_mod:      &ret_mod,
		num_globals: 0,
		num_funcs:   0,
	}
	for i := range ir.defs {
		if strEql(ir.defs[i].name, def_name) {
			_ = irToLL(&ctx, IrExprDefRef(i)).(LLExprIdentGlobal)
		}
	}
	ret_mod.globals = ret_mod.globals[0:ctx.num_globals]
	ret_mod.funcs = ret_mod.funcs[0:ctx.num_funcs]
	return ret_mod
}

func irToLL(ctx *CtxIrToLL, ir_expr IrExpr) Any {
	switch expr := ir_expr.(type) {
	case IrExprLitInt:
		return LLExprLitInt(expr)
	case IrExprLitStr:
		return LLExprLitStr(expr)
	case IrExprTag:
		return LLExprIdentLocal(expr)
	case IrExprDefRef:
		ref_def := &ctx.ir.defs[expr]
		if ident := llTopLevelNameFrom(ctx.ll_mod, ref_def, ctx.num_globals, ctx.num_funcs); ident != nil {
			return ident
		}
		switch ll_top_level_item := irToLL(ctx, ref_def.body).(type) {
		case LLGlobal:
			ll_top_level_item.anns.orig_ir_def = ref_def
			if 0 == len(ll_top_level_item.name) {
				ll_top_level_item.name = ref_def.name
			}
			ctx.ll_mod.globals[ctx.num_globals] = ll_top_level_item
			ctx.num_globals++
			return LLExprIdentGlobal(ll_top_level_item.name)
		case LLFunc:
			ll_top_level_item.anns.orig_ir_def = ref_def
			if 0 == len(ll_top_level_item.name) {
				ll_top_level_item.name = ref_def.name
			}
			ctx.ll_mod.funcs[ctx.num_funcs] = ll_top_level_item
			ctx.num_funcs++
			return LLExprIdentGlobal(ll_top_level_item.name)
		default:
			fail("invalid reference to '", ref_def.name, "'")
		}
	case IrExprForm:
		switch callee := expr[0].(type) {
		case IrExprSlashed:
			kwd := callee[0].(IrExprIdent)
			if len(callee) == 1 {
				if strEq(kwd, "define") {
					return irToLLFuncDef(ctx, expr)
				} else if strEq(kwd, "declare") {
					return irToLLFuncDecl(ctx, expr)
				} else if strEq(kwd, "global") {
					return irToLLGlobal(ctx, expr)
				} else if strEq(kwd, "let") {
					return irToLLInstrLet(ctx, expr)
				} else if strEq(kwd, "ret") {
					return irToLLInstrRet(ctx, expr)
				} else if strEq(kwd, "gep") {
					return irToLLInstrGep(ctx, expr)
				} else if strEq(kwd, "phi") {
					return irToLLInstrPhi(ctx, expr)
				} else if strEq(kwd, "switch") {
					return irToLLInstrSwitch(ctx, expr)
				} else if strEq(kwd, "call") {
					return irToLLInstrCall(ctx, expr)
				} else if strEq(kwd, "load") {
					return irToLLInstrLoad(ctx, expr)
				} else if strEq(kwd, "store") {
					return irToLLInstrStore(ctx, expr)
				} else if strEq(kwd, "op") {
					return irToLLInstrBinOp(ctx, expr)
				} else if strEq(kwd, "icmp") {
					return irToLLInstrCmpI(ctx, expr)
				} else if strEq(kwd, "brIf") {
					return irToLLInstrBrIf(ctx, expr)
				} else if strEq(kwd, "brTo") {
					return irToLLInstrBrTo(ctx, expr)
				} else if strEq(kwd, "as") {
					return irToLLInstrConvert(ctx, expr)
				} else if strEq(kwd, "alloca") {
					return irToLLInstrAlloca(ctx, expr)
				}
			}
			switch ll_sth := irToLL(ctx, callee).(type) {
			case LLType:
				assert(len(expr) == 2)
				return LLExprTyped{ty: ll_sth, expr: irToLL(ctx, expr[1]).(LLExpr)}
			default:
				panic(ll_sth)
			}
		case IrExprIdent:
			if strEq(callee, "/@") {
				assert(len(expr) == 2)
				return irToLL(ctx, expr[1].(IrExprDefRef)).(LLExprIdentGlobal)
			}
			panic("CALLEE-Ident\t" + string(callee))
		default:
			panic(callee)
		}
	case IrExprSlashed:
		kwd := expr[0].(IrExprIdent)
		if kwd[0] >= 'A' && kwd[0] <= 'Z' || (kwd[0] == '_' && len(kwd) == 1) {
			return irToLLType(ctx, expr)
		} else if strEq(kwd, "unreachable") {
			return LLInstrUnreachable{}
		} else if strEq(kwd, "ret") {
			return LLInstrRet{expr: LLExprTyped{ty: LLTypeVoid{}, expr: LLExprLitVoid{}}}
		}
		panic("KWD\t/" + string(kwd))
	case IrExprIdent:
		panic("IDENT\t" + string(expr))
	default:
		panic(expr)
	}
	panic(ir_expr)
}

func irToLLType(ctx *CtxIrToLL, expr IrExprSlashed) LLType {
	kwd := expr[0].(IrExprIdent)
	if len(kwd) == 1 || kwd[0] == 'I' {
		if kwd[0] == 'V' {
			return LLTypeVoid{}
		} else if kwd[0] == '_' {
			return LLTypeAuto{}
		} else if kwd[0] == 'I' {
			assert(len(expr) == 1)
			if len(kwd) == 1 {
				return LLTypeInt{bit_width: ll_target_word_bit_width}
			}
			return LLTypeInt{bit_width: uintFromStr(kwd[1:])}
		} else if kwd[0] == 'A' {
			assert(len(expr) == 3)
			size := expr[1].(IrExprLitInt)
			assert(size > 0)
			return LLTypeArr{size: int(size), ty: irToLL(ctx, IrExprSlashed(expr[2:])).(LLType)}
		} else if kwd[0] == 'P' {
			ret_ty := LLTypePtr{}
			if len(expr) > 1 {
				ret_ty.ty = irToLL(ctx, IrExprSlashed(expr[1:])).(LLType)
			} else {
				ret_ty.ty = LLTypeInt{bit_width: 8}
			}
			return ret_ty
		}
	}
	panic("Ty-KWD\t" + string(kwd))
}

func irToLLGlobal(ctx *CtxIrToLL, expr_form IrExprForm) LLGlobal {
	assert(len(expr_form) == 3)
	lit_obj_opts := expr_form[2].(IrExprObj)
	maybe_constant := irExprsFindKeyedValue(lit_obj_opts, ":", IrExprTag("constant"))
	maybe_external := irExprsFindKeyedValue(lit_obj_opts, ":", IrExprTag("external"))
	assert(!(maybe_constant == nil && maybe_external == nil))
	assert(!(maybe_constant != nil && maybe_external != nil))

	ret_global := LLGlobal{
		external: maybe_external != nil,
		constant: maybe_constant != nil,
		name:     nil, // consumer must set it when receiving ret_global, except for external case below
		ty:       irToLL(ctx, expr_form[1]).(LLType),
	}
	if maybe_external != nil {
		ret_global.name = maybe_external.(IrExprTag)
		assert(len(ret_global.name) != 0)
	} else if maybe_constant != nil {
		ret_global.initializer = irToLL(ctx, maybe_constant).(LLExpr)
		_, is_lit_str := ret_global.initializer.(LLExprLitStr)
		assert(is_lit_str)
	}
	return ret_global
}

func irToLLFuncDecl(ctx *CtxIrToLL, expr_form IrExprForm) LLFunc {
	assert(len(expr_form) == 5)
	_ = expr_form[2].(IrExprObj)
	lit_obj_params := expr_form[4].(IrExprObj)
	ret_func := LLFunc{
		external:     true,
		name:         expr_form[1].(IrExprTag),
		ty:           irToLL(ctx, expr_form[3]).(LLType),
		params:       allocˇLLFuncParam(len(lit_obj_params)),
		basic_blocks: nil,
	}
	irToLLFuncPopulateParams(ctx, lit_obj_params, ret_func.params)
	assert(len(ret_func.name) != 0)
	return ret_func
}

func irToLLFuncDef(ctx *CtxIrToLL, expr_form IrExprForm) LLFunc {
	assert(len(expr_form) == 5)
	_ = expr_form[1].(IrExprObj)
	lit_obj_params := expr_form[3].(IrExprObj)
	lit_obj_blocks := expr_form[4].(IrExprObj)
	ret_func := LLFunc{
		external:     false,
		name:         nil, // consumer must set it when receiving ret_func
		ty:           irToLL(ctx, expr_form[2]).(LLType),
		params:       allocˇLLFuncParam(len(lit_obj_params)),
		basic_blocks: allocˇLLBasicBlock(len(lit_obj_blocks)),
	}
	irToLLFuncPopulateParams(ctx, lit_obj_params, ret_func.params)
	for i := range lit_obj_blocks {
		pair := lit_obj_blocks[i].(IrExprInfix)
		assert(strEq(pair.kind, ":"))
		block_name := pair.lhs.(IrExprTag)
		if len(block_name) == 0 {
			counter++
			block_name = uintToStr(counter, 10, 1, Str("b."))
		}
		block_instrs := pair.rhs.(IrExprArr)
		ret_func.basic_blocks[i] = LLBasicBlock{
			name:   block_name,
			instrs: allocˇLLInstr(len(block_instrs)),
		}
		for j, ir_expr_instr := range block_instrs {
			ret_func.basic_blocks[i].instrs[j] = irToLL(ctx, ir_expr_instr).(LLInstr)
		}
	}
	return ret_func
}

func irToLLInstrRet(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrRet {
	assert(len(expr_form) == 2)
	return LLInstrRet{expr: llExprToTyped(irToLL(ctx, expr_form[1]).(LLExpr), LLTypeAuto{})}
}

func irToLLInstrBrTo(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrBrTo {
	assert(len(expr_form) == 2)
	ret_br := LLInstrBrTo{
		block_name: irToLL(ctx, expr_form[1]).(LLExprIdentLocal),
	}
	return ret_br
}

func irToLLInstrBrIf(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrBrIf {
	assert(len(expr_form) == 4)
	ret_br := LLInstrBrIf{
		cond:                irToLL(ctx, expr_form[1]).(LLExpr),
		block_name_if_true:  irToLL(ctx, expr_form[2]).(LLExprIdentLocal),
		block_name_if_false: irToLL(ctx, expr_form[3]).(LLExprIdentLocal),
	}
	return ret_br
}

func irToLLInstrAlloca(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrAlloca {
	assert(len(expr_form) == 3)
	ret_alloca := LLInstrAlloca{
		ty:        irToLL(ctx, expr_form[1]).(LLType),
		num_elems: irToLL(ctx, expr_form[2]).(LLExprTyped),
	}
	return ret_alloca
}

func irToLLInstrConvert(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrConvert {
	assert(len(expr_form) == 3)
	ret_conv := LLInstrConvert{
		convert_kind: 0,
		ty:           irToLL(ctx, expr_form[1]).(LLType),
		expr:         irToLL(ctx, expr_form[2]).(LLExprTyped),
	}
	ty_src := ret_conv.expr.ty
	switch ty_dst := ret_conv.ty.(type) {
	case LLTypePtr:
		ret_conv.convert_kind = ll_convert_int_to_ptr
	case LLTypeInt:
		if ty_src_int, is_ty_src_int := ty_src.(LLTypeInt); is_ty_src_int {
			if ty_src_int.bit_width > ty_dst.bit_width {
				ret_conv.convert_kind = ll_convert_trunc
			}
		} else if _, is_ty_src_ptr := ty_src.(LLTypePtr); is_ty_src_ptr {
			ret_conv.convert_kind = ll_convert_ptr_to_int
		}
	}
	assert(ret_conv.convert_kind != 0)
	return ret_conv
}

func irToLLInstrBinOp(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrBinOp {
	assert(len(expr_form) == 5)
	ret_op2 := LLInstrBinOp{
		op_kind: 0,
		ty:      irToLL(ctx, expr_form[2]).(LLType),
		lhs:     irToLL(ctx, expr_form[3]).(LLExpr),
		rhs:     irToLL(ctx, expr_form[4]).(LLExpr),
	}
	kind_tag := expr_form[1].(IrExprTag)
	if strEq(kind_tag, "add") {
		ret_op2.op_kind = ll_bin_op_add
	} else if strEq(kind_tag, "mul") {
		ret_op2.op_kind = ll_bin_op_mul
	} else if strEq(kind_tag, "sub") {
		ret_op2.op_kind = ll_bin_op_sub
	} else if strEq(kind_tag, "udiv") {
		ret_op2.op_kind = ll_bin_op_udiv
	}
	assert(ret_op2.op_kind != 0)
	return ret_op2
}

func irToLLInstrCmpI(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrCmpI {
	assert(len(expr_form) == 5)
	ret_cmp := LLInstrCmpI{
		cmp_kind: 0,
		ty:       irToLL(ctx, expr_form[2]).(LLType),
		lhs:      irToLL(ctx, expr_form[3]).(LLExpr),
		rhs:      irToLL(ctx, expr_form[4]).(LLExpr),
	}
	kind_tag := expr_form[1].(IrExprTag)
	if strEq(kind_tag, "eq") {
		ret_cmp.cmp_kind = ll_cmp_i_eq
	} else if strEq(kind_tag, "ne") {
		ret_cmp.cmp_kind = ll_cmp_i_ne
	} else if strEq(kind_tag, "ugt") {
		ret_cmp.cmp_kind = ll_cmp_i_ugt
	} else if strEq(kind_tag, "uge") {
		ret_cmp.cmp_kind = ll_cmp_i_uge
	} else if strEq(kind_tag, "ult") {
		ret_cmp.cmp_kind = ll_cmp_i_ult
	} else if strEq(kind_tag, "ule") {
		ret_cmp.cmp_kind = ll_cmp_i_ule
	} else if strEq(kind_tag, "sgt") {
		ret_cmp.cmp_kind = ll_cmp_i_sgt
	} else if strEq(kind_tag, "sge") {
		ret_cmp.cmp_kind = ll_cmp_i_sge
	} else if strEq(kind_tag, "slt") {
		ret_cmp.cmp_kind = ll_cmp_i_slt
	} else if strEq(kind_tag, "sle") {
		ret_cmp.cmp_kind = ll_cmp_i_sle
	}
	assert(ret_cmp.cmp_kind != 0)
	return ret_cmp
}

func irToLLInstrStore(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrStore {
	assert(len(expr_form) == 3)
	expr := irToLL(ctx, expr_form[2]).(LLExprTyped)
	ret_store := LLInstrStore{
		expr: expr,
		dst:  llExprToTyped(irToLL(ctx, expr_form[1]).(LLExpr), LLTypePtr{ty: expr.ty}),
	}
	return ret_store
}

func irToLLInstrLoad(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrLoad {
	assert(len(expr_form) == 3)
	ty := irToLL(ctx, expr_form[1]).(LLType)
	ret_load := LLInstrLoad{
		ty:   ty,
		expr: llExprToTyped(irToLL(ctx, expr_form[2]).(LLExpr), LLTypePtr{ty: ty}),
	}
	return ret_load
}

func irToLLInstrCall(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrCall {
	assert(len(expr_form) == 3)
	lit_arr_args := expr_form[2].(IrExprArr)
	ret_call := LLInstrCall{
		callee: irToLL(ctx, expr_form[1]).(LLExprIdentGlobal),
		ty:     LLTypeAuto{},
		args:   allocˇLLExprTyped(len(lit_arr_args)),
	}
	for i := range lit_arr_args {
		ret_call.args[i] = llExprToTyped(irToLL(ctx, lit_arr_args[i]).(LLExpr), LLTypeAuto{})
	}
	return ret_call
}

func irToLLInstrSwitch(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrSwitch {
	assert(len(expr_form) == 4)
	lit_obj_cases := expr_form[3].(IrExprObj)
	ret_switch := LLInstrSwitch{
		comparee:           irToLL(ctx, expr_form[1]).(LLExprTyped),
		default_block_name: expr_form[2].(IrExprTag),
		cases:              allocˇLLSwitchCase(len(lit_obj_cases)),
	}
	for i := range lit_obj_cases {
		pair := lit_obj_cases[i].(IrExprInfix)
		assert(strEq(pair.kind, ":"))
		ret_switch.cases[i] = LLSwitchCase{
			block_name: pair.rhs.(IrExprTag),
			expr:       irToLL(ctx, pair.lhs).(LLExprTyped),
		}
	}
	return ret_switch
}

func irToLLInstrPhi(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrPhi {
	assert(len(expr_form) == 3)
	lit_obj_preds := expr_form[2].(IrExprObj)
	ret_phi := LLInstrPhi{
		ty:           irToLL(ctx, expr_form[1]).(LLType),
		predecessors: allocˇLLPhiPred(len(lit_obj_preds)),
	}
	for i := range lit_obj_preds {
		pair := lit_obj_preds[i].(IrExprInfix)
		assert(strEq(pair.kind, ":"))
		ret_phi.predecessors[i] = LLPhiPred{
			block_name: pair.lhs.(IrExprTag),
			expr:       irToLL(ctx, pair.rhs).(LLExpr),
		}
	}
	return ret_phi
}

func irToLLInstrGep(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrGep {
	assert(len(expr_form) == 4)
	lit_arr_idxs := expr_form[3].(IrExprArr)
	ret_gep := LLInstrGep{
		ty:       irToLL(ctx, expr_form[1]).(LLType),
		base_ptr: irToLL(ctx, expr_form[2]).(LLExprTyped),
		indices:  allocˇLLExprTyped(len(lit_arr_idxs)),
	}
	for i := range lit_arr_idxs {
		ret_gep.indices[i] = irToLL(ctx, lit_arr_idxs[i]).(LLExprTyped)
	}
	return ret_gep
}

func irToLLInstrLet(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrLet {
	assert(len(expr_form) >= 4)
	assert(strEq(expr_form[2].(IrExprIdent), "="))

	var instr LLInstr
	ll_sth := irToLL(ctx, IrExprForm(expr_form[3:]))
	if instr, _ = ll_sth.(LLInstr); instr == nil {
		expr_ty := ll_sth.(LLExprTyped)
		instr = LLInstrBinOp{
			ty:      expr_ty.ty,
			lhs:     expr_ty.expr,
			rhs:     LLExprLitInt(0),
			op_kind: ll_bin_op_add,
		}
	}
	return LLInstrLet{
		name:  expr_form[1].(IrExprTag),
		instr: instr,
	}
}

func irToLLFuncPopulateParams(ctx *CtxIrToLL, lit IrExprObj, dst []LLFuncParam) {
	for i := range lit {
		pair := lit[i].(IrExprInfix)
		assert(strEq(pair.kind, ":"))
		dst[i].name = pair.lhs.(IrExprTag)
		assert(len(dst[i].name) != 0)
		dst[i].ty = irToLL(ctx, pair.rhs).(LLType)
	}
}
