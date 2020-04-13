package main

type LLType interface{ implementsLLType() }
type LLExpr interface{ implementsLLExpr() }
type LLInstr interface{ implementsLLInstr() }

type LLModule struct {
	target_datalayout Str
	target_triple     Str
	globals           []LLGlobal
	funcs             []LLFunc
	anns              struct {
		global_names []Str
	}
}

type LLGlobal struct {
	name        Str
	constant    bool
	external    bool
	ty          LLType
	initializer LLExpr
	anns        struct {
		orig_ir_def *IrDef
		idx         int
	}
}

type LLFunc struct {
	external     bool
	ty           LLType
	name         Str
	params       []LLFuncParam
	basic_blocks []LLBasicBlock
	anns         struct {
		orig_ir_def             *IrDef
		local_temporaries_names []Str
	}
}

type LLFuncParam struct {
	name Str
	ty   LLType
}

type LLBasicBlock struct {
	name   Str
	instrs []LLInstr
}

type LLInstrLet struct {
	name  Str
	instr LLInstr
}

type LLInstrRet struct {
	expr LLExprTyped
}

type LLInstrUnreachable struct{}

type LLInstrSwitch struct {
	comparee           LLExprTyped
	default_block_name Str
	cases              []LLSwitchCase
}

type LLSwitchCase struct {
	expr       LLExprTyped
	block_name Str
}

type LLInstrBrTo struct {
	block_name Str
}

type LLInstrBrIf struct {
	cond                LLExpr
	block_name_if_true  Str
	block_name_if_false Str
}

type LLInstrConvert struct {
	ty           LLType
	expr         LLExprTyped
	convert_kind LLConvertKind
}

type LLConvertKind int

const (
	_ LLConvertKind = iota
	ll_convert_int_to_ptr
	ll_convert_ptr_to_int
	ll_convert_trunc
)

type LLInstrComment struct {
	comment_text Str
}

type LLInstrAlloca struct {
	ty        LLType
	num_elems LLExprTyped
}

type LLInstrLoad struct {
	ty   LLType
	expr LLExprTyped
}

type LLInstrStore struct {
	dst  LLExprTyped
	expr LLExprTyped
}

type LLInstrCall struct {
	ty     LLType
	callee LLExpr
	args   []LLExprTyped
}

type LLInstrBinOp struct {
	ty      LLType
	lhs     LLExpr
	rhs     LLExpr
	op_kind LLBinOpKind
}

type LLBinOpKind int

const (
	_ LLBinOpKind = iota
	ll_bin_op_add
	ll_bin_op_mul
	ll_bin_op_sub
	ll_bin_op_udiv
)

type LLInstrCmpI struct {
	ty       LLType
	lhs      LLExpr
	rhs      LLExpr
	cmp_kind LLCmpIKind
}

type LLCmpIKind int

