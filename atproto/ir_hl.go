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
type IrHLExprStr Str
type IrHLExprIdent Str
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
	params []IrHLType
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
	scrut IrHLExpr
}
type IrHLExprPrimCallExt struct {
	callee IrHLExpr
	args   []IrHLExpr
}

type IrHLTypeBag struct {
	field_names      []Str
	field_types      []IrHLType
	is_union_or_enum bool
}
type IrHLTypeInt struct {
	min int64
	max int64
}
type IrHLTypeNever struct{}
type IrHLTypeTag struct{}
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
func (IrHLTypeNever) implementsIrHLType()        {}
func (IrHLTypeTag) implementsIrHLType()          {}

func (IrHLExprType) implementsIrHLExprVariant()        {}
func (IrHLExprTag) implementsIrHLExprVariant()         {}
func (IrHLExprInt) implementsIrHLExprVariant()         {}
func (IrHLExprStr) implementsIrHLExprVariant()         {}
func (IrHLExprIdent) implementsIrHLExprVariant()       {}
func (IrHLExprRefDef) implementsIrHLExprVariant()      {}
func (IrHLExprRefArg) implementsIrHLExprVariant()      {}
func (IrHLExprCall) implementsIrHLExprVariant()        {}
func (IrHLExprList) implementsIrHLExprVariant()        {}
func (IrHLExprBag) implementsIrHLExprVariant()         {}
func (IrHLExprInfix) implementsIrHLExprVariant()       {}
func (IrHLExprFunc) implementsIrHLExprVariant()        {}
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
	return ctx.ir
}

func irHLTopDefsFromAstTopDef(ctx *CtxIrHLFromAst, top_def *AstDef) {
	old_def := ctx.cur_def
	ctx.cur_def = top_def

	def := &ctx.ir.defs[ctx.num_defs]
	ctx.num_defs++
	def.anns.origin_def = top_def
	def.anns.name = top_def.anns.name
	def.body = irHlExprFrom(ctx, &top_def.body)

	ctx.cur_def = old_def
}

func irHlExprFrom(ctx *CtxIrHLFromAst, expr *AstExpr) (ret_expr IrHLExpr) {
	ret_expr.anns.origin_def, ret_expr.anns.origin_expr = ctx.cur_def, expr
	switch it := expr.variant.(type) {
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
	default:
		panic(it)
	}
	panic("newly introduced bug: should be unreachable here")
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
	case IrHLExprStr:
		print('"')
		print(string(it))
		print('"')
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
		if it.is_union_or_enum {
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
	case IrHLTypeNever:
		print("/Never")
	default:
		panic(it)
	}
}
