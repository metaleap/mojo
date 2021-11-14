package main

import (
	"io"
)

func (me *LlExprLitInt) expr() (LlType, interface{})     { return me.ty, me.value }
func (me *LlExprLitFloat) expr() (LlType, interface{})   { return me.ty, me.value }
func (me *LlExprLitCStr) expr() (LlType, interface{})    { return me.ty, me.value }
func (me *LlExprLitBuiltin) expr() (LlType, interface{}) { return me.ty, me.name }
func (me *LlExprRefLocal) expr() (LlType, interface{})   { return me.ty, me.name }
func (me *LlExprRefGlobal) expr() (LlType, interface{})  { return me.ty, me.name }

func llIrSrc(buf io.StringWriter, llvmIr interface{}) {
	push := func(strs ...string) {
		for _, s := range strs {
			_, _ = buf.WriteString(s)
		}
	}
	switch it := llvmIr.(type) {

	case LlTypeVoid:
		push("void")

	case LlTypeInt:
		push("i", itoa(it.bitWidth))

	case LlTypeFloat:
		switch it.bitWidth {
		case 16:
			push("half")
		case 32:
			push("float")
		case 64:
			push("double")
		case 128:
			push("fp128")
		default:
			panic(it.bitWidth)
		}

	case LlTypePtr:
		llIrSrc(buf, it.elemTy)
		push("*")

	case LlTypeArr:
		push("[", itoa(it.numElems), " x ")
		llIrSrc(buf, it.elemTy)
		push("]")

	case LlTypeStruct:
		for i, fieldtype := range it.fields {
			push(ifStr(i == 0, "{ ", ", "))
			llIrSrc(buf, fieldtype)
		}
		push(" }")

	case LlTypeFunc:
		llIrSrc(buf, it.ret.ty)
		push("(")
		for i, param := range it.params {
			if i > 0 {
				push(", ")
			}
			llIrSrc(buf, param.ty)
		}
		push(")")

	case *LlTopLevel:
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

	case *LlTopLevelGlobalVar:
		push("@", it.name, " = ", ifStr(it.constant, "constant ", "global "))
		llIrSrc(buf, it.ty)
		if it.init != nil {
			push(" ")
			llIrSrc(buf, it.init)
		}
		if it.comment != "" {
			push("\t\t; ", it.comment)
		}
		push("\n")

	case *LlTopLevelExtDecl:
		push("declare ")
		if it.intrinsic != 0 {
			// var decl LlTopLevelExtDecl
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
		for i := range fnty.params {
			if i > 0 {
				push(", ")
			}
			llIrSrc(buf, fnty.params[i].ty)
		}
		push(")")
		if it.comment != "" {
			push("\t\t; ", it.comment)
		}
		push("\n")

	case *LlTopLevelFuncDef:
		push("define ")
		fnty := &it.ty
		llIrSrc(buf, fnty.ret.ty)
		push(" @", it.name, "(")
		for i := range fnty.params {
			if i > 0 {
				push(", ")
			}
			llIrSrc(buf, fnty.params[i].ty)
			push(" %", fnty.params[i].name)
		}
		push(") {")
		if it.comment != "" {
			push("\t\t; ", it.comment)
		}
		push("\n")
		for _, block := range it.blocks {
			if name := block.name; name != "" {
				push("  ", name, ":")
				if block.comment != "" {
					push("\t\t; ", block.comment)
				}
				push("\n")
			}
			for i := range block.instrs {
				push("    ")
				llIrSrc(buf, block.instrs[i])
				if commented, ok := block.instrs[i].(interface{ commented() string }); ok && commented != nil {
					if comment := commented.commented(); comment != "" {
						push("\t\t; ", comment)
					}
				}
				push("\n")
			}
		}
		push("}\n")

	case LlExprLitInt:
		push(itoa(it.value))
	case LlExprLitCStr:
		push("c", strQuote(it.value))
	case LlExprLitBuiltin:
		push(it.name)
	case LlExprRefLocal:
		push("%", it.name)
	case LlExprRefGlobal:
		push("@", it.name)

	default:
		panic(it)
	}
}
