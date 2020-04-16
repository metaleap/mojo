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
type IrHLExprRefArg int
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
type IrHLExprPrimBranch struct {
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

type IrHLTypeTag struct{}
type IrHlTypeSlice struct{ payload IrHLType }
type IrHLTypeBag struct {
	field_names []Str
	field_types []IrHLType
	is_union    bool
}
type IrHLTypeInt struct {
	min int64
	max int64
}
type IrHLTypeVoid struct{}
type IrHLTypeExternal Str
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

func (IrHLTypeArr) implementsIrHLType()      {}
func (IrHLTypeFunc) implementsIrHLType()     {}
func (IrHLTypeExternal) implementsIrHLType() {}
func (IrHLTypePtr) implementsIrHLType()      {}
func (IrHLTypeVoid) implementsIrHLType()     {}
func (IrHLTypeBag) implementsIrHLType()      {}
func (IrHLTypeInt) implementsIrHLType()      {}
func (IrHLTypeTag) implementsIrHLType()      {}
func (IrHlTypeSlice) implementsIrHLType()    {}

func (IrHLExprType) implementsIrHLExprVariant()        {}
func (IrHLExprTag) implementsIrHLExprVariant()         {}
func (IrHLExprInt) implementsIrHLExprVariant()         {}
func (IrHLExprIdent) implementsIrHLExprVariant()       {}
func (IrHLExprRefDef) implementsIrHLExprVariant()      {}
func (IrHLExprRefArg) implementsIrHLExprVariant()      {}
func (IrHLExprCall) implementsIrHLExprVariant()        {}
func (IrHLExprList) implementsIrHLExprVariant()        {}
func (IrHLExprBag) implementsIrHLExprVariant()         {}
func (IrHLExprInfix) implementsIrHLExprVariant()       {}
func (IrHLExprFunc) implementsIrHLExprVariant()        {}
func (IrHLExprVoid) implementsIrHLExprVariant()        {}
func (IrHLExprRefTmp) implementsIrHLExprVariant()      {}
func (IrHLExprPrimArith) implementsIrHLExprVariant()   {}
func (IrHLExprPrimCallExt) implementsIrHLExprVariant() {}
func (IrHLExprPrimBranch) implementsIrHLExprVariant()  {}
func (IrHLExprPrimCmp) implementsIrHLExprVariant()     {}
func (IrHLExprPrimLen) implementsIrHLExprVariant()     {}
func (IrHLExprLet) implementsIrHLExprVariant()         {}
func (IrHLExprPrimCallee) implementsIrHLExprVariant()  {}

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
		irHlExprResolveTmpRefs(&ctx, &ctx.ir.defs[i].body)
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
	switch def_head := top_def.head.variant.(type) {
	case AstExprIdent:
		def.body = irHlExprFromAstDef(ctx, top_def)
	case AstExprForm:
		fn := IrHLExprFunc{params: ªIrHLExpr(len(def_head) - 1)}
		for i := 1; i < len(def_head); i++ {
			fn.params[i-1] = IrHLExpr{variant: IrHLExprIdent(def_head[i].variant.(AstExprIdent))}
			fn.params[i-1].anns.origin_def = top_def
			fn.params[i-1].anns.origin_expr = &def_head[i]
		}
		fn.body = irHlExprFromAstDef(ctx, top_def)
		def.body = IrHLExpr{variant: fn}
	default:
		panic("def head not supported in this prototype: " + string(astNodeSrc(&top_def.head.base, ctx.ir.anns.origin_ast)))
	}

	ctx.cur_def = old_def
}

