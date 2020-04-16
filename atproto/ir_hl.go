package main

type IrHL struct {
	defs []IrHLDef
	anns struct {
		origin_ast *Ast
	}
}

type IrHLDef struct {
	body IrHLExpr
	anns struct {
		origin_def *AstDef
		name       Str
	}
}

type IrHLExpr struct {
	variant IrHLExprVariant
	anns    struct {
		origin_def  *AstDef
		origin_expr *AstExpr
		ty          IrHLType
	}
}

type IrHLExprVariant interface{ implementsIrHLExprVariant() }
type IrHLType interface{ implementsIrHLType() }

type IrHLExprTag Str
type IrHLExprInt int64
type IrHLExprVoid struct{}
type IrHLExprIdent Str
type IrHLExprRefTmp struct{ ast_ref *AstNameRef }
type IrHLExprRefDef int
type IrHLExprRefLocal struct {
	let       *IrHLExpr
	let_idx   int
	param_idx int
}
type IrHLExprCall []IrHLExpr
type IrHLExprList []IrHLExpr
type IrHLExprBag []IrHLExpr
type IrHLExprType struct{ ty IrHLType }
type IrHLExprInfix struct {
	kind Str
	lhs  IrHLExpr
	rhs  IrHLExpr
}
type IrHLExprLet struct {
	names []Str
	exprs []IrHLExpr
	body  IrHLExpr
}
type IrHLExprFunc struct {
	params []IrHLExpr
	body   IrHLExpr
}
type IrHLExprPrimCallee Str
type IrHLExprPrimCase struct {
	scrut IrHLExpr
	cases []IrHLExpr
}
type IrHLExprPrimCmp struct {
	kind IrHLExprPrimCmpKind
	lhs  IrHLExpr
	rhs  IrHLExpr
}
type IrHLExprPrimArith struct {
	kind IrHLExprPrimArithKind
	lhs  IrHLExpr
	rhs  IrHLExpr
}
type IrHLExprPrimLen struct {
	subj IrHLExpr
}
type IrHLExprPrimCallExt struct {
	callee IrHLExpr
	args   []IrHLExpr
}
type IrHLExprPrimExtRef struct {
	name_tag  IrHLExpr
	opts      IrHLExpr
	ty        IrHLExpr
	fn_params []IrHLExpr
}

type IrHLExprPrimArithKind int

const (
	_ IrHLExprPrimArithKind = iota
	hl_arith_add
	hl_arith_mul
	hl_arith_sub
	hl_arith_div
	hl_arith_mod
)

type IrHLExprPrimCmpKind int

const (
	_ IrHLExprPrimCmpKind = iota
	hl_cmp_eq
	hl_cmp_neq
	hl_cmp_lt
	hl_cmp_gt
	hl_cmp_leq
	hl_cmp_geq
)

type IrHLTypeTag Str
type IrHLTypeBag struct {
	field_names []Str
	field_types []IrHLType
	is_union    bool
}
type IrHLTypeInt struct {
	min  int64
	max  int64
	c    bool
	word bool
}
type IrHLTypeVoid struct{}
type IrHLTypeExt Str
type IrHLTypePtr struct{ payload IrHLType }
type IrHLTypeArr struct {
	size    int
	payload IrHLType
}
type IrHLTypeFunc struct {
	returns IrHLType
	params  []IrHLType
	aborts  struct {
		always      bool
		potentially bool
	}
}

func (IrHLTypeArr) implementsIrHLType()  {}
func (IrHLTypeFunc) implementsIrHLType() {}
func (IrHLTypeExt) implementsIrHLType()  {}
func (IrHLTypePtr) implementsIrHLType()  {}
func (IrHLTypeVoid) implementsIrHLType() {}
func (IrHLTypeBag) implementsIrHLType()  {}
func (IrHLTypeInt) implementsIrHLType()  {}
func (IrHLTypeTag) implementsIrHLType()  {}

