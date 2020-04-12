package main

type Ir struct {
	origin_ast *Ast
	defs       []IrDef
}

type IrDef struct {
	origin_def *AstDef
	name       Str
	params     []Str
	body       IrExpr
}

type IrExpr interface{ implementsIrExpr() }

type IrExprLitInt int64

type IrExprLitStr Str

type IrExprDefRef int

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
	cur_args []IrExpr
}

func irFromAst(ast *Ast, expr *AstExpr) Ir {
	ret_ir := Ir{origin_ast: ast, defs: allocˇIrDef(len(ast.defs))}
	ctx := CtxAstToIr{scope: &ast.scope, dst: &ret_ir, num_defs: 0}
	_ = irFromExpr(&ctx, expr).(IrExprDefRef)
	ret_ir.defs = ret_ir.defs[0:ctx.num_defs]
	return ret_ir
}

func irFromExpr(ctx *CtxAstToIr, ast_expr *AstExpr) IrExpr {
	switch expr := ast_expr.kind.(type) {

	case AstExprLitInt:
		return IrExprLitInt(expr)

	case AstExprLitStr:
		return IrExprLitStr(expr)

	case AstExprLitClip:
		ret_arr := allocˇIrExpr(len(expr))
		for i := range expr {
			ret_arr[i] = irFromExpr(ctx, &expr[i])
		}
		return IrExprArr(ret_arr)

	case AstExprLitCurl:
		ret_obj := allocˇIrExpr(len(expr))
		for i := range expr {
			ret_obj[i] = irFromExpr(ctx, &expr[i])
		}
		return IrExprObj(ret_obj)

	case AstExprIdent:
		if resolved := astScopesResolve(ctx.scope, expr, -1); resolved == nil {
			fail("unresolved identifier '", expr, "' near:\n", astNodeSrcStr(&ast_expr.base, ctx.dst.origin_ast))
		} else if resolved.param_idx >= 0 {
			panic("TODO ARG\t" + string(expr))
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
					params:     nil,
				}
				if head_form, _ := resolved.ref_def.head.kind.(AstExprForm); head_form != nil {
					ir_def.params = allocˇStr(len(head_form) - 1)
					for i := range head_form {
						ir_def.params[i] = Str(head_form[i].kind.(AstExprIdent))
					}
				}
				ir_def_idx = ctx.num_defs
				ctx.dst.defs[ctx.num_defs] = ir_def
				ctx.num_defs++
				ctx.dst.defs[ctx.num_defs].body = irFromExpr(ctx, &resolved.ref_def.body)
			}
			return IrExprDefRef(ir_def_idx)
		}
		panic("unreachable")

	case AstExprForm:
		for _, supported_infix := range []string{":"} {
			if lhs, rhs := astExprFormSplit(ast_expr, supported_infix, false, false, false, ctx.dst.origin_ast); lhs != nil && rhs != nil {
				return IrExprInfix{
					kind: Str(supported_infix),
					lhs:  irFromExpr(ctx, lhs),
					rhs:  irFromExpr(ctx, rhs),
				}
			}
		}
		// TODO: tweak astExprTaggedIdent and astExprSlashed to accept AstExprForm instead of *AstExpr
		if tagged_ident := astExprTaggedIdent(ast_expr); tagged_ident != nil {
			return IrExprIdentTagged(tagged_ident)
		} else if slashed := astExprSlashed(ast_expr); slashed != nil {
			ret_slashed := IrExprSlashed(allocˇIrExpr(len(slashed)))
			for i, sub_expr := range slashed {
				ret_slashed[i] = irFromExpr(ctx, sub_expr)
			}
			return ret_slashed
		}

		panic("FORM\t" + string(astNodeSrcStr(&ast_expr.base, ctx.dst.origin_ast)))
	default:
		panic(expr)
	}
}
