#pragma once
#include "metaleap.c"
#include "std_io.c"
#include "at_toks.c"


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
typedef ·SliceOf(AstExpr) AstExprs;
struct AstExpr {
    AstNodeBase node_base;
    AstExprKind kind;
    union {
        U64 of_lit_int;      // 123
        Str of_lit_str;      // "123"
        Str of_ident;        // anyIdentifier                         (also operators)
        AstExprs of_form;    // expr1 expr2 expr3 ... exprN           (always: .len >= 2)
        AstExprs of_bracket; // [expr1, expr2, expr3, ..., exprN]     (always: .len >= 0)
        AstExprs of_braces;  // {expr1, expr2, expr3, ..., exprN}     (always: .len >= 0)
    };
    struct {
        Uint parensed;
        Bool toks_throng;
    } anns;
};
typedef ·Maybe(AstExpr) ºAstExpr;
typedef struct AstExpr² {
    ºAstExpr lhs;
    ºAstExpr rhs;
    AstExpr const* operator;
} AstExpr²;


typedef struct AstDef AstDef;
typedef ·SliceOf(AstDef) AstDefs;
struct AstDef {
    AstNodeBase node_base;
    AstExpr head;
    AstExpr body;
    AstDefs sub_defs;
    AstDef* parent_def;
    struct {
        Str name;
    } anns;
};


typedef struct Ast {
    Str src;
    Tokens toks;
    AstDefs top_defs;
} Ast;




static AstNodeBase astNodeBaseFrom(Uint const toks_idx, Uint const toks_len) {
    return (AstNodeBase) {.toks_idx = toks_idx, .toks_len = toks_len};
}

static Tokens astNodeToks(AstNodeBase const* const node, Ast const* const ast) {
    return ·slice(Token, ast->toks, node->toks_idx, node->toks_idx + node->toks_len);
}

static Str astNodeMsg(Str const msg_prefix, AstNodeBase const* const node, Ast const* const ast) {
    Tokens const node_toks = astNodeToks(node, ast);
    Str const line_nr = uintToStr(1 + node_toks.at[0].line_nr, 1, 10);
    Str const toks_src = toksSrc(node_toks, ast->src);
    return str5(msg_prefix, str(" in line "), line_nr, str(":\n"), toks_src);
}

static AstDef astDef(AstDef* const parent_def, Uint const all_toks_idx, Uint const toks_len) {
    return (AstDef) {
        .parent_def = parent_def,
        .node_base = astNodeBaseFrom(all_toks_idx, toks_len),
    };
}

static AstExpr astExpr(Uint const toks_idx, Uint const toks_len, AstExprKind const expr_kind) {
    return (AstExpr) {
        .node_base = astNodeBaseFrom(toks_idx, toks_len),
        .kind = expr_kind,
        .anns = {.parensed = 0, .toks_throng = false},
    };
}

static AstExpr astExprFormSub(AstExpr const* const ast_expr, Uint const idx_start, Uint const idx_end) {
    ·assert(!(idx_start == 0 && idx_end == ast_expr->of_form.len));
    ·assert(idx_end > idx_start);
    if (idx_end == idx_start + 1)
        return ast_expr->of_form.at[idx_start];

    AstExpr ret_expr = astExpr(ast_expr->of_form.at[idx_start].node_base.toks_idx, 0, ast_expr_form);
    ret_expr.anns.toks_throng = ast_expr->anns.toks_throng;
    ret_expr.of_form = ·slice(AstExpr, ast_expr->of_form, idx_start, idx_end);
    for (Uint i = idx_start; i < idx_end; i += 1)
        ret_expr.node_base.toks_len += ast_expr->of_form.at[i].node_base.toks_len;
    return ret_expr;
}

static ºUint astExprFormIndexOfIdent(AstExpr const* const ast_expr, Str const ident) {
    ·assert(ast_expr->kind == ast_expr_form);
    ·forEach(AstExpr, expr, ast_expr->of_form, {
        if (expr->kind == ast_expr_ident && strEql(ident, expr->of_ident))
            return ·ok(Uint, iˇexpr);
    });
    return ·none(Uint);
}

