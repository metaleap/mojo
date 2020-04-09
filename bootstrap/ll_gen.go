package main

const llmodule_default_target_datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
const llmodule_default_target_triple = "x86_64-unknown-linux-gnu"

func llModuleFrom(ast *Ast) LLModule {
	ret_mod := LLModule{
		target_datalayout: Str(llmodule_default_target_datalayout),
		target_triple:     Str(llmodule_default_target_triple),
		globals:           allocˇLLGlobal(len(ast.defs)),
		funcs:             allocˇLLFunc(len(ast.defs)),
	}
	num_globals, num_funcs := 0, 0
	for i := range ast.defs {
		top_def := &ast.defs[i]
		switch ll_sth := llTopLevelFrom(&top_def.body, top_def, ast).(type) {
		case LLGlobal:
			ret_mod.globals[num_globals] = ll_sth
			num_globals++
			println(string(astDefName(top_def)), "\tVS\t", string(ll_sth.name))
		case LLFunc:
			ret_mod.funcs[num_funcs] = ll_sth
			num_funcs++
			println(string(astDefName(top_def)), "\tVS\t", string(ll_sth.name))
		default:
			fail(string(astDefName(top_def)))
		}
	}
	ret_mod.globals = ret_mod.globals[0:num_globals]
	ret_mod.funcs = ret_mod.funcs[0:num_funcs]
	return ret_mod
}

func llTopLevelFrom(expr *AstExpr, top_def *AstDef, ast *Ast) Any {
	switch it := expr.kind.(type) {
	case AstExprForm:
		callee := astExprSlashed(&it[0])
		kwd := callee[0]
		if 1 == len(callee) {
			if astExprIsIdent(kwd, "global") {
				return llGlobalFrom(expr, top_def, ast)
			} else if astExprIsIdent(kwd, "declare") {
				return llFuncDeclFrom(expr, top_def, ast)
			} else if astExprIsIdent(kwd, "define") {
				return llFuncDefFrom(expr, top_def, ast)
			}
		}
	}
	return nil
}

func llGlobalFrom(full_expr *AstExpr, top_def *AstDef, ast *Ast) LLGlobal {
	expr_form := full_expr.kind.(AstExprForm)
	assert(len(expr_form) == 3)
	lit_curl := expr_form[2].kind.(AstExprLitCurl)
	maybe_constant := astExprsFindKeyedValue(lit_curl, "#constant", ast)
	maybe_external := astExprsFindKeyedValue(lit_curl, "#external", ast)
	assert(maybe_constant != nil || maybe_external != nil)
	assert(!(maybe_constant != nil && maybe_external != nil))

	ret_global := LLGlobal{
		external: maybe_external != nil,
		constant: maybe_constant != nil,
		name:     astDefName(top_def),
		ty:       llTypeFrom(astExprSlashed(&expr_form[1]), full_expr, ast),
	}
	if maybe_external != nil {
		ret_global.name = astExprTaggedIdent(maybe_external)
		assert(len(ret_global.name) != 0)
	} else if maybe_constant != nil {
		ret_global.initializer = llExprFrom(maybe_constant, ast)
		_, is_lit_str := ret_global.initializer.(LLExprLitStr)
		assert(is_lit_str)
	}
	return ret_global
}

func llFuncDeclFrom(full_expr *AstExpr, top_def *AstDef, ast *Ast) LLFunc {
	expr_form := full_expr.kind.(AstExprForm)
	lit_curl_opts, _ := expr_form[2].kind.(AstExprLitCurl)
	assert(lit_curl_opts != nil)
	lit_curl_params := expr_form[4].kind.(AstExprLitCurl)
	ret_func := LLFunc{
		external:     true,
		ty:           llTypeFrom(astExprSlashed(&expr_form[3]), full_expr, ast),
		name:         astExprTaggedIdent(&expr_form[1]),
		params:       allocˇLLFuncParam(len(lit_curl_params)),
		basic_blocks: nil,
	}
	for i := range lit_curl_params {
		lhs, rhs := astExprFormSplit(&lit_curl_params[i], ":", true, true, true, ast)
		ret_func.params[i].name = astExprTaggedIdent(lhs)
		assert(len(ret_func.params[i].name) != 0) // source lang must provide param names for self-doc reasons...
		ret_func.params[i].name = nil             // ...but destination lang (LLVM-IR) doesn't need them
		ret_func.params[i].ty = llTypeFrom(astExprSlashed(rhs), rhs, ast)
	}
	assert(len(ret_func.name) != 0)
	return ret_func
}