func (IrHLExprType) implementsIrHLExprVariant()        {}
func (IrHLExprTag) implementsIrHLExprVariant()         {}
func (IrHLExprInt) implementsIrHLExprVariant()         {}
func (IrHLExprIdent) implementsIrHLExprVariant()       {}
func (IrHLExprRefDef) implementsIrHLExprVariant()      {}
func (IrHLExprRefLocal) implementsIrHLExprVariant()    {}
func (IrHLExprCall) implementsIrHLExprVariant()        {}
func (IrHLExprList) implementsIrHLExprVariant()        {}
func (IrHLExprBag) implementsIrHLExprVariant()         {}
func (IrHLExprInfix) implementsIrHLExprVariant()       {}
func (IrHLExprFunc) implementsIrHLExprVariant()        {}
func (IrHLExprVoid) implementsIrHLExprVariant()        {}
func (IrHLExprRefTmp) implementsIrHLExprVariant()      {}
func (IrHLExprPrimArith) implementsIrHLExprVariant()   {}
func (IrHLExprPrimCallExt) implementsIrHLExprVariant() {}
func (IrHLExprPrimCase) implementsIrHLExprVariant()    {}
func (IrHLExprPrimCmp) implementsIrHLExprVariant()     {}
func (IrHLExprPrimLen) implementsIrHLExprVariant()     {}
func (IrHLExprLet) implementsIrHLExprVariant()         {}
func (IrHLExprPrimCallee) implementsIrHLExprVariant()  {}
func (IrHLExprPrimExtRef) implementsIrHLExprVariant()  {}

type CtxIrHLFromAst struct {
	ir       IrHL
	cur_def  *AstDef
	num_defs int
}

func irHLFrom(ast *Ast) IrHL {
	ctx := CtxIrHLFromAst{num_defs: 0, ir: IrHL{
		defs: ªIrHLDef(ast.anns.num_def_toks),
	}}
	ctx.ir.anns.origin_ast = ast

	for i := range ast.defs {
		irHLTopDefsFromAstTopDef(&ctx, &ast.defs[i])
	}
	ctx.ir.defs = ctx.ir.defs[0:ctx.num_defs]
	for i := range ctx.ir.defs {
		irHlDefResolveTmpRefs(&ctx, &ctx.ir.defs[i])
	}
	return ctx.ir
}

func irHLTopDefsFromAstTopDef(ctx *CtxIrHLFromAst, top_def *AstDef) {
	old_def := ctx.cur_def
	ctx.cur_def = top_def

	def := &ctx.ir.defs[ctx.num_defs]
	ctx.num_defs++
	def.anns.origin_def = top_def
	def.anns.name = top_def.anns.name
	def.body = irHlExprMaybeFuncFromAstDef(ctx, top_def)
	ctx.cur_def = old_def
}

func irHlExprMaybeFuncFromAstDef(ctx *CtxIrHLFromAst, def *AstDef) IrHLExpr {
	switch def_head := def.head.variant.(type) {
	case AstExprIdent:
		return irHlExprMaybeLetFromAstDef(ctx, def)
	case AstExprForm:
		fn := IrHLExprFunc{params: ªIrHLExpr(len(def_head) - 1)}
		for i := 1; i < len(def_head); i++ {
			fn.params[i-1] = IrHLExpr{variant: IrHLExprIdent(def_head[i].variant.(AstExprIdent))}
			fn.params[i-1].anns.origin_def = def
			fn.params[i-1].anns.origin_expr = &def_head[i]
		}
		fn.body = irHlExprMaybeLetFromAstDef(ctx, def)
		return IrHLExpr{variant: fn}
	default:
		panic("def head not supported in this prototype: " + string(astNodeSrc(&def.head.base, ctx.ir.anns.origin_ast)))
	}
}

func irHlExprMaybeLetFromAstDef(ctx *CtxIrHLFromAst, def *AstDef) (ret_expr IrHLExpr) {
	if len(def.defs) == 0 {
		return irHlExprFrom(ctx, &def.body)
	}
	ret_let := IrHLExprLet{
		names: ªStr(len(def.defs)),
		exprs: ªIrHLExpr(len(def.defs)),
	}
	for i := range def.defs {
		this_def := &def.defs[i]
		ret_let.names[i] = this_def.anns.name
		ret_let.exprs[i] = irHlExprMaybeFuncFromAstDef(ctx, this_def)
	}
	ret_let.body = irHlExprFrom(ctx, &def.body)
	ret_expr.variant = ret_let
	ret_expr.anns.origin_def, ret_expr.anns.origin_expr = def, &def.body
	return
}

