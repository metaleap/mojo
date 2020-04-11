package main

type CtxLowerToLL struct {
	ast         *Ast
	scope       *AstScopes
	ll_mod      *LLModule
	cur_def     *AstDef
	num_globals int
	num_funcs   int
}

func llModuleFromAst(ast *Ast) LLModule {
	ret_mod := LLModule{
		target_datalayout: Str(llmodule_default_target_datalayout),
		target_triple:     Str(llmodule_default_target_triple),
		globals:           allocˇLLGlobal(len(ast.defs)),
		funcs:             allocˇLLFunc(len(ast.defs)),
	}
	ctx := CtxLowerToLL{ll_mod: &ret_mod, ast: ast, scope: &ast.scope}
	ret_mod.globals = ret_mod.globals[0:ctx.num_globals]
	ret_mod.funcs = ret_mod.funcs[0:ctx.num_funcs]
	_ = lowerToLL(&ctx, &AstExpr{kind: AstExprIdent("main")})
	return ret_mod
}

func lowerToLL(ctx *CtxLowerToLL, ast_expr *AstExpr) Any {
	if tagged_ident := astExprTaggedIdent(ast_expr); tagged_ident != nil {
		println("TAGGED\t", string(astNodeSrcStr(&ast_expr.base, ctx.ast)))
		panic(tagged_ident)
	} else if slashed := astExprSlashed(ast_expr); slashed != nil {
		kwd := slashed[0].kind.(AstExprIdent)
		if strEq(kwd, "define") {
			func_def := LLFunc{external: false}

			ret_ptr := &ctx.ll_mod.funcs[ctx.num_funcs]
			*ret_ptr = func_def
			ctx.num_funcs++
			return ret_ptr
		} else if strEq(kwd, "declare") {
			func_decl := LLFunc{external: true}

			ret_ptr := &ctx.ll_mod.funcs[ctx.num_funcs]
			*ret_ptr = func_decl
			ctx.num_funcs++
			return ret_ptr
		}
		println("SLASHED\t", string(astNodeSrcStr(&ast_expr.base, ctx.ast)))
		panic(slashed)
	}

	switch expr := ast_expr.kind.(type) {
	case AstExprLitCurl:
		return expr
	case AstExprIdent:
		if resolved := astScopesResolve(ctx.scope, expr, -1); resolved != nil {
			if found_global := llModuleFindGlobal(ctx.ll_mod, nil, resolved.ref_def, ctx.num_globals); found_global != nil {
				return found_global
			}
			if found_func := llModuleFindFunc(ctx.ll_mod, nil, resolved.ref_def, ctx.num_funcs); found_func != nil {
				return found_func
			}
			old_scope, old_def := ctx.scope, ctx.cur_def
			ctx.scope, ctx.cur_def = &resolved.ref_def.scope, resolved.ref_def
			evald := lowerToLL(ctx, &resolved.ref_def.body)
			ctx.scope, ctx.cur_def = old_scope, old_def
			switch it := evald.(type) {
			case LLGlobal:
				if len(it.name) == 0 {
					it.name = expr
				}
				if llModuleFindGlobal(ctx.ll_mod, it.name, nil, ctx.num_globals) != nil || llModuleFindFunc(ctx.ll_mod, it.name, nil, ctx.num_globals) != nil {
					fail("shadowing of identifier '", it.name, "' near:\n", astNodeSrcStr(&ast_expr.base, ctx.ast))
				}
				ctx.ll_mod.globals[ctx.num_globals] = it
				ctx.num_globals++
				return &ctx.ll_mod.globals[ctx.num_globals-1]
			case LLFunc:
				if len(it.name) == 0 {
					it.name = expr
				}
				if llModuleFindGlobal(ctx.ll_mod, it.name, nil, ctx.num_globals) != nil || llModuleFindFunc(ctx.ll_mod, it.name, nil, ctx.num_globals) != nil {
					fail("shadowing of identifier '", it.name, "' near:\n", astNodeSrcStr(&ast_expr.base, ctx.ast))
				}
				ctx.ll_mod.funcs[ctx.num_funcs] = it
				ctx.num_funcs++
				return &ctx.ll_mod.funcs[ctx.num_funcs-1]
			}
		}
	case AstExprForm:
		args_exprs := expr[1:]
		switch callee := lowerToLL(ctx, &expr[0]).(type) {
		case LLFunc:
			assert(len(args_exprs) == 4)
			if callee.external {
				panic("EXT_FN")
			} else {
				_ = args_exprs[0].kind.(AstExprLitCurl)
				lit_curl_params := args_exprs[2].kind.(AstExprLitCurl)
				_ = lit_curl_params
			}
		default:
			panic(callee)
		}
		args := allocˇAny(len(expr) - 1)
		for i := range args {
			args[i] = lowerToLL(ctx, &expr[i+1])
		}
	}
	println(string(astNodeSrcStr(&ast_expr.base, ctx.ast)))
	panic(ast_expr.kind)
}
