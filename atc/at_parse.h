#pragma once
#include "metaleap.h"
#include "at_toks.h"
#include "at_ast.h"


AstExpr parseExpr(Tokens const toks, Uint const all_toks_idx, Ast const* const ast) {
    return (AstExpr) {};
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
                Str base10digits = slice(U8, lit_src, i + 1, idx_end);
                i += 3;
                ºUint maybe = uintParse(base10digits);
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
