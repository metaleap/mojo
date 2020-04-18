#pragma once
#include "metaleap.h"
#include "at_toks.h"


typedef struct AstNodeBase {
    Uint toks_idx;
    Uint toks_len;
} AstNodeBase;

typedef enum AstExprKind {
    ast_expr_lit_int = 1,
    ast_expr_lit_str = 2,
    ast_expr_ident = 3,
    ast_expr_form = 4,
    ast_expr_lit_bracket = 5,
    ast_expr_lit_braces = 6,
} AstExprKind;

typedef struct AstExpr AstExpr;
typedef SliceOf(AstExpr) AstExprs;
struct AstExpr {
    AstNodeBase node_base;
    AstExprKind kind;
    union {
        U64 kind_lit_int;      // 123
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
typedef Maybe(AstExpr) ºAstExpr;
typedef struct AstExpr² {
    ºAstExpr lhs;
    ºAstExpr rhs;
} AstExpr²;

typedef struct AstDef AstDef;
typedef SliceOf(AstDef) AstDefs;
struct AstDef {
    AstNodeBase node_base;
    AstExpr head;
    AstExpr body;
    AstDefs sub_defs;
    AstDef* parent_def;
    struct {
        Str name;
        Uint total_sub_defs_count;
    } anns;
};

typedef struct Ast {
    Str src;
    Tokens toks;
    AstDefs top_defs;
} Ast;



AstNodeBase astNodeBaseFrom(Uint const toks_idx, Uint const toks_len) {
    return (AstNodeBase) {.toks_idx = toks_idx, .toks_len = toks_len};
}

Tokens astNodeToks(AstNodeBase const* const node, Ast const* const ast) {
    return slice(Token, ast->toks, node->toks_idx, node->toks_idx + node->toks_len);
}

String astNodeMsg(Str const msg_prefix, AstNodeBase const* const node, Ast const* const ast) {
    Tokens const node_toks = astNodeToks(node, ast);
    Str const line_nr = uintToStr(1 + node_toks.at[0].line_nr, 10);
    Str const toks_src = toksSrc(node_toks, ast->src);
    return (String)str5(msg_prefix, str(" in line "), line_nr, str(":\n"), toks_src).at;
}

AstExpr astExpr(Uint const toks_idx, Uint const toks_len, AstExprKind const expr_kind) {
    return (AstExpr) {
        .node_base = astNodeBaseFrom(toks_idx, toks_len),
        .kind = expr_kind,
        .anns = {.parensed = 0, .toks_throng = false},
    };
}

AstExpr astExprFormSub(AstExpr const* const ast_expr, Uint const idx_start, Uint const idx_end) {
    assert(!(idx_start == 0 && idx_end == ast_expr->kind_form.len));
    assert(idx_end > idx_start);
    if (idx_end == idx_start + 1)
        return ast_expr->kind_form.at[idx_start];

    AstExpr ret_expr = astExpr(ast_expr->kind_form.at[idx_start].node_base.toks_idx, 0, ast_expr_form);
    ret_expr.anns.toks_throng = ast_expr->anns.toks_throng;
    ret_expr.kind_form = slice(AstExpr, ast_expr->kind_form, idx_start, idx_end);
    for (Uint i = idx_start; i < idx_end; i += 1)
        ret_expr.node_base.toks_len += ast_expr->kind_form.at[i].node_base.toks_len;
    return ret_expr;
}

ºUint astExprFormIndexOfIdent(AstExpr const* const ast_expr, Str const ident) {
    forEach(AstExpr, expr, ast_expr->kind_form, {
        if (expr->kind == ast_expr_ident && strEql(ident, expr->kind_ident))
            return ok(Uint, exprˇidx);
    });
    return none(Uint);
}

AstExpr² astExprFormBreakOn(AstExpr const* const ast_expr, Str const ident, Bool const must_lhs, Bool const must_rhs, Ast const* const ast) {
    AstExpr² ret_tup = (AstExpr²) {.lhs = none(AstExpr), .rhs = none(AstExpr)};

    ºUint const pos = astExprFormIndexOfIdent(ast_expr, ident);
    if (pos.ok) {
        if (pos.it > 0)
            ret_tup.lhs = ok(AstExpr, astExprFormSub(ast_expr, 0, pos.it));
        if (pos.it < ast_expr->kind_form.len - 1)
            ret_tup.rhs = ok(AstExpr, astExprFormSub(ast_expr, 1 + pos.it, ast_expr->kind_form.len));
    }
    Bool const must_both = must_lhs && must_rhs;
    if (must_both && !pos.ok)
        panic(astNodeMsg(str3(str("expected '"), ident, str("'")), &ast_expr->node_base, ast));
    if (must_lhs && !ret_tup.lhs.ok)
        panic(astNodeMsg(str3(str("expected expression before '"), ident, str("'")), &ast_expr->node_base, ast));
    if (must_rhs && !ret_tup.rhs.ok)
        panic(astNodeMsg(str3(str("expected expression following '"), ident, str("'")), &ast_expr->node_base, ast));
    return ret_tup;
}