static AstExpr² astExprFormBreakOn(AstExpr const* const ast_expr, Str const ident, Bool const must_lhs, Bool const must_rhs,
                                   Ast const* const ast) {
    if (must_lhs || must_rhs)
        ·assert(ast != NULL);
    ·assert(ast_expr->kind == ast_expr_form);

    AstExpr² ret_tup = (AstExpr²) {.lhs = ·none(AstExpr), .rhs = ·none(AstExpr), .operator= NULL };
    ºUint const pos = astExprFormIndexOfIdent(ast_expr, ident);
    if (pos.ok) {
        ret_tup.operator=& ast_expr->of_form.at[pos.it];
        if (pos.it > 0)
            ret_tup.lhs = ·ok(AstExpr, astExprFormSub(ast_expr, 0, pos.it));
        if (pos.it < ast_expr->of_form.len - 1)
            ret_tup.rhs = ·ok(AstExpr, astExprFormSub(ast_expr, 1 + pos.it, ast_expr->of_form.len));
    }
    Bool const must_both = must_lhs && must_rhs;
    if (must_both && !pos.ok)
        ·fail(astNodeMsg(str3(str("expected '"), ident, str("'")), &ast_expr->node_base, ast));
    if (must_lhs && !ret_tup.lhs.ok)
        ·fail(astNodeMsg(str3(str("expected expression before '"), ident, str("'")), &ast_expr->node_base, ast));
    if (must_rhs && !ret_tup.rhs.ok)
        ·fail(astNodeMsg(str3(str("expected expression following '"), ident, str("'")), &ast_expr->node_base, ast));
    return ret_tup;
}



static Bool astExprHasIdent(AstExpr const* const expr, Str const ident) {
    switch (expr->kind) {
        case ast_expr_ident: {
            return strEql(ident, expr->of_ident);
        } break;
        case ast_expr_form: {
            if (expr->of_form.len == 2 && expr->of_form.at[0].kind == ast_expr_ident && strEql(strL("#", 1), expr->of_form.at[0].of_ident))
                return false;
            ·forEach(AstExpr, sub_expr, expr->of_form, {
                if (astExprHasIdent(sub_expr, ident))
                    return true;
            });
        } break;
        case ast_expr_lit_bracket: {
            ·forEach(AstExpr, sub_expr, expr->of_bracket, {
                if (astExprHasIdent(sub_expr, ident))
                    return true;
            });
        } break;
        case ast_expr_lit_braces: {
            ·forEach(AstExpr, sub_expr, expr->of_braces, {
                if (astExprHasIdent(sub_expr, ident))
                    return true;
            });
        } break;
        default: break;
    }
    return false;
}

static Bool astDefHasIdent(AstDef const* const def, Str const ident) {
    ·forEach(AstDef, sub_def, def->sub_defs, {
        if (astDefHasIdent(sub_def, ident))
            return true;
    });
    return astExprHasIdent(&def->body, ident);
}



static AstExpr astExprInstrOrTag(AstNodeBase const from, Str const name, Bool const tag) {
    AstExpr ret_expr = (AstExpr) {.kind = ast_expr_form, .of_form = ·make(AstExpr, 2, 2), .node_base = from, .anns = {.toks_throng = true}};
    ret_expr.of_form.at[0] = (AstExpr) {.kind = ast_expr_ident, .of_ident = strL(tag ? "#" : "@", 1), .node_base = from};
    ret_expr.of_form.at[1] = (AstExpr) {.kind = ast_expr_ident, .of_ident = name, .node_base = from};
    return ret_expr;
}

static AstExpr astExprFormEmpty(AstNodeBase const from) {
    return (AstExpr) {.kind = ast_expr_form, .of_form = ·make(AstExpr, 0, 0), .node_base = from};
}

static void astExprDesugarOperatorsIntoInstrs(AstExpr* const expr) {
#define num_operators 3
    static Strs operator_names = (Strs) {.at = NULL, .len = 0};
    static Strs instr_names = (Strs) {.at = NULL, .len = 0};
    if (operator_names.at == NULL && instr_names.at == NULL) {
        operator_names = ·make(Str, num_operators, num_operators);
        operator_names.at[0] = str(":");
        operator_names.at[1] = str("->");
        operator_names.at[2] = str(".");
        instr_names = ·make(Str, num_operators, num_operators);
        instr_names.at[0] = str("kvPair");
        instr_names.at[1] = str("func");
        instr_names.at[2] = str("fldAcc");
    }

    switch (expr->kind) {
        case ast_expr_form: {
            if (expr->of_form.len == 0)
                break;
            Bool matched = false;
            AstExpr² maybe;
            for (Uint i = 0; i < num_operators && !matched; i += 1) {
                maybe = astExprFormBreakOn(expr, operator_names.at[i], false, false, NULL);
                if (maybe.operator!= NULL) {
                    matched = true;
                    expr->anns.toks_throng = false;
                    expr->of_form = ·make(AstExpr, 3, 3);
                    expr->of_form.at[0] = astExprInstrOrTag(maybe.operator->node_base, instr_names.at[i], false);
                    expr->of_form.at[1] = (maybe.lhs.ok) ? maybe.lhs.it : astExprFormEmpty(expr->node_base);
                    expr->of_form.at[2] = (maybe.rhs.ok) ? maybe.rhs.it : astExprFormEmpty(expr->node_base);
                    if (strEql(operator_names.at[i], strL("->", 1))) {
                        // TODO!
                    }
                    if (strEql(operator_names.at[i], strL(".", 1)) && expr->of_form.at[2].kind == ast_expr_ident)
                        expr->of_form.at[2] = astExprInstrOrTag(expr->of_form.at[2].node_base, expr->of_form.at[2].of_ident, true);
                    else if (maybe.rhs.ok)
                        astExprDesugarOperatorsIntoInstrs(&expr->of_form.at[2]);
                }
            }
            if (!matched)
                ·forEach(AstExpr, sub_expr, expr->of_form, { astExprDesugarOperatorsIntoInstrs(sub_expr); });
        } break;
        case ast_expr_lit_bracket: {
            ·forEach(AstExpr, sub_expr, expr->of_bracket, { astExprDesugarOperatorsIntoInstrs(sub_expr); });
        } break;
        case ast_expr_lit_braces: {
            ·forEach(AstExpr, sub_expr, expr->of_braces, { astExprDesugarOperatorsIntoInstrs(sub_expr); });
        } break;
        default: break;
    }
}

