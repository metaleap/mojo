package main

type IrHL struct {
	defs []IrHLDef
	anns struct {
		origin_ast *Ast
	}
}

type IrHLDef struct {
	args []IrHLType
	body IrHLExpr
	anns struct {
		origin_def *AstDef
	}
}

type IrHLExpr struct {
	kind IrHLExprKind
	anns struct {
		origin_expr *AstExpr
		ty          IrHLType
	}
}

type IrHLExprKind interface{ implementsIrHLExprKind() }
type IrHLType interface{ implementsIrHLType() }
type IrHLTypePrim interface{ implementsIrHLTypePrim() }

type IrHLExprTag Str
type IrHLExprInt int64
type IrHLExprStr Str
type IrHLExprIdent Str
type IrHLExprRefDef int
type IrHLExprRefArg int
type IrHLExprPrim []IrHLExpr
type IrHLExprForm []IrHLExpr
type IrHLExprList []IrHLExpr
type IrHLExprBag []IrHLExpr
type IrHLExprInfix struct {
	kind Str
	lhs  IrHLExpr
	rhs  IrHLExpr
}

type IrHLTypePrimInt struct{ bit_width int }
type IrHLTypePrimArr struct{ payload IrHLType }
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
type IrHLTypeVoid struct{}
type IrHLTypeTag struct{}

func (IrHLTypePrimArr) implementsIrHLTypePrim()    {}
func (IrHLTypePrimFunc) implementsIrHLTypePrim()   {}
func (IrHLTypePrimInt) implementsIrHLTypePrim()    {}
func (IrHLTypePrimOpaque) implementsIrHLTypePrim() {}
func (IrHLTypePrimPtr) implementsIrHLTypePrim()    {}
func (IrHLTypePrimStruct) implementsIrHLTypePrim() {}

func (IrHLTypePrimArr) implementsIrHLType()    {}
func (IrHLTypePrimFunc) implementsIrHLType()   {}
func (IrHLTypePrimInt) implementsIrHLType()    {}
func (IrHLTypePrimOpaque) implementsIrHLType() {}
func (IrHLTypePrimPtr) implementsIrHLType()    {}
func (IrHLTypePrimStruct) implementsIrHLType() {}

func (IrHLTypeBag) implementsIrHLType()   {}
func (IrHLTypeInt) implementsIrHLType()   {}
func (IrHLTypeNever) implementsIrHLType() {}
func (IrHLTypeVoid) implementsIrHLType()  {}
func (IrHLTypeTag) implementsIrHLType()   {}

func (IrHLExprTag) implementsIrHLExprKind()    {}
func (IrHLExprInt) implementsIrHLExprKind()    {}
func (IrHLExprStr) implementsIrHLExprKind()    {}
func (IrHLExprIdent) implementsIrHLExprKind()  {}
func (IrHLExprRefDef) implementsIrHLExprKind() {}
func (IrHLExprRefArg) implementsIrHLExprKind() {}
func (IrHLExprPrim) implementsIrHLExprKind()   {}
func (IrHLExprForm) implementsIrHLExprKind()   {}
func (IrHLExprList) implementsIrHLExprKind()   {}
func (IrHLExprBag) implementsIrHLExprKind()    {}
func (IrHLExprInfix) implementsIrHLExprKind()  {}

func irHLFrom(ast *Ast) IrHL {
	astHoistSubDefs(ast)
	num_defs := 0
	ret_ir := IrHL{
		defs: ªIrHLDef(ast.anns.num_def_toks),
	}
	ret_ir.anns.origin_ast = ast
	ret_ir.defs = ret_ir.defs[0:num_defs]
	for i := range ast.defs {
		irHLTopDefsFromAstTopDef(&ret_ir, ast, &ast.defs[i])
	}
	return ret_ir
}

func irHLTopDefsFromAstTopDef(dst_ir *IrHL, ast *Ast, top_def *AstDef) {

}

func irHLDump(ir *IrHL) {
	for i := range ir.defs {
		irHLDumpDef(ir, &ir.defs[i])
	}
}

func irHLDumpDef(ir *IrHL, def *IrHLDef) {}

func irHLDumpExpr(ir *IrHL, expr *IrHLExpr) {}
