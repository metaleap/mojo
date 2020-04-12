package main

import (
	"os"
	"unsafe"
)

type LLCtxEval struct {
	ll_mod *LLModule
	mem    struct {
		bytes         []byte
		globals_addrs []int
		stack_addr    int
	}
	cur_frame struct {
		fn              *LLFunc
		locals          []LLExpr
		pred_block_name Str
	}
}

func llEvalCtx(ll_mod *LLModule, stack_size int) LLCtxEval {
	ret_ctx := LLCtxEval{ll_mod: ll_mod}
	ret_ctx.mem.globals_addrs = allocˇint(len(ll_mod.anns.global_names))
	for i := range ret_ctx.mem.globals_addrs {
		ret_ctx.mem.globals_addrs[i] = -1
	}
	prep := allocˇStr(len(ll_mod.anns.global_names))
	global_mem_size := 0
	for i := range ll_mod.globals {
		the_global := &ll_mod.globals[i]
		if !the_global.external {
			switch ty := the_global.ty.(type) {
			case LLTypeArr:
				ty_int := ty.ty.(LLTypeInt)
				assert(ty_int.bit_width == 8)
				prep[i] = the_global.initializer.(LLExprLitStr)
				global_mem_size += len(prep[i])
			default:
				panic(ty)
			}
		}
	}
	ret_ctx.mem.bytes = allocˇbyte(global_mem_size + stack_size)
	ret_ctx.mem.stack_addr = 0
	for i, bytes := range prep {
		ret_ctx.mem.globals_addrs[i] = ret_ctx.mem.stack_addr
		for i_mem := range bytes {
			ret_ctx.mem.bytes[ret_ctx.mem.stack_addr+i_mem] = bytes[i_mem]
		}
		ret_ctx.mem.stack_addr += len(bytes)
	}
	return ret_ctx
}

func llEvalRunMain(ctx *LLCtxEval) {
	var main_func *LLFunc
	for i := range ctx.ll_mod.funcs {
		this_func := &ctx.ll_mod.funcs[i]
		if !this_func.external {
			if strEq(this_func.name, "main") {
				main_func = this_func
				break
			}
		}
	}
	if main_func == nil {
		panic("no main func")
	}

	switch exit_code := llEvalCall(ctx, main_func, nil).(type) {
	case LLExprLitInt:
		os.Exit(int(exit_code))
	default:
		os.Exit(0)
	}
}

func llEvalCall(ctx *LLCtxEval, fn *LLFunc, args []LLExpr) Any {
	ctx_old := *ctx
	ctx.cur_frame.fn = fn
	ctx.cur_frame.locals = allocˇLLExpr(len(fn.anns.local_temporaries_names))
	ctx.cur_frame.pred_block_name = nil

	for idx_block, idx_instr := 0, 0; true; idx_instr++ {
		eval_result, is_ret, is_jump := llEvalInstr(ctx, fn.basic_blocks[idx_block].instrs[idx_instr])
		if is_ret {
			*ctx = ctx_old
			return eval_result
		} else if is_jump {
			ctx.cur_frame.pred_block_name = fn.basic_blocks[idx_block].name
			idx_instr = -1
			idx_block = eval_result.(int)
		}
	}
	panic("unreachable")
}

