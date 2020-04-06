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
				ret_mod.globals[num_globals] = llGlobalFromExtVar(&top_def.scope, body)
				num_globals++
			} else if strEql(callee, Str("/extFun")) {
				ret_mod.funcs[num_funcs] = llFuncDeclFrom(&top_def.scope, body)
				num_funcs++
			} else if strEql(callee, Str("/defFun")) {
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

func llGlobalFromExtVar(scope *AstScopes, form AstExprForm) LLGlobal {
	assert(len(form) == 3)
	name := astScopesResolve(scope, form[1].kind.(AstExprIdent), -1).(*AstDef).body.kind.(AstExprLitStr)
	return LLGlobal{
		name:     name,
		external: true,
		ty:       llTypeFrom(form[2].kind.(AstExprIdent)),
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

func llFuncDeclFrom(scope *AstScopes, form AstExprForm) LLFunc {
	assert(len(form) == 4)
	name := astScopesResolve(scope, form[1].kind.(AstExprIdent), -1).(*AstDef).body.kind.(AstExprLitStr)
	lit_curl := form[3].kind.(AstExprLitCurl)
	ret_decl := LLFunc{
		name:   name,
		ty:     llTypeFrom(form[2].kind.(AstExprIdent)),
		params: allocˇLLFuncParam(len(lit_curl)),
	}
	for i := range lit_curl {
		name_and_expr := lit_curl[i].kind.(AstExprForm)
		assert(len(name_and_expr) == 3 && strEql(name_and_expr[1].kind.(AstExprIdent), Str(":")))
		ret_decl.params[i].name = llIdentFrom(&name_and_expr[0])
		ret_decl.params[i].ty = llTypeFrom(name_and_expr[2].kind.(AstExprIdent))
	}
	return ret_decl
}

func llIdentFrom(expr *AstExpr) Str {
	name := expr.kind.(AstExprIdent)
	assert(name[0] == '#')
	name = name[1:]
	return name
}

func llTypeFrom(ident AstExprIdent) LLType {
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
