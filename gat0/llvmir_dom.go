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
	ty   LlType
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

type LlExpr interface{}
