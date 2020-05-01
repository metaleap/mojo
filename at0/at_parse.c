#pragma once
#include "utils_and_libc_deps.c"
#include "fs_io.c"
#include "at_toks.c"
#include "at_ast.c"


#define asts_capacity 4
typedef struct CtxParseAsts {
    Asts asts;
    Strs src_file_paths;
} CtxParseAsts;


Ast astParse(Tokens const all_toks, Str const full_src, Str const src_file_path);
AstExpr astParseExpr(Tokens const toks, UInt const all_toks_idx, Ast* const ast, Bool);
void astParseDef(AstDef* const dst_def, Ast* const ast);

void loadAndParseRootSourceFileAndImports(CtxParseAsts* const ctx) {
    while (ctx->src_file_paths.len > ctx->asts.len) {
        Str const next_src_file_path = ctx->src_file_paths.at[ctx->asts.len];
        Str const full_src = readFile(next_src_file_path);

        Tokens const toks = tokenize(full_src, false, next_src_file_path);
        toksVerifyBrackets(toks);

        Ast ast = astParse(toks, full_src, next_src_file_path);
        if (ctx->asts.len == asts_capacity)
            ·fail(str("TODO: increase asts_capacity"));
        ·append(ctx->asts, ast);

        for (UInt i = 0; i < ast.anns.incl_file_paths.len; i += 1) {
            Str const incl_file_path = relPathFromRelPath(next_src_file_path, ast.anns.incl_file_paths.at[i]);
            if (!strIn(incl_file_path, ctx->src_file_paths)) {
                if (ctx->src_file_paths.len == asts_capacity)
                    ·fail(str("TODO: increase asts_capacity"));
                ·append(ctx->src_file_paths, incl_file_path);
            }
        }
    }
}

Ast astParse(Tokens const all_toks, Str const full_src, Str const src_file_path) {
    Tokenss const chunks = toksIndentBasedChunks(all_toks);

    Ast ret_ast = (Ast) {.src = full_src,
                         .toks = all_toks,
                         .top_defs = ·make(AstDef, 0, chunks.len),
                         .anns = {.total_nr_of_def_toks = 0,
                                  .incl_file_paths = ·make(Str, 0, asts_capacity),
                                  .src_file_path = src_file_path,
                                  .path_based_ident_prefix = ident(src_file_path)}};

    // guesstimate `total_nr_of_def_toks` by counting `:=` and `->` and `@->`
    ·forEach(Token, tok, all_toks, {
        if (tok->kind == tok_kind_ident) {
            Str const tok_src = tokSrc(tok, full_src);
            if (strEql(tok_src, strL(":=", 2)) || strEql(tok_src, strL("->", 2)) || strEql(tok_src, strL("@->", 3)))
                ret_ast.anns.total_nr_of_def_toks += 1;
        }
    });

    UInt toks_idx = 0;
    ·forEach(Tokens, chunk_toks, chunks, {
        AstDef* const dst_def = &ret_ast.top_defs.at[ret_ast.top_defs.len];
        *dst_def = astDef(NULL, toks_idx, chunk_toks->len);
        astParseDef(dst_def, &ret_ast);
        ret_ast.top_defs.len += 1;
        toks_idx += chunk_toks->len;
    });

    ·forEach(AstDef, top_def, ret_ast.top_defs, {
        astDefRewriteGlyphsIntoInstrs(top_def, &ret_ast);
        astSubDefsReorder(top_def->sub_defs);
    });

    return ret_ast;
}

