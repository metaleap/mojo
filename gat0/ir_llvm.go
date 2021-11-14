package main

func (me *LlExprLitInt) expr() (LlType, interface{})     { return me.ty, me.value }
func (me *LlExprLitFloat) expr() (LlType, interface{})   { return me.ty, me.value }
func (me *LlExprLitCStr) expr() (LlType, interface{})    { return me.ty, me.value }
func (me *LlExprLitBuiltin) expr() (LlType, interface{}) { return me.ty, me.name }
func (me *LlExprRefLocal) expr() (LlType, interface{})   { return me.ty, me.name }
func (me *LlExprRefGlobal) expr() (LlType, interface{})  { return me.ty, me.name }

func llIrSrc(llvmIr interface{}) []byte {
	buf := make([]byte, 0, 1024)
	push := func(strs ...string) {
		for _, s := range strs {
			buf = append(buf, s...)
		}
	}
	switch it := llvmIr.(type) {

	case *LlModule:
		if it.source_filename != "" {
			push("source_filename=\"", it.source_filename, "\"\n")
		}
		for i := range it.GlobalVars {
			buf = append(buf, llIrSrc(&it.GlobalVars[i])...)
		}
		for i := range it.ExtDecls {
			buf = append(buf, llIrSrc(&it.ExtDecls[i])...)
		}
		for i := range it.FuncDefs {
			buf = append(buf, llIrSrc(&it.FuncDefs[i])...)
		}

	case *LlGlobalVar:

	case *LlExtDecl:

	case *LlFuncDef:

	default:
		panic(it)
	}
	return buf
}
