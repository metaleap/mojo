package main

type CtxLowerToLL struct {
	ll_mod      *LLModule
	num_globals int
	num_funcs   int
}

func llModuleFromAst(ast *Ast) LLModule {
	for i := range ast.defs {
		if top_def := &ast.defs[i]; strEql(astDefName(top_def), Str("main")) {
			ret_mod := LLModule{
				target_datalayout: Str(llmodule_default_target_datalayout),
				target_triple:     Str(llmodule_default_target_triple),
				globals:           allocˇLLGlobal(len(ast.defs)),
				funcs:             allocˇLLFunc(len(ast.defs)),
			}
			ctx := CtxLowerToLL{ll_mod: &ret_mod}
			ret_mod.globals = ret_mod.globals[0:ctx.num_globals]
			ret_mod.funcs = ret_mod.funcs[0:ctx.num_funcs]
			lowerToLL(&ctx, top_def)
			return ret_mod
		}
	}
	panic("main not found")
}

func lowerToLL(ctx *CtxLowerToLL, top_def *AstDef) {

}
