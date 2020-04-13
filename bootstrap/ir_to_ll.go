package main

const llmodule_default_target_datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
const llmodule_default_target_triple = "x86_64-unknown-linux-gnu"

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
	case IrExprDefRef:
		ref_def := &ctx.ir.defs[expr]
		if ident := llTopLevelNameFrom(ctx.ll_mod, ref_def, ctx.num_globals, ctx.num_funcs); ident != nil {
			return ident
		}
		switch ll_top_level_item := irToLL(ctx, ref_def.body).(type) {
		case LLGlobal:
			if 0 == len(ll_top_level_item.name) {
				ll_top_level_item.name = ref_def.name
			}
			ctx.ll_mod.globals[ctx.num_globals] = ll_top_level_item
			ctx.num_globals++
			return LLExprIdentGlobal(ll_top_level_item.name)
		case LLFunc:
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
			if strEq(kwd, "define") {
				assert(len(callee) == 1)
				return irToLLFuncDef(ctx, expr)
			} else if strEq(kwd, "declare") {
				assert(len(callee) == 1)
				return irToLLFuncDecl(ctx, expr)
			} else if strEq(kwd, "global") {
				assert(len(callee) == 1)
				return irToLLGlobal(ctx, expr)
			} else if strEq(kwd, "let") {
				assert(len(callee) == 1)
				return irToLLInstrLet(ctx, expr)
			} else if strEq(kwd, "ret") {
				assert(len(callee) == 1)
				return irToLLInstrRet(ctx, expr)
			} else if strEq(kwd, "gep") {
				assert(len(callee) == 1)
				return irToLLInstrGep(ctx, expr)
			} else {
				switch ll_sth := irToLL(ctx, callee).(type) {
				case LLType:
					assert(len(expr) == 2)
					return LLExprTyped{ty: ll_sth, expr: irToLL(ctx, expr[1]).(LLExpr)}
				default:
					panic(ll_sth)
				}
			}
		default:
			panic(callee)
		}
	case IrExprSlashed:
		kwd := expr[0].(IrExprIdent)
		if kwd[0] >= 'A' && kwd[0] <= 'Z' {
			return irToLLType(ctx, expr)
		}
		panic("KWD\t" + string(kwd))
	case IrExprIdent:
		panic("IDENT\t" + string(expr))
	default:
		panic(expr)
	}
	return nil
}

func irToLLType(ctx *CtxIrToLL, expr IrExprSlashed) LLType {
	kwd := expr[0].(IrExprIdent)
	if kwd[0] == 'I' {
		assert(len(expr) == 1)
		return LLTypeInt{bit_width: uintFromStr(kwd[1:])}
	} else if kwd[0] == 'A' {
		assert(len(expr) == 3)
		size := expr[1].(IrExprLitInt)
		assert(size > 0)
		return LLTypeArr{size: uint64(size), ty: irToLL(ctx, IrExprSlashed(expr[2:])).(LLType)}
	} else if kwd[0] == 'P' {
		ret_ty := LLTypePtr{}
		if len(expr) > 1 {
			ret_ty.ty = irToLL(ctx, IrExprSlashed(expr[1:])).(LLType)
		}
		return ret_ty
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

func irToLLInstrRet(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrRet {
	assert(len(expr_form) == 2)
	return LLInstrRet{expr: irToLL(ctx, expr_form[1]).(LLExprTyped)}
}

func irToLLInstrLet(ctx *CtxIrToLL, expr_form IrExprForm) LLInstrLet {
	assert(len(expr_form) == 2)
	lit_obj_pair := expr_form[1].(IrExprObj)
	assert(len(lit_obj_pair) == 1)
	pair := lit_obj_pair[0].(IrExprInfix)
	assert(strEq(pair.kind, ":"))

	var instr LLInstr
	ll_sth := irToLL(ctx, pair.rhs)
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
		name:  pair.lhs.(IrExprTag),
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
