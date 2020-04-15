package main

type IrHL struct {
	defs []IrHLDef
	anns struct {
		origin_ast *Ast
	}
}

type IrHLDef struct {
	params []IrHLType
	body   IrHLExpr
	anns   struct {
		origin_def *AstDef
		name       Str
	}
}

type IrHLExpr struct {
	variant IrHLExprVariant
	anns    struct {
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
type IrHLExprInfix struct {
	kind Str
	lhs  IrHLExpr
	rhs  IrHLExpr
}
type IrHLExprType struct{ ty IrHLType }

type IrHLTypePrimInt struct{ bit_width int }
type IrHLTypePrimArr struct {
	size    int
	payload IrHLType
}
type IrHLTypePrimPtr struct{ payload IrHLType }
type IrHLTypePrimStruct struct{ fields []IrHLType }
type IrHLTypePrimFunc struct {
	returns IrHLType
	params  []IrHLType
}
type IrHLTypePrimOpaque Str

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
type IrHLTypePrimVoid struct{}
type IrHLTypeTag struct{}

func (IrHLTypePrimArr) implementsIrHLTypePrim()    {}
func (IrHLTypePrimFunc) implementsIrHLTypePrim()   {}
func (IrHLTypePrimInt) implementsIrHLTypePrim()    {}
func (IrHLTypePrimOpaque) implementsIrHLTypePrim() {}
func (IrHLTypePrimPtr) implementsIrHLTypePrim()    {}
func (IrHLTypePrimStruct) implementsIrHLTypePrim() {}
func (IrHLTypePrimVoid) implementsIrHLTypePrim()   {}

func (IrHLTypePrimArr) implementsIrHLType()    {}
func (IrHLTypePrimFunc) implementsIrHLType()   {}
func (IrHLTypePrimInt) implementsIrHLType()    {}
func (IrHLTypePrimOpaque) implementsIrHLType() {}
func (IrHLTypePrimPtr) implementsIrHLType()    {}
func (IrHLTypePrimStruct) implementsIrHLType() {}
func (IrHLTypePrimVoid) implementsIrHLType()   {}
func (IrHLTypeBag) implementsIrHLType()        {}
func (IrHLTypeInt) implementsIrHLType()        {}
func (IrHLTypeNever) implementsIrHLType()      {}
func (IrHLTypeTag) implementsIrHLType()        {}

func (IrHLExprType) implementsIrHLExprVariant()   {}
func (IrHLExprTag) implementsIrHLExprVariant()    {}
func (IrHLExprInt) implementsIrHLExprVariant()    {}
func (IrHLExprStr) implementsIrHLExprVariant()    {}
func (IrHLExprIdent) implementsIrHLExprVariant()  {}
func (IrHLExprRefDef) implementsIrHLExprVariant() {}
func (IrHLExprRefArg) implementsIrHLExprVariant() {}
func (IrHLExprCall) implementsIrHLExprVariant()   {}
func (IrHLExprList) implementsIrHLExprVariant()   {}
func (IrHLExprBag) implementsIrHLExprVariant()    {}
func (IrHLExprInfix) implementsIrHLExprVariant()  {}

type CtxIrHLFromAst struct {
	ir       IrHL
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

}

func irHLDump(ir *IrHL) {
	for i := range ir.defs {
		irHLDumpDef(ir, &ir.defs[i])
	}
}

func irHLDumpDef(ir *IrHL, def *IrHLDef) {
	print(string(def.anns.name))
	for i := range def.params {
		print(' ')
		irHLDumpType(ir, def.params[i])
	}
	print(" :=\n    ")
	irHLDumpExpr(ir, &def.body)
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
	case IrHLTypePrimOpaque:
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
