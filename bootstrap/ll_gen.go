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
	return LLGlobal{
		name:     astScopesResolve(scope, form[1].kind.(AstExprIdent), -1).(*AstDef).body.kind.(AstExprLitStr),
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

func llTypeFrom(ident AstExprIdent) LLType {
	if strEql(ident, Str("/P")) {
		return LLTypePtr{ty: LLTypeInt{bit_width: 8}}
	}
	panic("TODO: llTypeFrom " + string(ident))
}