func llFuncDefFrom(full_expr *AstExpr, top_def *AstDef, ast *Ast) LLFunc {
	expr_form := full_expr.kind.(AstExprForm)
	assert(len(expr_form) == 5)
	lit_curl_opts, _ := expr_form[1].kind.(AstExprLitCurl)
	assert(lit_curl_opts != nil)
	lit_curl_params := expr_form[3].kind.(AstExprLitCurl)
	lit_curl_blocks := expr_form[4].kind.(AstExprLitCurl)
	ret_func := LLFunc{
		name:         astDefName(top_def),
		external:     false,
		ty:           llTypeFrom(astExprSlashed(&expr_form[2]), &expr_form[2], ast),
		params:       allocˇLLFuncParam(len(lit_curl_params)),
		basic_blocks: allocˇLLBasicBlock(len(lit_curl_blocks)),
	}
	for i := range lit_curl_params {
		lhs, rhs := astExprFormSplit(&lit_curl_params[i], ":", true, true, true, ast)
		ret_func.params[i].name = astExprTaggedIdent(lhs)
		assert(len(ret_func.params[i].name) != 0)
		ret_func.params[i].ty = llTypeFrom(astExprSlashed(rhs), rhs, ast)
	}
	for i := range lit_curl_blocks {
		ret_func.basic_blocks[i] = llBlockFrom(&lit_curl_blocks[i], top_def, ast)
	}
	assert(len(ret_func.name) != 0)
	return ret_func
}

func llBlockFrom(pair_expr *AstExpr, top_def *AstDef, ast *Ast) LLBasicBlock {
	lhs, rhs := astExprFormSplit(pair_expr, ":", true, true, true, ast)
	lit_clip := rhs.kind.(AstExprLitClip)
	ret_block := LLBasicBlock{name: astExprTaggedIdent(lhs), instrs: allocˇLLInstr(len(lit_clip))}
	if len(ret_block.name) == 0 {
		counter++
		ret_block.name = uintToStr(counter, 10, 1, Str("b."))
	}
	for i := range lit_clip {
		ret_block.instrs[i] = llInstrFrom(&lit_clip[i], top_def, ast)
	}
	return ret_block
}

func llInstrFrom(expr *AstExpr, top_def *AstDef, ast *Ast) LLInstr {
	switch it := expr.kind.(type) {
	case AstExprForm:
		callee := astExprSlashed(&it[0])
		kwd := callee[0]
		if 1 == len(callee) {
			if astExprIsIdent(kwd, "ret") {
				assert(len(it) == 2)
				return LLInstrRet{expr: llExprFrom(&it[1], ast).(LLExprTyped)}
			} else if astExprIsIdent(kwd, "let") {
				assert(len(it) == 2)
				lit_curl := it[1].kind.(AstExprLitCurl)
				assert(len(lit_curl) == 1)
				lhs, rhs := astExprFormSplit(&lit_curl[0], ":", true, true, true, ast)
				return LLInstrLet{
					name:  astExprTaggedIdent(lhs),
					instr: llInstrFrom(rhs, top_def, ast),
				}
			} else if astExprIsIdent(kwd, "load") {
				assert(len(it) == 4)
				_ = it[1].kind.(AstExprLitCurl)
				ret_load := LLInstrLoad{
					ty:   llTypeFrom(astExprSlashed(&it[2]), expr, ast),
					expr: llExprFrom(&it[3], ast).(LLExprTyped),
				}
				return ret_load
			} else if astExprIsIdent(kwd, "call") {
				assert(len(it) == 5)
				_ = it[1].kind.(AstExprLitCurl)
				lit_clip_args := it[4].kind.(AstExprLitClip)
				ret_call := LLInstrCall{
					ty:     llTypeFrom(astExprSlashed(&it[3]), expr, ast),
					callee: llExprFrom(&it[2], ast),
					args:   allocˇLLExprTyped(len(lit_clip_args)),
				}
				for i := range lit_clip_args {
					ret_call.args[i] = llExprFrom(&lit_clip_args[i], ast).(LLExprTyped)
				}
				return ret_call
			} else if astExprIsIdent(kwd, "switch") {
				assert(len(it) == 4)
				lit_curl := it[3].kind.(AstExprLitCurl)
				ret_switch := LLInstrSwitch{
					comparee:           llExprFrom(&it[1], ast).(LLExprTyped),
					default_block_name: astExprTaggedIdent(&it[2]),
					cases:              allocˇLLSwitchCase(len(lit_curl)),
				}
				for i := range lit_curl {
					lhs, rhs := astExprFormSplit(&lit_curl[i], ":", true, true, true, ast)
					ret_switch.cases[i].expr = llExprFrom(lhs, ast).(LLExprTyped)
					ret_switch.cases[i].block_name = astExprTaggedIdent(rhs)
				}
				return ret_switch
			}
		}
	}
	return LLInstrComment{comment_text: astNodeSrcStr(&expr.base, ast)}
}

