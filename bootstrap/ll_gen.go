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
		case AstExprForm:
			callee := body[0].kind.(AstExprIdent)
			if strEql(callee, Str("/global")) {
				// ret_mod.globals[num_globals] = llGlobalFrom(top_def, ast)
				// num_globals++
			} else if strEql(callee, Str("/declare")) {
				// ret_mod.funcs[num_funcs] = llFuncDeclFrom(top_def, ast)
				// num_funcs++
			} else if strEql(callee, Str("/define")) {
				// ret_mod.funcs[num_funcs] = llFuncDefFrom(top_def, ast)
				// num_funcs++
			} else {
				fail(astNodeSrcStr(&top_def.base, ast.src, ast.toks))
			}
		default:
			fail(astNodeSrcStr(&top_def.base, ast.src, ast.toks))
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

func llGlobalFrom(top_def *AstDef, ast *Ast) LLGlobal {
	form := top_def.body.kind.(AstExprForm)
	assert(len(form) == 3)
	// lit_curl = form[2].kind.(AstExprLitCurl)
	// var maybe_extern Str

	c_name := astScopesResolve(&top_def.scope, form[1].kind.(AstExprIdent), -1).(*AstDef).body.kind.(AstExprLitStr)
	return LLGlobal{
		name:     c_name,
		external: true,
		ty:       nil,
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
		ty:       nil,
		params:   allocˇLLFuncParam(len(lit_curl)),
	}
	for i := range lit_curl {
		ret_decl.params[i].name = nil
		ret_decl.params[i].ty = nil
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
		ty:           nil,
		params:       allocˇLLFuncParam(len(lit_args)),
		basic_blocks: allocˇLLBasicBlock(len(lit_body)),
	}
	for i := range lit_args {
		arg_name, _ := astExprFormSplit(&lit_args[i], Str(":"), true, true, true, ast)
		ret_def.params[i].name = llIdentFrom(arg_name, '#', ast)
		ret_def.params[i].ty = nil
	}
	for i := range lit_body {
		block_name, _ := astExprFormSplit(&lit_body[i], Str(":"), true, true, true, ast)
		ret_def.basic_blocks[i].name = llIdentFrom(block_name, '#', ast)
		ret_def.basic_blocks[i].stmts = nil // created later in `llFuncDefPopulateBlocks`
	}
	return ret_def
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