func irHlExprFromAstDef(ctx *CtxIrHLFromAst, def *AstDef) (ret_expr IrHLExpr) {
	if len(def.defs) == 0 {
		return irHlExprFrom(ctx, &def.body)
	}
	ret_let := IrHLExprLet{
		names: ªStr(len(def.defs)),
		exprs: ªIrHLExpr(len(def.defs)),
	}
	for i := range def.defs {
		ret_let.names[i] = def.defs[i].anns.name
		ret_let.exprs[i] = irHlExprFromAstDef(ctx, &def.defs[i])
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
		tmp_list := ªAstExpr(len(it))
		for i, char := range it {
			tmp_list[i].base = expr.base
			tmp_list[i].variant = AstExprLitInt(char)
		}
		ret_expr = irHlExprFrom(ctx, &AstExpr{variant: AstExprLitList(tmp_list), base: expr.base, anns: expr.anns})

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
		}
		ret_expr.variant = IrHLExprIdent(it)

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
		if ret_expr.variant == nil {
			if prefix, _ := it[0].variant.(AstExprIdent); len(prefix) == 1 {
				if ident, _ := it[1].variant.(AstExprIdent); ident != nil {
					if prefix[0] == '#' {
						ret_expr.variant = IrHLExprTag(ident)
					} else if prefix[0] == '/' {
						if ident[0] >= 'A' && ident[0] <= 'Z' {
							println("TODO: type-expr " + string(astNodeSrc(&expr.base, ast)))
						} else {
							ret_expr.variant = IrHLExprPrimCallee(ident)
						}
					}
				}
			}
		}
		if ret_expr.variant == nil {
			ret_call := ªIrHLExpr(len(it))
			for i := range it {
				ret_call[i] = irHlExprFrom(ctx, &it[i])
			}
			switch callee := ret_call[0].variant.(type) {
			case IrHLExprPrimCallee:
				if strEq(callee, "call") && len(ret_call) > 1 {
					ret_expr.variant = IrHLExprPrimCallExt{callee: ret_call[1], args: ret_call[2:]}
				} else if strEq(callee, "len") && len(ret_call) == 2 {
					ret_expr.variant = IrHLExprPrimLen{subj: ret_call[1]}
				} else if is_or := strEq(callee, "or"); (is_or || strEq(callee, "and")) && len(ret_call) == 3 {
					ret_branch := IrHLExprPrimBranch{
						scrut: ret_call[1],
						cases: ªIrHLExpr(2),
					}
					if is_or { // or a b -> case a [true: true, false: b]
						ret_branch.cases[0].variant = IrHLExprInfix{kind: Str(":"), lhs: IrHLExpr{variant: IrHLExprTag("true")}, rhs: IrHLExpr{variant: IrHLExprTag("true")}}
						ret_branch.cases[1].variant = IrHLExprInfix{kind: Str(":"), lhs: IrHLExpr{variant: IrHLExprTag("false")}, rhs: ret_call[2]}
					} else { // and a b -> case a [true: b, false: false]
						ret_branch.cases[0].variant = IrHLExprInfix{kind: Str(":"), lhs: IrHLExpr{variant: IrHLExprTag("true")}, rhs: ret_call[2]}
						ret_branch.cases[1].variant = IrHLExprInfix{kind: Str(":"), lhs: IrHLExpr{variant: IrHLExprTag("false")}, rhs: IrHLExpr{variant: IrHLExprTag("false")}}
					}
					ret_expr.variant = ret_branch
				} else if strEq(callee, "cmp") && len(ret_call) == 4 {
					if tag, _ := ret_call[1].variant.(IrHLExprTag); tag != nil {
						cmp_kind := IrHLExprPrimCmpKind(0)
						if strEq(tag, "eq") {
							cmp_kind = hl_cmp_eq
						} else if strEq(tag, "neq") {
							cmp_kind = hl_cmp_neq
						} else if strEq(tag, "lt") {
							cmp_kind = hl_cmp_lt
						} else if strEq(tag, "gt") {
							cmp_kind = hl_cmp_gt
						} else if strEq(tag, "leq") {
							cmp_kind = hl_cmp_leq
						} else if strEq(tag, "geq") {
							cmp_kind = hl_cmp_geq
						}
						if cmp_kind != 0 {
							ret_expr.variant = IrHLExprPrimCmp{
								kind: cmp_kind,
								lhs:  ret_call[2],
								rhs:  ret_call[3],
							}
						}
					}
				}
			default:
				ret_expr.variant = IrHLExprCall(ret_call)
			}
		}
	default:
		panic(it)
	}
	if ret_expr.variant == nil {
		fail(astNodeMsg("Syntax error in line ", &expr.base, ast))
	}
	return
}

func irHlExprResolveTmpRefs(ctx *CtxIrHLFromAst, expr *IrHLExpr) {
	switch it := expr.variant.(type) {
	case IrHLExprRefTmp:
		assert(it.ast_ref.top_def == it.ast_ref.ref_def)
		expr.variant = nil
		for i := range ctx.ir.defs {
			if ctx.ir.defs[i].anns.origin_def == it.ast_ref.top_def {
				expr.variant = IrHLExprRefDef(i)
				break
			}
		}
		if expr.variant == nil {
			panic("newly introduced bug: could not late-resolve def ref '" + string(it.ast_ref.top_def.anns.name) + "'")
		}
	case IrHLExprType, IrHLExprVoid, IrHLExprTag, IrHLExprInt, IrHLExprIdent, IrHLExprRefDef, IrHLExprRefArg:
		// no further traversal from here
	case IrHLExprFunc:
		irHlExprResolveTmpRefs(ctx, &it.body)
		expr.variant = it
	case IrHLExprCall:
		for i := range it {
			irHlExprResolveTmpRefs(ctx, &it[i])
		}
		expr.variant = it
	case IrHLExprList:
		for i := range it {
			irHlExprResolveTmpRefs(ctx, &it[i])
		}
		expr.variant = it
	case IrHLExprBag:
		for i := range it {
			irHlExprResolveTmpRefs(ctx, &it[i])
		}
		expr.variant = it
	case IrHLExprInfix:
		irHlExprResolveTmpRefs(ctx, &it.lhs)
		irHlExprResolveTmpRefs(ctx, &it.rhs)
		expr.variant = it
	case IrHLExprPrimCmp:
		irHlExprResolveTmpRefs(ctx, &it.lhs)
		irHlExprResolveTmpRefs(ctx, &it.rhs)
		expr.variant = it
	case IrHLExprPrimArith:
		irHlExprResolveTmpRefs(ctx, &it.lhs)
		irHlExprResolveTmpRefs(ctx, &it.rhs)
		expr.variant = it
	case IrHLExprPrimLen:
		irHlExprResolveTmpRefs(ctx, &it.subj)
		expr.variant = it
	case IrHLExprPrimCallExt:
		irHlExprResolveTmpRefs(ctx, &it.callee)
		for i := range it.args {
			irHlExprResolveTmpRefs(ctx, &it.args[i])
		}
		expr.variant = it
	case IrHLExprPrimBranch:
		irHlExprResolveTmpRefs(ctx, &it.scrut)
		for i := range it.cases {
			irHlExprResolveTmpRefs(ctx, &it.cases[i])
		}
		expr.variant = it
	case IrHLExprLet:
		for i := range it.exprs {
			irHlExprResolveTmpRefs(ctx, &it.exprs[i])
		}
		irHlExprResolveTmpRefs(ctx, &it.body)
	default:
		panic(it)
	}
}