func llTypeFrom(expr_callee_slashed []*AstExpr, full_expr_for_err_msg *AstExpr, ast *Ast) LLType {
	if len(expr_callee_slashed) == 0 {
		return nil
	}
	ident, _ := expr_callee_slashed[0].kind.(AstExprIdent)
	if len(ident) == 0 {
		return nil
	}
	if ident[0] == 'P' && len(ident) == 1 {
		sub_type := llTypeFrom(expr_callee_slashed[1:], full_expr_for_err_msg, ast)
		if sub_type == nil {
			sub_type = LLTypeInt{bit_width: 8}
		}
		return LLTypePtr{ty: sub_type}
	} else if ident[0] == 'V' && len(ident) == 1 {
		return LLTypeVoid{}
	} else if ident[0] == 'A' && len(ident) == 1 && len(expr_callee_slashed) >= 3 {
		sub_type := llTypeFrom(expr_callee_slashed[2:], full_expr_for_err_msg, ast)
		if sub_type != nil {
			size_part_lit_int_str := astNodeSrcStr(&expr_callee_slashed[1].base, ast)
			return LLTypeArr{ty: sub_type, size: uintFromStr(size_part_lit_int_str)}
		}
	} else if ident[0] == 'I' {
		var bit_width uint64 = 64
		if len(ident) > 1 {
			bit_width = uintFromStr(ident[1:])
		}
		return LLTypeInt{bit_width: bit_width}
	}
	fail("expected prim type in: ", astNodeSrcStr(&full_expr_for_err_msg.base, ast))
	return nil
}

func llExprFrom(expr *AstExpr, ast *Ast) LLExpr {
	switch it := expr.kind.(type) {
	case AstExprLitInt:
		return LLExprLitInt(it)
	case AstExprLitStr:
		return LLExprLitStr(it)
	case AstExprForm:
		slashed := astExprSlashed(&it[0])
		if len(slashed) != 0 {
			kwd := slashed[0]
			ident, _ := kwd.kind.(AstExprIdent)
			if len(ident) != 0 {
				if len(it) == 2 && ident[0] >= 'A' && ident[0] <= 'Z' {
					ret_expr := LLExprTyped{ty: llTypeFrom(slashed, expr, ast), expr: llExprFrom(&it[1], ast)}
					_, is_void := ret_expr.ty.(LLTypeVoid)
					assert((ret_expr.expr != nil && !is_void) || (is_void && ret_expr.expr == nil))
					return ret_expr
				}
			}
		}

		if strEql(astNodeSrcStr(&it[0].base, ast), Str("/@/")) {
			assert(len(it) == 2)
			return LLExprIdentGlobal(it[1].kind.(AstExprIdent))
		}

		tag_lit := astExprTaggedIdent(expr)
		if tag_lit != nil {
			return LLExprIdentLocal(tag_lit)
		}
	}
	println("TODO Expr:\t", string(astNodeSrcStr(&expr.base, ast)))
	return nil
}

func llIdentFrom(expr *AstExpr, must_prefix byte, ast *Ast) Str {
	src_str := astNodeSrcStr(&expr.base, ast)
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
