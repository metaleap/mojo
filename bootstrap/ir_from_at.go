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

type IrExprTag Str

type IrExprSlashed []IrExpr

type IrExprForm []IrExpr

type IrExprArr []IrExpr

type IrExprObj []IrExpr

type IrExprInfix struct {
	kind Str
	lhs  IrExpr
	rhs  IrExpr
}

func (IrExprLitInt) implementsIrExpr()  {}
func (IrExprLitStr) implementsIrExpr()  {}
func (IrExprDefRef) implementsIrExpr()  {}
func (IrExprArgRef) implementsIrExpr()  {}
func (IrExprIdent) implementsIrExpr()   {}
func (IrExprTag) implementsIrExpr()     {}
func (IrExprSlashed) implementsIrExpr() {}
func (IrExprForm) implementsIrExpr()    {}
func (IrExprArr) implementsIrExpr()     {}
func (IrExprObj) implementsIrExpr()     {}
func (IrExprInfix) implementsIrExpr()   {}

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
			if len(expr) == 1 && expr[0] == '#' {
				return IrExprTag("")
			}
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
				old_scope := ctx.scope
				ctx.scope = &resolved.ref_def.scope
				ctx.dst.defs[ir_def_idx].body = irExprFrom(ctx, &resolved.ref_def.body)
				ctx.scope = old_scope
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
			return IrExprTag(tagged_ident)
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

func irExprEquiv(lhs IrExpr, rhs IrExpr) bool {
	if lhs != nil && rhs != nil {
		switch l := lhs.(type) {
		case IrExprLitInt:
			r, ok := rhs.(IrExprLitInt)
			return ok && l == r
		case IrExprLitStr:
			r, ok := rhs.(IrExprLitStr)
			return ok && strEql(l, r)
		case IrExprDefRef:
			r, ok := rhs.(IrExprDefRef)
			return ok && l == r
		case IrExprArgRef:
			r, ok := rhs.(IrExprArgRef)
			return ok && l == r
		case IrExprIdent:
			r, ok := rhs.(IrExprIdent)
			return ok && strEql(l, r)
		case IrExprTag:
			r, ok := rhs.(IrExprTag)
			return ok && strEql(l, r)
		case IrExprSlashed:
			r, ok := rhs.(IrExprSlashed)
			return ok && irExprsEquiv(l, r)
		case IrExprForm:
			r, ok := rhs.(IrExprForm)
			return ok && irExprsEquiv(l, r)
		case IrExprArr:
			r, ok := rhs.(IrExprArr)
			return ok && irExprsEquiv(l, r)
		case IrExprObj:
			r, ok := rhs.(IrExprObj)
			return ok && irExprsEquiv(l, r)
		case IrExprInfix:
			r, ok := rhs.(IrExprInfix)
			return ok && irExprEquiv(l.lhs, r.lhs) && irExprEquiv(l.rhs, r.rhs) && strEql(l.kind, r.kind)
		}
	}
	return (lhs == nil && rhs == nil)
}