func irHLDump(ir *IrHL) {
	for i := range ir.defs {
		irHLDumpDef(ir, &ir.defs[i])
	}
}

func irHLDumpDef(ir *IrHL, def *IrHLDef) {
	print(string(def.anns.name))
	print(" :=\n  ")
	irHLDumpExpr(ir, &def.body, 4)
	print("\n\n")
}

func irHLDumpExpr(ir *IrHL, expr *IrHLExpr, ind int) {
	switch it := expr.variant.(type) {
	case IrHLExprType:
		irHLDumpType(ir, it.ty)
	case IrHLExprTag:
		print("#")
		print(string(it))
	case IrHLExprInt:
		print(it)
	case IrHLExprIdent:
		print(string(it))
	case IrHLExprRefDef:
		print(string(ir.defs[it].anns.name))
	case IrHLExprRefArg:
		print("@")
		print(it)
	case IrHLExprCall:
		if expr.anns.origin_expr.anns.toks_throng {
			for i := range it {
				irHLDumpExpr(ir, &it[i], ind)
			}
		} else {
			print("(")
			for i := range it {
				if i > 0 {
					print(" ")
				}
				irHLDumpExpr(ir, &it[i], ind)
			}
			print(")")
		}
	case IrHLExprList:
		print("[ ")
		for i := range it {
			irHLDumpExpr(ir, &it[i], ind)
			print(", ")
		}
		print("]")
	case IrHLExprBag:
		if len(it) <= 2 { // gettin nasty now!
			print("{")
			for i := range it {
				if i > 0 {
					print(", ")
				}
				irHLDumpExpr(ir, &it[i], ind+2)
			}
			print("}")
		} else {
			print("{\n")
			for i := range it {
				for j := 0; j < ind; j++ {
					print(" ")
				}
				irHLDumpExpr(ir, &it[i], ind+2)
				print(",\n")
			}
			for i := 0; i < ind; i++ {
				print(" ")
			}
			print("}")
		}
	case IrHLExprInfix:
		irHLDumpExpr(ir, &it.lhs, ind)
		print(string(it.kind))
		print(" ")
		irHLDumpExpr(ir, &it.rhs, ind)
	case IrHLExprLet:
		print("(")
		irHLDumpExpr(ir, &it.body, ind)
		assert(len(it.names) == len(it.exprs))
		for i, name := range it.names {
			print(", ")
			print(string(name))
			print(" := ")
			irHLDumpExpr(ir, &it.exprs[i], ind)
		}
		print(")")
	default:
		panic(it)
	}
}

func irHLDumpType(ir *IrHL, ty IrHLType) {
	switch it := ty.(type) {
	case IrHLTypeVoid:
		print("/V")
	case IrHLTypeInt:
		print("/I/")
		print(it.min)
		print("/")
		print(it.max)
	case IrHLTypeExternal:
		print("/Extern #")
		print(string(it))
	case IrHLTypePtr:
		print("/P")
		irHLDumpType(ir, it.payload)
	case IrHLTypeArr:
		print("/A/")
		print(it.size)
		irHLDumpType(ir, it.payload)
	case IrHLTypeFunc:
		print("/F")
		irHLDumpType(ir, it.returns)
		print("/")
		print(len(it.params))
		for i := range it.params {
			irHLDumpType(ir, it.params[i])
		}
	case IrHLTypeTag:
		print("/Tag")
	case IrHLTypeBag:
		if it.is_union {
			print("/Union{ ")
		} else {
			print("/Struct{ ")
		}
		for i := range it.field_names {
			print(string(it.field_names[i]))
			print(": ")
			irHLDumpType(ir, it.field_types[i])
			print(", ")
		}
		print("}")
	default:
		panic(it)
	}
}
