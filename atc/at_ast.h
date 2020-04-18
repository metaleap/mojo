#pragma once
#include "at_toks.h"
#include "std.h"

typedef enum AstExprKind {
    ast_expr_lit_int,
    ast_expr_lit_str,
    ast_expr_ident,
    ast_expr_form,
    ast_expr_lit_bracket,
    ast_expr_lit_braces,
} AstExprKind;

typedef struct AstNode {
    Uint toks_idx;
    Uint toks_len;
} AstNode;

typedef struct AstExpr AstExpr;
typedef SliceOf(AstExpr) AstExprs;
struct AstExpr {
    AstNode base;
    AstExprKind kind;
    union {
        Uint kind_lit_int;     // 123
        Str kind_lit_str;      // "123"
        Str kind_ident;        // anyIdentifier                         (also operators)
        AstExprs kind_form;    // expr1 expr2 expr3 ... exprN           (always: .len >= 2)
        AstExprs kind_bracket; // [expr1, expr2, expr3, ..., exprN]     (always: .len >= 0)
        AstExprs kind_braces;  // {expr1, expr2, expr3, ..., exprN}     (always: .len >= 0)
    };
    struct {
        Uint parensed;
        Bool toks_throng;
    } anns;
};

typedef struct AstDef AstDef;
typedef SliceOf(AstDef) AstDefs;
struct AstDef {
    AstNode base;
    AstExpr head;
    AstExpr body;
    AstDefs sub_defs;
    struct {
        Bool is_top_def;
        Str name;
    } anns;
};

typedef struct Ast {
    Str src;
    Tokens toks;
    AstDefs top_defs;
} Ast;

typedef struct AstNameRef {
    Str name;
    AstDef* top_def;
    Uints sub_def_path;
    ºUint param_idx;
} AstNameRef;
typedef SliceOf(AstNameRef) AstNameRefs;

struct AstScopes;
typedef struct AstScopes AstScopes;
struct AstScopes {
    AstNameRefs names;
    AstScopes* parent;
};

AstNode astNodeFrom(Uint const toks_idx, Uint const toks_len) {
    return (AstNode) {.toks_idx = toks_idx, .toks_len = toks_len};
}

Tokens astNodeToks(AstNode const* const node, Ast const* const ast) {
    return slice(Token, ast->toks, node->toks_idx, node->toks_idx + node->toks_len);
}

Str astNodeMsg(String const msg_prefix, AstNode const* const node, Ast const* const ast) {
    Tokens const node_toks = astNodeToks(node, ast);
    Str const line_nr = uintToStr(1 + node_toks.at[0].line_nr, 10);
    Str const toks_src = toksSrc(node_toks, ast->src);
    Str arr[5] = {str(msg_prefix), str(" in line "), line_nr, str(":\n"), toks_src};
    return strConcat((Strs) {.len = 5, .at = arr});
}

AstExpr astExprFormSub(AstExpr const* const ast_expr, Uint const idx_start, Uint const idx_end) {
    assert(!(idx_start == 0 && idx_end == ast_expr->kind_form.len));
    assert(idx_end > idx_start);
    if (idx_end == idx_start + 1)
        return ast_expr->kind_form.at[idx_start];

    AstExpr ret_expr = (AstExpr) {
        .kind = ast_expr_form,
        .anns = {.parensed = false, .toks_throng = ast_expr->anns.toks_throng},
        .base = {.toks_len = 0, .toks_idx = ast_expr->kind_form.at[idx_start].base.toks_idx},
        .kind_form = slice(AstExpr, ast_expr->kind_form, idx_start, idx_end),
    };
    for (Uint i = idx_start; i < idx_end; i += 1)
        ret_expr.base.toks_len += ast_expr->kind_form.at[i].base.toks_len;
    return ret_expr;
}