static void astDefDesugarOperatorsIntoInstrs(AstDef* const def) {
    ·forEach(AstDef, sub_def, def->sub_defs, { astDefDesugarOperatorsIntoInstrs(sub_def); });
    astExprDesugarOperatorsIntoInstrs(&def->body);
}

static void astDesugarOperatorsIntoInstrs(Ast const* const ast) {
    ·forEach(AstDef, def, ast->top_defs, { astDefDesugarOperatorsIntoInstrs(def); });
}



static void astDefPrint(AstDef const* const, Ast const* const, Uint const);
static void astExprPrint(AstExpr const* const, AstDef const* const, Ast const* const, Bool const, Uint const);

static void astPrint(Ast const* const ast) {
    ·forEach(AstDef, top_def, ast->top_defs, {
        astDefPrint(top_def, ast, 0);
        printChr('\n');
    });
}

static void astDefPrint(AstDef const* const def, Ast const* const ast, Uint const ind) {
    printChr('\n');
    for (Uint i = 0; i < ind; i += 1)
        printChr(' ');
    astExprPrint(&def->head, def, ast, false, ind);
    printStr(str(" :=\n"));
    for (Uint i = 0; i < 2 + ind; i += 1)
        printChr(' ');
    astExprPrint(&def->body, def, ast, false, ind + 2);

    ·forEach(AstDef, sub_def, def->sub_defs, { astDefPrint(sub_def, ast, 2 + ind); });
}

static void astExprPrint(AstExpr const* const expr, AstDef const* const def, Ast const* const ast, Bool const is_form_item, Uint const ind) {
    for (Uint i = 0; i < expr->anns.parensed; i++)
        printChr('(');
    switch (expr->kind) {
        case ast_expr_ident: {
            printStr(expr->of_ident);
        } break;

        case ast_expr_lit_int: {
            printStr(uintToStr(expr->of_lit_int, 1, 10));
        } break;

        case ast_expr_lit_str: {
            printStr(strQuot(expr->of_lit_str));
        } break;

        case ast_expr_form: {
            if (expr->of_form.len == 0)
                break;
            Bool const parens = is_form_item && expr->anns.parensed == 0 && !expr->anns.toks_throng;
            if (parens)
                printChr('(');
            ·forEach(AstExpr, sub_expr, expr->of_form, {
                if (iˇsub_expr != 0 && !expr->anns.toks_throng)
                    printChr(' ');
                astExprPrint(sub_expr, def, ast, true, ind);
            });
            if (parens)
                printChr(')');
        } break;

        case ast_expr_lit_bracket: {
            printChr('[');
            ·forEach(AstExpr, sub_expr, expr->of_bracket, {
                if (iˇsub_expr != 0)
                    printStr(str(", "));
                astExprPrint(sub_expr, def, ast, false, ind);
            });
            printChr(']');
        } break;

        case ast_expr_lit_braces: {
            printStr(str("{\n"));
            Uint const ind_next = 2 + ind;
            ·forEach(AstExpr, sub_expr, expr->of_braces, {
                for (Uint i = 0; i < ind_next; i += 1)
                    printChr(' ');
                astExprPrint(sub_expr, def, ast, false, ind_next);
                printStr(str(",\n"));
            });
            for (Uint i = 0; i < ind; i += 1)
                printChr(' ');
            printChr('}');
        } break;

        default: {
            ·fail(str2(str("TODO: astExprPrint for .kind of "), uintToStr(expr->kind, 1, 10)));
        } break;
    }
    for (Uint i = 0; i < expr->anns.parensed; i++)
        printChr(')');
}