void astParseDef(AstDef* const dst_def, Ast* const ast) {
    Tokens const toks = astNodeToks(&dst_def->node_base, ast);
    ºUInt const idx_tok_def = toksIndexOfIdent(toks, str(":="), ast->src);
    if ((!idx_tok_def.ok) || idx_tok_def.it == 0 || idx_tok_def.it == toks.len - 1)
        ·fail(astNodeMsg(str("expected '<head_expr> := <body_expr>'"), &dst_def->node_base, ast));

    Tokenss const def_body_chunks = toksIndentBasedChunks(·slice(Token, toks, idx_tok_def.it + 1, toks.len));
    dst_def->sub_defs = ·make(AstDef, 0, def_body_chunks.len - 1);
    UInt all_toks_idx = dst_def->node_base.toks_idx + idx_tok_def.it + 1;
    dst_def->body = astParseExpr(def_body_chunks.at[0], all_toks_idx, ast, false);

    AstExpr head = astParseExpr(·slice(Token, toks, 0, idx_tok_def.it), dst_def->node_base.toks_idx, ast, false);
    dst_def->anns.head_node_base = head.node_base;
    switch (head.kind) {
        case ast_expr_ident: {
            dst_def->name = head.of_ident;
        } break;
        case ast_expr_form: {
            if (head.of_exprs.at[0].kind != ast_expr_ident)
                ·fail(astNodeMsg(str("unsupported def header form"), &head.node_base, ast));
            dst_def->name = head.of_exprs.at[0].of_ident;

            AstExpr fn = astExpr(dst_def->node_base.toks_idx, dst_def->node_base.toks_len, ast_expr_form, 3);
            fn.of_exprs.at[0] = astExprInstrOrTag(head.node_base, strL("->", 2), false);
            fn.of_exprs.at[2] = dst_def->body;
            fn.of_exprs.at[1] =
                astExpr(head.of_exprs.at[1].node_base.toks_idx, head.node_base.toks_len - 1, ast_expr_lit_bracket, head.of_exprs.len - 1);
            dst_def->anns.param_names = ·make(Str, fn.of_exprs.at[1].of_exprs.len, 0);
            for (UInt i = 1; i < head.of_exprs.len; i += 1)
                if (head.of_exprs.at[i].kind != ast_expr_ident)
                    ·fail(astNodeMsg(str("unsupported def header form"), &head.node_base, ast));
                else {
                    dst_def->anns.param_names.at[i - 1] = head.of_exprs.at[i].of_ident;
                    fn.of_exprs.at[1].of_exprs.at[i - 1] =
                        astExprInstrOrTag(head.of_exprs.at[i].node_base, head.of_exprs.at[i].of_ident, true);
                }
            dst_def->body = fn;
        } break;
        default: {
            ·fail(astNodeMsg(str("unsupported def header form"), &head.node_base, ast));
        } break;
    }
    dst_def->anns.qname =
        (dst_def->anns.parent_def != NULL) ? str3(dst_def->anns.parent_def->anns.qname, strL("-", 1), dst_def->name) : dst_def->name;

    ·forEach(Tokens, chunk_toks, def_body_chunks, {
        if (iˇchunk_toks != 0) {
            AstDef* const sub_def = &dst_def->sub_defs.at[dst_def->sub_defs.len];
            dst_def->sub_defs.len += 1;
            *sub_def = astDef(dst_def, all_toks_idx, chunk_toks->len);
            astParseDef(sub_def, ast);
        }
        all_toks_idx += chunk_toks->len;
    });
}

AstExpr astParseExprLitInt(UInt const all_toks_idx, Ast const* const ast, Token const* const tok) {
    AstExpr ret_expr = astExpr(all_toks_idx, 1, ast_expr_lit_int, 0);
    ºU64 const maybe = uInt64Parse(tokSrc(tok, ast->src));
    if (!maybe.ok)
        ·fail(astNodeMsg(str("malformed or not-yet-supported integer literal"), &ret_expr.node_base, ast));
    ret_expr.of_lit_int = maybe.it;
    return ret_expr;
}

