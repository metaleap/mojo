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
				ret_mod.globals[num_globals] = llGlobalFromExtVar(body, &top_def.scope)
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

func llGlobalFromExtVar(form AstExprForm, scope *AstScopes) LLGlobal {
	assert(len(form) == 3)
	c_name := astScopesResolve(scope, form[1].kind.(AstExprIdent), -1).(*AstDef).body.kind.(AstExprLitStr)
	return LLGlobal{
		name:     c_name,
		external: true,
		ty:       llTypeFrom(&form[2]),
	}
}

func llGlobalFromLitStr(name Str, body AstExprLitStr) LLGlobal {
	return LLGlobal{
		name:        name,
		constant:    true,
		ty:          LLTypeArr{size: len(body), ty: LLTypeInt{bit_width: 8}},
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
		ty:       llTypeFrom(&form[2]),
		params:   allocˇLLFuncParam(len(lit_curl)),
	}
	for i := range lit_curl {
		_, ty := astExprFormSplit(&lit_curl[i], Str(":"), true, true, true, ast)
		ret_decl.params[i].name = nil
		ret_decl.params[i].ty = llTypeFrom(ty)
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
		ty:           llTypeFrom(&form[1]),
		params:       allocˇLLFuncParam(len(lit_args)),
		basic_blocks: allocˇLLBasicBlock(len(lit_body)),
	}
	for i := range lit_args {
		arg_name, arg_ty := astExprFormSplit(&lit_args[i], Str(":"), true, true, true, ast)
		ret_def.params[i].name = llIdentFrom(arg_name)
		ret_def.params[i].ty = llTypeFrom(arg_ty)
	}
	for i := range lit_body {
		block_name, block_stmts := astExprFormSplit(&lit_body[i], Str(":"), true, true, true, ast)
		ret_def.basic_blocks[i].name = llIdentFrom(block_name)
		stmts := block_stmts.kind.(AstExprLitClip)
		ret_def.basic_blocks[i].stmts = allocˇLLStmt(len(stmts))
		ret_def.basic_blocks[i].stmts = ret_def.basic_blocks[i].stmts[0:0]
	}
	return ret_def
}

func llTypeFrom(expr *AstExpr) LLType {
	ident := expr.kind.(AstExprIdent)
	assert(ident[0] == '/')
	if ident[1] == 'V' {
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
}

func llIdentFrom(expr *AstExpr) Str {
	name := expr.kind.(AstExprIdent)
	assert(name[0] == '#')
	name = name[1:]
	return name
}
