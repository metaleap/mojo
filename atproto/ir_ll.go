package main

type IrLLProg struct {
	defs []IrLLDef
	anns struct {
		origin *IrHL
	}
}

type IrLLDef struct {
	args []IrLLType
	body IrLLExpr
	anns struct {
		origin *IrHLDef
	}
}

type IrLLExpr struct {
	variant IrLLExprVariant
	anns    struct {
		origin *IrHLExpr
		ty     IrLLType
	}
}

type IrLLType interface{ implementsIrLLType() }
type IrLLExprVariant interface{ implementsIrLLExprVariant() }

type IrLLTypeInt struct{ bit_width int }
type IrLLTypeArr struct {
	size    int
	payload IrLLType
}
type IrLLTypePtr struct{ payload IrLLType }
type IrLLTypeStruct struct{ fields []IrLLType }
type IrLLTypeVoid struct{}
type IrLLTypeFunc struct {
	returns IrLLType
	params  []IrLLType
}
type IrLLTypeExtern Str

type IrLLExprInt int64

type IrLLExprPtr uintptr

type IrLLExprArgRef int

type IrLLExprCall struct {
	callee int
	args   []IrLLExpr
	is_ext bool
	is_ptr bool
}

type IrLLExprSelect struct {
	cond     IrLLExpr
	if_true  IrLLExpr
	if_false IrLLExpr
}

type IrLLExprOpInt struct {
	kind IrLLOpIntKind
	lhs  IrLLExpr
	rhs  IrLLExpr
}

type IrLLOpBoolKind int

const (
	_ IrLLOpBoolKind = iota
	ll_bool_op_and
	ll_bool_op_or
)

type IrLLOpIntKind int

const (
	_ IrLLOpIntKind = iota
	ll_iop_add
	ll_iop_mul
	ll_iop_sub
	ll_iop_udiv
	ll_iop_sdiv
	ll_iop_urem
	ll_iop_srem
)

type IrLLExprCmpInt struct {
	kind IrLLCmpIntKind
	lhs  IrLLExpr
	rhs  IrLLExpr
}

type IrLLCmpIntKind int

const (
	_ IrLLCmpIntKind = iota
	ll_icmp_eq
	ll_icmp_ne
	ll_icmp_ugt
	ll_icmp_uge
	ll_icmp_ult
	ll_icmp_ule
	ll_icmp_sgt
	ll_icmp_sge
	ll_icmp_slt
	ll_icmp_sle
)

func (IrLLTypeInt) implementsIrLLType()    {}
func (IrLLTypeArr) implementsIrLLType()    {}
func (IrLLTypePtr) implementsIrLLType()    {}
func (IrLLTypeStruct) implementsIrLLType() {}
func (IrLLTypeVoid) implementsIrLLType()   {}
func (IrLLTypeFunc) implementsIrLLType()   {}
func (IrLLTypeExtern) implementsIrLLType() {}

func (IrLLExprInt) implementsIrLLExprVariant()    {}
func (IrLLExprPtr) implementsIrLLExprVariant()    {}
func (IrLLExprArgRef) implementsIrLLExprVariant() {}
func (IrLLExprCall) implementsIrLLExprVariant()   {}
func (IrLLExprSelect) implementsIrLLExprVariant() {}
func (IrLLExprOpInt) implementsIrLLExprVariant()  {}
func (IrLLExprCmpInt) implementsIrLLExprVariant() {}