func irHlExprFrom(ctx *CtxIrHLFromAst, expr *AstExpr) (ret_expr IrHLExpr) {
	ast := ctx.ir.anns.origin_ast
	ret_expr.anns.origin_def, ret_expr.anns.origin_expr = ctx.cur_def, expr
	switch it := expr.variant.(type) {

	case AstExprLitInt:
		ret_expr.variant = IrHLExprInt(it)
		ret_expr.anns.ty = IrHLTypeInt{min: int64(it), max: int64(it)}

	case AstExprLitStr:
		ret_list := ªAstExpr(len(it))
		for i, char := range it {
			ret_list[i].base = expr.base
			ret_list[i].variant = AstExprLitInt(char)
		}
		ret_expr = irHlExprFrom(ctx, &AstExpr{variant: AstExprLitList(ret_list), base: expr.base, anns: expr.anns})

	case AstExprLitList:
		ret_list := ªIrHLExpr(len(it))
		for i := range it {
			ret_list[i] = irHlExprFrom(ctx, &it[i])
		}
		ret_expr.variant = IrHLExprList(ret_list)

	case AstExprLitObj:
		ret_obj := ªIrHLExpr(len(it))
		for i := range it {
			ret_obj[i] = irHlExprFrom(ctx, &it[i])
		}
		ret_expr.variant = IrHLExprBag(ret_obj)

	case AstExprIdent:
		if strEq(it, "()") {
			ret_expr.variant = IrHLExprVoid{}
			ret_expr.anns.ty = IrHLTypeVoid{}
		} else if ref := astScopesResolve(&ctx.cur_def.scope, it, -1); ref != nil {
			ret_expr.variant = IrHLExprRefTmp{ast_ref: ref}
		} else {
			println(string(astNodeMsg("unresolved in line ", &expr.base, ast)))
			ret_expr.variant = IrHLExprIdent(it)
		}
	case AstExprForm:
		for _, supported_infix := range []string{":", "."} {
			if lhs, rhs := astExprFormSplit(expr, supported_infix, false, false, false, ast); lhs.variant != nil && rhs.variant != nil {
				ret_expr.variant = IrHLExprInfix{
					kind: Str(supported_infix),
					lhs:  irHlExprFrom(ctx, &lhs),
					rhs:  irHlExprFrom(ctx, &rhs),
				}
				return
			}
		}
		if ret_expr.variant == nil && len(it) == 2 {
			if prefix, _ := it[0].variant.(AstExprIdent); len(prefix) == 1 {
				if ident, _ := it[1].variant.(AstExprIdent); ident != nil {
					if prefix[0] == '#' {
						ret_expr.variant = IrHLExprTag(ident)
					} else if prefix[0] == '/' {
						ret_expr.variant = IrHLExprPrimCallee(ident)
					}
				}
			}
		}
		if ret_expr.variant == nil {
			ret_call := ªIrHLExpr(len(it))
			for i := range it {
				ret_call[i] = irHlExprFrom(ctx, &it[i])
			}
			ret_expr.variant = IrHLExprCall(ret_call)
		}
	default:
		panic(it)
	}
	if ret_expr.variant == nil {
		fail(astNodeMsg("Syntax error in line ", &expr.base, ast))
	}
	return
}

