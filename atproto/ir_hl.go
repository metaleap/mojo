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
type IrHLTypePrim interface{ implementsIrHLTypePrim() }

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
type IrHLExprFunc struct {
	params []IrHLExpr
	body   IrHLExpr
}
type IrHLExprPrimCase struct {
	scrut IrHLExpr
	cases []IrHLExpr
}
type IrHLExprPrimCmpInt struct {
	kind IrLLCmpIntKind
	lhs  IrHLExpr
	rhs  IrHLExpr
}
type IrHLExprPrimOpInt struct {
	kind IrLLOpIntKind
	lhs  IrHLExpr
	rhs  IrHLExpr
}
type IrHLExprPrimOpBool struct {
	kind IrLLOpBoolKind
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

type IrHLTypeBag struct {
	field_names []Str
	field_types []IrHLType
	is_union    bool
}
type IrHLTypeInt struct {
	min int64
	max int64
}
type IrHLTypeAborts struct{}
type IrHLTypeTag struct{}
type IrHlTypeSlice struct{ payload IrHLType }
type IrHLTypePrimVoid struct{}
type IrHLTypePrimExternal Str
type IrHLTypePrimInt struct{ bit_width int }
type IrHLTypePrimPtr struct{ payload IrHLType }
type IrHLTypePrimStruct struct{ fields []IrHLType }
type IrHLTypePrimArr struct {
	size    int
	payload IrHLType
}
type IrHLTypePrimFunc struct {
	returns IrHLType
	params  []IrHLType
}

func (IrHLTypePrimArr) implementsIrHLTypePrim()      {}
func (IrHLTypePrimFunc) implementsIrHLTypePrim()     {}
func (IrHLTypePrimInt) implementsIrHLTypePrim()      {}
func (IrHLTypePrimExternal) implementsIrHLTypePrim() {}
func (IrHLTypePrimPtr) implementsIrHLTypePrim()      {}
func (IrHLTypePrimStruct) implementsIrHLTypePrim()   {}
func (IrHLTypePrimVoid) implementsIrHLTypePrim()     {}

func (IrHLTypePrimArr) implementsIrHLType()      {}
func (IrHLTypePrimFunc) implementsIrHLType()     {}
func (IrHLTypePrimInt) implementsIrHLType()      {}
func (IrHLTypePrimExternal) implementsIrHLType() {}
func (IrHLTypePrimPtr) implementsIrHLType()      {}
func (IrHLTypePrimStruct) implementsIrHLType()   {}
func (IrHLTypePrimVoid) implementsIrHLType()     {}
func (IrHLTypeBag) implementsIrHLType()          {}
func (IrHLTypeInt) implementsIrHLType()          {}
func (IrHLTypeAborts) implementsIrHLType()       {}
func (IrHLTypeTag) implementsIrHLType()          {}
func (IrHlTypeSlice) implementsIrHLType()        {}

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
func (IrHLExprPrimOpBool) implementsIrHLExprVariant()  {}
func (IrHLExprPrimOpInt) implementsIrHLExprVariant()   {}
func (IrHLExprPrimCallExt) implementsIrHLExprVariant() {}
func (IrHLExprPrimCase) implementsIrHLExprVariant()    {}
func (IrHLExprPrimCmpInt) implementsIrHLExprVariant()  {}
func (IrHLExprPrimLen) implementsIrHLExprVariant()     {}

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
		def.body = irHlExprFrom(ctx, &top_def.body)
	case AstExprForm:
		fn := IrHLExprFunc{params: ªIrHLExpr(len(def_head) - 1)}
		for i := 1; i < len(def_head); i++ {
			fn.params[i-1] = IrHLExpr{variant: IrHLExprIdent(def_head[i].variant.(AstExprIdent))}
			fn.params[i-1].anns.origin_def = top_def
			fn.params[i-1].anns.origin_expr = &def_head[i]
		}
		fn.body = irHlExprFrom(ctx, &top_def.body)
		def.body = IrHLExpr{variant: fn}
	default:
		panic("def head not supported in this prototype: " + string(astNodeSrc(&top_def.head.base, ctx.ir.anns.origin_ast)))
	}

	ctx.cur_def = old_def
}

