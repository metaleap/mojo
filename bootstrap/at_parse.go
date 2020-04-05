package main

func parse(all_toks Tokens, full_src Str) Ast {
	chunks := toksIndentBasedChunks(all_toks)
	ret_ast := Ast{
		src:  full_src,
		toks: all_toks,
		defs: allocˇAstDef(len(chunks) * 2),
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
		gathered := allocˇStrNamed(4)
		num := astDefGatherAndRewriteLitStrs(this_top_def, gathered, 0)
		for j := range gathered[0:num] {
			ret_ast.defs[len(chunks)+num_str_lits] = AstDef{
				head:       AstExpr{kind: AstExprIdent(gathered[j].name)},
				body:       AstExpr{kind: AstExprLitStr(gathered[j].value)},
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
	tok_idx_def := toksIndexOfKind(toks, tok_kind_sep_def)
	if tok_idx_def <= 0 || tok_idx_def == len(toks)-1 {
		fail("expected '<head_expr> := <body_expr>', near:\n", astNodeSrcStr(&dst_def.base, full_src, all_toks))
	}

	dst_def.head = parseExpr(full_src, all_toks, toks[0:tok_idx_def], dst_def.base.toks_idx)
	chunks_body := toksIndentBasedChunks(toks[tok_idx_def+1:])
	dst_def.defs = allocˇAstDef(len(chunks_body) - 1)
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

func parseExpr(full_src Str, all_toks Tokens, expr_toks Tokens, all_toks_idx int) AstExpr {
	acc_ret := allocˇAstExpr(len(expr_toks))
	acc_len := 0
	for i := 0; i < len(expr_toks); i++ {
		switch tok_kind := expr_toks[i].kind; tok_kind {
		case tok_kind_lit_int:
			tok_str := toksSrcStr(expr_toks[i:i+1], full_src)
			acc_ret[acc_len] = AstExpr{
				base: astNodeFrom(all_toks_idx+i, 1),
				kind: AstExprLitInt(parseExprLitInt(tok_str)),
			}
		case tok_kind_lit_str:
			tok_str := toksSrcStr(expr_toks[i:i+1], full_src)
			acc_ret[acc_len] = AstExpr{
				base: astNodeFrom(all_toks_idx+i, 1),
				kind: AstExprLitStr(parseExprLitStr(tok_str)),
			}
		case tok_kind_sep_bcurly_open, tok_kind_sep_bsquare_open, tok_kind_sep_bparen_open:
			idx_close := i + toksIndexOfMatchingBracket(expr_toks[i:])
			if idx_close < 0 {
				fail("no matching closing bracket near:", toksSrcStr(expr_toks, full_src))
			}
			if tok_kind == tok_kind_sep_bparen_open {
				acc_ret[acc_len] = parseExpr(full_src, all_toks,
					expr_toks[i+1:idx_close], all_toks_idx+i+1)
			} else {
				acc_ret[acc_len] = AstExpr{base: astNodeFrom(all_toks_idx+i, 1+(idx_close-i))}
				bracketed_exprs := parseExprsDelimited(full_src, all_toks,
					expr_toks[i+1:idx_close], all_toks_idx+i+1, tok_kind_sep_comma)
				switch tok_kind {
				case tok_kind_sep_bcurly_open:
					acc_ret[acc_len].kind = AstExprLitCurl(bracketed_exprs)
				case tok_kind_sep_bsquare_open:
					acc_ret[acc_len].kind = AstExprLitClip(bracketed_exprs)
				default:
					unreachable()
				}
			}
			i = idx_close // loop header will increment
		default:
			tok_str := toksSrcStr(expr_toks[i:i+1], full_src)
			acc_ret[acc_len] = AstExpr{
				base: astNodeFrom(all_toks_idx+i, 1),
				kind: AstExprIdent(tok_str),
			}
		}
		acc_len++
	}

	assert(acc_len != 0)
	if acc_len == 1 {
		return acc_ret[0]
	}
	return AstExpr{
		base: astNodeFrom(all_toks_idx, len(expr_toks)),
		kind: AstExprForm(acc_ret[0:acc_len]),
	}
}

func parseExprLitInt(lit_src Str) uint64 {
	return uintFromStr(lit_src)
}

func parseExprLitStr(lit_src Str) Str {
	assert(len(lit_src) >= 2 && lit_src[0] == '"' && lit_src[len(lit_src)-1] == '"')
	ret_str := allocˇu8(len(lit_src) - 2)
	ret_len := 0
	for i := 1; i < len(lit_src)-1; i++ {
		if lit_src[i] != '\\' {
			ret_str[ret_len] = lit_src[i]
		} else {
			int10str := lit_src[i+1 : i+4]
			i += 3
			integer := uintFromStr(int10str)
			assert(integer < 256)
			ret_str[ret_len] = byte(integer)
		}
		ret_len++
	}
	return ret_str[0:ret_len]
}

func parseExprsDelimited(full_src Str, all_toks Tokens, toks Tokens, all_toks_idx int, tok_kind_sep TokenKind) []AstExpr {
	per_item_toks := toksSplit(toks, full_src, tok_kind_sep)
	ret_exprs := allocˇAstExpr(len(per_item_toks))
	toks_idx := all_toks_idx
	for i, this_item_toks := range per_item_toks {
		ret_exprs[i] = parseExpr(full_src, all_toks, this_item_toks, toks_idx)
		toks_idx += len(this_item_toks)
	}
	return ret_exprs
}