func irHlExprTraverse(expr *IrHLExpr, def *IrHLDef, visitor func(*IrHLExpr) IrHLExprVariant) IrHLExprVariant {
	switch it := expr.variant.(type) {
	case IrHLExprRefTmp, IrHLExprType, IrHLExprVoid, IrHLExprTag, IrHLExprInt, IrHLExprIdent, IrHLExprRefDef, IrHLExprRefLocal, IrHLExprPrimCallee:
		// no further traversal from here
	case IrHLExprPrimExtRef:
		it.name_tag.variant = irHlExprTraverse(&it.name_tag, def, visitor)
		it.opts.variant = irHlExprTraverse(&it.opts, def, visitor)
		it.ty.variant = irHlExprTraverse(&it.ty, def, visitor)
		for i := range it.fn_params {
			it.fn_params[i].variant = irHlExprTraverse(&it.fn_params[i], def, visitor)
		}
		expr.variant = it
	case IrHLExprFunc:
		it.body.variant = irHlExprTraverse(&it.body, def, visitor)
		expr.variant = it
	case IrHLExprCall:
		for i := range it {
			it[i].variant = irHlExprTraverse(&it[i], def, visitor)
		}
		expr.variant = it
	case IrHLExprList:
		for i := range it {
			it[i].variant = irHlExprTraverse(&it[i], def, visitor)
		}
		expr.variant = it
	case IrHLExprBag:
		for i := range it {
			it[i].variant = irHlExprTraverse(&it[i], def, visitor)
		}
		expr.variant = it
	case IrHLExprInfix:
		it.lhs.variant = irHlExprTraverse(&it.lhs, def, visitor)
		it.rhs.variant = irHlExprTraverse(&it.rhs, def, visitor)
		expr.variant = it
	case IrHLExprPrimCmp:
		it.lhs.variant = irHlExprTraverse(&it.lhs, def, visitor)
		it.rhs.variant = irHlExprTraverse(&it.rhs, def, visitor)
		expr.variant = it
	case IrHLExprPrimArith:
		it.lhs.variant = irHlExprTraverse(&it.lhs, def, visitor)
		it.rhs.variant = irHlExprTraverse(&it.rhs, def, visitor)
		expr.variant = it
	case IrHLExprPrimLen:
		it.subj.variant = irHlExprTraverse(&it.subj, def, visitor)
		expr.variant = it
	case IrHLExprPrimCallExt:
		it.callee.variant = irHlExprTraverse(&it.callee, def, visitor)
		for i := range it.args {
			it.args[i].variant = irHlExprTraverse(&it.args[i], def, visitor)
		}
		expr.variant = it
	case IrHLExprPrimCase:
		it.scrut.variant = irHlExprTraverse(&it.scrut, def, visitor)
		for i := range it.cases {
			it.cases[i].variant = irHlExprTraverse(&it.cases[i], def, visitor)
		}
		expr.variant = it
	case IrHLExprLet:
		for i := range it.exprs {
			it.exprs[i].variant = irHlExprTraverse(&it.exprs[i], def, visitor)
		}
		it.body.variant = irHlExprTraverse(&it.body, def, visitor)
		expr.variant = it
	default:
		panic(it)
	}
	expr.variant = visitor(expr)
	return expr.variant
}