const (
	_ LLCmpIKind = iota
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

type LLInstrPhi struct {
	ty           LLType
	predecessors []LLPhiPred
}

type LLPhiPred struct {
	block_name Str
	expr       LLExpr
}

type LLInstrGep struct {
	ty       LLType
	base_ptr LLExprTyped
	indices  []LLExprTyped
}

type LLExprIdentLocal Str

type LLExprIdentGlobal Str

type LLExprLitInt int64

type LLExprLitStr Str

type LLExprLitVoid struct{}

type LLExprTyped struct {
	ty   LLType
	expr LLExpr
}

type LLTypeInt struct {
	bit_width uint64 // u23 really.. we save us some casts here
}

type LLTypeVoid struct{}

type LLTypeAuto struct{}

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

type LLTypeFunc struct {
	ty     LLType
	params []LLType
}

func llTopLevelNameFrom(ll_mod *LLModule, ir_def *IrDef, num_globals int, num_funcs int) LLExprIdentGlobal {
	for i := range ll_mod.globals[0:num_globals] {
		if ll_mod.globals[i].anns.orig_ir_def == ir_def {
			return LLExprIdentGlobal(ll_mod.globals[i].name)
		}
	}
	for i := range ll_mod.funcs[0:num_funcs] {
		if ll_mod.funcs[i].anns.orig_ir_def == ir_def {
			return LLExprIdentGlobal(ll_mod.funcs[i].name)
		}
	}
	return nil
}

func llTopLevelNameFind(ll_mod *LLModule, name Str) Any {
	for i := range ll_mod.globals {
		if strEql(name, ll_mod.globals[i].name) {
			return &ll_mod.globals[i]
		}
	}
	for i := range ll_mod.funcs {
		if strEql(name, ll_mod.funcs[i].name) {
			return &ll_mod.funcs[i]
		}
	}
	return nil
}

func llExprToTyped(expr LLExpr, ty_fallback LLType) LLExprTyped {
	if expr_typed, is := expr.(LLExprTyped); is {
		return expr_typed
	}
	return LLExprTyped{ty: ty_fallback, expr: expr}
}

func llPopulateAutoTypes(ll_mod *LLModule) {
	for i := range ll_mod.globals {
		llPopulateAutoTypesInGlobal(ll_mod, &ll_mod.globals[i])
	}
	for i := range ll_mod.funcs {
		llPopulateAutoTypesInFunc(ll_mod, &ll_mod.funcs[i])
	}
}

func llPopulateAutoTypesInGlobal(ll_mod *LLModule, ll_global *LLGlobal) {
	if !llTypeIsOrHasAuto(ll_global.ty) {
		return
	}
	switch initer := ll_global.initializer.(type) {
	case LLExprLitStr:
		ll_global.ty = LLTypeArr{size: len(initer), ty: LLTypeInt{bit_width: 8}}
	default:
		panic(initer)
	}
}

func llPopulateAutoTypesInFunc(ll_mod *LLModule, ll_func *LLFunc) {
	for i := range ll_func.basic_blocks {
		for j, instr := range ll_func.basic_blocks[i].instrs {
			ll_func.basic_blocks[i].instrs[j] =
				llPopulateAutoTypesInInstr(ll_mod, ll_func, i, instr)
		}
	}
}

func llPopulateAutoTypesInInstr(ll_mod *LLModule, ll_func *LLFunc, idx_block int, ll_instr LLInstr) LLInstr {
	switch instr := ll_instr.(type) {
	case LLInstrAlloca:
		if llTypeIsAuto(instr.num_elems.ty) {
			instr.num_elems.ty = LLTypeInt{bit_width: ll_target_word_bit_width}
		}
		instr.num_elems.expr = llExprUnAutoTyped(ll_mod, ll_func, instr.num_elems.expr)
		ll_instr = instr
	case LLInstrCall:
		callee, is_ident_global := instr.callee.(LLExprIdentGlobal)
		if is_ident_global {
			switch found := llTopLevelNameFind(ll_mod, callee).(type) {
			case *LLFunc:
				if llTypeIsAuto(instr.ty) {
					instr.ty = found.ty
				}
				for i, arg := range instr.args {
					if llTypeIsAuto(arg.ty) {
						arg.ty = found.params[i].ty
						instr.args[i] = arg
					}
				}
			}
		}
		instr.callee = llExprUnAutoTyped(ll_mod, ll_func, instr.callee)
		for i, arg := range instr.args {
			instr.args[i] = llExprUnAutoTyped(ll_mod, ll_func, arg).(LLExprTyped)
		}
		ll_instr = instr
	case LLInstrBinOp:
		if llTypeIsAuto(instr.ty) {
			ty := llExprTypeIfIdentLocalReferringToFuncParam(ll_func, instr.lhs)
			if ty == nil {
				ty = llExprTypeIfIdentLocalReferringToFuncParam(ll_func, instr.rhs)
			}
			if ty != nil {
				instr.ty = ty
			}
		}
		instr.lhs = llExprUnAutoTyped(ll_mod, ll_func, instr.lhs)
		instr.rhs = llExprUnAutoTyped(ll_mod, ll_func, instr.rhs)
		ll_instr = instr
	case LLInstrCmpI:
		if llTypeIsAuto(instr.ty) {
			ty := llExprTypeIfIdentLocalReferringToFuncParam(ll_func, instr.lhs)
			if ty == nil {
				ty = llExprTypeIfIdentLocalReferringToFuncParam(ll_func, instr.rhs)
			}
			if ty != nil {
				instr.ty = ty
			}
		}
		instr.lhs = llExprUnAutoTyped(ll_mod, ll_func, instr.lhs)
		instr.rhs = llExprUnAutoTyped(ll_mod, ll_func, instr.rhs)
		ll_instr = instr
	case LLInstrGep:
		instr.base_ptr = llExprUnAutoTyped(ll_mod, ll_func, instr.base_ptr).(LLExprTyped)
		for i, index := range instr.indices {
			instr.indices[i] = llExprUnAutoTyped(ll_mod, ll_func, index).(LLExprTyped)
		}
		ll_instr = instr
	case LLInstrLoad:
		instr.expr = llExprUnAutoTyped(ll_mod, ll_func, instr.expr).(LLExprTyped)
		ll_instr = instr
	case LLInstrStore:
		instr.dst = llExprUnAutoTyped(ll_mod, ll_func, instr.dst).(LLExprTyped)
		instr.expr = llExprUnAutoTyped(ll_mod, ll_func, instr.expr).(LLExprTyped)
		ll_instr = instr
	case LLInstrPhi:
		for i := range instr.predecessors {
			instr.predecessors[i].expr = llExprUnAutoTyped(ll_mod, ll_func, instr.predecessors[i].expr)
		}
		ll_instr = instr
	case LLInstrBrIf:
		instr.cond = llExprUnAutoTyped(ll_mod, ll_func, instr.cond)
		ll_instr = instr
	case LLInstrLet:
		instr.instr = llPopulateAutoTypesInInstr(ll_mod, ll_func, idx_block, instr.instr)
		ll_instr = instr
	case LLInstrRet:
		instr.expr.expr = llExprUnAutoTyped(ll_mod, ll_func, instr.expr.expr)
		if llTypeIsAuto(instr.expr.ty) {
			instr.expr.ty = ll_func.ty
		}
		ll_instr = instr
	case LLInstrConvert:
		instr.expr = llExprUnAutoTyped(ll_mod, ll_func, instr.expr).(LLExprTyped)
		ll_instr = instr
	case LLInstrSwitch:
		instr.comparee = llExprUnAutoTyped(ll_mod, ll_func, instr.comparee).(LLExprTyped)
		for i := range instr.cases {
			instr.cases[i].expr = llExprUnAutoTyped(ll_mod, ll_func, instr.cases[i].expr).(LLExprTyped)
		}
		ll_instr = instr
	case LLInstrBrTo, LLInstrComment, LLInstrUnreachable:
	default:
		panic(instr)
	}
	return ll_instr
}

func llExprTypeIfIdentLocalReferringToFuncParam(ll_func *LLFunc, ll_expr LLExpr) LLType {
	if ident, _ := ll_expr.(LLExprIdentLocal); ident != nil {
		for i := range ll_func.params {
			if strEql(ident, ll_func.params[i].name) {
				return ll_func.params[i].ty
			}
		}
	}
	return nil
}

func llExprUnAutoTyped(ll_mod *LLModule, ll_func *LLFunc, ll_expr LLExpr) LLExpr {
	// switch expr:=ll_expr
	return ll_expr
}

func llTypeIsAuto(ll_type LLType) bool {
	_, is := ll_type.(LLTypeAuto)
	return is
}

func llTypeIsOrHasAuto(ll_type LLType) bool {
	switch ty := ll_type.(type) {
	case LLTypeVoid, LLTypeInt:
		return false
	case LLTypeAuto:
		return true
	case LLTypeArr:
		return llTypeIsOrHasAuto(ty.ty)
	case LLTypeFunc:
		for i := range ty.params {
			if llTypeIsOrHasAuto(ty.params[i]) {
				return true
			}
		}
		return llTypeIsOrHasAuto(ty.ty)
	case LLTypePtr:
		return llTypeIsOrHasAuto(ty.ty)
	case LLTypeStruct:
		for i := range ty.fields {
			if llTypeIsOrHasAuto(ty.fields[i]) {
				return true
			}
		}
	default:
		panic(ty)
	}
	return false
}

func llTypeEql(ty1 LLType, ty2 LLType) bool {
	assert(ty1 != nil && ty2 != nil)
	switch tl := ty1.(type) {
	case LLTypeVoid:
		_, ok := ty2.(LLTypeVoid)
		return ok
	case LLTypeInt:
		if tr, ok := ty2.(LLTypeInt); ok {
			return tl.bit_width == tr.bit_width
		}
	case LLTypePtr:
		if tr, ok := ty2.(LLTypePtr); ok {
			return llTypeEql(tl.ty, tr.ty)
		}
	case LLTypeArr:
		if tr, ok := ty2.(LLTypeArr); ok {
			return tl.size == tr.size && llTypeEql(tl.ty, tr.ty)
		}
	case LLTypeStruct:
		if tr, ok := ty2.(LLTypeStruct); ok && len(tl.fields) == len(tr.fields) {
			for i, tl_field_ty := range tl.fields {
				if !llTypeEql(tl_field_ty, tr.fields[i]) {
					return false
				}
			}
			return true
		}
	case LLTypeFunc:
		if tr, ok := ty2.(LLTypeFunc); ok && len(tl.params) == len(tr.params) && llTypeEql(tl.ty, tr.ty) {
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

func (LLTypeArr) implementsLLType()    {}
func (LLTypeFunc) implementsLLType()   {}
func (LLTypeInt) implementsLLType()    {}
func (LLTypePtr) implementsLLType()    {}
func (LLTypeStruct) implementsLLType() {}
func (LLTypeVoid) implementsLLType()   {}
func (LLTypeAuto) implementsLLType()   {}

func (LLExprIdentGlobal) implementsLLExpr() {}
func (LLExprIdentLocal) implementsLLExpr()  {}
func (LLExprLitInt) implementsLLExpr()      {}
func (LLExprLitStr) implementsLLExpr()      {}
func (LLExprLitVoid) implementsLLExpr()     {}
func (LLExprTyped) implementsLLExpr()       {}

func (LLInstrAlloca) implementsLLInstr()      {}
func (LLInstrBinOp) implementsLLInstr()       {}
func (LLInstrCall) implementsLLInstr()        {}
func (LLInstrCmpI) implementsLLInstr()        {}
func (LLInstrGep) implementsLLInstr()         {}
func (LLInstrLoad) implementsLLInstr()        {}
func (LLInstrStore) implementsLLInstr()       {}
func (LLInstrPhi) implementsLLInstr()         {}
func (LLInstrBrIf) implementsLLInstr()        {}
func (LLInstrBrTo) implementsLLInstr()        {}
func (LLInstrComment) implementsLLInstr()     {}
func (LLInstrLet) implementsLLInstr()         {}
func (LLInstrRet) implementsLLInstr()         {}
func (LLInstrSwitch) implementsLLInstr()      {}
func (LLInstrUnreachable) implementsLLInstr() {}
func (LLInstrConvert) implementsLLInstr()     {}

// instrs that can be LLVM-IR "constant-expressions":

func (LLInstrConvert) implementsLLExpr() {}
func (LLInstrGep) implementsLLExpr()     {}
func (LLInstrCmpI) implementsLLExpr()    {}
func (LLInstrBinOp) implementsLLExpr()   {}
