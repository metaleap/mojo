package main

type CtxIrLLEval struct {
	prog *IrLLProg
	args []IrLLExpr
}

// TODO: wastefully recursive right now, later pre-alloc stack of manually pushed/popped call frames
func irLLEval(ctx *CtxIrLLEval, expr IrLLExpr) IrLLExpr {
	var ret_expr IrLLExpr
	switch it := expr.kind.(type) {
	case IrLLExprInt, IrLLExprPtr:
		ret_expr = expr
	case IrLLExprArgRef:
		ret_expr = ctx.args[it]
	case IrLLExprCall:
		if it.is_ext {
			panic("TODO: fn-ext call")
		} else if it.is_ptr {
			panic("TODO: fn-ptr call")
		}
		cur_args := ªIrLLExpr(len(it.args))
		for i := range it.args {
			cur_args[i] = irLLEval(ctx, it.args[i])
		}
		old_args := ctx.args
		ctx.args = cur_args
		ret_expr = irLLEval(ctx, ctx.prog.defs[it.callee].body)
		ctx.args = old_args
	case IrLLExprSelect:
		boolish_int := irLLEval(ctx, it.cond)
		if boolish_int.kind.(IrLLExprInt) == 1 {
			ret_expr = irLLEval(ctx, it.if_true)
		} else {
			ret_expr = irLLEval(ctx, it.if_false)
		}
	case IrLLExprOpInt:
		lhs := irLLEval(ctx, it.lhs)
		rhs := irLLEval(ctx, it.rhs)
		// assert(lhs.anns.ty == rhs.anns.ty)
		l, r := lhs.kind.(IrLLExprInt), rhs.kind.(IrLLExprInt)
		var result IrLLExprInt
		switch it.kind {
		case ll_bin_op_add:
			result = l + r
		case ll_bin_op_mul:
			result = l * r
		case ll_bin_op_sub:
			result = l - r
		case ll_bin_op_udiv:
			result = IrLLExprInt(uint64(l) / uint64(r))
		case ll_bin_op_sdiv:
			result = l / r
		case ll_bin_op_urem:
			result = IrLLExprInt(uint64(l) % uint64(r))
		case ll_bin_op_srem:
			result = l % r
		default:
			panic(it.kind)
		}
		ret_expr.anns.ty = lhs.anns.ty
		ret_expr.kind = result
	case IrLLExprCmpInt:
		result := false
		lhs := irLLEval(ctx, it.lhs).kind.(IrLLExprInt)
		rhs := irLLEval(ctx, it.rhs).kind.(IrLLExprInt)
		switch it.kind {
		case ll_cmp_i_eq:
			result = (lhs == rhs)
		case ll_cmp_i_ne:
			result = (lhs != rhs)
		case ll_cmp_i_ugt:
			result = (uint64(lhs) > uint64(rhs))
		case ll_cmp_i_uge:
			result = (uint64(lhs) >= uint64(rhs))
		case ll_cmp_i_ult:
			result = (uint64(lhs) < uint64(rhs))
		case ll_cmp_i_ule:
			result = (uint64(lhs) <= uint64(rhs))
		case ll_cmp_i_sgt:
			result = (lhs > rhs)
		case ll_cmp_i_sge:
			result = (lhs >= rhs)
		case ll_cmp_i_slt:
			result = (lhs < rhs)
		case ll_cmp_i_sle:
			result = (lhs <= rhs)
		default:
			panic(it.kind)
		}
		ret_expr.anns.ty = IrLLTypeInt{bit_width: 1}
		ret_expr.kind = IrLLExprInt(0)
		if result {
			ret_expr.kind = IrLLExprInt(1)
		}
	default:
		panic(it)
	}
	// assert(ret_expr.kind != nil)
	return ret_expr
}
