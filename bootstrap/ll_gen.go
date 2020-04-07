package main

const llmodule_default_target_datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
const llmodule_default_target_triple = "x86_64-unknown-linux-gnu"

func llModule(ast *Ast) LLModule {
	ret_mod := LLModule{
		target_datalayout: Str(llmodule_default_target_datalayout),
		target_triple:     Str(llmodule_default_target_triple),
		globals:           allocˇLLGlobal(len(ast.defs)),
		funcs:             allocˇLLFunc(len(ast.defs)),
	}
	num_globals, num_funcs := 0, 0
	for i := range ast.defs {
		top_def := &ast.defs[i]
		switch body := top_def.body.kind.(type) {
		case AstExprLitStr:
			ret_mod.globals[num_globals] = llGlobalFromLitStr(top_def, body)
			num_globals++
		case AstExprForm:
			callee := body[0].kind.(AstExprIdent)
			if strEql(callee, Str("/extVar")) {
				ret_mod.globals[num_globals] = llGlobalFromExtVar(top_def, ast)
				num_globals++
			} else if strEql(callee, Str("/extFun")) {
				ret_mod.funcs[num_funcs] = llFuncDeclFrom(top_def, ast)
				num_funcs++
			} else if strEql(callee, Str("/defFun")) {
				ret_mod.funcs[num_funcs] = llFuncDefFrom(top_def, ast)
				num_funcs++
			} else {
				fail(callee)
			}
		default:
			panic(body)
		}
	}
	ret_mod.globals = ret_mod.globals[0:num_globals]
	ret_mod.funcs = ret_mod.funcs[0:num_funcs]
	for i := range ast.defs {
		top_def := &ast.defs[i]
		top_def_name := astDefName(top_def)
		switch body := top_def.body.kind.(type) {
		case AstExprForm:
			callee := body[0].kind.(AstExprIdent)
			if strEql(callee, Str("/defFun")) {
				done := false
				for j := range ret_mod.funcs {
					dst_def := &ret_mod.funcs[j]
					if strEql(dst_def.name, top_def_name) {
						llFuncDefSetAllExprTypes(top_def, dst_def, ast)
						llFuncDefPopulateBlocks(top_def, dst_def, ast)
						done = true
						break
					}
				}
				assert(done)
			}
		}
	}
	return ret_mod
}

func llGlobalFromLitStr(top_def *AstDef, body AstExprLitStr) LLGlobal {
	name := astDefName(top_def)

	top_def.base.anns.ll_ty = LLTypeArr{size: uint64(len(body)), ty: LLTypeInt{bit_width: 8}}
	top_def.body.base.anns.ll_ty = top_def.base.anns.ll_ty
	return LLGlobal{
		name:        name,
		constant:    true,
		ty:          top_def.base.anns.ll_ty,
		initializer: LLExprLitStr(body),
	}
}

func llGlobalFromExtVar(top_def *AstDef, ast *Ast) LLGlobal {
	form := top_def.body.kind.(AstExprForm)
	assert(len(form) == 3)

	c_name := astScopesResolve(&top_def.scope, form[1].kind.(AstExprIdent), -1).(*AstDef).body.kind.(AstExprLitStr)
	top_def.base.anns.ll_ty = llTypeFrom(&form[2], top_def, ast)
	top_def.body.base.anns.ll_ty = top_def.base.anns.ll_ty
	return LLGlobal{
		name:     c_name,
		external: true,
		ty:       top_def.base.anns.ll_ty,
	}
}

func llFuncDeclFrom(top_def *AstDef, ast *Ast) LLFunc {
	form := top_def.body.kind.(AstExprForm)
	assert(len(form) == 4)

	top_def.base.anns.ll_ty = llTypeFrom(&top_def.body, top_def, ast)
	top_def.body.base.anns.ll_ty = top_def.base.anns.ll_ty
	fn_ty := top_def.base.anns.ll_ty.(LLTypeFunc)

	c_name := astScopesResolve(&top_def.scope, form[1].kind.(AstExprIdent), -1).(*AstDef).body.kind.(AstExprLitStr)
	lit_curl := form[3].kind.(AstExprLitCurl)
	ret_decl := LLFunc{
		external: true,
		name:     c_name,
		ty:       fn_ty.ty,
		params:   allocˇLLFuncParam(len(lit_curl)),
	}
	for i := range lit_curl {
		ret_decl.params[i].name = nil
		ret_decl.params[i].ty = fn_ty.params[i]
	}
	return ret_decl
}

