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
	predecessors []LLPhiPred
}

type LLPhiPred struct {
	expr       LLExpr
	block_name Str
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
	size uint64
}

type LLTypeStruct struct {
	fields []LLType
}

type LLTypeFunc struct {
	ty     LLType
	params []LLType
}

func llTypeEql(t1 LLType, t2 LLType) bool {
	assert(t1 != nil && t2 != nil)
	switch tl := t1.(type) {
	case LLTypeVoid:
		_, ok := t2.(LLTypeVoid)
		return ok
	case LLTypeInt:
		if tr, ok := t2.(LLTypeInt); ok {
			return tl.bit_width == tr.bit_width
		}
	case LLTypePtr:
		if tr, ok := t2.(LLTypePtr); ok {
			return llTypeEql(tl.ty, tr.ty)
		}
	case LLTypeArr:
		if tr, ok := t2.(LLTypeArr); ok {
			return tl.size == tr.size && llTypeEql(tl.ty, tr.ty)
		}
	case LLTypeStruct:
		if tr, ok := t2.(LLTypeStruct); ok && len(tl.fields) == len(tr.fields) {
			for i, tl_field_ty := range tl.fields {
				if !llTypeEql(tl_field_ty, tr.fields[i]) {
					return false
				}
			}
			return true
		}
	case LLTypeFunc:
		if tr, ok := t2.(LLTypeFunc); ok && len(tl.params) == len(tr.params) && llTypeEql(tl.ty, tr.ty) {
			for i, tl_param_ty := range tl.params {
				if !llTypeEql(tl_param_ty, tr.params[i]) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func llTypeToStr(ll_ty LLType) Str {
	switch t := ll_ty.(type) {
	case LLTypeVoid:
		return Str("/V")
	case LLTypeInt:
		return uintToStr(uint64(t.bit_width), 10, 1, Str("/I"))
	case LLTypePtr:
		return strConcat([]Str{Str("/P"), llTypeToStr(t.ty)})
	case LLTypeArr:
		return strConcat([]Str{uintToStr(t.size, 10, 1, Str("/A/")), llTypeToStr(t.ty)})
	case LLTypeFunc:
		strs := allocˇStr(3 + len(t.params))
		strs[0] = Str("/F")
		strs[1] = llTypeToStr(t.ty)
		strs[2] = uintToStr(uint64(len(t.params)), 10, 1, Str("/"))
		for i := range t.params {
			strs[3+i] = llTypeToStr(t.params[i])
		}
		return strConcat(strs)
	}
	panic(ll_ty)
}
