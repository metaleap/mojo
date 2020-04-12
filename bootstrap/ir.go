package main

type Ir struct {
	origin_ast *Ast
	defs       []IrDef
}

type IrDef struct {
	origin_def *AstDef
	name       Str
	num_params int
	body       IrExpr
}

type IrExpr interface{ implementsIrExpr() }

type IrExprLitInt int64

type IrExprLitStr Str

type IrExprDefRef int

type IrExprArgRef int

type IrExprIdent Str

type IrExprIdentTagged Str

type IrExprSlashed []IrExpr

type IrExprForm []IrExpr

type IrExprArr []IrExpr

type IrExprObj []IrExpr

type IrExprInfix struct {
	kind Str
	lhs  IrExpr
	rhs  IrExpr
}

func (IrExprLitInt) implementsIrExpr()      {}
func (IrExprLitStr) implementsIrExpr()      {}
func (IrExprDefRef) implementsIrExpr()      {}
func (IrExprArgRef) implementsIrExpr()      {}
func (IrExprIdent) implementsIrExpr()       {}
func (IrExprIdentTagged) implementsIrExpr() {}
func (IrExprSlashed) implementsIrExpr()     {}
func (IrExprForm) implementsIrExpr()        {}
func (IrExprArr) implementsIrExpr()         {}
func (IrExprObj) implementsIrExpr()         {}
func (IrExprInfix) implementsIrExpr()       {}

type CtxAstToIr struct {
	scope    *AstScopes
	dst      *Ir
	num_defs int
}

func irFromAst(ast *Ast, expr *AstExpr) Ir {
	ret_ir := Ir{origin_ast: ast, defs: allocˇIrDef(len(ast.defs))}
	ctx := CtxAstToIr{scope: &ast.scope, dst: &ret_ir, num_defs: 0}
	_ = irExprFrom(&ctx, expr).(IrExprDefRef)
	ret_ir.defs = ret_ir.defs[0:ctx.num_defs]
	return ret_ir
}

func irExprFrom(ctx *CtxAstToIr, ast_expr *AstExpr) IrExpr {
	switch expr := ast_expr.kind.(type) {

	case AstExprLitInt:
		return IrExprLitInt(expr)

	case AstExprLitStr:
		return IrExprLitStr(expr)

	case AstExprLitClip:
		ret_arr := allocˇIrExpr(len(expr))
		for i := range expr {
			ret_arr[i] = irExprFrom(ctx, &expr[i])
		}
		return IrExprArr(ret_arr)

	case AstExprLitCurl:
		ret_obj := allocˇIrExpr(len(expr))
		for i := range expr {
			ret_obj[i] = irExprFrom(ctx, &expr[i])
		}
		return IrExprObj(ret_obj)

	case AstExprIdent:
		if resolved := astScopesResolve(ctx.scope, expr, -1); resolved == nil {
			return IrExprIdent(expr)
		} else if resolved.param_idx >= 0 {
			return IrExprArgRef(resolved.param_idx)
		} else {
			assert(resolved.top_def == resolved.ref_def)
			ir_def_idx := -1
			for i := range ctx.dst.defs[0:ctx.num_defs] {
				if ctx.dst.defs[i].origin_def == resolved.ref_def {
					ir_def_idx = i
					break
				}
			}
			if ir_def_idx < 0 {
				ir_def := IrDef{
					origin_def: resolved.ref_def,
					name:       resolved.ref_def.anns.name,
					num_params: 0,
				}
				if head_form, _ := resolved.ref_def.head.kind.(AstExprForm); head_form != nil {
					ir_def.num_params = len(head_form) - 1
				}
				ir_def_idx = ctx.num_defs
				ctx.num_defs++
				ctx.dst.defs[ir_def_idx] = ir_def
				ctx.dst.defs[ir_def_idx].body = irExprFrom(ctx, &resolved.ref_def.body)
			}
			return IrExprDefRef(ir_def_idx)
		}

	case AstExprForm:
		for _, supported_infix := range []string{":"} {
			if lhs, rhs := astExprFormSplit(ast_expr, supported_infix, false, false, false, ctx.dst.origin_ast); lhs != nil && rhs != nil {
				return IrExprInfix{
					kind: Str(supported_infix),
					lhs:  irExprFrom(ctx, lhs),
					rhs:  irExprFrom(ctx, rhs),
				}
			}
		}
		// TODO: tweak astExprTaggedIdent and astExprSlashed to accept AstExprForm instead of *AstExpr
		if tagged_ident := astExprTaggedIdent(ast_expr); tagged_ident != nil {
			return IrExprIdentTagged(tagged_ident)
		} else if slashed := astExprSlashed(ast_expr); slashed != nil {
			ret_slashed := IrExprSlashed(allocˇIrExpr(len(slashed)))
			for i, sub_expr := range slashed {
				ret_slashed[i] = irExprFrom(ctx, sub_expr)
			}
			return ret_slashed
		}

		ir_form := allocˇIrExpr(len(expr))
		for i := range expr {
			ir_form[i] = irExprFrom(ctx, &expr[i])
		}
		return IrExprForm(ir_form)

	default:
		panic(expr)
	}
}

