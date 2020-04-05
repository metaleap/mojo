package main

type LLModule struct {
	target_datalayout Str
	target_triple     Str
	globals           []LLGlobal
	funcs             []LLFunc
}

type LLType interface{} // just for doc purposes, no methods
type LLExpr interface{} // just for doc purposes, no methods
type LLStmt interface{} // just for doc purposes, no methods

type LLGlobal struct {
	name        Str
	constant    bool
	external    bool
	ty          LLType
	initializer LLExpr
}

type LLFunc struct {
	ty           LLType
	name         Str
	params       []LLFuncParam
	basic_blocks LLBasicBlock
}

type LLFuncParam struct {
	name Str
	ty   LLType
}

type LLBasicBlock struct {
	name       Str
	statements []LLStmt
}

type LLStmtLet struct {
	name Str
	expr LLExpr
}

type LLStmtRet LLExprTyped

type LLStmtSwitch struct {
	comparee           LLExprTyped
	default_block_name Str
	cases              []struct {
		expr       LLExprTyped
		block_name Str
	}
}

type LLStmtBr struct {
	block_name Str
}

type LLExprIdentLocal Str

type LLExprIdentGlobal Str

type LLExprLitInt uint64

type LLExprLitStr Str

type LLExprTyped struct {
	ty   LLType
	expr LLExpr
}

type LLExprAlloca struct {
	ty        LLType
	num_elems LLExprTyped
}

type LLExprLoad LLExprTyped

type LLExprCall struct {
	Callee LLExprTyped
	Args   []LLExprTyped
}

type LLExprBinOp struct {
	ty  LLType
	lhs LLExpr
	rhs LLExpr
	op  LLExprBinOpKind
}

type LLExprBinOpKind int

const (
	_ LLExprBinOpKind = iota
	ll_bin_op_add
	ll_bin_op_sub
)

type LLExprCmpI struct {
	ty  LLType
	lhs LLExpr
	rhs LLExpr
	cmp LLExprCmpIKind
}

type LLExprCmpIKind int

const (
	_ LLExprCmpIKind = iota
	ll_cmpi_eq
	ll_cmpi_ne
	ll_cmpi_ugt
	ll_cmpi_uge
	ll_cmpi_ult
	ll_cmpi_ule
	ll_cmpi_sgt
	ll_cmpi_sge
	ll_cmpi_slt
	ll_cmpi_sle
)

type LLExprPhi struct {
	ty    LLType
	preds []struct {
		expr       LLExpr
		block_name Str
	}
}

type LLExprGep struct {
	ty       LLType
	base_ptr LLExprTyped
	idxs     []LLExprTyped
}

type LLTypeInt struct {
	bit_width int
}

type LLTypePtr struct {
	ty LLType
}

type LLTypeArr struct {
	ty   LLType
	size int
}

type LLTypeAgg struct {
	fields []LLType
}

type LLTypeFun struct {
	ty     LLType
	params []LLType
}