func irHlExprFrom(ctx *CtxIrHLFromAst, expr *AstExpr) (ret_expr IrHLExpr) {
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
			ret_expr.anns.ty = IrHLTypePrimVoid{}
		} else if ref := astScopesResolve(&ctx.cur_def.scope, it, -1); ref != nil {
			ret_expr.variant = IrHLExprRefTmp{ast_ref: ref}
		}
		ret_expr.variant = IrHLExprIdent(it)

	case AstExprForm:
		for _, supported_infix := range []string{":", "."} {
			if lhs, rhs := astExprFormSplit(expr, supported_infix, false, false, false, ctx.ir.anns.origin_ast); lhs.variant != nil && rhs.variant != nil {
				ret_expr.variant = IrHLExprInfix{
					kind: Str(supported_infix),
					lhs:  irHlExprFrom(ctx, &lhs),
					rhs:  irHlExprFrom(ctx, &rhs),
				}
				return
			}
		}

		ret_call := ªIrHLExpr(len(it))
		for i := range it {
			ret_call[i] = irHlExprFrom(ctx, &it[i])
		}
		ret_expr.variant = IrHLExprCall(ret_call)

	default:
		panic(it)
	}
	if ret_expr.variant == nil {
		panic("newly introduced bug: should be unreachable here")
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
	case IrHLExprPrimCmpInt:
		irHlExprResolveTmpRefs(ctx, &it.lhs)
		irHlExprResolveTmpRefs(ctx, &it.rhs)
		expr.variant = it
	case IrHLExprPrimOpBool:
		irHlExprResolveTmpRefs(ctx, &it.lhs)
		irHlExprResolveTmpRefs(ctx, &it.rhs)
		expr.variant = it
	case IrHLExprPrimOpInt:
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
	case IrHLExprPrimCase:
		irHlExprResolveTmpRefs(ctx, &it.scrut)
		for i := range it.cases {
			irHlExprResolveTmpRefs(ctx, &it.cases[i])
		}
		expr.variant = it
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
	print(" := ")
	irHLDumpExpr(ir, &def.body)
	print("\n\n")
}

func irHLDumpExpr(ir *IrHL, expr *IrHLExpr) {
	switch it := expr.variant.(type) {
	case IrHLExprType:
		irHLDumpType(ir, it.ty)
	case IrHLExprTag:
		print('#')
		print(string(it))
	case IrHLExprInt:
		print(it)
	case IrHLExprIdent:
		print(string(it))
	case IrHLExprRefDef:
		print(string(ir.defs[it].anns.name))
	case IrHLExprRefArg:
		print('@')
		print(it)
	case IrHLExprCall:
		print('(')
		for i := range it {
			irHLDumpExpr(ir, &it[i])
		}
		print(')')
	case IrHLExprList:
		print("[ ")
		for i := range it {
			irHLDumpExpr(ir, &it[i])
			print(", ")
		}
		print(']')
	case IrHLExprBag:
		print("{ ")
		for i := range it {
			irHLDumpExpr(ir, &it[i])
			print(", ")
		}
		print('}')
	case IrHLExprInfix:
		irHLDumpExpr(ir, &it.lhs)
		print(string(it.kind))
		irHLDumpExpr(ir, &it.rhs)
	default:
		panic(it)
	}
}

func irHLDumpType(ir *IrHL, ty IrHLType) {
	switch it := ty.(type) {
	case IrHLTypePrimVoid:
		print("/V")
	case IrHLTypePrimInt:
		print("/I")
		print(it.bit_width)
	case IrHLTypePrimExternal:
		print("/Extern #")
		print(string(it))
	case IrHLTypePrimPtr:
		print("/P")
		irHLDumpType(ir, it.payload)
	case IrHLTypePrimArr:
		print("/A/")
		print(it.size)
		irHLDumpType(ir, it.payload)
	case IrHLTypePrimFunc:
		print("/F")
		irHLDumpType(ir, it.returns)
		print('/')
		print(len(it.params))
		for i := range it.params {
			irHLDumpType(ir, it.params[i])
		}
	case IrHLTypePrimStruct:
		print("/S/")
		print(len(it.fields))
		for i := range it.fields {
			irHLDumpType(ir, it.fields[i])
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
		print('}')
	case IrHLTypeInt:
		print("Int/")
		print(it.min)
		print('/')
		print(it.max)
	case IrHLTypeAborts:
		print("/Aborts")
	default:
		panic(it)
	}
}
