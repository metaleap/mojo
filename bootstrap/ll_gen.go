package main

const llmodule_default_target_datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
const llmodule_default_target_triple = "x86_64-unknown-linux-gnu"

func llGenModule(ast *Ast) LLModule {
	ret_mod := LLModule{
		target_datalayout: Str(llmodule_default_target_datalayout),
		target_triple:     Str(llmodule_default_target_triple),
		globals:           allocˇLLGlobal(len(ast.defs)),
		funcs:             allocˇLLFunc(len(ast.defs)),
	}
	num_globals, num_funcs := 0, 0
	for i := range ast.defs {
		top_def := &ast.defs[i]
		switch ll_sth := llGen(&top_def.body, top_def, ast).(type) {
		case LLGlobal:
			ret_mod.globals[num_globals] = ll_sth
			num_globals++
		case LLFunc:
			ret_mod.funcs[num_funcs] = ll_sth
			num_funcs++
		default:
			println(string(astDefName(top_def)), ll_sth)
		}
	}
	ret_mod.globals = ret_mod.globals[0:num_globals]
	ret_mod.funcs = ret_mod.funcs[0:num_funcs]
	return ret_mod
}

func llGen(expr *AstExpr, top_def *AstDef, ast *Ast) Any {
	switch it := expr.kind.(type) {
	case AstExprForm:
		callee := astExprSlashed(&it[0])
		if len(callee) == 0 {
			fail("expected some '/...' prim call in:\n", astNodeSrcStr(&expr.base, ast))
		}
		if 1 == len(callee) {
			if astExprIsIdent(callee[0], "global") {
				return llGenGlobal(expr, top_def, ast)
				// } else if astExprIsIdent(callee[0], "declare") {
				// 	return llGenFuncDecl(expr, top_def, ast)
				// } else if astExprIsIdent(callee[0], "define") {
				// 	return llGenFuncDef(expr, top_def, ast)
			}
		}
		if ident, is_ident := callee[0].kind.(AstExprIdent); is_ident && ident[0] >= 'A' && ident[0] <= 'Z' {
			return llGenType(callee, ast)
		}
	}
	return nil
}

func llGenGlobal(full_expr *AstExpr, top_def *AstDef, ast *Ast) LLGlobal {
	expr_form := full_expr.kind.(AstExprForm)
	ret_global := LLGlobal{
		external: true,
		name:     astDefName(top_def),
		ty:       llGenType(astExprSlashed(&expr_form[1]), ast),
	}
	if ret_global.ty == nil {
		fail("expected prim type in:\n", astNodeSrcStr(&full_expr.base, ast))
	}
	return ret_global
}

func llGenFuncDecl(full_expr *AstExpr, top_def *AstDef, ast *Ast) LLFunc {
	return LLFunc{name: astDefName(top_def)}
}

func llGenFuncDef(full_expr *AstExpr, top_def *AstDef, ast *Ast) LLFunc {
	return LLFunc{name: astDefName(top_def)}
}

func llGenType(expr_callee_slashed []*AstExpr, ast *Ast) LLType {
	if len(expr_callee_slashed) != 0 && astExprIsIdent(expr_callee_slashed[0], "P") {
		sub_type := llGenType(expr_callee_slashed[1:], ast)
		if sub_type == nil {
			sub_type = LLTypeInt{bit_width: 8}
		}
		return LLTypePtr{ty: sub_type}
	}
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