func irExprsEquiv(lhs []IrExpr, rhs []IrExpr) bool {
	if len(lhs) == len(rhs) {
		for i := range lhs {
			if !irExprEquiv(lhs[i], rhs[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func irExprsFindKeyedValue(exprs []IrExpr, infix_kind string, key IrExpr) IrExpr {
	kind := Str(infix_kind)
	for i := range exprs {
		if expr_infix, ok := exprs[i].(IrExprInfix); ok && strEql(expr_infix.kind, kind) {
			if irExprEquiv(expr_infix.lhs, key) {
				return expr_infix.rhs
			}
		}
	}
	return nil
}

type CtxReduce struct {
	ir   *Ir
	args []IrExpr
	done []bool
}

func irReduceDefs(ir *Ir) {
	ctx := CtxReduce{ir: ir, done: allocˇbool(len(ir.defs)), args: nil}
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
	ret_expr := ir_expr
	switch expr := ir_expr.(type) {
	case IrExprSlashed:
		ret_slashed := allocˇIrExpr(len(expr))
		for i := range expr {
			ret_slashed[i] = irReduceExpr(ctx, expr[i])
		}
		ret_expr = IrExprSlashed(ret_slashed)
	case IrExprArr:
		ret_arr := allocˇIrExpr(len(expr))
		for i := range expr {
			ret_arr[i] = irReduceExpr(ctx, expr[i])
		}
		ret_expr = IrExprArr(ret_arr)
	case IrExprObj:
		ret_obj := allocˇIrExpr(len(expr))
		for i := range expr {
			ret_obj[i] = irReduceExpr(ctx, expr[i])
		}
		ret_expr = IrExprObj(ret_obj)
	case IrExprInfix:
		ret_expr = IrExprInfix{
			kind: expr.kind,
			lhs:  irReduceExpr(ctx, expr.lhs),
			rhs:  irReduceExpr(ctx, expr.rhs),
		}
	case IrExprArgRef:
		if ctx.args != nil {
			ret_expr = ctx.args[expr]
		}
	case IrExprDefRef:
		irReduceDef(ctx, int(expr))
		// ref_def := &ctx.ir.defs[expr]
		// if ref_def.num_params == 0 {
		// 	ret_expr = ref_def.body
		// }
	case IrExprForm:
		ret_form := allocˇIrExpr(len(expr))
		for i := range expr {
			ret_form[i] = irReduceExpr(ctx, expr[i])
		}
		ret_expr = IrExprForm(ret_form)
		switch callee := ret_form[0].(type) {
		case IrExprDefRef:
			ref_def := &ctx.ir.defs[callee]
			irReduceDef(ctx, int(callee))
			old_args := ctx.args
			ctx.args = ret_form[1:]
			ret_expr = irReduceExpr(ctx, ref_def.body)
			ctx.args = old_args
		}
	}

	if !irExprEquiv(ir_expr, ret_expr) {
		print("\n\n// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=\n// EVAL'D:\n  ")
		irExprDbgPrint(ctx.ir, ir_expr, 2)
		print("\n// TO:\n  ")
		irExprDbgPrint(ctx.ir, ret_expr, 2)
	}

	return ret_expr
}

func irDbgPrint(ir *Ir) {
	for i := range ir.defs {
		irDefDbgPrint(ir, &ir.defs[i])
	}
}

func irDefDbgPrint(ir *Ir, def *IrDef) {
	print(string(def.name))
	for i := 0; i < def.num_params; i++ {
		print(string(uintToStr(uint64(i), 10, 1, Str(" @"))))
	}
	print(" :=\n  ")
	irExprDbgPrint(ir, def.body, 2)
	println("\n")
}

func irExprDbgPrint(ir *Ir, expr IrExpr, ind int) {
	switch it := expr.(type) {
	case IrExprLitInt:
		print(it)
	case IrExprLitStr:
		print("\"" + string(it) + "\"")
	case IrExprDefRef:
		assert(it >= 0)
		print(string(ir.defs[it].name))
	case IrExprArgRef:
		assert(it >= 0)
		print("@" + string(uintToStr(uint64(it), 10, 1, nil)))
	case IrExprIdent:
		print(string(it))
	case IrExprTag:
		print("#" + string(it))
	case IrExprSlashed:
		for i := range it {
			print("/")
			irExprDbgPrint(ir, it[i], ind)
		}
	case IrExprForm:
		print("(")
		for i := range it {
			if i > 0 {
				print(" ")
			}
			irExprDbgPrint(ir, it[i], ind)
		}
		print(")")
	case IrExprInfix:
		irExprDbgPrint(ir, it.lhs, ind)
		print(" " + string(it.kind) + " ")
		irExprDbgPrint(ir, it.rhs, ind)
	case IrExprArr:
		print("[")
		irExprsDbgPrint(ir, it, ind, len(it) > 2)
		print("]")
	case IrExprObj:
		print("{")
		irExprsDbgPrint(ir, it, ind, len(it) > 2)
		print("}")
	}
}

func irExprsDbgPrint(ir *Ir, exprs []IrExpr, ind int, multiple_lines bool) {
	if multiple_lines {
		println()
	}
	for i := range exprs {
		if multiple_lines {
			for j := 0; j < ind; j++ {
				print(" ")
			}
		}
		irExprDbgPrint(ir, exprs[i], ind+2)
		if multiple_lines {
			print(",\n")
		} else {
			print(", ")
		}
	}
	if multiple_lines {
		for j := 0; j < ind; j++ {
			print(" ")
		}
	}
}
