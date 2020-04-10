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

	case LLInstrRet:
		eval_result = llEvalExpr(ctx, instr.expr)
		is_ret = true

	case LLInstrLet:
		ctx.locals[string(instr.name)], _, _ = llEvalInstr(ctx, instr.instr)

	case LLInstrSwitch:
		comparee := llEvalExpr(ctx, instr.comparee.expr).(LLExprLitInt)
		block_name := instr.default_block_name
		for i := range instr.cases {
			this_case := &instr.cases[i]
			case_value := llEvalExpr(ctx, this_case.expr.expr).(LLExprLitInt)
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

	case LLInstrPhi:

	case LLInstrBinOp:

	case LLInstrCmpI:

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
	return nil
}
