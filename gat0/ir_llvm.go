package main

import (
	"bytes"
)

func (me *LlExprLitInt) expr() (LlType, interface{})     { return me.ty, me.value }
func (me *LlExprLitFloat) expr() (LlType, interface{})   { return me.ty, me.value }
func (me *LlExprLitCStr) expr() (LlType, interface{})    { return me.ty, me.value }
func (me *LlExprLitBuiltin) expr() (LlType, interface{}) { return me.ty, me.name }
func (me *LlExprRefLocal) expr() (LlType, interface{})   { return me.ty, me.name }
func (me *LlExprRefGlobal) expr() (LlType, interface{})  { return me.ty, me.name }

func llIrSrc(buf *bytes.Buffer, llvmIr interface{}) {
	push := func(strs ...string) {
		for _, s := range strs {
			buf.WriteString(s)
		}
	}
	switch it := llvmIr.(type) {

	case *LlModule:
		if it.source_filename != "" {
			push("source_filename=\"", it.source_filename, "\"\n")
		}
		for i := range it.GlobalVars {
			llIrSrc(buf, &it.GlobalVars[i])
		}
		push("\n")
		for i := range it.ExtDecls {
			llIrSrc(buf, &it.ExtDecls[i])
		}
		push("\n")
		for i := range it.FuncDefs {
			llIrSrc(buf, &it.FuncDefs[i])
			push("\n")
		}

	case *LlGlobalVar:
		push("@", it.name, " = ", ifStr(it.constant, "constant ", "global "))
		llIrSrc(buf, it.ty)
		if it.init != nil {
			push(" ")
			llIrSrc(buf, it.init)
		}
		push("\n")

	case *LlExtDecl:
		push("declare ")
		if it.intrinsic != 0 {
			// var decl LlExtDecl
			switch it.intrinsic {
			default:
				panic(it.intrinsic)
			}
			// llIrSrc(buf, &decl)
			// return
		}
		fnty := it.ty.(*LlTypeFunc)
		llIrSrc(buf, fnty.ret.ty)
		push(" @", it.name, "(")
		for i := range fnty.args {
			if i > 0 {
				push(", ")
			}
			llIrSrc(buf, fnty.args[i].ty)
		}
		push(")\n")

	case *LlFuncDef:
		push("define ")
		fnty := &it.ty
		llIrSrc(buf, fnty.ret.ty)
		push(" @", it.name, "(")
		for i := range fnty.args {
			if i > 0 {
				push(", ")
			}
			llIrSrc(buf, fnty.args[i].ty)
			push(" %", fnty.args[i].name)
		}
		push(") {\n")

		push("}\n")

	default:
		panic(it)
	}
}