AstExpr astParseExprLitStr(UInt const all_toks_idx, Ast const* const ast, Token const* const tok, U8 const quote_char) {
    AstExpr ret_expr = astExpr(all_toks_idx + 1, 1, ast_expr_lit_str, 0);
    Str const lit_src = tokSrc(tok, ast->src);

    ·assert(lit_src.len >= 2 && lit_src.at[0] == quote_char && lit_src.at[lit_src.len - 1] == quote_char);
    Str ret_str = newStr(0, lit_src.len - 1);
    for (UInt i = 1; i < lit_src.len - 1; i += 1) {
        if (lit_src.at[i] != '\\')
            ret_str.at[ret_str.len] = lit_src.at[i];
        else {
            UInt const idx_end = i + 4;
            Bool bad_esc = idx_end > lit_src.len - 1;
            if (!bad_esc) {
                Str const base10digits = ·slice(U8, lit_src, i + 1, idx_end);
                i += 3;
                ºU64 const maybe = uInt64Parse(base10digits);
                bad_esc = (!maybe.ok) || maybe.it >= 256;
                ret_str.at[ret_str.len] = (U8)maybe.it;
            }
            if (bad_esc)
                ·fail(
                    astNodeMsg(str("expected 3-digit base-10 integer decimal 000-255 following backslash escape"), &ret_expr.node_base, ast));
        }
        ret_str.len += 1;
    }
    ret_str.at[ret_str.len] = 0; // we ensured the extra capacity up in `newStr` call

    ret_expr.of_lit_str = ret_str;
    return ret_expr;
}

AstExprs astParseExprsDelimited(Tokens const toks, UInt const all_toks_idx, TokenKind const tok_kind_sep, Ast* const ast) {
    if (toks.len == 0)
        return ·len0(AstExpr);
    Tokenss const per_elem_toks = toksSplit(toks, tok_kind_sep);
    AstExprs ret_exprs = ·make(AstExpr, 0, per_elem_toks.len);
    UInt toks_idx = all_toks_idx;
    ·forEach(Tokens, this_elem_toks, per_elem_toks, {
        if (this_elem_toks->len == 0)
            toks_idx += 1; // 1 for eaten delimiter
        else {
            ·append(ret_exprs, astParseExpr(*this_elem_toks, toks_idx, ast, false));
            toks_idx += (1 + this_elem_toks->len); // 1 for eaten delimiter
        }
    });
    return ret_exprs;
}

