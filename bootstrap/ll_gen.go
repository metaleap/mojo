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
		top_def_name := astDefName(top_def)
		switch body := top_def.body.kind.(type) {
		case AstExprLitStr:
			ret_mod.globals[num_globals] = llGlobalFromLitStr(top_def_name, body)
			num_globals++
		case AstExprForm:
			callee := body[0].kind.(AstExprIdent)
			if strEql(callee, Str("/extVar")) {
				ret_mod.globals[num_globals] = llGlobalFromExtVar(body, top_def, ast)
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
	return ret_mod
}

func llGlobalFromExtVar(form AstExprForm, cur_def *AstDef, ast *Ast) LLGlobal {
	assert(len(form) == 3)
	c_name := astScopesResolve(&cur_def.scope, form[1].kind.(AstExprIdent), -1).(*AstDef).body.kind.(AstExprLitStr)
	return LLGlobal{
		name:     c_name,
		external: true,
		ty:       llTypeFrom(&form[2], cur_def, ast),
	}
}

func llGlobalFromLitStr(name Str, body AstExprLitStr) LLGlobal {
	return LLGlobal{
		name:        name,
		constant:    true,
		ty:          LLTypeArr{size: uint64(len(body)), ty: LLTypeInt{bit_width: 8}},
		initializer: LLExprLitStr(body),
	}
}

func llFuncDeclFrom(top_def *AstDef, ast *Ast) LLFunc {
	form := top_def.body.kind.(AstExprForm)
	assert(len(form) == 4)
	c_name := astScopesResolve(&top_def.scope, form[1].kind.(AstExprIdent), -1).(*AstDef).body.kind.(AstExprLitStr)
	lit_curl := form[3].kind.(AstExprLitCurl)
	ret_decl := LLFunc{
		external: true,
		name:     c_name,
		ty:       llTypeFrom(&form[2], top_def, ast),
		params:   allocˇLLFuncParam(len(lit_curl)),
	}
	for i := range lit_curl {
		_, ty := astExprFormSplit(&lit_curl[i], Str(":"), true, true, true, ast)
		ret_decl.params[i].name = nil
		ret_decl.params[i].ty = llTypeFrom(ty, top_def, ast)
	}
	return ret_decl
}

func llFuncDefFrom(top_def *AstDef, ast *Ast) LLFunc {
	form := top_def.body.kind.(AstExprForm)
	assert(len(form) == 4)
	lit_args := form[2].kind.(AstExprLitCurl)
	lit_body := form[3].kind.(AstExprLitCurl)
	ret_def := LLFunc{
		external:     false,
		name:         astDefName(top_def),
		ty:           llTypeFrom(&form[1], top_def, ast),
		params:       allocˇLLFuncParam(len(lit_args)),
		basic_blocks: allocˇLLBasicBlock(len(lit_body)),
	}
	for i := range lit_args {
		arg_name, arg_ty := astExprFormSplit(&lit_args[i], Str(":"), true, true, true, ast)
		ret_def.params[i].name = llIdentFrom(arg_name, ast)
		ret_def.params[i].ty = llTypeFrom(arg_ty, top_def, ast)
	}
	for i := range lit_body {
		block_name, block_stmts := astExprFormSplit(&lit_body[i], Str(":"), true, true, true, ast)
		ret_def.basic_blocks[i].name = llIdentFrom(block_name, ast)
		stmts := block_stmts.kind.(AstExprLitClip)
		ret_def.basic_blocks[i].stmts = allocˇLLStmt(len(stmts) * 2)
		num_stmts := 0
		for j := range stmts {
			ll_stmts := llStmtsFrom(&stmts[j], top_def, ast, true)
			for _, ll_stmt := range ll_stmts {
				ret_def.basic_blocks[i].stmts[num_stmts] = ll_stmt
				num_stmts++
			}
		}
		ret_def.basic_blocks[i].stmts = ret_def.basic_blocks[i].stmts[0:num_stmts]
	}
	return ret_def
}

func llStmtsFrom(expr *AstExpr, cur_def *AstDef, ast *Ast, prepend_comment bool) []LLStmt {
	form := expr.kind.(AstExprForm)
	ident := form[0].kind.(AstExprIdent)
	ret_stmts := allocˇLLStmt(2)
	num_stmts := 0

	if prepend_comment {
		ret_stmts[num_stmts] = LLStmtComment{comment_text: trimAt(astNodeSrcStr(&expr.base, ast.src, ast.toks), '\n')}
		num_stmts++
	}

	if !strEql(ident, Str("/let")) {
		for i := 1; i < len(form); i++ {
			if _, is := form[i].kind.(AstExprForm); is {
				tmp_name := llIdentFrom(&form[i], ast)
				tmp_let := AstExpr{
					kind: AstExprForm{
						AstExpr{kind: AstExprIdent("/let")},
						AstExpr{kind: AstExprIdent(tmp_name)},
						form[i],
					},
					base: form[i].base,
				}
				for _, sub_stmt := range llStmtsFrom(&tmp_let, cur_def, ast, false) {
					ret_stmts[num_stmts] = sub_stmt
					num_stmts++
				}
				form[i].kind = AstExprIdent(tmp_name)
			}
		}
	}

	if strEql(ident, Str("/br")) {
		assert(len(form) == 2)
		ret_stmts[num_stmts] = LLStmtBr{block_name: llIdentFrom(&form[1], ast)}
		num_stmts++
	} else if strEql(ident, Str("/let")) {
		assert(len(form) == 3)
		name := llIdentFrom(&form[1], ast)
		if len(name) == 1 && name[0] == '.' {
			counter++
			name = uintToStr(counter, 10, 1, Str(".."))
		}
		ret_stmts[num_stmts] = LLStmtLet{
			name: name,
			expr: llExprFrom(&form[2], nil, cur_def, ast),
		}
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
	switch it := expr.kind.(type) {
	case AstExprLitStr:
		return LLTypeArr{size: uint64(len(it)), ty: LLTypeInt{bit_width: 8}}
	case AstExprIdent:
		ident := expr.kind.(AstExprIdent)
		if ident[0] != '/' {
			ref_def := astScopesResolve(&cur_def.scope, ident, -1).(*AstDef)
			return llTypeFrom(&ref_def.body, cur_def, ast)
		} else if ident[1] == 'V' {
			assert(len(ident) == 2)
			return LLTypeVoid{}
		} else if ident[1] == 'P' {
			assert(len(ident) == 2)
			return LLTypePtr{ty: LLTypeInt{bit_width: 8}}
		} else if ident[1] == 'I' {
			bit_width := uintFromStr(ident[2:])
			return LLTypeInt{bit_width: uint32(bit_width)}
		}
		panic("TODO: llTypeFrom " + string(ident))
	case AstExprForm:
		switch callee := it[0].kind.(type) {
		case AstExprIdent:
			if strEql(callee, Str("/extFun")) {
				lit_args := it[3].kind.(AstExprLitCurl)
				ret_ty := LLTypeFunc{ty: llTypeFrom(&it[2], cur_def, ast), params: allocˇLLType(len(lit_args))}
				for i := range lit_args {
					_, ty := astExprFormSplit(&lit_args[i], Str(":"), true, true, true, ast)
					ret_ty.params[i] = llTypeFrom(ty, cur_def, ast)
				}
				return ret_ty
			} else if strEql(callee, Str("/defFun")) {
				lit_args := it[2].kind.(AstExprLitCurl)
				ret_ty := LLTypeFunc{ty: llTypeFrom(&it[1], cur_def, ast), params: allocˇLLType(len(lit_args))}
				for i := range lit_args {
					_, ty := astExprFormSplit(&lit_args[i], Str(":"), true, true, true, ast)
					ret_ty.params[i] = llTypeFrom(ty, cur_def, ast)
				}
				return ret_ty
			} else if strEql(callee, Str("/extVar")) {
				return llTypeFrom(&it[2], cur_def, ast)
			}
			panic(string(callee))
		default:
			panic(callee)
		}
	default:
		panic(it)
	}
}

func llIdentFrom(expr *AstExpr, ast *Ast) Str {
	src_str := astNodeSrcStr(&expr.base, ast.src, ast.toks)
	if len(src_str) == 0 {
		counter++
		src_str = uintToStr(counter, 10, 1, Str(".tmp."))
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
