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
        U8 parensed;
        Bool toks_throng;
    } anns;
};
typedef ·Maybe(AstExpr) ºAstExpr;
typedef struct AstExpr² {
    ºAstExpr lhs_form;
    ºAstExpr rhs_form;
    AstExpr const* glyph;
} AstExpr²;


typedef struct AstDef AstDef;
typedef ·SliceOf(AstDef) AstDefs;
struct AstDef {
    AstNodeBase node_base;
    AstExpr head;
    AstExpr body;
    AstDefs sub_defs;
    struct {
        AstDef* parent_def;
        Str name;
        Str qname;
    } anns;
};


typedef struct Ast {
    Str src;
    Tokens toks;
    AstDefs top_defs;
} Ast;



static void astDefPrint(AstDef const* const, Uint const);
static void astExprPrint(AstExpr const* const, Bool const, Uint const);



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
    return str8(node_toks.at[0].file_name, str(":"), line_nr, str(": "), msg_prefix, str(":\n"), toks_src, str("\n"));
}

static Str astNodeSrc(AstNodeBase const* const node, Ast const* const ast) {
    return toksSrc(astNodeToks(node, ast), ast->src);
}

static AstDef astDef(AstDef* const parent_def, Uint const all_toks_idx, Uint const toks_len) {
    return (AstDef) {
        .anns = {.parent_def = parent_def},
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

static AstExpr² astExprFormBreakOn(AstExpr const* const expr, Str const ident, Bool const must_lhs, Bool const must_rhs, Ast const* const ast) {
    if (must_lhs || must_rhs)
        ·assert(ast != NULL);
    ·assert(expr->kind == ast_expr_form);

    AstExpr² ret_tup = (AstExpr²) {.lhs_form = ·none(AstExpr), .rhs_form = ·none(AstExpr), .glyph = NULL};
    ºUint const pos = astExprFormIndexOfIdent(expr, ident);
    if (pos.ok) {
        ret_tup.glyph = &expr->of_form.at[pos.it];
        if (pos.it > 0)
            ret_tup.lhs_form = ·ok(AstExpr, astExprFormSub(expr, 0, pos.it));
        if (pos.it < expr->of_form.len - 1)
            ret_tup.rhs_form = ·ok(AstExpr, astExprFormSub(expr, 1 + pos.it, expr->of_form.len));
    }
    Bool const must_both = must_lhs && must_rhs;
    if (must_both && !pos.ok)
        ·fail(astNodeMsg(str3(str("expected '"), ident, str("'")), &expr->node_base, ast));
    if (must_lhs && !ret_tup.lhs_form.ok)
        ·fail(astNodeMsg(str3(str("expected expression before '"), ident, str("'")), &expr->node_base, ast));
    if (must_rhs && !ret_tup.rhs_form.ok)
        ·fail(astNodeMsg(str3(str("expected expression following '"), ident, str("'")), &expr->node_base, ast));
    return ret_tup;
}


static Bool astExprIsTagOrInstr(AstExpr const* const expr, Bool const check_is_tag, Bool const check_is_instr) {
    AstExpr* const gly = &expr->of_form.at[0];
    return expr->kind == ast_expr_form && expr->of_form.len == 2 && gly->kind == ast_expr_ident && gly->of_ident.len == 1
           && ((check_is_instr && gly->of_ident.at[0] == '@' && expr->of_form.at[1].kind == ast_expr_ident)
               || (check_is_tag && gly->of_ident.at[0] == '#'));
}

static Bool astExprHasIdent(AstExpr const* const expr, Str const ident) {
    switch (expr->kind) {
        case ast_expr_ident: {
            return strEql(ident, expr->of_ident);
        } break;
        case ast_expr_form: {
            if (astExprIsTagOrInstr(expr, true, true))
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

static void astExprRewriteGlyphsIntoInstrs(AstExpr* const expr) {
    enum InstrGlyph {
        glyph_kvpair = 0,
        glyph_func = 1,
        glyph_fldacc = 2,

        num_glyphs = 3,
    };
    static Strs glyph_names = (Strs) {.at = NULL, .len = 0};
    if (glyph_names.at == NULL) {
        glyph_names = ·make(Str, num_glyphs, num_glyphs);
        glyph_names.at[glyph_kvpair] = str(":");
        glyph_names.at[glyph_func] = str("->");
        glyph_names.at[glyph_fldacc] = str(".");
    }

    if (expr->kind == ast_expr_lit_bracket)
        ·forEach(AstExpr, sub_expr, expr->of_bracket, { astExprRewriteGlyphsIntoInstrs(sub_expr); });
    else if (expr->kind == ast_expr_lit_braces)
        ·forEach(AstExpr, sub_expr, expr->of_braces, { astExprRewriteGlyphsIntoInstrs(sub_expr); });
    else if (expr->kind == ast_expr_form && expr->of_form.len != 0) {
        Bool matched = false;
        AstExpr² maybe;
        for (Uint i = 0; i < num_glyphs && !matched; i += 1) {
            maybe = astExprFormBreakOn(expr, glyph_names.at[i], false, false, NULL);
            if (!(maybe.glyph != NULL && (maybe.lhs_form.ok || maybe.rhs_form.ok)))
                continue;
            matched = true;
            expr->anns.toks_throng = false;
            expr->of_form = ·make(AstExpr, 3, 3);
            expr->of_form.at[0] = astExprInstrOrTag(maybe.glyph->node_base, glyph_names.at[i], false);
            expr->of_form.at[1] = (maybe.lhs_form.ok) ? maybe.lhs_form.it : astExprFormEmpty(expr->node_base);
            expr->of_form.at[2] = (maybe.rhs_form.ok) ? maybe.rhs_form.it : astExprFormEmpty(expr->node_base);
            if (i == glyph_func) {
                ·forEach(AstExpr, param_expr, expr->of_form.at[1].of_form, {
                    if (param_expr->kind == ast_expr_ident)
                        *param_expr = astExprInstrOrTag(param_expr->node_base, param_expr->of_ident, true);
                });
                expr->of_form.at[1].kind = ast_expr_lit_bracket;
            } else if (i == glyph_fldacc) {
                if (expr->of_form.at[2].of_form.len == 1 && expr->of_form.at[2].of_form.at[0].kind == ast_expr_ident)
                    expr->of_form.at[2].of_form.at[0] =
                        astExprInstrOrTag(expr->of_form.at[2].of_form.at[0].node_base, expr->of_form.at[2].of_form.at[0].of_ident, true);
            }
            if (maybe.lhs_form.ok) {
                if (i != glyph_func && expr->of_form.at[1].of_form.len == 1)
                    expr->of_form.at[1] = expr->of_form.at[1].of_form.at[0];
                astExprRewriteGlyphsIntoInstrs(&expr->of_form.at[1]);
            }
            if (maybe.rhs_form.ok) {
                if (expr->of_form.at[2].of_form.len == 1)
                    expr->of_form.at[2] = expr->of_form.at[2].of_form.at[0];
                astExprRewriteGlyphsIntoInstrs(&expr->of_form.at[2]);
            }
        }
        if (!matched)
            ·forEach(AstExpr, sub_expr, expr->of_form, { astExprRewriteGlyphsIntoInstrs(sub_expr); });
    }
}

static void astDefRewriteGlyphsIntoInstrs(AstDef* const def) {
    ·forEach(AstDef, sub_def, def->sub_defs, { astDefRewriteGlyphsIntoInstrs(sub_def); });
    astExprRewriteGlyphsIntoInstrs(&def->body);
}

static void astRewriteGlyphsIntoInstrs(Ast const* const ast) {
    ·forEach(AstDef, def, ast->top_defs, { astDefRewriteGlyphsIntoInstrs(def); });
}

static void astSubDefsReorder(AstDefs const defs) {
    ·forEach(AstDef, the_def, defs, { astSubDefsReorder(the_def->sub_defs); });

    Uint num_rounds = 0;
    for (Bool again = true; again; num_rounds += 1) {
        again = false;
        if (num_rounds > 42 * defs.len)
            ·fail(str2(str("Circular sub-def dependencies inside "),
                       (defs.at[0].anns.parent_def == NULL) ? str("<top-level>") : defs.at[0].anns.parent_def->anns.name));
        ·forEach(AstDef, the_def, defs, {
            for (Uint i = iˇthe_def + 1; i < defs.len; i += 1) {
                Bool const has = astDefHasIdent(&defs.at[i], the_def->anns.name);
                if (has) {
                    AstDef dependant = defs.at[i];
                    defs.at[i] = *the_def;
                    defs.at[iˇthe_def] = dependant;
                    again = true;
                    break;
                }
            }
            if (again)
                break;
        });
    }
}

static void astReorderSubDefs(Ast const* const ast) {
    ·forEach(AstDef, top_def, ast->top_defs, { astSubDefsReorder(top_def->sub_defs); });
}

static void astDefHoistFuncsExprsToNewTopDefs(AstDef const* const def) {
    ·forEach(AstDef, sub_def, def->sub_defs, { astDefHoistFuncsExprsToNewTopDefs(sub_def); });
}

static void astHoistFuncsExprsToNewTopDefs(Ast const* const ast) {
    ·forEach(AstDef, top_def, ast->top_defs, { astDefHoistFuncsExprsToNewTopDefs(top_def); });
}

static void astExprVerifyNoShadowings(AstExpr const* const expr, Strs names_stack, Uint const names_stack_capacity, Ast const* const ast) {
    if (astExprIsTagOrInstr(expr, true, true))
        return;
    switch (expr->kind) {
        case ast_expr_lit_bracket: {
            ·forEach(AstExpr, sub_expr, expr->of_bracket, { astExprVerifyNoShadowings(sub_expr, names_stack, names_stack_capacity, ast); });
        } break;
        case ast_expr_lit_braces: {
            ·forEach(AstExpr, sub_expr, expr->of_braces, { astExprVerifyNoShadowings(sub_expr, names_stack, names_stack_capacity, ast); });
        } break;
        case ast_expr_form: {
            AstExpr* const callee = &expr->of_form.at[0];
            if (astExprIsTagOrInstr(callee, false, true) && strEql(strL("->", 2), callee->of_form.at[1].of_ident)) {
                ·assert(expr->of_form.at[1].kind == ast_expr_lit_bracket);
                AstExprs const params = expr->of_form.at[1].of_bracket;
                ·forEach(AstExpr, param, params, {
                    ·assert(astExprIsTagOrInstr(param, true, false));
                    Str const param_name = param->of_form.at[1].of_ident;
                    for (Uint j = 0; j < names_stack.len; j += 1)
                        if (strEql(names_stack.at[j], param_name))
                            ·fail(astNodeMsg(str("shadowing earlier definition of the same name"), &param->node_base, ast));
                    ·append(names_stack, param_name);
                });
                astExprVerifyNoShadowings(&expr->of_form.at[2], names_stack, names_stack_capacity, ast);
            } else
                ·forEach(AstExpr, sub_expr, expr->of_form, { astExprVerifyNoShadowings(sub_expr, names_stack, names_stack_capacity, ast); });
        } break;
        default: break;
    }
}

static void astDefsVerifyNoShadowings(AstDefs const defs, Strs names_stack, Uint const names_stack_capacity, Ast const* const ast) {
    ·forEach(AstDef, def, defs, {
        for (Uint i = 0; i < names_stack.len; i += 1)
            if (strEql(names_stack.at[i], def->anns.name))
                ·fail(astNodeMsg(str("shadowing earlier definition of the same name"), &def->head.node_base, ast));
        if (names_stack_capacity == names_stack.len)
            ·fail(str("astDefsVerifyNoShadowings: TODO pre-allocate a bigger names_stack"));
        ·append(names_stack, def->anns.name);
    });
    ·forEach(AstDef, def, defs, {
        Uint num_params = 0;
        if (def->head.kind == ast_expr_form) {
            num_params = def->head.of_form.len - 1;
            for (Uint i = 1; i <= num_params; i += 1) {
                ·assert(def->head.of_form.at[i].kind == ast_expr_ident);
                Str const param_name = def->head.of_form.at[i].of_ident;
                for (Uint j = 0; j < names_stack.len; j += 1)
                    if (strEql(names_stack.at[j], param_name))
                        ·fail(astNodeMsg(str("shadowing earlier definition of the same name"), &def->head.of_form.at[i].node_base, ast));

                ·append(names_stack, param_name);
            }
        }
        astDefsVerifyNoShadowings(def->sub_defs, names_stack, names_stack_capacity, ast);
        astExprVerifyNoShadowings(&def->body, names_stack, names_stack_capacity, ast);
        names_stack.len -= num_params;
    });
}



static void astPrint(Ast const* const ast) {
    ·forEach(AstDef, top_def, ast->top_defs, {
        astDefPrint(top_def, 0);
        printChr('\n');
    });
}

static void astDefPrint(AstDef const* const def, Uint const ind) {
    printChr('\n');
    for (Uint i = 0; i < ind; i += 1)
        printChr(' ');
    astExprPrint(&def->head, false, ind);
    printStr(str(" :=\n"));
    for (Uint i = 0; i < 2 + ind; i += 1)
        printChr(' ');
    astExprPrint(&def->body, false, ind + 2);

    ·forEach(AstDef, sub_def, def->sub_defs, {
        printChr('\n');
        astDefPrint(sub_def, 2 + ind);
    });
}

static void astExprPrint(AstExpr const* const expr, Bool const is_form_item, Uint const ind) {
    for (Uint i = 0; i < expr->anns.parensed; i += 1)
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
                astExprPrint(sub_expr, true, ind);
            });
            if (parens)
                printChr(')');
        } break;

        case ast_expr_lit_bracket: {
            printChr('[');
            ·forEach(AstExpr, sub_expr, expr->of_bracket, {
                if (iˇsub_expr != 0)
                    printStr(str(", "));
                astExprPrint(sub_expr, false, ind);
            });
            printChr(']');
        } break;

        case ast_expr_lit_braces: {
            printStr(str("{\n"));
            Uint const ind_next = 2 + ind;
            ·forEach(AstExpr, sub_expr, expr->of_braces, {
                for (Uint i = 0; i < ind_next; i += 1)
                    printChr(' ');
                astExprPrint(sub_expr, false, ind_next);
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
    for (Uint i = 0; i < expr->anns.parensed; i += 1)
        printChr(')');
}
