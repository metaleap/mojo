package main

type LLModule struct {
	target_datalayout Str
	target_triple     Str
	globals           []LLGlobal
	funcs             []LLFunc
}

type LLType interface{} // just for doc purposes: to remain empty with no methods
type LLExpr interface{} // just for doc purposes: to remain empty with no methods
type LLStmt interface{} // just for doc purposes: to remain empty with no methods

type LLGlobal struct {
	name        Str
	constant    bool
	external    bool
	ty          LLType
	initializer LLExpr
}

type LLFunc struct {
	external     bool
	ty           LLType
	name         Str
	params       []LLFuncParam
	basic_blocks []LLBasicBlock
}

type LLFuncParam struct {
	name Str
	ty   LLType
}

type LLBasicBlock struct {
	name  Str
	stmts []LLStmt
}

type LLStmtLet struct {
	name Str
	expr LLExpr
}

type LLStmtRet struct {
	expr LLExprTyped
}

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

type LLStmtComment struct {
	comment_text Str
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
	callee LLExprTyped
	args   []LLExprTyped
}

type LLExprBinOp struct {
	ty      LLType
	lhs     LLExpr
	rhs     LLExpr
	op_kind LLExprBinOpKind
}

type LLExprBinOpKind int

const (
	_ LLExprBinOpKind = iota
	ll_bin_op_add
)

type LLExprCmpI struct {
	ty       LLType
	lhs      LLExpr
	rhs      LLExpr
	cmp_kind LLExprCmpIKind
}

type LLExprCmpIKind int

const (
	_ LLExprCmpIKind = iota
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

type LLExprPhi struct {
	ty           LLType
	predecessors []struct {
		expr       LLExpr
		block_name Str
	}
}

type LLExprGep struct {
	ty       LLType
	base_ptr LLExprTyped
	indices  []LLExprTyped
}

type LLTypeInt struct {
	bit_width uint32 // u23 really..
}

type LLTypeVoid struct{}

type LLTypePtr struct {
	ty LLType
}

type LLTypeArr struct {
	ty   LLType
	size int
}

type LLTypeStruct struct {
	fields []LLType
}

type LLTypeFun struct {
	ty     LLType
	params []LLType
}