func llEvalInstr(ctx *LLCtxEval, ll_instr LLInstr) (eval_result Any, is_ret bool, is_jump bool) {
	switch instr := ll_instr.(type) {

	case LLInstrBrTo:
		for i := range ctx.cur_frame.fn.basic_blocks {
			if strEql(instr.block_name, ctx.cur_frame.fn.basic_blocks[i].name) {
				is_jump = true
				eval_result = i
				break
			}
		}
		assert(eval_result != nil)

	case LLInstrBrIf:
		expr_bool := llEvalExpr(ctx, instr.cond).(LLExprLitInt)
		var block_name Str
		if expr_bool == 0 {
			block_name = instr.block_name_if_true
		} else if expr_bool == 1 {
			block_name = instr.block_name_if_false
		}
		for i := range ctx.cur_frame.fn.basic_blocks {
			if strEql(block_name, ctx.cur_frame.fn.basic_blocks[i].name) {
				is_jump = true
				eval_result = i
				break
			}
		}
		assert(eval_result != nil)

	case LLInstrRet:
		eval_result = llEvalExpr(ctx, instr.expr)
		is_ret = true

	case LLInstrLet:
		for i, name := range ctx.cur_frame.fn.anns.local_temporaries_names {
			if strEql(name, instr.name) {
				instr_result, _, _ := llEvalInstr(ctx, instr.instr)
				if expr, _ := instr_result.(LLExpr); expr != nil {
					ctx.cur_frame.locals[i] = expr
				} else {
					panic(instr.instr)
				}
				break
			}
		}

	case LLInstrSwitch:
		comparee := llEvalExpr(ctx, instr.comparee).(LLExprLitInt)
		block_name := instr.default_block_name
		for i := range instr.cases {
			this_case := &instr.cases[i]
			case_value := llEvalExpr(ctx, this_case.expr).(LLExprLitInt)
			if case_value == comparee {
				block_name = this_case.block_name
				break
			}
		}
		for i := range ctx.cur_frame.fn.basic_blocks {
			if strEql(block_name, ctx.cur_frame.fn.basic_blocks[i].name) {
				is_jump = true
				eval_result = i
				break
			}
		}
		assert(eval_result != nil)

	case LLInstrBinOp:
		lhs := llEvalExpr(ctx, instr.lhs).(LLExprLitInt)
		rhs := llEvalExpr(ctx, instr.rhs).(LLExprLitInt)
		switch instr.op_kind {
		case ll_bin_op_add:
			eval_result = lhs + rhs
		case ll_bin_op_mul:
			eval_result = lhs * rhs
		case ll_bin_op_sub:
			eval_result = lhs - rhs
		case ll_bin_op_udiv:
			eval_result = uint64(lhs) / uint64(rhs)
		default:
			panic(instr.op_kind)
		}

	case LLInstrCmpI:
		expr_bool := LLExprLitInt(0)
		lhs := llEvalExpr(ctx, instr.lhs).(LLExprLitInt)
		rhs := llEvalExpr(ctx, instr.rhs).(LLExprLitInt)
		switch instr.cmp_kind {
		case ll_cmp_i_eq:
			if lhs == rhs {
				expr_bool = 1
			}
		case ll_cmp_i_ne:
			if lhs != rhs {
				expr_bool = 1
			}
		case ll_cmp_i_ugt:
			if uint64(lhs) > uint64(rhs) {
				expr_bool = 1
			}
		case ll_cmp_i_uge:
			if uint64(lhs) >= uint64(rhs) {
				expr_bool = 1
			}
		case ll_cmp_i_ult:
			if uint64(lhs) < uint64(rhs) {
				expr_bool = 1
			}
		case ll_cmp_i_ule:
			if uint64(lhs) <= uint64(rhs) {
				expr_bool = 1
			}
		case ll_cmp_i_sgt:
			if lhs > rhs {
				expr_bool = 1
			}
		case ll_cmp_i_sge:
			if lhs >= rhs {
				expr_bool = 1
			}
		case ll_cmp_i_slt:
			if lhs < rhs {
				expr_bool = 1
			}
		case ll_cmp_i_sle:
			if lhs <= rhs {
				expr_bool = 1
			}
		default:
			panic(instr.cmp_kind)
		}
		eval_result = expr_bool

	case LLInstrPhi:
		for i := range instr.predecessors {
			if strEql(ctx.cur_frame.pred_block_name, instr.predecessors[i].block_name) {
				eval_result = llEvalExpr(ctx, instr.predecessors[i].expr)
				break
			}
		}
		assert(eval_result != nil)

	case LLInstrCall:
		callee := llEvalExpr(ctx, instr.callee).(*LLFunc)
		args := allocˇLLExpr(len(instr.args))
		for i, arg := range instr.args {
			args[i] = llEvalExpr(ctx, arg).(LLExpr)
		}
		eval_result = llEvalCall(ctx, callee, args)

	case LLInstrGep:
		switch base_ptr := llEvalExpr(ctx, instr.base_ptr).(type) {
		case *LLGlobal:
			addr := ctx.mem.globals_addrs[base_ptr.anns.idx]
			eval_result = uintptr(unsafe.Pointer(&ctx.mem.bytes[addr]))
			panic(instr.ty)
		default:
			panic(base_ptr)
		}

	case LLInstrConvert:

	case LLInstrStore:

	case LLInstrLoad:

	case LLInstrAlloca:

	case LLInstrUnreachable:
		unreachable()
	case LLInstrComment:
	default:
		panic(instr)
	}
	return
}

func llEvalExpr(ctx *LLCtxEval, ll_expr LLExpr) Any {
	switch expr := ll_expr.(type) {
	case LLExprLitInt:
		return expr
	case LLExprLitVoid:
		return expr
	case LLExprIdentLocal:
		for i, name := range ctx.cur_frame.fn.anns.local_temporaries_names {
			if strEql(name, expr) {
				return ctx.cur_frame.locals[i]
			}
		}
		panic(expr)
	case LLExprTyped:
		return llEvalExpr(ctx, expr.expr)
	case LLExprIdentGlobal:
		for i := range ctx.ll_mod.funcs {
			if strEql(ctx.ll_mod.funcs[i].name, expr) {
				return &ctx.ll_mod.funcs[i]
			}
		}
		for i := range ctx.ll_mod.globals {
			if strEql(ctx.ll_mod.globals[i].name, expr) {
				return &ctx.ll_mod.globals[i]
			}
		}
		panic(expr)
	default:
		panic(expr)
	}
}