func llFuncDefFrom(top_def *AstDef, ast *Ast) LLFunc {
	form := top_def.body.kind.(AstExprForm)
	assert(len(form) == 4)

	top_def.base.anns.ll_ty = llTypeFrom(&top_def.body, top_def, ast)
	top_def.body.base.anns.ll_ty = top_def.base.anns.ll_ty
	fn_ty := top_def.base.anns.ll_ty.(LLTypeFunc)

	lit_args := form[2].kind.(AstExprLitCurl)
	lit_body := form[3].kind.(AstExprLitCurl)
	ret_def := LLFunc{
		external:     false,
		name:         astDefName(top_def),
		ty:           fn_ty.ty,
		params:       allocˇLLFuncParam(len(lit_args)),
		basic_blocks: allocˇLLBasicBlock(len(lit_body)),
	}
	for i := range lit_args {
		arg_name, _ := astExprFormSplit(&lit_args[i], Str(":"), true, true, true, ast)
		ret_def.params[i].name = llIdentFrom(arg_name, '#', ast)
		ret_def.params[i].ty = fn_ty.params[i]
	}
	for i := range lit_body {
		block_name, _ := astExprFormSplit(&lit_body[i], Str(":"), true, true, true, ast)
		ret_def.basic_blocks[i].name = llIdentFrom(block_name, '#', ast)
		ret_def.basic_blocks[i].stmts = nil // created later in `llFuncDefPopulateBlocks`
	}
	return ret_def
}

func llFuncDefSetAllExprTypes(top_def *AstDef, dst_def *LLFunc, ast *Ast) {
	form := top_def.body.kind.(AstExprForm)
	assert(len(form) == 4)
	lit_body := form[3].kind.(AstExprLitCurl)
	for i := range lit_body {
		_, block_stmts := astExprFormSplit(&lit_body[i], Str(":"), true, true, true, ast)
		stmt_exprs := block_stmts.kind.(AstExprLitClip)
		for j := range stmt_exprs {
			stmt_expr := stmt_exprs[j].kind.(AstExprForm)
			ident := stmt_expr[0].kind.(AstExprIdent)
			if strEql(ident, Str("/ret")) {
				assert(len(stmt_expr) == 2)
				_ = llExprEnsureTypeAnnotation(&stmt_expr[1], dst_def.ty, top_def, ast)
			}
		}
	}
}

func llExprEnsureTypeAnnotation(expr *AstExpr, maybe_ty LLType, cur_def *AstDef, ast *Ast) LLType {
	if expr.base.anns.ll_ty == nil {
		switch this := expr.kind.(type) {
		case AstExprForm:
			ident := this[0].kind.(AstExprIdent)
			if strEql(ident, Str("/phi")) {
				assert(len(this) == 2)
				lit_curl := this[1].kind.(AstExprLitCurl)
				phi_ty := maybe_ty
				for i := range lit_curl {
					_, rhs := astExprFormSplit(&lit_curl[i], Str(":"), true, true, true, ast)
					if nil != llExprEnsureTypeAnnotation(rhs, phi_ty, cur_def, ast) {
						if phi_ty == nil {
							phi_ty = rhs.base.anns.ll_ty
						} else if !llTypeEql(phi_ty, rhs.base.anns.ll_ty) {
							fail("type mismatch: got ", phi_ty, " but had ", rhs.base.anns.ll_ty, " in:\n", astNodeSrcStr(&expr.base, ast.src, ast.toks))
						}
					}
				}
				expr.base.anns.ll_ty = phi_ty
				for i := range lit_curl {
					_, rhs := astExprFormSplit(&lit_curl[i], Str(":"), true, true, true, ast)
					if rhs.base.anns.ll_ty == nil {
						rhs.base.anns.ll_ty = phi_ty
					}
				}
			} else if strEql(ident, Str("/call")) {
				// callee := this[1].kind.(AstExprIdent)
				// ref_def := astScopesResolve(&cur_def.scope, callee, -1).(*AstDef)
				// func_ty := ref_def.base.anns.ll_ty.(LLTypeFunc)
			}
		}
	}

	if maybe_ty != nil {
		if expr.base.anns.ll_ty != nil && !llTypeEql(maybe_ty, expr.base.anns.ll_ty) {
			fail("type mismatch: got ", maybe_ty, " but had ", expr.base.anns.ll_ty, " in:\n", astNodeSrcStr(&expr.base, ast.src, ast.toks))
		}
		expr.base.anns.ll_ty = maybe_ty
	}
	return expr.base.anns.ll_ty
}

