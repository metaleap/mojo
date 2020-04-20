package main

func irHLPrint(ir *IrHL) {
	for i := range ir.defs {
		irHLPrintDef(ir, &ir.defs[i])
	}
}

func irHLPrintDef(ir *IrHL, def *IrHLDef) {
	print("\n\n")
	print(string(def.anns.name))
	print(" :=\n  ")
	irHLPrintExpr(ir, &def.body, 4)
	print("\n\n")
}

func irHLPrintExpr(ir *IrHL, expr *IrHLExpr, ind int) {
	switch it := expr.variant.(type) {
	case IrHLExprType:
		irHLPrintType(ir, it.ty)
	case IrHLExprTag:
		print("#")
		print(string(it))
	case IrHLExprInt:
		print(it)
	case IrHLExprIdent:
		print(string(it))
	case IrHLExprRefDef:
		print(string(ir.defs[it].anns.name))
	case IrHLExprRefLocal:
		if it.let != nil {
			print("%")
			print(string(it.let.variant.(IrHLExprLet).names[it.let_idx]))
		}
		if it.param_idx >= 0 {
			print("@")
			print(it.param_idx)
		}
	case IrHLExprCall:
		if expr.anns.origin_expr.anns.toks_throng {
			for i := range it {
				irHLPrintExpr(ir, &it[i], ind)
			}
		} else {
			print("(")
			for i := range it {
				if i > 0 {
					print(" ")
				}
				irHLPrintExpr(ir, &it[i], ind)
			}
			print(")")
		}
	case IrHLExprList:
		print("[ ")
		for i := range it {
			irHLPrintExpr(ir, &it[i], ind)
			print(", ")
		}
		print("]")
	case IrHLExprBag:
		if ind == 0 || len(it) < 2 {
			print("{")
			for i := range it {
				if i > 0 {
					print(", ")
				}
				irHLPrintExpr(ir, &it[i], ind+2)
			}
			print("}")
		} else {
			print("{\n")
			for i := range it {
				for j := 0; j < ind; j++ {
					print(" ")
				}
				irHLPrintExpr(ir, &it[i], 2+ind)
				print(",\n")
			}
			for i := 0; i < ind; i++ {
				print(" ")
			}
			print("}")
		}
	case IrHLExprInfix:
		irHLPrintExpr(ir, &it.lhs, ind)
		print(string(it.kind))
		if !strEq(it.kind, ".") {
			print(" ")
		}
		irHLPrintExpr(ir, &it.rhs, ind)
	case IrHLExprLet:
		print("(")
		irHLPrintExpr(ir, &it.body, ind)
		print(",\n\n")
		for i, name := range it.names {
			for j := 0; j < ind; j++ {
				print(" ")
			}
			print(string(name))
			print(" := ")
			irHLPrintExpr(ir, &it.exprs[i], 2+ind)
			print(",\n\n")
		}
		for i := 0; i < ind; i++ {
			print(" ")
		}
		print(")")
	case IrHLExprPrimCallee:
		print("⟨")
		print(string(it))
		print("⟩")
	case IrHLExprPrimCase:
		print("⟨case ")
		irHLPrintExpr(ir, &it.scrut, ind)
		irHLPrintExpr(ir, &IrHLExpr{variant: IrHLExprBag(it.cases)}, ind)
		print("⟩")
	case IrHLExprPrimCmp:
		print("⟨cmp ")
		switch it.kind {
		case hl_cmp_eq:
			print("#eq ")
		case hl_cmp_neq:
			print("#neq ")
		case hl_cmp_lt:
			print("#lt ")
		case hl_cmp_gt:
			print("#gt ")
		case hl_cmp_leq:
			print("#leq ")
		case hl_cmp_geq:
			print("#geq ")
		default:
			panic(it.kind)
		}
		irHLPrintExpr(ir, &it.lhs, ind)
		print(" ")
		irHLPrintExpr(ir, &it.rhs, ind)
		print("⟩")
	case IrHLExprPrimCallExt:
		print("⟨call ")
		irHLPrintExpr(ir, &it.callee, ind)
		for i := range it.args {
			print(" ")
			irHLPrintExpr(ir, &it.args[i], ind)
		}
		print("⟩")
	case IrHLExprPrimLen:
		print("⟨len ")
		irHLPrintExpr(ir, &it.subj, ind)
		print("⟩")
	case IrHLExprPrimExtRef:
		print("⟨extern ")
		irHLPrintExpr(ir, &it.name_tag, ind)
		print(" ")
		irHLPrintExpr(ir, &it.opts, ind)
		print(" ")
		irHLPrintExpr(ir, &it.ty, ind)
		if it.fn_params != nil {
			print(" ")
			irHLPrintExpr(ir, &IrHLExpr{variant: IrHLExprBag(it.fn_params)}, 0)
		}
		print("⟩")
	case IrHLExprFunc:
		print("(")
		for i := range it.params {
			irHLPrintExpr(ir, &it.params[i], ind)
			print(" ")
		}
		print("-> ")
		irHLPrintExpr(ir, &it.body, ind)
		print(")")
	case nil:
		print("⟨?NIL?⟩")
	default:
		panic(it)
	}
}

func irHLPrintType(ir *IrHL, ty IrHLType) {
	print("⟨T ")
	switch it := ty.(type) {
	case IrHLTypeVoid:
		print("#void")
	case IrHLTypeInt:
		print("#int ")
		if it.c {
			print("#c")
		} else if it.word {
			print("#word")
		} else {
			print(it.min)
			print(" ")
			print(it.max)
		}
	case IrHLTypeExt:
		print("#extern #")
		print(string(it))
	case IrHLTypePtr:
		print("#ptr ")
		irHLPrintType(ir, it.payload)
	case IrHLTypeArr:
		print("#arr ")
		irHLPrintType(ir, it.payload)
		if it.size >= 0 {
			print(" ")
			print(it.size)
		}
	case IrHLTypeFunc:
		print("#func ")
		irHLPrintType(ir, it.returns)
		for i := range it.params {
			print(" ")
			irHLPrintType(ir, it.params[i])
		}
	case IrHLTypeTag:
		print("#tag")
		if len(it) != 0 {
			print(" #")
			print(string(it))
		}
	case IrHLTypeBag:
		if it.is_union {
			print("#union {")
		} else {
			print("#struct {")
		}
		for i := range it.field_names {
			if i > 0 {
				print(", ")
			}
			print(string(it.field_names[i]))
			print(": ")
			irHLPrintType(ir, it.field_types[i])
		}
		print("}")
	default:
		panic(it)
	}
	print("⟩")
}