func irReduceDefs(ir *Ir) {
	ctx := CtxReduce{ir: ir, done: allocˇbool(len(ir.defs))}
	for i := range ir.defs {
		irReduceDef(&ctx, i)
	}
}

func irReduceDef(ctx *CtxReduce, def_idx int) {
	if ctx.done[def_idx] {
		return
	}
	ctx.done[def_idx] = true
	old_args := ctx.args
	ctx.args = nil
	def := &ctx.ir.defs[def_idx]
	def.body = irReduceExpr(ctx, def.body)
	ctx.args = old_args
}

func irReduceExpr(ctx *CtxReduce, ir_expr IrExpr) IrExpr {
	switch expr := ir_expr.(type) {
	case IrExprSlashed:
		ret_slashed := allocˇIrExpr(len(expr))
		for i := range expr {
			ret_slashed[i] = irReduceExpr(ctx, expr[i])
		}
		return IrExprSlashed(ret_slashed)
	case IrExprArr:
		ret_arr := allocˇIrExpr(len(expr))
		for i := range expr {
			ret_arr[i] = irReduceExpr(ctx, expr[i])
		}
		return IrExprObj(ret_arr)
	case IrExprObj:
		ret_obj := allocˇIrExpr(len(expr))
		for i := range expr {
			ret_obj[i] = irReduceExpr(ctx, expr[i])
		}
		return IrExprObj(ret_obj)
	case IrExprInfix:
		return IrExprInfix{
			kind: expr.kind,
			lhs:  irReduceExpr(ctx, expr.lhs),
			rhs:  irReduceExpr(ctx, expr.rhs),
		}
	case IrExprArgRef:
		if ctx.args != nil {
			return ctx.args[expr]
		}
	case IrExprDefRef:
		irReduceDef(ctx, int(expr))
		ref_def := &ctx.ir.defs[expr]
		if ref_def.num_params == 0 {
			return ref_def.body
		}
	case IrExprForm:
		ret_form := allocˇIrExpr(len(expr))
		var ret_expr IrExpr = IrExprForm(ret_form)
		for i := range expr {
			ret_form[i] = irReduceExpr(ctx, expr[i])
		}
		switch callee := ret_form[0].(type) {
		case IrExprDefRef:
			ref_def := &ctx.ir.defs[callee]
			irReduceDef(ctx, int(callee))
			old_args := ctx.args
			ctx.args = ret_form[1:]
			ret_expr = irReduceExpr(ctx, ref_def.body)
			ctx.args = old_args
		}
		return ret_expr
	}
	return ir_expr
}

type CtxReduce struct {
	ir   *Ir
	args []IrExpr
	done []bool
}