func llFuncDefPopulateBlocks(top_def *AstDef, dst_def *LLFunc, ast *Ast) {
	form := top_def.body.kind.(AstExprForm)
	assert(len(form) == 4)
	lit_body := form[3].kind.(AstExprLitCurl)
	for i := range lit_body {
		_, block_stmts := astExprFormSplit(&lit_body[i], Str(":"), true, true, true, ast)
		stmt_exprs := block_stmts.kind.(AstExprLitClip)
		dst_def.basic_blocks[i].stmts = allocˇLLStmt(len(stmt_exprs) * 2)
		num_stmts := 0
		for j := range stmt_exprs {
			ll_stmts := llStmtsFrom(&stmt_exprs[j], top_def, dst_def, ast, true)
			for _, ll_stmt := range ll_stmts {
				dst_def.basic_blocks[i].stmts[num_stmts] = ll_stmt
				num_stmts++
			}
		}
		dst_def.basic_blocks[i].stmts = dst_def.basic_blocks[i].stmts[0:num_stmts]
	}
}

func llStmtsFrom(stmt_expr *AstExpr, top_def *AstDef, dst_def *LLFunc, ast *Ast, prepend_comment bool) []LLStmt {
	form := stmt_expr.kind.(AstExprForm)
	ident := form[0].kind.(AstExprIdent)
	ret_stmts := allocˇLLStmt(2)
	num_stmts := 0

	if prepend_comment {
		ret_stmts[num_stmts] = LLStmtComment{comment_text: trimAt(astNodeSrcStr(&stmt_expr.base, ast.src, ast.toks), '\n')}
		num_stmts++
	}

	if strEql(ident, Str("/br")) {
		assert(len(form) == 2)
		ret_stmts[num_stmts] = LLStmtBr{block_name: llIdentFrom(&form[1], '#', ast)}
		num_stmts++
	}

	return ret_stmts[0:num_stmts]
}

func llExprFrom(expr *AstExpr, maybe_ty LLType, cur_def *AstDef, ast *Ast) LLExpr {
	switch it := expr.kind.(type) {
	case AstExprLitInt:
		return LLExprBinOp{ty: maybe_ty, op_kind: ll_bin_op_add, lhs: LLExprLitInt(0), rhs: LLExprLitInt(it)}
	case AstExprIdent:
		ref_def := astScopesResolve(&cur_def.scope, it, -1).(*AstDef)
		def_ty := llTypeFrom(&ref_def.body, cur_def, ast)
		return LLExprLoad{ty: def_ty, expr: LLExprIdentGlobal(astDefName(ref_def))}
	case AstExprForm:
		ident := it[0].kind.(AstExprIdent)
		if strEql(ident, Str("/call")) {
			assert(len(it) >= 3)
			callee := astScopesResolve(&cur_def.scope, it[1].kind.(AstExprIdent), -1).(*AstDef)
			func_ty := llTypeFrom(&callee.body, cur_def, ast).(LLTypeFunc)
			ret_call := LLExprCall{
				callee: LLExprTyped{
					ty:   func_ty.ty,
					expr: LLExprIdentGlobal(astDefName(callee)),
				},
				args: allocˇLLExprTyped(len(it) - 3),
			}
			assert(len(ret_call.args) == len(func_ty.params))
			for i := range ret_call.args {
				ret_call.args[i] = LLExprTyped{
					ty:   func_ty.params[i],
					expr: llExprFrom(&it[i+3], func_ty.params[i], cur_def, ast),
				}
			}
			return ret_call
		} else if strEql(ident, Str("/len")) {
			assert(len(it) == 2)
			arr_ty := llTypeFrom(&it[1], cur_def, ast).(LLTypeArr)
			return LLExprLitInt(arr_ty.size)
		} else if strEql(ident, Str("/phi")) {
			assert(len(it) == 2)
			assert(maybe_ty != nil)
			lit_curl := it[1].kind.(AstExprLitCurl)
			ret_phi := LLExprPhi{ty: maybe_ty, predecessors: allocˇLLPhiPred(len(lit_curl))}
			for i := range lit_curl {
				name, expr := astExprFormSplit(&lit_curl[i], Str(":"), true, true, true, ast)
				ret_phi.predecessors[i] = LLPhiPred{
					block_name: name.kind.(AstExprIdent),
					expr:       llExprFrom(expr, maybe_ty, cur_def, ast),
				}
			}
			return ret_phi
		}
		panic(string(ident))
	default:
		panic(it)
	}
}

