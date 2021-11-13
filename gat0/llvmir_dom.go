package main

type LlLinkage int

const (
	_ LlLinkage = iota
	private
	internal
	available_externally
	linkonce
	weak
	common
	appending
	extern_weak
	linkonce_odr
	weak_odr
	external
)

type LlCallConv int

const (
	ccc LlCallConv = iota
	fastcc
	coldcc
	preserve_mostcc
	preserve_allcc
	tailcc
)

type LlVisibility int

const (
	_ LlVisibility = iota
	target_default
	hidden
	protected
)

type LlDllStorage int

const (
	_ LlDllStorage = iota
	dllimport
	dllexport
)

type LlTlsModel int

const (
	_ LlTlsModel = iota
	localdynamic
	initialexec
	localexec
)

type LlRuntimePreemption int

const (
	dso_preemptable LlRuntimePreemption = iota
	dso_local
)

type LlModule struct {
	Decls []LlExtDecl
	Funcs []LlFuncDef
	Vars  []LlGlobalVar
}

type LlNamed struct {
	name string
}

type LlGlobalVar struct {
	LlNamed
	constant               bool
	init                   *LlExpr
	linkage                LlLinkage
	visibility             LlVisibility
	unnamed_addr           bool
	local_unnamed_addr     bool
	externally_initialized bool
	align                  int
	addrSpace              int
	dllStorage             LlDllStorage
	runtimePreemption      LlRuntimePreemption
	attrs                  []LlAttr
	meta                   []LlMeta
	tlsModel               LlTlsModel
	ty                     LlType
	section                string
	comdat                 string
}

type LlFuncDef struct {
	LlNamed
	linkage           LlLinkage
	runtimePreemption LlRuntimePreemption
	visibility        LlVisibility
	dllStorage        LlDllStorage
	callConv          LlCallConv
	unnamed_addr      bool
	ret               *LlParam
	args              []LlParam
	attrs             []LlAttr
	meta              []LlMeta
	align             int
	addrSpace         int
	section           string
	comdat            string
	blocks            []LlBlock
	gcName            string
	prefix            interface{}
	prologue          interface{}
	personality       interface{}
}

type LlExtDecl struct {
	LlNamed
	linkage            LlLinkage
	visibility         LlVisibility
	dllStorage         LlDllStorage
	callConv           LlCallConv
	unnamed_addr       bool
	local_unnamed_addr bool
	addrSpace          int
	ret                *LlParam
	args               []LlParam
	align              int
	gcName             string
	prefix             interface{}
	prologue           interface{}
}

type LlAlias struct {
	LlNamed
	ty                 LlType
	aliaseeName        string
	aliaseeTy          LlType
	linkage            LlLinkage
	runtimePreemption  LlRuntimePreemption
	visibility         LlVisibility
	dllStorage         LlDllStorage
	tlsModel           LlTlsModel
	unnamed_addr       bool
	local_unnamed_addr bool
}

type LlParam struct {
	LlNamed
	ty    LlType
	attrs []LlAttr
}

type LlType struct{}

type LlBlock struct {
	LlNamed
	instrs []LlInstr
}

type LlInstr struct{}

type LlExpr struct{}

type LlAttr struct{}

type LlMeta struct{}
