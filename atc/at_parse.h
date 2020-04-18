#pragma once
#include "metaleap.h"
#include "at_toks.h"
#include "at_ast.h"


AstExpr parseExpr(Tokens const expr_toks, Uint const all_toks_idx, Ast const* const ast) {
    assert(expr_toks.len != 0);
    AstExprs ret_acc = make(AstExpr, 0, expr_toks.len);
    Bool const whole_form_throng = (expr_toks.len > 1) && (tokThrong(expr_toks, 0, ast->src) == expr_toks.len - 1);

    for (Uint i = 0; i < expr_toks.len; i += 1) {
        Uint const idx_throng_end = whole_form_throng ? i : tokThrong(expr_toks, i, ast->src);
        if (idx_throng_end > i) {
            append(ret_acc, parseExpr(slice(Token, expr_toks, i, idx_throng_end + 1), all_toks_idx + i, ast));
            i = idx_throng_end; // loop header will increment
        } else
            switch (expr_toks.at[i].kind) {
                case tok_kind_comment: panic("unreachable"); break;
                case tok_kind_lit_num_prefixed: panic("TODO"); break;
                case tok_kind_lit_str_qdouble: panic("TODO"); break;
                case tok_kind_lit_str_qsingle: panic("TODO"); break;
                case tok_kind_sep_bcurly_open:
                case tok_kind_sep_bsquare_open:
                case tok_kind_sep_bparen_open: panic("TODO"); break;
                default: {
                    AstExpr expr_ident = astExpr(all_toks_idx + i, 1, ast_expr_ident);
                    expr_ident.kind_ident = tokSrc(&expr_toks.at[i], ast->src);
                    append(ret_acc, expr_ident);
                } break;
            }
    }

    assert(ret_acc.len != 0);
    if (ret_acc.len == 1)
        return ret_acc.at[0];

    AstExpr ret_expr = astExpr(all_toks_idx, expr_toks.len, ast_expr_form);
    ret_expr.kind_form = ret_acc;
    ret_expr.anns.toks_throng = whole_form_throng;
    return ret_expr;
}

U64 parseExprLitInt(AstNodeBase const* const node_base, Ast const* const ast, Str const lit_src) {
    ºU64 const maybe = uintParse(lit_src);
    if (maybe.ok)
        panic(astNodeMsg(str("malformed integer literal"), node_base, ast));
    return maybe.it;
}

Str parseExprLitStr(AstNodeBase const* const node_base, Ast const* const ast, Str const lit_src, U8 const quote_char) {
    assert(lit_src.len >= 2 && lit_src.at[0] == quote_char && lit_src.at[lit_src.len - 1] == quote_char);
    Str ret_str = newStr(0, lit_src.len - 2);
    for (Uint i = 0; i < lit_src.len; i += 1) {
        if (lit_src.at[i] != '\\')
            ret_str.at[ret_str.len] = lit_src.at[i];
        else {
            Uint const idx_end = i + 4;
            Bool bad_esc = idx_end > lit_src.len - 1;
            if (!bad_esc) {
                Str const base10digits = slice(U8, lit_src, i + 1, idx_end);
                i += 3;
                ºU64 const maybe = uintParse(base10digits);
                bad_esc = (!maybe.ok) || maybe.it >= 256;
                ret_str.at[ret_str.len] = (U8)maybe.it;
            }
            if (bad_esc)
                panic(astNodeMsg(str("expected 3-digit base-10 integer decimal 000-255 following backslash escape"), node_base, ast));
        }
        ret_str.len += 1;
    }
    return ret_str;
}

AstExprs parseExprsDelimited(Tokens const toks, Uint const all_toks_idx, TokenKind const tok_kind_sep, Ast const* const ast) {
    if (toks.len == 0)
        return (AstExprs) {.len = 0, .at = NULL};
    Tokenss per_elem_toks = toksSplit(toks, tok_kind_sep);
    AstExprs ret_exprs = make(AstExpr, 0, per_elem_toks.len);
    Uint toks_idx = all_toks_idx;
    forEach(Tokens, this_elem_toks, per_elem_toks, {
        if (this_elem_toks->len == 0)
            toks_idx += 1; // 1 for eaten delimiter
        else {
            append(ret_exprs, parseExpr(*this_elem_toks, toks_idx, ast));
            toks_idx += (1 + this_elem_toks->len); // 1 for eaten delimiter
        }
    });
    return ret_exprs;
}
