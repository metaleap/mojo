package main

import (
	"os"
)

type LLCtxRun struct {
	ll_mod *LLModule
	fn     *LLFunc
	locals map[string]Any
}

func llRun(ll_mod *LLModule) {
	var main_func *LLFunc
	for i := range ll_mod.funcs {
		this_func := &ll_mod.funcs[i]
		if !this_func.external {
			if strEql(this_func.name, Str("main")) {
				main_func = this_func
				break
			}
		}
	}
	if main_func == nil {
		panic("no main func")
	}

	ctx := LLCtxRun{}
	switch exit_code := llCall(&ctx, main_func, nil).(type) {
	case LLExprLitInt:
		os.Exit(int(exit_code))
	default:
		os.Exit(0)
	}
}

func llCall(ctx *LLCtxRun, fn *LLFunc, args []LLExpr) Any {
	old_fn := ctx.fn
	ctx.fn = fn
	old_locals := ctx.locals
	ctx.locals = make(map[string]Any, 8)

	for idx_block, idx_instr := 0, 0; true; idx_instr++ {
		eval_result, is_ret, is_jump := llEvalInstr(ctx, fn.basic_blocks[idx_block].instrs[idx_instr])
		if is_ret {
			ctx.fn = old_fn
			ctx.locals = old_locals
			return eval_result
		} else if is_jump {
			idx_instr = -1
			idx_block = eval_result.(int)
		}
	}
	panic("unreachable")
}

func llEvalInstr(ctx *LLCtxRun, ll_instr LLInstr) (eval_result Any, is_ret bool, is_jump bool) {
	switch instr := ll_instr.(type) {

	case LLInstrBrTo:
		for i := range ctx.fn.basic_blocks {
			if strEql(instr.block_name, ctx.fn.basic_blocks[i].name) {
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
		for i := range ctx.fn.basic_blocks {
			if strEql(block_name, ctx.fn.basic_blocks[i].name) {
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
		ctx.locals[string(instr.name)], _, _ = llEvalInstr(ctx, instr.instr)

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
		for i := range ctx.fn.basic_blocks {
			if strEql(block_name, ctx.fn.basic_blocks[i].name) {
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

	case LLInstrCall:

	case LLInstrConvert:

	case LLInstrGep:

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

func llEvalExpr(ctx *LLCtxRun, ll_expr LLExpr) Any {
	switch expr := ll_expr.(type) {
	case LLExprLitInt:
		return expr
	case LLExprLitVoid:
		return expr
	case LLExprIdentLocal:
		return ctx.locals[string(expr)]
	case LLExprTyped:
		return llEvalExpr(ctx, expr.expr)
	default:
		panic(expr)
	}
}
