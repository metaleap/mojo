package main

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
	init     *LlExpr
	ty       LlType
}

type LlFuncDef struct {
	LlNamed
	ret    *LlParam
	args   []LlParam
	blocks []LlBlock
}

type LlExtDecl struct {
	LlNamed
	ret  *LlParam
	args []LlParam
}

type LlParam struct {
	LlNamed
	ty LlType
}

type LlType struct{}

type LlBlock struct {
	LlNamed
	instrs []LlInstr
}

type LlInstr struct{}

type LlExpr struct{}
