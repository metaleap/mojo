package main

type CtxIrLLEval struct {
	prog *IrLLProg
	args struct {
		stash [8]IrLLExpr // fixed-size for this prototype
		idx   int
	}
}

// TODO: wastefully recursive right now, later pre-alloc a stack of manually pushed/popped call frames
func irLLEval(ctx *CtxIrLLEval, expr *IrLLExpr) IrLLExpr {
	var ret_expr IrLLExpr
	switch it := expr.variant.(type) {
	case IrLLExprInt, IrLLExprPtr:
		ret_expr = *expr
	case IrLLExprArgRef:
		ret_expr = ctx.args.stash[ctx.args.idx+int(it)]
	case IrLLExprCall:
		old_idx := ctx.args.idx
		new_idx := old_idx + len(it.args)
		for i := range it.args {
			ctx.args.stash[new_idx+i] = irLLEval(ctx, &it.args[i])
		}
		if it.is_ext {
			panic("TODO: fn-ext call")
		} else if it.is_ptr {
			panic("TODO: fn-ptr call")
		}
		ctx.args.idx = new_idx
		ret_expr = irLLEval(ctx, &ctx.prog.defs[it.callee].body)
		ctx.args.idx = old_idx
	case IrLLExprSelect:
		boolish_int := irLLEval(ctx, &it.cond)
		if boolish_int.variant.(IrLLExprInt) == 1 {
			ret_expr = irLLEval(ctx, &it.if_true)
		} else {
			ret_expr = irLLEval(ctx, &it.if_false)
		}
	case IrLLExprOpInt:
		lhs := irLLEval(ctx, &it.lhs)
		rhs := irLLEval(ctx, &it.rhs)
		// assert(lhs.anns.ty == rhs.anns.ty)
		l, r := lhs.variant.(IrLLExprInt), rhs.variant.(IrLLExprInt)
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
		ret_expr.variant = result
	case IrLLExprCmpInt:
		result := false
		lhs := irLLEval(ctx, &it.lhs).variant.(IrLLExprInt)
		rhs := irLLEval(ctx, &it.rhs).variant.(IrLLExprInt)
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
		ret_expr.variant = IrLLExprInt(0)
		if result {
			ret_expr.variant = IrLLExprInt(1)
		}
	default:
		panic(it)
	}
	// assert(ret_expr.variant != nil)
	return ret_expr
}