func irHlExprResolvePrimCalls() {
	// if strEq(callee, "call") && len(ret_call) > 1 {
	// 	ret_expr.variant = IrHLExprPrimCallExt{callee: ret_call[1], args: ret_call[2:]}
	// } else if strEq(callee, "len") && len(ret_call) == 2 {
	// 	ret_expr.variant = IrHLExprPrimLen{subj: ret_call[1]}
	// } else if is_or := strEq(callee, "or"); (is_or || strEq(callee, "and")) && len(ret_call) == 3 {
	// 	ret_branch := IrHLExprPrimCase{
	// 		scrut: ret_call[1],
	// 		cases: ªIrHLExpr(2),
	// 	}
	// 	if is_or { // or a b -> case a [false: b, true: true]
	// 		ret_branch.cases[0].variant = IrHLExprInfix{kind: Str(":"), lhs: IrHLExpr{variant: IrHLExprTag("false")}, rhs: ret_call[2]}
	// 		ret_branch.cases[1].variant = IrHLExprInfix{kind: Str(":"), lhs: IrHLExpr{variant: IrHLExprTag("true")}, rhs: IrHLExpr{variant: IrHLExprTag("true")}}
	// 	} else { // and a b -> case a [true: b, false: false]
	// 		ret_branch.cases[0].variant = IrHLExprInfix{kind: Str(":"), lhs: IrHLExpr{variant: IrHLExprTag("true")}, rhs: ret_call[2]}
	// 		ret_branch.cases[1].variant = IrHLExprInfix{kind: Str(":"), lhs: IrHLExpr{variant: IrHLExprTag("false")}, rhs: IrHLExpr{variant: IrHLExprTag("false")}}
	// 	}
	// 	ret_expr.variant = ret_branch
	// } else if strEq(callee, "case") && len(ret_call) == 3 {
	// 	if lit_cases, _ := ret_call[2].variant.(IrHLExprBag); len(lit_cases) != 0 {
	// 		ret_expr.variant = IrHLExprPrimCase{
	// 			scrut: ret_call[1],
	// 			cases: lit_cases,
	// 		}
	// 	}
	// } else if strEq(callee, "extern") && (len(ret_call) == 4 || len(ret_call) == 5) {
	// 	ret_ext := IrHLExprPrimExtRef{
	// 		name_tag:  ret_call[1],
	// 		opts:      ret_call[2],
	// 		ty:        ret_call[3],
	// 		fn_params: nil,
	// 	}
	// 	if len(ret_call) == 5 {
	// 		if lit_params, _ := ret_call[4].variant.(IrHLExprBag); lit_params != nil {
	// 			ret_ext.fn_params = []IrHLExpr(lit_params)
	// 		}
	// 	}
	// 	if len(ret_call) == 4 || ret_ext.fn_params != nil {
	// 		ret_expr.variant = ret_ext
	// 	}
	// } else if strEq(callee, "cmp") && len(ret_call) == 4 {
	// 	if tag, _ := ret_call[1].variant.(IrHLExprTag); tag != nil {
	// 		cmp_kind := IrHLExprPrimCmpKind(0)
	// 		if strEq(tag, "eq") {
	// 			cmp_kind = hl_cmp_eq
	// 		} else if strEq(tag, "neq") {
	// 			cmp_kind = hl_cmp_neq
	// 		} else if strEq(tag, "lt") {
	// 			cmp_kind = hl_cmp_lt
	// 		} else if strEq(tag, "gt") {
	// 			cmp_kind = hl_cmp_gt
	// 		} else if strEq(tag, "leq") {
	// 			cmp_kind = hl_cmp_leq
	// 		} else if strEq(tag, "geq") {
	// 			cmp_kind = hl_cmp_geq
	// 		}
	// 		if cmp_kind != 0 {
	// 			ret_expr.variant = IrHLExprPrimCmp{
	// 				kind: cmp_kind,
	// 				lhs:  ret_call[2],
	// 				rhs:  ret_call[3],
	// 			}
	// 		}
	// 	}
	// }
}

func irHlDefResolveTmpRefs(ctx *CtxIrHLFromAst, def *IrHLDef) {
	def.body.variant = irHlExprTraverse(&def.body, def, func(expr *IrHLExpr) IrHLExprVariant {
		if it, ok := expr.variant.(IrHLExprRefTmp); ok {
			expr.variant = nil
			if it.ast_ref.top_def == it.ast_ref.ref_def && it.ast_ref.param_idx == -1 {
				for i := range ctx.ir.defs {
					if ctx.ir.defs[i].anns.origin_def == it.ast_ref.top_def {
						expr.variant = IrHLExprRefDef(i)
						break
					}
				}
			} else {
				ret_local := IrHLExprRefLocal{let: nil, let_idx: -1, param_idx: it.ast_ref.param_idx}
				if it.ast_ref.top_def != it.ast_ref.ref_def {
					_ = irHlExprTraverse(&def.body, def, func(this_expr *IrHLExpr) IrHLExprVariant {
						if ret_local.let == nil {
							if let, is := this_expr.variant.(IrHLExprLet); is {
								for i, name := range let.names {
									if strEql(name, it.ast_ref.ref_def.anns.name) {
										ret_local.let_idx = i
										ret_local.let = this_expr
										break
									}
								}
							}
						}
						return this_expr.variant
					})
					assert(ret_local.let != nil && ret_local.let_idx >= 0)
				}
				expr.variant = ret_local
			}
			assert(expr.variant != nil)
		}
		return expr.variant
	})
}
