package main

type IrAst struct {
	defs []IrDef
	anns struct {
		origin_ast *Ast
	}
}

type IrDef struct {
	args []IrType
	body IrExpr
	anns struct {
		origin_def *AstDef
	}
}

type IrExpr struct {
	kind IrExprKind
	anns struct {
		origin_expr *AstExpr
		ty          IrType
	}
}

type IrExprKind interface{ implementsIrExprKind() }
type IrType interface{ implementsIrType() }
type IrTypePrim interface{ implementsIrTypePrim() }

type IrExprTag Str
type IrExprInt int64
type IrExprStr Str
type IrExprIdent Str
type IrExprRefDef int
type IrExprRefArg int
type IrExprPrim []IrExpr
type IrExprForm []IrExpr
type IrExprList []IrExpr
type IrExprBag []IrExpr
type IrExprInfix struct {
	kind Str
	lhs  IrExpr
	rhs  IrExpr
}

type IrTypePrimInt struct{ bit_width int }
type IrTypePrimArr struct{ payload IrType }
type IrTypePrimPtr struct{ payload IrType }
type IrTypePrimStruct struct{ fields []IrType }
type IrTypePrimFunc struct {
	returns IrType
	params  []IrType
}
type IrTypePrimOpaque Str

type IrTypeBag struct {
	field_names      []Str
	field_types      []IrType
	is_union_or_enum bool
}
type IrTypeInt struct {
	min int64
	max int64
}
type IrTypeNever struct{}
type IrTypeVoid struct{}
type IrTypeTag struct{}

func (IrTypePrimArr) implementsIrTypePrim()    {}
func (IrTypePrimFunc) implementsIrTypePrim()   {}
func (IrTypePrimInt) implementsIrTypePrim()    {}
func (IrTypePrimOpaque) implementsIrTypePrim() {}
func (IrTypePrimPtr) implementsIrTypePrim()    {}
func (IrTypePrimStruct) implementsIrTypePrim() {}

func (IrTypePrimArr) implementsIrType()    {}
func (IrTypePrimFunc) implementsIrType()   {}
func (IrTypePrimInt) implementsIrType()    {}
func (IrTypePrimOpaque) implementsIrType() {}
func (IrTypePrimPtr) implementsIrType()    {}
func (IrTypePrimStruct) implementsIrType() {}

func (IrTypeBag) implementsIrType()   {}
func (IrTypeInt) implementsIrType()   {}
func (IrTypeNever) implementsIrType() {}
func (IrTypeVoid) implementsIrType()  {}
func (IrTypeTag) implementsIrType()   {}

func (IrExprTag) implementsIrExprKind()    {}
func (IrExprInt) implementsIrExprKind()    {}
func (IrExprStr) implementsIrExprKind()    {}
func (IrExprIdent) implementsIrExprKind()  {}
func (IrExprRefDef) implementsIrExprKind() {}
func (IrExprRefArg) implementsIrExprKind() {}
func (IrExprPrim) implementsIrExprKind()   {}
func (IrExprForm) implementsIrExprKind()   {}
func (IrExprList) implementsIrExprKind()   {}
func (IrExprBag) implementsIrExprKind()    {}
func (IrExprInfix) implementsIrExprKind()  {}
