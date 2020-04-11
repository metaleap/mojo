package main

type CtxLowerToLL struct {
	ast         *Ast
	ll_mod      *LLModule
	num_globals int
	num_funcs   int
}

func llModuleFromAst(ast *Ast) LLModule {
	for i := range ast.defs {
		if top_def := &ast.defs[i]; strEql(top_def.anns.name, Str("main")) {
			ret_mod := LLModule{
				target_datalayout: Str(llmodule_default_target_datalayout),
				target_triple:     Str(llmodule_default_target_triple),
				globals:           allocˇLLGlobal(len(ast.defs)),
				funcs:             allocˇLLFunc(len(ast.defs)),
			}
			ctx := CtxLowerToLL{ll_mod: &ret_mod, ast: ast}
			ret_mod.globals = ret_mod.globals[0:ctx.num_globals]
			ret_mod.funcs = ret_mod.funcs[0:ctx.num_funcs]
			_ = lowerToLL(&ctx, top_def)
			return ret_mod
		}
	}
	panic("main not found")
}

func lowerToLL(ctx *CtxLowerToLL, ast_something Any) Any {
	switch it := ast_something.(type) {
	case *AstDef:
		body := lowerToLL(ctx, it.body)
		switch ll_top_level_artifact := body.(type) {
		case LLGlobal:
		case LLFunc:
		default:
			panic(ll_top_level_artifact)
		}
		return body
	case AstExpr:
		if tagged_ident, _ := astExprTaggedIdent(&it); tagged_ident != nil {

		} else if slashed, _ := astExprSlashed(&it); slashed != nil {
		}

		switch expr := it.kind.(type) {
		case AstExprLitCurl:
			return expr
		case AstExprForm:
			args := allocˇAny(len(expr) - 1)
			for i := range args {
				args[i] = lowerToLL(ctx, expr[i+1])
			}
			switch callee := lowerToLL(ctx, expr[0]).(type) {
			default:
				panic(callee)
			}
		}
		println(string(astNodeSrcStr(&it.base, ctx.ast)))
		panic(it.kind)
	}
	panic(ast_something)
}
