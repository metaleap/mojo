package main

func parse(all_toks []Token, full_src Str) Ast {
	chunks := toksIndentBasedChunks(all_toks)
	ret_ast := Ast{
		src:  full_src,
		toks: all_toks,
		defs: ªAstDef(len(chunks)),
	}
	ret_ast.anns.num_def_toks = toksCount(all_toks, Str(":="), full_src)
	toks_idx := 0
	for i, this_chunk_toks := range chunks {
		ret_ast.defs[i].base.toks_idx = toks_idx
		ret_ast.defs[i].base.toks_len = len(this_chunk_toks)
		toks_idx += len(this_chunk_toks)
	}
	for i := range ret_ast.defs {
		this_top_def := &ret_ast.defs[i]
		this_top_def.anns.is_top_def = true
		parseDef(this_top_def, &ret_ast)
	}
	return ret_ast
}

func parseDef(dst_def *AstDef, dst_ast *Ast) {
	toks := astNodeToks(&dst_def.base, dst_ast)
	tok_idx_def := toksIndexOfIdent(toks, Str(":="), dst_ast.src)
	if tok_idx_def <= 0 || tok_idx_def == len(toks)-1 {
		fail(astNodeMsg("expected '<head_expr> := <body_expr>' in line ", &dst_def.base, dst_ast))
	}

	dst_def.head = parseExpr(toks[0:tok_idx_def], dst_def.base.toks_idx, dst_ast)
	{
		switch head_expr := dst_def.head.variant.(type) {
		case AstExprIdent:
			dst_def.anns.name = head_expr
		case AstExprForm:
			dst_def.anns.name = head_expr[0].variant.(AstExprIdent)
			for i := 1; i < len(head_expr); i++ {
				switch param_name := head_expr[i].variant.(type) {
				case AstExprIdent:
					if param_name[0] != '_' || param_name[1] == '_' {
						fail(astNodeMsg("param name doesn't begin with exactly one underscore in line ", &head_expr[i].base, dst_ast))
					}
				default:
					fail(astNodeMsg("unsupported def header in line ", &head_expr[i].base, dst_ast))
				}
			}
		default:
			fail(astNodeMsg("unsupported def header in line ", &dst_def.head.base, dst_ast))
		}
	}
	chunks_body := toksIndentBasedChunks(toks[tok_idx_def+1:])
	dst_def.defs = ªAstDef(len(chunks_body) - 1)
	toks_idx := dst_def.base.toks_idx + tok_idx_def + 1
	for i, this_chunk_toks := range chunks_body {
		if i == 0 {
			dst_def.body = parseExpr(this_chunk_toks, toks_idx, dst_ast)
		} else {
			sub_def := &dst_def.defs[i-1]
			sub_def.base.toks_idx = toks_idx
			sub_def.base.toks_len = len(this_chunk_toks)
			sub_def.anns.is_top_def = false
			parseDef(sub_def, dst_ast)
		}
		toks_idx += len(this_chunk_toks)
	}
}

