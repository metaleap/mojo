package main

const (
	llTrue     = "true"
	llFalse    = "false"
	llNull     = "null"
	llZeroInit = "zeroinitializer"
	llUndef    = "undef"
	llPoison   = "poison"
)

type LlModule struct {
	source_filename string

	Decls []LlExtDecl
	Funcs []LlFuncDef
	Vars  []LlGlobalVar
}

type LlNamed struct {
	name string
}

type LlGlobalVar struct {
	LlNamed
	constant bool
	init     LlExpr
	ty       LlType
}

type LlFuncDef struct {
	LlNamed
	ty     LlTypeFunc
	blocks []LlBlock
}

type LlExtDecl struct {
	LlNamed
	ty LlTypeFunc
}

type LlParam struct {
	LlNamed
	ty LlType
}

type LlType interface{}

type LlTypeFunc struct {
	ret  LlParam
	args []LlParam
}

type LlTypeInt struct {
	bitWidth int
}

type LlTypeFloat struct {
	bitWidth int
}

type LlTypePtr struct {
	ty LlType
}

type LlTypeArr struct {
	numElems int
	ty       LlType
}

type LlTypeStruct struct {
	fields []LlType
}

type LlBlock struct {
	LlNamed
	instrs []LlInstr
}

type LlInstr interface{}

type LlInstrRet struct {
	expr LlExpr
}

type LlInstrBr struct {
	cond    LlExpr
	ifTrue  *LlBlock
	ifFalse *LlBlock
}

type LlInstrSwitch struct {
	scrut LlExpr
	def   *LlBlock
	cases []struct {
		intConst LlExpr
		dest     *LlBlock
	}
}

type LlInstrUnreachable struct {
}

type LlInstrOp1Fneg struct {
	op1 LlExpr
}

type LlInstrOp2 struct {
	op1 LlExpr
	op2 LlExpr
}

type LlInstrOp2Wrappable struct {
	LlInstrOp2
	noUnsignedWrap bool
	noSignedWrap   bool
}

type LlInstrOp2Add LlInstrOp2Wrappable
type LlInstrOp2Sub LlInstrOp2Wrappable
type LlInstrOp2Mul LlInstrOp2Wrappable
type LlInstrOp2Fadd LlInstrOp2
type LlInstrOp2Fsub LlInstrOp2
type LlInstrOp2Fmul LlInstrOp2
type LlInstrOp2Fdiv LlInstrOp2
type LlInstrOp2Urem LlInstrOp2
type LlInstrOp2Srem LlInstrOp2
type LlInstrOp2Frem LlInstrOp2

type LlInstrOp2Exactable struct {
	LlInstrOp2
	exact bool
}

type LlInstrOp2Udiv LlInstrOp2Exactable
type LlInstrOp2Sdiv LlInstrOp2Exactable

type LlInstrOp2Shl LlInstrOp2Wrappable
type LlInstrOp2Lshr LlInstrOp2Exactable
type LlInstrOp2Ashr LlInstrOp2Exactable

type LlInstrOp2And LlInstrOp2
type LlInstrOp2Or LlInstrOp2
type LlInstrOp2Xor LlInstrOp2

type LlInstrExtractValue struct {
	aggr LlExpr
	idxs []LlExpr
}

type LlInstrInsertValue struct {
	aggr LlExpr
	idxs []LlExpr
	elem LlExpr
}

type LlInstrLoad struct {
	ty     LlType
	ptrSrc LlExpr
}

type LlInstrStore struct {
	LlExpr
	ptrDst LlExpr
}

type LlInstrGetElementPtr struct {
	ty       LlTypeFunc
	ptrSrc   LlExpr
	inbounds bool
	idxs     []LlExpr
}

type LlExpr interface {
	ty() LlType
}
