package main

func parse(all_toks Tokens, full_src Str) Ast {
	chunks := toksIndentBasedChunks(all_toks)
	ret_ast := Ast{
		src:  full_src,
		toks: all_toks,
		defs: make([]AstDef, len(chunks)*2),
	}
	toks_idx := 0
	for i, this_chunk_toks := range chunks {
		ret_ast.defs[i].base.toks_idx = toks_idx
		ret_ast.defs[i].base.toks_len = len(this_chunk_toks)
		toks_idx += len(this_chunk_toks)
	}
	for i := range ret_ast.defs[0:len(chunks)] {
		this_top_def := &ret_ast.defs[i]
		this_top_def.is_top_def = true
		parseDef(full_src, all_toks, this_top_def)
	}

	num_str_lits := 0
	for i := range ret_ast.defs[0:len(chunks)] {
		this_top_def := &ret_ast.defs[i]
		str_lit_and_new_name_pairs := make([][2]Str, 4)
		num := astDefGatherAndRewriteLitStrs(this_top_def, str_lit_and_new_name_pairs, 0)
		for j := range str_lit_and_new_name_pairs[0:num] {
			this_str_lit_and_new_name := &str_lit_and_new_name_pairs[j]
			ret_ast.defs[len(chunks)+num_str_lits] = AstDef{
				head:       AstExpr{kind: AstExprIdent(this_str_lit_and_new_name[1])},
				body:       AstExpr{kind: AstExprLitStr(this_str_lit_and_new_name[0])},
				is_top_def: true,
			}
			num_str_lits++
		}
	}

	ret_ast.defs = ret_ast.defs[0 : len(chunks)+num_str_lits]
	return ret_ast
}

func parseDef(full_src Str, all_toks Tokens, dst_def *AstDef) {
	toks := astNodeToks(&dst_def.base, all_toks)
	tok_idx_def := toksIndexOfFirst(toks, tok_kind_sep_def)
	if tok_idx_def <= 0 || tok_idx_def == len(toks)-1 {
		fail("expected '<head_expr> := <body_expr>', near:\n", astNodeSrcStr(&dst_def.base, full_src, all_toks))
	}

	dst_def.head = parseExpr(full_src, all_toks, toks[0:tok_idx_def], dst_def.base.toks_idx)
	chunks_body := toksIndentBasedChunks(toks[tok_idx_def+1:])
	dst_def.defs = make([]AstDef, len(chunks_body)-1)
	toks_idx := dst_def.base.toks_idx + tok_idx_def + 1
	for i, this_chunk_toks := range chunks_body {
		if i == 0 {
			dst_def.body = parseExpr(full_src, all_toks, this_chunk_toks, toks_idx)
		} else {
			sub_def := &dst_def.defs[i-1]
			sub_def.base.toks_idx = toks_idx
			sub_def.base.toks_len = len(this_chunk_toks)
			sub_def.is_top_def = false
			parseDef(full_src, all_toks, sub_def)
		}
		toks_idx += len(this_chunk_toks)
	}
}

func parseExpr(full_src Str, all_toks Tokens, expr_toks Tokens, toks_idx int) AstExpr {
	panic("TODO")
}