AstExpr astParseExpr(Tokens const expr_toks, UInt const all_toks_idx, Ast* const ast, Bool const known_whole_form_throng) {
    ·assert(expr_toks.len != 0);
    AstExprs ret_acc = ·make(AstExpr, 0, expr_toks.len);
    Bool const whole_form_throng =
        known_whole_form_throng || ((expr_toks.len > 1) && (tokThrong(expr_toks, 0, ast->src) == expr_toks.len - 1));

    for (UInt i = 0; i < expr_toks.len; i += 1) {
        UInt const idx_throng_end = whole_form_throng ? i : tokThrong(expr_toks, i, ast->src);
        if (idx_throng_end > i) {
            ·append(ret_acc, astParseExpr(·slice(Token, expr_toks, i, idx_throng_end + 1), all_toks_idx + i, ast, true));
            i = idx_throng_end; // loop header will increment
        } else {
            switch (expr_toks.at[i].kind) {

                case tok_kind_comment: {
                    ·fail(str("unreachable"));
                } break;

                case tok_kind_lit_num_prefixed: {
                    ·append(ret_acc, astParseExprLitInt(all_toks_idx + 1, ast, &expr_toks.at[i]));
                } break;

                case tok_kind_lit_str_qdouble: {
                    ·append(ret_acc, astParseExprLitStr(all_toks_idx + 1, ast, &expr_toks.at[i], '\"'));
                } break;

                case tok_kind_lit_str_qsingle: {
                    AstExpr expr_lit = astParseExprLitStr(all_toks_idx + 1, ast, &expr_toks.at[i], '\"');
                    if (expr_lit.of_lit_str.len != 1)
                        ·fail(astNodeMsg(str("currently only supporting single-byte char literals"), &expr_lit.node_base, ast));

                    expr_lit.kind = ast_expr_lit_int;
                    expr_lit.of_lit_int = expr_lit.of_lit_str.at[0];
                    ·append(ret_acc, expr_lit);
                } break;

                case tok_kind_sep_bcurly_open:  // fall through to:
                case tok_kind_sep_bsquare_open: // fall through to:
                case tok_kind_sep_bparen_open: {
                    TokenKind const tok_kind = expr_toks.at[i].kind;
                    ºUInt idx_closing = toksIndexOfMatchingBracket(·slice(Token, expr_toks, i, expr_toks.len));
                    ·assert(idx_closing.ok); // the other-case will have been caught already by toksCheckBrackets
                    idx_closing.it += i;
                    if (tok_kind == tok_kind_sep_bparen_open) {
                        Tokens const toks_inside_parens = ·slice(Token, expr_toks, i + 1, idx_closing.it);
                        if (toks_inside_parens.len == 0) {
                            AstExpr const expr_ident = astExprIdent(all_toks_idx + i, 2, strL("()", 2));
                            ·append(ret_acc, expr_ident);
                        } else {
                            AstExpr expr_inside_parens = astParseExpr(toks_inside_parens, all_toks_idx + i + 1, ast, false);
                            expr_inside_parens.anns.parensed += (expr_inside_parens.anns.parensed < 255) ? 1 : 0;
                            // still want the parens toks captured in node base:
                            expr_inside_parens.node_base.toks_idx = all_toks_idx + i;
                            expr_inside_parens.node_base.toks_len += 2;
                            ·append(ret_acc, expr_inside_parens);
                        }
                    } else { // no parens: either square brackets or curly braces
                        AstExprs const exprs_inside = astParseExprsDelimited(·slice(Token, expr_toks, i + 1, idx_closing.it),
                                                                             all_toks_idx + i + 1, tok_kind_sep_comma, ast);
                        Bool const is_braces = (tok_kind == tok_kind_sep_bcurly_open);
                        Bool const is_bracket = (tok_kind == tok_kind_sep_bsquare_open);
                        ·assert(is_braces || is_bracket); // always true right now obviously, but for future overpaced refactorers..
                        AstExpr expr_brac =
                            astExpr(all_toks_idx + i, 1 + (idx_closing.it - i), is_bracket ? ast_expr_lit_bracket : ast_expr_lit_braces, 0);
                        expr_brac.of_exprs = exprs_inside;
                        ·append(ret_acc, expr_brac);
                    }
                    i = idx_closing.it;
                } break;

                case tok_kind_ident: {
                    AstExpr const expr_ident = astExprIdent(all_toks_idx + i, 1, tokSrc(&expr_toks.at[i], ast->src));
                    ·append(ret_acc, expr_ident);
                } break;

                default: {
                    ·fail(str4(str("unrecognized token in line "), uIntToStr(expr_toks.at[i].line_nr + 1, 1, 10), str(": "),
                               tokSrc(&expr_toks.at[i], ast->src)));
                } break;
            }
        }
    }

    ·assert(ret_acc.len != 0);
    if (ret_acc.len == 1)
        return ret_acc.at[0];

    AstExpr ret_expr = astExpr(all_toks_idx, expr_toks.len, ast_expr_form, 0);
    ret_expr.of_exprs = ret_acc;
    ret_expr.anns.toks_throng = whole_form_throng;

    Str const incl_file_path = astExprIsIncl(&ret_expr);
    if (incl_file_path.at != NULL) {
        if (!strIn(incl_file_path, ast->anns.incl_file_paths)) {
            if (ast->anns.incl_file_paths.len == asts_capacity)
                ·fail(str("TODO: increase asts_capacity"));
            ·append(ast->anns.incl_file_paths, incl_file_path);
        }
        if (ret_expr.of_exprs.len > 2) {
            AstExpr sub_expr = astExpr(ret_expr.node_base.toks_idx, 2, ast_expr_form, 2);
            sub_expr.anns.parensed = true;
            sub_expr.anns.toks_throng = true;
            sub_expr.of_exprs.at[0] = ret_expr.of_exprs.at[0];
            sub_expr.of_exprs.at[1] = ret_expr.of_exprs.at[1];
            for (UInt i = 1; i < ret_expr.of_exprs.len; i += 1)
                ret_expr.of_exprs.at[i - 1] = ret_expr.of_exprs.at[i];
            ret_expr.of_exprs.at[0] = sub_expr;
            ret_expr.of_exprs.len -= 1;
        }
    }

    return ret_expr;
}
