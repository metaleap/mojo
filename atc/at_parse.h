#pragma once
#include "metaleap.h"
#include "at_toks.h"
#include "at_ast.h"


AstExpr parseExpr(Tokens const toks, Uint const all_toks_idx, Ast const* const ast) {
    return AstExpr {};
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
