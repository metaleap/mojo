package main

type IrLLProg struct {
	defs []IrLLDef
	ext  struct {
		vars  Any
		funcs Any
	}
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
	kind IrLLExprKind
	anns struct {
		origin *IrHLExpr
		ty     IrLLType
	}
}

type IrLLType interface{ implementsIrLLType() }
type IrLLExprKind interface{ implementsIrLLExprKind() }

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

type IrLLOpIntKind int

const (
	_ IrLLOpIntKind = iota
	ll_bin_op_add
	ll_bin_op_mul
	ll_bin_op_sub
	ll_bin_op_udiv
	ll_bin_op_sdiv
	ll_bin_op_urem
	ll_bin_op_srem
)

type IrLLExprCmpInt struct {
	kind IrLLCmpIntKind
	lhs  IrLLExpr
	rhs  IrLLExpr
}

type IrLLCmpIntKind int

const (
	_ IrLLCmpIntKind = iota
	ll_cmp_i_eq
	ll_cmp_i_ne
	ll_cmp_i_ugt
	ll_cmp_i_uge
	ll_cmp_i_ult
	ll_cmp_i_ule
	ll_cmp_i_sgt
	ll_cmp_i_sge
	ll_cmp_i_slt
	ll_cmp_i_sle
)

func (IrLLTypeInt) implementsIrLLType()    {}
func (IrLLTypeArr) implementsIrLLType()    {}
func (IrLLTypePtr) implementsIrLLType()    {}
func (IrLLTypeStruct) implementsIrLLType() {}
func (IrLLTypeVoid) implementsIrLLType()   {}
func (IrLLTypeFunc) implementsIrLLType()   {}
func (IrLLTypeExtern) implementsIrLLType() {}

func (IrLLExprInt) implementsIrLLExprKind()    {}
func (IrLLExprPtr) implementsIrLLExprKind()    {}
func (IrLLExprArgRef) implementsIrLLExprKind() {}
func (IrLLExprCall) implementsIrLLExprKind()   {}
func (IrLLExprSelect) implementsIrLLExprKind() {}
func (IrLLExprOpInt) implementsIrLLExprKind()  {}
func (IrLLExprCmpInt) implementsIrLLExprKind() {}