func parseExpr(expr_toks []Token, all_toks_idx int, ast *Ast) AstExpr {
	assert(len(expr_toks) != 0)
	acc_ret := ªAstExpr(len(expr_toks))
	acc_len := 0
	expr_is_throng := (len(expr_toks)-1 == tokThrong(expr_toks, 0, ast.src))
	for i := 0; i < len(expr_toks); i++ {
		idx_throng_end := i
		if !expr_is_throng {
			idx_throng_end = tokThrong(expr_toks, i, ast.src)
		}
		if idx_throng_end > i {
			acc_ret[acc_len] = parseExpr(expr_toks[i:idx_throng_end+1], all_toks_idx+i, ast)
			acc_ret[acc_len].anns.toks_throng = true
			i = idx_throng_end // loop header will increment
		} else {
			switch tok_kind := expr_toks[i].kind; tok_kind {
			case tok_kind_lit_int:
				tok_str := toksSrcStr(expr_toks[i:i+1], ast.src)
				node_base := astNodeFrom(all_toks_idx+i, 1)
				acc_ret[acc_len] = AstExpr{
					base:    node_base,
					variant: AstExprLitInt(parseExprLitInt(&node_base, ast, tok_str)),
				}
			case tok_kind_lit_str_double:
				tok_str := toksSrcStr(expr_toks[i:i+1], ast.src)
				node_base := astNodeFrom(all_toks_idx+i, 1)
				acc_ret[acc_len] = AstExpr{
					base:    node_base,
					variant: AstExprLitStr(parseExprLitStr(&node_base, ast, tok_str, '"')),
				}
			case tok_kind_lit_str_single:
				tok_str := toksSrcStr(expr_toks[i:i+1], ast.src)
				node_base := astNodeFrom(all_toks_idx+i, 1)
				char_str := parseExprLitStr(&node_base, ast, tok_str, '\'')
				char_rune := rune(-1)
				for _, r := range string(char_str) {
					if char_rune < 0 {
						char_rune = r
					} else {
						fail(astNodeMsg("character literal too long in line ", &node_base, ast))
					}
				}
				if char_rune < 0 {
					fail(astNodeMsg("character literal too short in line ", &node_base, ast))
				} else {
					acc_ret[acc_len] = AstExpr{base: node_base, variant: AstExprLitInt(char_rune)}
				}
			case tok_kind_sep_bcurly_open, tok_kind_sep_bsquare_open, tok_kind_sep_bparen_open:
				idx_close := toksIndexOfMatchingBracket(expr_toks[i:])
				assert(idx_close > 0) // the other-case will have been caught already by toksCheckBrackets
				idx_close += i
				if tok_kind == tok_kind_sep_bparen_open {
					inside_parens_toks := expr_toks[i+1 : idx_close]
					if len(inside_parens_toks) == 0 {
						acc_ret[acc_len] = AstExpr{variant: AstExprIdent("()"), base: astNodeFrom(all_toks_idx+i, 2)}
					} else {
						acc_ret[acc_len] = parseExpr(inside_parens_toks, all_toks_idx+i+1, ast)
						acc_ret[acc_len].anns.parensed++
						// still want the paren toks captured in node base:
						acc_ret[acc_len].base.toks_idx = all_toks_idx + i
						acc_ret[acc_len].base.toks_len += 2
					}
				} else {
					acc_ret[acc_len] = AstExpr{base: astNodeFrom(all_toks_idx+i, 1+(idx_close-i))}
					bracketed_exprs := parseExprsDelimited(expr_toks[i+1:idx_close], all_toks_idx+i+1, tok_kind_sep_comma, ast)
					switch tok_kind {
					case tok_kind_sep_bcurly_open:
						acc_ret[acc_len].variant = AstExprLitObj(bracketed_exprs)
					case tok_kind_sep_bsquare_open:
						acc_ret[acc_len].variant = AstExprLitList(bracketed_exprs)
					default:
						unreachable()
					}
				}
				i = idx_close // loop header will increment
			case tok_kind_comment:
				unreachable()
			default:
				tok_str := toksSrcStr(expr_toks[i:i+1], ast.src)
				acc_ret[acc_len] = AstExpr{
					base:    astNodeFrom(all_toks_idx+i, 1),
					variant: AstExprIdent(tok_str),
				}
			}
		}
		acc_len++
	}

	assert(acc_len != 0)
	if acc_len == 1 {
		return acc_ret[0]
	}
	return AstExpr{
		base:    astNodeFrom(all_toks_idx, len(expr_toks)),
		variant: AstExprForm(acc_ret[0:acc_len]),
	}
}

func parseExprLitInt(node_base *AstNode, ast *Ast, lit_src Str) uint64 {
	r, ok := uintFromStr(lit_src)
	if !ok {
		fail(astNodeMsg("malformed integer literal in line ", node_base, ast))
	}
	return r
}

func parseExprLitStr(node_base *AstNode, ast *Ast, lit_src Str, delim_char byte) Str {
	assert(len(lit_src) >= 2 && lit_src[0] == delim_char && lit_src[len(lit_src)-1] == delim_char) // tokenizer-guaranteed
	ret_str := ªbyte(len(lit_src) - 2)
	ret_len := 0
	for i, i_cry := 1, false; i < len(lit_src)-1; i++ {
		if lit_src[i] != '\\' {
			ret_str[ret_len] = lit_src[i]
		} else {
			idx_end := i + 4
			i_cry = idx_end > len(lit_src)-1
			if !i_cry {
				base10digits := lit_src[i+1 : idx_end]
				i += 3
				integer, ok := uintFromStr(base10digits)
				i_cry = (!ok) || integer >= 256
				ret_str[ret_len] = byte(integer)
			}
		}
		if i_cry {
			fail(astNodeMsg("expected 3-digit base-10 integer decimal 000..256 following escape in line ", node_base, ast))
		}
		ret_len++
	}
	return ret_str[0:ret_len]
}

func parseExprsDelimited(toks []Token, all_toks_idx int, tok_kind_sep TokenKind, ast *Ast) []AstExpr {
	if len(toks) == 0 {
		return ªAstExpr(0)
	}
	per_item_toks := toksSplit(toks, ast.src, tok_kind_sep)
	ret_exprs := ªAstExpr(len(per_item_toks))
	ret_idx := 0
	toks_idx := all_toks_idx
	for _, this_item_toks := range per_item_toks {
		if len(this_item_toks) == 0 {
			toks_idx++ // the 1 for the comma
		} else {
			ret_exprs[ret_idx] = parseExpr(this_item_toks, toks_idx, ast)
			toks_idx += 1 + len(this_item_toks) // the 1 for the comma
			ret_idx++
		}
	}
	return ret_exprs[0:ret_idx]
}
