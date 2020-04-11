package main

type CtxAstToIr struct {
	scope    *AstScopes
	dst      *Ir
	cur_def  []*AstDef
	num_defs int
}

type Ir struct {
	origin_ast *Ast
	defs       []IrDef
}

type IrDef struct {
	origin []*AstDef
	name   Str
	body   []IrExpr
}

type IrExpr interface{ implementsIrExpr() }

type IrExprLitInt int64

type IrExprLitStr Str

type IrExprDefRef struct{ *IrDef }

type IrExprIdent Str

type IrExprIdentTagged Str

type IrExprSlashed []IrExpr

type IrExprForm []IrExpr

type IrExprArr []IrExpr

type IrExprObj []IrExpr

type IrExprPair struct {
	lhs IrExpr
	rhs IrExpr
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
func (IrExprPair) implementsIrExpr()        {}

func irFromAst(ast *Ast, expr *AstExpr) Ir {
	ret_ir := Ir{origin_ast: ast}
	ctx := CtxAstToIr{scope: &ast.scope, dst: &ret_ir, cur_def: nil, num_defs: 0}
	_ = irFromExpr(&ctx, expr).(IrExprDefRef)
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
		if resolved := astScopesResolve(ctx.scope, expr, -1); resolved != nil {

		}
		panic("IDENT\t" + string(expr))

	case AstExprForm:
		if lhs, rhs := astExprFormSplit(ast_expr, ":", false, false, false, ctx.dst.origin_ast); lhs != nil && rhs != nil {
			return IrExprPair{
				lhs: irFromExpr(ctx, lhs),
				rhs: irFromExpr(ctx, rhs),
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