func llTypeFrom(expr *AstExpr, cur_def *AstDef, ast *Ast) LLType {
	ret_ty := expr.base.anns.ll_ty
	if ret_ty == nil {
		switch it := expr.kind.(type) {
		case AstExprLitStr:
			ret_ty = LLTypeArr{size: uint64(len(it)), ty: LLTypeInt{bit_width: 8}}
		case AstExprIdent:
			ident := expr.kind.(AstExprIdent)
			if ident[0] != '/' {
				ref_def := astScopesResolve(&cur_def.scope, ident, -1).(*AstDef)
				if ret_ty = ref_def.base.anns.ll_ty; ret_ty == nil {
					ret_ty = llTypeFrom(&ref_def.body, cur_def, ast)
					ref_def.base.anns.ll_ty = ret_ty
				}
			} else if ident[1] == 'V' {
				assert(len(ident) == 2)
				ret_ty = LLTypeVoid{}
			} else if ident[1] == 'P' {
				assert(len(ident) == 2)
				ret_ty = LLTypePtr{ty: LLTypeInt{bit_width: 8}}
			} else if ident[1] == 'I' {
				bit_width := uintFromStr(ident[2:])
				ret_ty = LLTypeInt{bit_width: uint32(bit_width)}
			}
			if ret_ty == nil {
				panic("TODO: llTypeFrom " + string(ident))
			}
		case AstExprForm:
			switch callee := it[0].kind.(type) {
			case AstExprIdent:
				if strEql(callee, Str("/extFun")) {
					lit_args := it[3].kind.(AstExprLitCurl)
					fn_ty := LLTypeFunc{ty: llTypeFrom(&it[2], cur_def, ast), params: allocˇLLType(len(lit_args))}
					for i := range lit_args {
						_, ty := astExprFormSplit(&lit_args[i], Str(":"), true, true, true, ast)
						fn_ty.params[i] = llTypeFrom(ty, cur_def, ast)
					}
					ret_ty = fn_ty
				} else if strEql(callee, Str("/defFun")) {
					lit_args := it[2].kind.(AstExprLitCurl)
					fn_ty := LLTypeFunc{ty: llTypeFrom(&it[1], cur_def, ast), params: allocˇLLType(len(lit_args))}
					for i := range lit_args {
						_, ty := astExprFormSplit(&lit_args[i], Str(":"), true, true, true, ast)
						fn_ty.params[i] = llTypeFrom(ty, cur_def, ast)
					}
					ret_ty = fn_ty
				} else if strEql(callee, Str("/extVar")) {
					ret_ty = llTypeFrom(&it[2], cur_def, ast)
				}
				if ret_ty == nil {
					panic(string(callee))
				}
			default:
				panic(callee)
			}
		default:
			panic(it)
		}
		assert(ret_ty != nil)
		expr.base.anns.ll_ty = ret_ty
	}
	return ret_ty
}

func llIdentFrom(expr *AstExpr, must_prefix byte, ast *Ast) Str {
	src_str := astNodeSrcStr(&expr.base, ast.src, ast.toks)
	if must_prefix != 0 && len(src_str) != 0 {
		assert(len(src_str) != 0)
		assert(src_str[0] == must_prefix)
	}
	if len(src_str) == 0 || (len(src_str) == 1 && src_str[0] == must_prefix) {
		counter++
		src_str = uintToStr(counter, 10, 1, Str("-$-"))
	}
	ret_str := allocˇu8(len(src_str))
	for i, c := range src_str {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c == '-' || c == '$' || (c >= '0' && c <= '9' && i > 0) {
			ret_str[i] = c
		} else {
			ret_str[i] = '.'
		}
	}
	return ret_str
}
