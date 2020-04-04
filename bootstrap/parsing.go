package main

func parse(all_toks Tokens, full_src Str) Ast {
	chunks := toksIndentBasedChunks(all_toks)
	ret_ast := Ast{
		src:  full_src,
		toks: all_toks,
		defs: make([]AstDef, len(chunks)*2),
	}
	toks_idx := 0
	for i, chunk_toks := range chunks {
		ret_ast.defs[i].base.toks_idx = toks_idx
		ret_ast.defs[i].base.toks_len = len(chunk_toks)
		toks_idx += len(chunk_toks)
	}
	for i := range ret_ast.defs[0:len(chunks)] {
		top_def := &ret_ast.defs[i]
		top_def.is_top_def = true
		parseDef(full_src, all_toks, top_def)
	}

	num_str_lits := 0
	for i := range ret_ast.defs[0:len(chunks)] {
		top_def := &ret_ast.defs[i]
		str_lit_and_new_name_pairs := make([][2]Str, 4)
		num := astDefGatherAndRewriteLitStrs(top_def, str_lit_and_new_name_pairs, 0)
		for j := range str_lit_and_new_name_pairs[0:num] {
			str_lit_and_new_name := &str_lit_and_new_name_pairs[j]
			ret_ast.defs[len(chunks)+num_str_lits] = AstDef{
				head:       AstExpr{kind: AstExprIdent(str_lit_and_new_name[1])},
				body:       AstExpr{kind: AstExprLitStr(str_lit_and_new_name[0])},
				is_top_def: true,
			}
			num_str_lits++
		}
	}

	ret_ast.defs = ret_ast.defs[0 : len(chunks)+num_str_lits]
	return ret_ast
}

func parseDef(full_src Str, all_toks Tokens, dst_def *AstDef) {
	fail("TODO")
}
