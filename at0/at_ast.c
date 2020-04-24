#pragma once
#include "metaleap.c"
#include "std_io.c"
#include "at_toks.c"



typedef struct AstNodeBase {
    UInt toks_idx;
    UInt toks_len;
} AstNodeBase;


typedef enum AstExprKind {
    ast_expr_lit_int = 1,     // 123
    ast_expr_lit_str = 2,     // "123"
    ast_expr_ident = 3,       // anyIdentifier                         (also operators)
    ast_expr_form = 4,        // expr1 expr2 expr3 ... exprN           (always: .len >= 2)
    ast_expr_lit_bracket = 5, // [expr1, expr2, expr3, ..., exprN]     (always: .len >= 0)
    ast_expr_lit_braces = 6,  // {expr1, expr2, expr3, ..., exprN}     (always: .len >= 0)
} AstExprKind;

typedef struct AstExpr AstExpr;
typedef ·SliceOf(AstExpr) AstExprs;
struct AstExpr {
    AstNodeBase node_base;
    AstExprKind kind;
    union {
        U64 of_lit_int;
        Str of_lit_str;
        Str of_ident;
        AstExprs of_exprs;
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
    Str name;
    AstExpr body;
    AstDefs sub_defs;
    struct {
        Strs param_names;
        AstNodeBase head_node_base;
        AstDef const* parent_def;
        Str qname;
    } anns;
};


typedef struct Ast {
    Str src;
    Tokens toks;
    AstDefs top_defs;
} Ast;



void astPrintDef(AstDef const* const, UInt const);
void astPrintExpr(AstExpr const* const, Bool const, UInt const);



AstNodeBase astNodeBaseFrom(UInt const toks_idx, UInt const toks_len) {
    return (AstNodeBase) {.toks_idx = toks_idx, .toks_len = toks_len};
}

Tokens astNodeToks(AstNodeBase const* const node, Ast const* const ast) {
    return ·slice(Token, ast->toks, node->toks_idx, node->toks_idx + node->toks_len);
}

Str astNodeMsg(Str const msg_prefix, AstNodeBase const* const node, Ast const* const ast) {
    Tokens const node_toks = astNodeToks(node, ast);
    Str const line_nr = uIntToStr(1 + node_toks.at[0].line_nr, 1, 10);
    Str const toks_src = toksSrc(node_toks, ast->src);
    return str8(node_toks.at[0].file_name, str(":"), line_nr, str(": "), msg_prefix, str(":\n"), toks_src, str("\n"));
}

Str astNodeSrc(AstNodeBase const* const node, Ast const* const ast) {
    return toksSrc(astNodeToks(node, ast), ast->src);
}

AstDef astDef(AstDef const* const parent_def, UInt const all_toks_idx, UInt const toks_len) {
    AstNodeBase node_base = astNodeBaseFrom(all_toks_idx, toks_len);
    return (AstDef) {
        .sub_defs = (AstDefs) {.at = NULL, .len = 0},
        .anns = {.parent_def = parent_def, .head_node_base = node_base, .param_names = (Strs) {.len = 0, .at = NULL}},
        .node_base = node_base,
    };
}

AstExpr astExpr(UInt const toks_idx, UInt const toks_len, AstExprKind const expr_kind, UInt const len_if_non_atomic) {
    return (AstExpr) {
        .node_base = astNodeBaseFrom(toks_idx, toks_len),
        .kind = expr_kind,
        .anns = {.parensed = 0, .toks_throng = false},
        .of_exprs = ·make(AstExpr, len_if_non_atomic, len_if_non_atomic),
    };
}

AstExpr astExprFormEmpty(AstNodeBase const from) {
    return (AstExpr) {.kind = ast_expr_form, .of_exprs = (AstExprs) {.at = NULL, .len = 0}, .node_base = from};
}

AstExpr astExprFormSub(AstExpr const* const ast_expr, UInt const idx_start, UInt const idx_end) {
    ·assert(!(idx_start == 0 && idx_end == ast_expr->of_exprs.len));
    if (idx_end == idx_start)
        return astExprFormEmpty(ast_expr->node_base);

    AstExpr ret_expr = astExpr(ast_expr->of_exprs.at[idx_start].node_base.toks_idx, 0, ast_expr_form, 0);
    ret_expr.anns.toks_throng = ast_expr->anns.toks_throng;
    ret_expr.of_exprs = ·slice(AstExpr, ast_expr->of_exprs, idx_start, idx_end);
    for (UInt i = idx_start; i < idx_end; i += 1)
        ret_expr.node_base.toks_len += ast_expr->of_exprs.at[i].node_base.toks_len;
    return ret_expr;
}

ºUInt astExprFormIndexOfIdent(AstExpr const* const ast_expr, Str const ident) {
    ·assert(ast_expr->kind == ast_expr_form);
    ·forEach(AstExpr, expr, ast_expr->of_exprs, {
        if (expr->kind == ast_expr_ident && strEql(ident, expr->of_ident))
            return ·ok(UInt, iˇexpr);
    });
    return ·none(UInt);
}

AstExprs astExprFormSplit(AstExpr const* const expr, Str const ident, ºStr const ident_stop) {
    ·assert(expr->kind == ast_expr_form);
    UInts indices = ·make(UInt, 0, expr->of_exprs.len);
    ·forEach(AstExpr, sub_expr, expr->of_exprs, {
        if (sub_expr->kind == ast_expr_ident) {
            if (ident_stop.ok && strEql(ident_stop.it, sub_expr->of_ident))
                break;
            if (strEql(ident, sub_expr->of_ident))
                ·append(indices, iˇsub_expr);
        }
    });
    AstExprs ret_exprs = ·make(AstExpr, 0, 1 + indices.len);
    UInt idx_start = 0;
    for (UInt i = 0; i < indices.len; i += 1) {
        ·append(ret_exprs, astExprFormSub(expr, idx_start, indices.at[i]));
        ·assert(ret_exprs.len <= 1 + indices.len);
        idx_start = 1 + indices.at[i];
    }
    if (idx_start == 0)
        return (AstExprs) {.len = 1, .at = (AstExpr*)expr};
    ·append(ret_exprs, astExprFormSub(expr, idx_start, expr->of_exprs.len));
    ·assert(ret_exprs.len <= 1 + indices.len);
    return ret_exprs;
}

AstExpr² astExprFormBreakOn(AstExpr const* const expr, Str const ident, Bool const must_lhs, Bool const must_rhs,
                            ºUInt const must_be_before_idx, Ast const* const ast) {
    if (must_lhs || must_rhs)
        ·assert(ast != NULL);
    ·assert(expr->kind == ast_expr_form);

    AstExpr² ret_tup = (AstExpr²) {.lhs_form = ·none(AstExpr), .rhs_form = ·none(AstExpr), .glyph = NULL};
    ºUInt pos = astExprFormIndexOfIdent(expr, ident);
    if (pos.ok && must_be_before_idx.ok && pos.it >= must_be_before_idx.it)
        pos.ok = false;
    if (pos.ok) {
        ret_tup.glyph = &expr->of_exprs.at[pos.it];
        if (pos.it > 0)
            ret_tup.lhs_form = ·ok(AstExpr, astExprFormSub(expr, 0, pos.it));
        if (pos.it < expr->of_exprs.len - 1)
            ret_tup.rhs_form = ·ok(AstExpr, astExprFormSub(expr, 1 + pos.it, expr->of_exprs.len));
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

Bool astExprIsInstrOrTag(AstExpr const* const expr, Bool const check_is_instr, Bool const check_is_tag, Bool const check_is_tag_ident) {
    AstExpr* const gly = &expr->of_exprs.at[0];
    return expr->kind == ast_expr_form && expr->of_exprs.len == 2 && gly->kind == ast_expr_ident && gly->of_ident.len == 1
           && ((check_is_instr && gly->of_ident.at[0] == '@' && expr->of_exprs.at[1].kind == ast_expr_ident)
               || (check_is_tag && gly->of_ident.at[0] == '#' && (expr->of_exprs.at[1].kind == ast_expr_ident || !check_is_tag_ident)));
}

Bool astExprIsFunc(AstExpr const* const expr) {
    if (expr->kind != ast_expr_form)
        return false;
    AstExpr* const callee = &expr->of_exprs.at[0];
    if (callee->kind == ast_expr_ident)
        return strEql(strL("@->", 3), callee->of_ident);
    return astExprIsInstrOrTag(callee, true, false, false) && strEql(strL("->", 2), callee->of_exprs.at[1].of_ident);
}

Bool astExprHasIdent(AstExpr const* const expr, Str const ident) {
    switch (expr->kind) {
        case ast_expr_ident: {
            return strEql(ident, expr->of_ident);
        } break;
        case ast_expr_form: {
            if (astExprIsInstrOrTag(expr, true, true, true))
                return false;
            ·forEach(AstExpr, sub_expr, expr->of_exprs, {
                if (astExprHasIdent(sub_expr, ident))
                    return true;
            });
        } break;
        case ast_expr_lit_bracket: // fall through to:
        case ast_expr_lit_braces: {
            ·forEach(AstExpr, sub_expr, expr->of_exprs, {
                if (astExprHasIdent(sub_expr, ident))
                    return true;
            });
        } break;
        default: break;
    }
    return false;
}

Bool astDefHasIdent(AstDef const* const def, Str const ident) {
    ·forEach(AstDef, sub_def, def->sub_defs, {
        if (astDefHasIdent(sub_def, ident))
            return true;
    });
    return astExprHasIdent(&def->body, ident);
}

AstExpr astExprIdent(UInt const toks_idx, UInt const toks_len, Str const name) {
    AstExpr ret_expr = astExpr(toks_idx, toks_len, ast_expr_ident, 0);
    ret_expr.of_ident = name;
    return ret_expr;
}

AstExpr astExprInstrOrTag(AstNodeBase const from, Str const name, Bool const tag) {
    AstExpr ret_expr = (AstExpr) {.kind = ast_expr_form, .of_exprs = ·make(AstExpr, 2, 2), .node_base = from, .anns = {.toks_throng = true}};
    ret_expr.of_exprs.at[0] = astExprIdent(from.toks_idx, from.toks_len, strL(tag ? "#" : "@", 1));
    ret_expr.of_exprs.at[1] = astExprIdent(from.toks_idx, from.toks_len, name);
    return ret_expr;
}

void astExprFormNorm(AstExpr* const expr, ºAstExpr const if_empty) {
    if (expr->kind == ast_expr_form) {
        if (expr->of_exprs.len == 1)
            *expr = expr->of_exprs.at[0];
        if (expr->of_exprs.len == 0 && if_empty.ok)
            *expr = if_empty.it;
    }
}



void astExprRewriteGlyphsIntoInstrs(AstExpr* const expr, Ast const* const ast);

Bool astExprRewriteFirstGlyphIntoInstr(AstExpr* const expr, Str const glyph_name, Bool const must_lhs, Bool const must_rhs, Bool const is_func,
                                       Bool const is_sel, Ast const* const ast) {
    AstExpr² maybe;
    maybe = astExprFormBreakOn(expr, glyph_name, false, false, ·none(UInt), ast);
    if (maybe.glyph == NULL)
        return false;
    if (must_lhs && !maybe.lhs_form.ok)
        ·fail(astNodeMsg(str3(str("expected left-hand-side operand for '"), glyph_name, str("' in")), &expr->node_base, ast));
    if (must_rhs && !maybe.rhs_form.ok)
        ·fail(astNodeMsg(str3(str("expected right-hand-side operand for '"), glyph_name, str("' in")), &expr->node_base, ast));
    expr->anns.toks_throng = false;
    expr->of_exprs = ·make(AstExpr, 3, 3);
    expr->of_exprs.at[0] = astExprInstrOrTag(maybe.glyph->node_base, glyph_name, false);
    expr->of_exprs.at[1] = (maybe.lhs_form.ok) ? maybe.lhs_form.it : astExprFormEmpty(expr->node_base);
    expr->of_exprs.at[2] = (maybe.rhs_form.ok) ? maybe.rhs_form.it : astExprFormEmpty(expr->node_base);
    if (is_func) {
        ·forEach(AstExpr, param_expr, expr->of_exprs.at[1].of_exprs, {
            if (param_expr->kind == ast_expr_ident)
                *param_expr = astExprInstrOrTag(param_expr->node_base, param_expr->of_ident, true);
        });
        expr->of_exprs.at[1].kind = ast_expr_lit_bracket;
    } else if (is_sel) {
        if (expr->of_exprs.at[2].of_exprs.len == 1 && expr->of_exprs.at[2].of_exprs.at[0].kind == ast_expr_ident)
            expr->of_exprs.at[2].of_exprs.at[0] =
                astExprInstrOrTag(expr->of_exprs.at[2].of_exprs.at[0].node_base, expr->of_exprs.at[2].of_exprs.at[0].of_ident, true);
    }
    if (maybe.lhs_form.ok) {
        if ((!is_func) && expr->of_exprs.at[1].of_exprs.len == 1)
            expr->of_exprs.at[1] = expr->of_exprs.at[1].of_exprs.at[0];
        astExprRewriteGlyphsIntoInstrs(&expr->of_exprs.at[1], ast);
    }
    if (maybe.rhs_form.ok) {
        if (expr->of_exprs.at[2].of_exprs.len == 1)
            expr->of_exprs.at[2] = expr->of_exprs.at[2].of_exprs.at[0];
        astExprRewriteGlyphsIntoInstrs(&expr->of_exprs.at[2], ast);
    }
    return true;
}

void astExprRewriteOpIntoInstr(AstExpr* const expr, Str const op, Bool const can_nest, Str const op_desc, Ast const* const ast) {
    AstExprs const subjs = astExprFormSplit(expr, op, ·none(Str));
    if (subjs.len < 2)
        ·fail(astNodeMsg(str3(str("expected operands on both sides of '"), op, str("'")), &expr->node_base, ast));
    if (subjs.len > 2 && !can_nest)
        ·fail(astNodeMsg(str3(str("multiple "), op_desc, str(" operators, clarify intent with parens")), &expr->node_base, ast));
    ·forEach(AstExpr, subj, subjs, {
        if (subj->kind == ast_expr_form && subj->of_exprs.len == 0)
            ·fail(astNodeMsg(str3(str("expected operands on both sides of '"), op, str("'")), &expr->node_base, ast));
    });
    AstExpr instr = astExpr(expr->node_base.toks_idx, expr->node_base.toks_len, ast_expr_form, 3);
    instr.of_exprs.at[0] = astExprInstrOrTag(expr->node_base, op, false);
    instr.of_exprs.at[1] = subjs.at[0];
    instr.of_exprs.at[2] = subjs.at[1];
    astExprFormNorm(&instr.of_exprs.at[1], ·none(AstExpr));
    astExprFormNorm(&instr.of_exprs.at[2], ·none(AstExpr));
    for (UInt i = 2; i < subjs.len; i += 1) {
        AstExpr sub_instr = astExpr(expr->node_base.toks_idx, expr->node_base.toks_len, ast_expr_form, 3);
        sub_instr.of_exprs.at[0] = astExprInstrOrTag(expr->node_base, op, false);
        sub_instr.of_exprs.at[1] = instr;
        sub_instr.of_exprs.at[2] = subjs.at[i];
        astExprFormNorm(&sub_instr.of_exprs.at[2], ·none(AstExpr));
        instr = sub_instr;
    }
    astExprRewriteGlyphsIntoInstrs(&instr.of_exprs.at[1], ast);
    astExprRewriteGlyphsIntoInstrs(&instr.of_exprs.at[2], ast);
    *expr = instr;
}

void astExprRewriteGlyphsIntoInstrs(AstExpr* const expr, Ast const* const ast) {
    if (expr->kind == ast_expr_lit_bracket || expr->kind == ast_expr_lit_braces)
        ·forEach(AstExpr, sub_expr, expr->of_exprs, { astExprRewriteGlyphsIntoInstrs(sub_expr, ast); });

    else if (astExprIsFunc(expr))
        astExprRewriteGlyphsIntoInstrs(&expr->of_exprs.at[2], ast);

    else if (expr->kind == ast_expr_form && expr->of_exprs.len != 0 && !astExprIsInstrOrTag(expr, true, true, true)) {
        Bool matched = false;
        // check for `:` key-value-pair sugar { usually: inside, struct: literals, .. }
        if (!matched)
            matched = astExprRewriteFirstGlyphIntoInstr(expr, strL(":", 1), true, true, false, false, ast);
        // check for `.` field-selector sugar (foo.bar.baz)
        if (!matched) {
            ºUInt const idx_dot = astExprFormIndexOfIdent(expr, strL(".", 1));
            matched = idx_dot.ok;
            if (idx_dot.ok)
                astExprRewriteOpIntoInstr(expr, strL(".", 1), true, strL("", 0), ast);
        }
        // check for `? |` but any earlier `->` has prio
        ºUInt const idx_qmark = astExprFormIndexOfIdent(expr, strL("?", 1));
        ºUInt const idx_func = astExprFormIndexOfIdent(expr, strL("->", 2));
        if ((!matched) && idx_qmark.ok && idx_func.ok && idx_func.it < idx_qmark.it)
            matched = astExprRewriteFirstGlyphIntoInstr(expr, strL("->", 2), false, true, true, false, ast);
        if ((!matched) && idx_qmark.ok) {
            matched = true;
            AstExpr instr = astExpr(expr->node_base.toks_idx, expr->node_base.toks_len, ast_expr_form, 3);
            // @if-instr callee
            instr.of_exprs.at[0] = astExprInstrOrTag(expr->node_base, strL("?", 1), false);
            // @if-instr cond
            instr.of_exprs.at[1] = astExprFormSub(expr, 0, idx_qmark.it);
            astExprFormNorm(&instr.of_exprs.at[1], ·ok(AstExpr, astExprInstrOrTag(expr->node_base, strL("true", 4), true)));
            // @if-instr cases
            AstExpr const q_follow = astExprFormSub(expr, 1 + idx_qmark.it, expr->of_exprs.len);
            if (q_follow.of_exprs.len < 3)
                ·fail(astNodeMsg(str("insufficient cases following '?'"), &expr->node_base, ast));
            AstExprs const cases = astExprFormSplit(&q_follow, strL("|", 1), ·ok(Str, strL("?", 1)));
            if (cases.len <= 1)
                ·fail(astNodeMsg(str("insufficient cases following '?'"), &expr->node_base, ast));
            instr.of_exprs.at[2] = astExpr(expr->node_base.toks_idx, expr->node_base.toks_len, ast_expr_lit_braces, cases.len);
            UInt count_arrows = 0;
            ·forEach(AstExpr, case_expr, cases, {
                if (case_expr->kind == ast_expr_form && case_expr->of_exprs.len == 0)
                    ·fail(astNodeMsg(str("expected expression in case"), &case_expr->node_base, ast));
                AstExpr² const arrow =
                    astExprFormBreakOn(case_expr, strL("=>", 2), false, false, astExprFormIndexOfIdent(case_expr, strL("?", 1)), ast);
                AstExpr expr_case = *case_expr;
                if (arrow.glyph != NULL) {
                    count_arrows += 1;
                    expr_case = astExpr(case_expr->node_base.toks_idx, case_expr->node_base.toks_len, ast_expr_form, 3);
                    expr_case.of_exprs.at[0] = astExprInstrOrTag(arrow.glyph->node_base, strL("|", 1), false);
                    if (!arrow.lhs_form.ok)
                        ·fail(astNodeMsg(str("expected expression before '=>'"), &case_expr->node_base, ast));
                    expr_case.of_exprs.at[1] = arrow.lhs_form.it;
                    astExprFormNorm(&expr_case.of_exprs.at[1], ·none(AstExpr));
                    if (!arrow.rhs_form.ok)
                        ·fail(astNodeMsg(str("expected expression after '=>'"), &case_expr->node_base, ast));
                    expr_case.of_exprs.at[2] = arrow.rhs_form.it;
                    astExprFormNorm(&expr_case.of_exprs.at[2], ·none(AstExpr));
                }
                instr.of_exprs.at[2].of_exprs.at[iˇcase_expr] = expr_case;
            });
            if (count_arrows == 0 && cases.len == 2 && idx_qmark.it > 0) {
                AstExprs const* const new_cases = &instr.of_exprs.at[2].of_exprs;
                AstExpr const case_true = new_cases->at[0];
                AstExpr const case_false = new_cases->at[1];
                new_cases->at[0] = astExpr(case_true.node_base.toks_idx, case_true.node_base.toks_len, ast_expr_form, 3);
                new_cases->at[0].of_exprs.at[0] = astExprInstrOrTag(case_true.node_base, strL("|", 1), false);
                new_cases->at[0].of_exprs.at[1] = astExprInstrOrTag(expr->node_base, strL("true", 4), true);
                new_cases->at[0].of_exprs.at[2] = case_true;
                new_cases->at[1] = astExpr(case_false.node_base.toks_idx, case_false.node_base.toks_len, ast_expr_form, 3);
                new_cases->at[1].of_exprs.at[0] = astExprInstrOrTag(case_false.node_base, strL("|", 1), false);
                new_cases->at[1].of_exprs.at[1] = astExprInstrOrTag(expr->node_base, strL("false", 5), true);
                new_cases->at[1].of_exprs.at[2] = case_false;
            } else if (count_arrows != cases.len)
                ·fail(astNodeMsg(str("some cases are lacking '=>'"), &expr->node_base, ast));
            astExprRewriteGlyphsIntoInstrs(&instr.of_exprs.at[1], ast);
            astExprRewriteGlyphsIntoInstrs(&instr.of_exprs.at[2], ast);
            *expr = instr;
        }
        // check for anon-func sugar `->`
        if ((!matched) && idx_func.ok)
            matched = astExprRewriteFirstGlyphIntoInstr(expr, strL("->", 2), false, true, true, false, ast);
        // check for logical operators
        if (!matched) {
            ºUInt const idx_and = astExprFormIndexOfIdent(expr, strL("&&", 2));
            ºUInt const idx_or = astExprFormIndexOfIdent(expr, strL("||", 2));
            if (idx_and.ok && idx_or.ok)
                ·fail(astNodeMsg(str("same precedence for '&&' and '||', clarify intent with parens"), &expr->node_base, ast));
            if (idx_and.ok || idx_or.ok) {
                matched = true;
                Str const op = strL(idx_and.ok ? "&&" : "||", 2);
                astExprRewriteOpIntoInstr(expr, op, true, str("logical"), ast);
            }
        }
        // check for comparison operators
        if (!matched) {
            ºUInt const idx_eq = astExprFormIndexOfIdent(expr, strL("==", 2));
            ºUInt const idx_neq = astExprFormIndexOfIdent(expr, strL("/=", 2));
            ºUInt const idx_geq = astExprFormIndexOfIdent(expr, strL(">=", 2));
            ºUInt const idx_leq = astExprFormIndexOfIdent(expr, strL("<=", 2));
            ºUInt const idx_gt = astExprFormIndexOfIdent(expr, strL(">", 1));
            ºUInt const idx_lt = astExprFormIndexOfIdent(expr, strL("<", 1));
            CInt const n = idx_eq.ok + idx_neq.ok + idx_geq.ok + idx_leq.ok + idx_gt.ok + idx_lt.ok;
            if (n > 1)
                ·fail(astNodeMsg(str("mix of comparison operators, clarify intent with parens"), &expr->node_base, ast));
            else if (n != 0) {
                matched = true;
                Str const op = (idx_gt.ok || idx_lt.ok) ? strL(idx_gt.ok ? ">" : "<", 1)
                                                        : strL(idx_neq.ok ? "/=" : idx_leq.ok ? "<=" : idx_geq.ok ? ">=" : "==", 2);
                astExprRewriteOpIntoInstr(expr, op, false, str("comparison"), ast);
            }
        }
        // check for int arithmetic operators
        if (!matched) {
            ºUInt const idx_add = astExprFormIndexOfIdent(expr, strL("+", 1));
            ºUInt const idx_sub = astExprFormIndexOfIdent(expr, strL("-", 1));
            ºUInt const idx_mul = astExprFormIndexOfIdent(expr, strL("*", 1));
            ºUInt const idx_div = astExprFormIndexOfIdent(expr, strL("/", 1));
            ºUInt const idx_rem = astExprFormIndexOfIdent(expr, strL("\x25", 1));
            CInt const n = idx_add.ok + idx_sub.ok + idx_mul.ok + idx_div.ok + idx_rem.ok;
            if (n > 1)
                ·fail(astNodeMsg(str("mix of arithmetic operators, clarify intent with parens"), &expr->node_base, ast));
            else if (n != 0) {
                matched = true;
                Str op = strL(idx_add.ok ? "+" : idx_sub.ok ? "-" : idx_mul.ok ? "*" : idx_div.ok ? "/" : "\x25", 1);
                astExprRewriteOpIntoInstr(expr, op, idx_mul.ok || idx_add.ok, str("arithmetic"), ast);
            }
        }
        // nothing desugared, traverse normally into form
        if (!matched)
            ·forEach(AstExpr, sub_expr, expr->of_exprs, { astExprRewriteGlyphsIntoInstrs(sub_expr, ast); });
    }
}

void astDefRewriteGlyphsIntoInstrs(AstDef* const def, Ast const* const ast) {
    ·forEach(AstDef, sub_def, def->sub_defs, { astDefRewriteGlyphsIntoInstrs(sub_def, ast); });
    astExprRewriteGlyphsIntoInstrs(&def->body, ast);
}

void astRewriteGlyphsIntoInstrs(Ast const* const ast) {
    ·forEach(AstDef, def, ast->top_defs, { astDefRewriteGlyphsIntoInstrs(def, ast); });
}



void astSubDefsReorder(AstDefs const defs) {
    ·forEach(AstDef, the_def, defs, { astSubDefsReorder(the_def->sub_defs); });

    UInt num_rounds = 0;
    for (Bool again = true; again; num_rounds += 1) {
        again = false;
        if (num_rounds > 42 * defs.len)
            ·fail(str2(str("Circular sub-def dependencies inside "),
                       (defs.at[0].anns.parent_def == NULL) ? str("<top-level>") : defs.at[0].anns.parent_def->name));
        ·forEach(AstDef, the_def, defs, {
            for (UInt i = iˇthe_def + 1; i < defs.len; i += 1) {
                Bool const has = astDefHasIdent(&defs.at[i], the_def->name);
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

void astReorderSubDefs(Ast const* const ast) {
    ·forEach(AstDef, top_def, ast->top_defs, { astSubDefsReorder(top_def->sub_defs); });
}



void astPrint(Ast const* const ast) {
    ·forEach(AstDef, top_def, ast->top_defs, {
        astPrintDef(top_def, 0);
        printChr('\n');
    });
}

void astPrintDef(AstDef const* const def, UInt const ind) {
    printChr('\n');
    for (UInt i = 0; i < ind; i += 1)
        printChr(' ');
    printStr(def->name);
    printStr(str(" :=\n"));
    for (UInt i = 0; i < 2 + ind; i += 1)
        printChr(' ');
    astPrintExpr(&def->body, false, ind + 2);

    ·forEach(AstDef, sub_def, def->sub_defs, {
        printChr('\n');
        astPrintDef(sub_def, 2 + ind);
    });
}

void astPrintExpr(AstExpr const* const expr, Bool const is_form_item, UInt const ind) {
    for (UInt i = 0; i < expr->anns.parensed; i += 1)
        printChr('(');
    switch (expr->kind) {
        case ast_expr_ident: {
            printStr(expr->of_ident);
        } break;

        case ast_expr_lit_int: {
            printStr(uIntToStr(expr->of_lit_int, 1, 10));
        } break;

        case ast_expr_lit_str: {
            printStr(strQuot(expr->of_lit_str));
        } break;

        case ast_expr_form: {
            if (expr->of_exprs.len == 0)
                break;
            Bool const parens = (!astExprIsInstrOrTag(expr, true, true, true)) && is_form_item
                                && (expr->node_base.toks_len == 0 || (expr->anns.parensed == 0 && !expr->anns.toks_throng));
            if (parens)
                printChr('(');
            ·forEach(AstExpr, sub_expr, expr->of_exprs, {
                if (iˇsub_expr != 0 && (expr->node_base.toks_len == 0 || !expr->anns.toks_throng))
                    printChr(' ');
                astPrintExpr(sub_expr, true, ind);
            });
            if (parens)
                printChr(')');
        } break;

        case ast_expr_lit_bracket: {
            printChr('[');
            ·forEach(AstExpr, sub_expr, expr->of_exprs, {
                if (iˇsub_expr != 0)
                    printStr(str(", "));
                astPrintExpr(sub_expr, false, ind);
            });
            printChr(']');
        } break;

        case ast_expr_lit_braces: {
            if (expr->of_exprs.len == 0)
                printStr(str("{}"));
            else {
                printStr(str("{\n"));
                UInt const ind_next = 2 + ind;
                ·forEach(AstExpr, sub_expr, expr->of_exprs, {
                    for (UInt i = 0; i < ind_next; i += 1)
                        printChr(' ');
                    astPrintExpr(sub_expr, false, ind_next);
                    printStr(str(",\n"));
                });
                for (UInt i = 0; i < ind; i += 1)
                    printChr(' ');
                printChr('}');
            }
        } break;

        default: {
            ·fail(str2(str("TODO: astPrintExpr for .kind of "), uIntToStr(expr->kind, 1, 10)));
        } break;
    }
    for (UInt i = 0; i < expr->anns.parensed; i += 1)
        printChr(')');
}
