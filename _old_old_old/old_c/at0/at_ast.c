#pragma once
#include "utils_std_mem.c"
#include "utils_toks.c"



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
    SrcFileIssues issues;
    struct {
        Str src_file_path;
        Str path_based_ident_prefix;
        UInt total_nr_of_def_toks;
        UInt total_nr_of_tag_toks;
        Strs incl_file_paths;
        Strs all_top_def_names;
    } anns;
} Ast;
typedef ·ListOf(Ast) Asts;



void astPrintDef(AstDef const* const, UInt const);
void astPrintExpr(AstExpr const* const, Bool const, UInt const);



AstNodeBase astNodeBaseFrom(UInt const toks_idx, UInt const toks_len) {
    return (AstNodeBase) {.toks_idx = toks_idx, .toks_len = toks_len};
}

Tokens astNodeToks(AstNodeBase const* const node, Ast const* const ast) {
    return ·slice(Token, ast->toks, node->toks_idx, node->toks_idx + node->toks_len);
}

Str astNodeMsg(Str const msg_prefix, AstNodeBase const* const node, Ast const* const ast) {
    if (node == NULL || ast == NULL)
        return str2(NULL, msg_prefix, str("\n"));
    Tokens const node_toks = astNodeToks(node, ast);
    Str const toks_src = toksSrc(node_toks, ast->src);
    return str6(NULL, tokPosStr(NULL, &node_toks.at[0]), str(": "), msg_prefix, str(":\n"), toks_src, str("\n"));
}

Str astNodeSrc(AstNodeBase const* const node, Ast const* const ast) {
    return toksSrc(astNodeToks(node, ast), ast->src);
}

AstDef astDef(AstDef const* const parent_def, UInt const all_toks_idx, UInt const toks_len) {
    AstNodeBase node_base = astNodeBaseFrom(all_toks_idx, toks_len);
    return (AstDef) {
        .sub_defs = ·len0(AstDef),
        .anns = {.parent_def = parent_def, .head_node_base = node_base, .param_names = ·len0(Str)},
        .node_base = node_base,
    };
}

AstExpr astExpr(UInt const toks_idx, UInt const toks_len, AstExprKind const expr_kind, UInt const len_if_non_atomic) {
    return (AstExpr) {
        .node_base = astNodeBaseFrom(toks_idx, toks_len),
        .kind = expr_kind,
        .anns = {.parensed = 0, .toks_throng = false},
        .of_exprs = ·sliceOf(AstExpr, NULL, len_if_non_atomic, len_if_non_atomic),
    };
}

AstExpr astExprFormEmpty(AstNodeBase const from) {
    return (AstExpr) {.kind = ast_expr_form, .of_exprs = ·len0(AstExpr), .node_base = from};
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
            return ·got(UInt, iˇexpr);
    });
    return ·none(UInt);
}

AstExprs astExprFormSplit(AstExpr const* const expr, Str const ident, ºStr const ident_stop) {
    ·assert(expr->kind == ast_expr_form);
    UInts indices = ·sliceOf(UInt, NULL, 0, expr->of_exprs.len);
    ·forEach(AstExpr, sub_expr, expr->of_exprs, {
        if (sub_expr->kind == ast_expr_ident) {
            if (ident_stop.got && strEql(ident_stop.it, sub_expr->of_ident))
                break;
            if (strEql(ident, sub_expr->of_ident))
                ·push(indices, iˇsub_expr);
        }
    });
    AstExprs ret_exprs = ·sliceOf(AstExpr, NULL, 0, 1 + indices.len);
    UInt idx_start = 0;
    for (UInt i = 0; i < indices.len; i += 1) {
        ·push(ret_exprs, astExprFormSub(expr, idx_start, indices.at[i]));
        ·assert(ret_exprs.len <= 1 + indices.len);
        idx_start = 1 + indices.at[i];
    }
    if (idx_start == 0)
        return (AstExprs) {.len = 1, .at = (AstExpr*)expr};
    ·push(ret_exprs, astExprFormSub(expr, idx_start, expr->of_exprs.len));
    ·assert(ret_exprs.len <= 1 + indices.len);
    return ret_exprs;
}

AstExpr² astExprFormBreakOn(AstExpr const* const expr, Str const ident, Bool const must_lhs, Bool const must_rhs,
                            ºUInt const must_be_before_idx1, ºUInt const must_be_before_idx2, Ast const* const ast,
                            ºSrcFileIssue* const issue) {
    if (must_lhs || must_rhs)
        ·assert(ast != NULL);
    ·assert(expr->kind == ast_expr_form);

    AstExpr² ret_tup = (AstExpr²) {.lhs_form = ·none(AstExpr), .rhs_form = ·none(AstExpr), .glyph = NULL};
    ºUInt pos = astExprFormIndexOfIdent(expr, ident);
    if (pos.got
        && ((must_be_before_idx2.got && pos.it >= must_be_before_idx2.it) || (must_be_before_idx1.got && pos.it >= must_be_before_idx1.it)))
        pos.got = false;
    if (pos.got) {
        ret_tup.glyph = &expr->of_exprs.at[pos.it];
        if (pos.it > 0)
            ret_tup.lhs_form = ·got(AstExpr, astExprFormSub(expr, 0, pos.it));
        if (pos.it < expr->of_exprs.len - 1)
            ret_tup.rhs_form = ·got(AstExpr, astExprFormSub(expr, 1 + pos.it, expr->of_exprs.len));
    }
    Bool const must_both = must_lhs && must_rhs;
    if (must_both && !pos.got)
        *issue = srcFileErr(issue_parse, str3(NULL, str("expected '"), ident, str("'")), ast->toks.at[expr->node_base.toks_idx]);
    else if (must_lhs && !ret_tup.lhs_form.got)
        *issue =
            srcFileErr(issue_parse, str3(NULL, str("expected expression before '"), ident, str("'")), ast->toks.at[expr->node_base.toks_idx]);
    else if (must_rhs && !ret_tup.rhs_form.got)
        *issue = srcFileErr(issue_parse, str3(NULL, str("expected expression following '"), ident, str("'")),
                            ast->toks.at[expr->node_base.toks_idx]);
    return ret_tup;
}

Bool astExprIsInstrOrTag(AstExpr const* const expr, Bool const check_is_instr, Bool const check_is_tag, Bool const check_is_tag_ident) {
    AstExpr* const gly = &expr->of_exprs.at[0];
    return expr->kind == ast_expr_form && expr->of_exprs.len == 2 && gly->kind == ast_expr_ident && gly->of_ident.len == 1
           && ((check_is_instr && gly->of_ident.at[0] == '@' && expr->of_exprs.at[1].kind == ast_expr_ident)
               || (check_is_tag && gly->of_ident.at[0] == '#' && (expr->of_exprs.at[1].kind == ast_expr_ident || !check_is_tag_ident)));
}

Str astExprIsIncl(AstExpr const* const expr) {
    if (expr->kind == ast_expr_form && expr->of_exprs.at[1].kind == ast_expr_lit_str && expr->of_exprs.at[0].kind == ast_expr_ident
        && expr->of_exprs.at[0].of_ident.at[0] == '@' && expr->of_exprs.at[0].of_ident.len == 1)
        return expr->of_exprs.at[1].of_lit_str;
    return ·len0(U8);
}

Bool astExprsHaveNonAtoms(AstExprs const exprs) {
    ·forEach(AstExpr, expr, exprs, {
        if (expr->kind == ast_expr_lit_bracket || expr->kind == ast_expr_lit_braces
            || (expr->kind == ast_expr_form && !astExprIsInstrOrTag(expr, true, true, true)))
            return true;
    });
    return false;
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
    AstExpr ret_expr =
        (AstExpr) {.kind = ast_expr_form, .of_exprs = ·sliceOf(AstExpr, NULL, 2, 2), .node_base = from, .anns = {.toks_throng = true}};
    ret_expr.of_exprs.at[0] = astExprIdent(from.toks_idx, from.toks_len, strL(tag ? "#" : "@", 1));
    ret_expr.of_exprs.at[1] = astExprIdent(from.toks_idx, from.toks_len, name);
    return ret_expr;
}

void astExprFormNorm(AstExpr* const expr, ºAstExpr const if_empty) {
    if (expr->kind == ast_expr_form) {
        if (expr->of_exprs.len == 1)
            *expr = expr->of_exprs.at[0];
        if (expr->of_exprs.len == 0 && if_empty.got)
            *expr = if_empty.it;
    }
}

Bool astExprsThronged(AstExpr* const former, AstExpr* const latter, Ast const* const ast) {
    Tokens const toks_lhs = astNodeToks(&former->node_base, ast);
    Tokens const toks_rhs = astNodeToks(&latter->node_base, ast);
    Token* const tok_lhs = ·last(toks_lhs);
    Token* const tok_rhs = &toks_rhs.at[0];
    return tok_rhs->char_pos == (tok_lhs->char_pos + tok_lhs->str_len);
}



ºSrcFileIssue astExprRewriteGlyphsIntoInstrs(AstExpr* const expr, Ast const* const ast, Bool const colon_to_call);

Bool astExprRewriteFirstGlyphIntoInstr(AstExpr* const expr, Str const glyph_name, Bool const must_lhs, Bool const must_rhs,
                                       Bool const is_func, Bool const is_sel, ºUInt const must_be_before_idx1,
                                       ºUInt const must_be_before_idx2, Ast const* const ast, ºSrcFileIssue* const issue) {
    AstExpr² const split = astExprFormBreakOn(expr, glyph_name, false, false, must_be_before_idx1, must_be_before_idx2, ast, issue);
    if (!(issue->got || split.glyph == NULL)) {
        if (must_lhs && !split.lhs_form.got)
            *issue = srcFileErr(issue_parse, str3(NULL, str("expected left-hand-side operand for '"), glyph_name, str("'")),
                                ast->toks.at[expr->node_base.toks_idx]);
        else if (must_rhs && !split.rhs_form.got)
            *issue = srcFileErr(issue_parse, str3(NULL, str("expected right-hand-side operand for '"), glyph_name, str("'")),
                                ast->toks.at[expr->node_base.toks_idx]);
    }
    if (issue->got || split.glyph == NULL)
        return false;
    expr->anns.toks_throng = false;
    expr->of_exprs = ·sliceOf(AstExpr, NULL, 3, 3);
    expr->of_exprs.at[0] = astExprInstrOrTag(split.glyph->node_base, glyph_name, false);
    expr->of_exprs.at[1] = (split.lhs_form.got) ? split.lhs_form.it : astExprFormEmpty(expr->node_base);
    expr->of_exprs.at[2] = (split.rhs_form.got) ? split.rhs_form.it : astExprFormEmpty(expr->node_base);
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
    if (split.lhs_form.got) {
        if ((!is_func) && expr->of_exprs.at[1].of_exprs.len == 1)
            expr->of_exprs.at[1] = expr->of_exprs.at[1].of_exprs.at[0];
        *issue = astExprRewriteGlyphsIntoInstrs(&expr->of_exprs.at[1], ast, true);
    }
    if (split.rhs_form.got && !issue->got) {
        if (expr->of_exprs.at[2].of_exprs.len == 1)
            expr->of_exprs.at[2] = expr->of_exprs.at[2].of_exprs.at[0];
        *issue = astExprRewriteGlyphsIntoInstrs(&expr->of_exprs.at[2], ast, true);
    }
    return true;
}

ºSrcFileIssue astExprRewriteOpIntoInstr(AstExpr* const expr, Str const op, Bool const can_nest, Str const op_desc, Ast const* const ast) {
    ºSrcFileIssue issue = ·none(SrcFileIssue);
    AstExprs const subjs = astExprFormSplit(expr, op, ·none(Str));
    if (subjs.len < 2)
        return srcFileErr(issue_parse, str3(NULL, str("expected operands on both sides of '"), op, str("'")),
                          ast->toks.at[expr->node_base.toks_idx]);
    if (subjs.len > 2 && !can_nest)
        return srcFileErr(issue_parse, str3(NULL, str("multiple "), op_desc, str(" operators, clarify intent with parens")),
                          ast->toks.at[expr->node_base.toks_idx]);
    ·forEach(AstExpr, subj, subjs, {
        if (subj->kind == ast_expr_form && subj->of_exprs.len == 0)
            return srcFileErr(issue_parse, str3(NULL, str("expected operands on both sides of '"), op, str("'")),
                              ast->toks.at[expr->node_base.toks_idx]);
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
    issue = astExprRewriteGlyphsIntoInstrs(&instr.of_exprs.at[1], ast, true);
    if (!issue.got)
        issue = astExprRewriteGlyphsIntoInstrs(&instr.of_exprs.at[2], ast, true);
    if (!issue.got)
        *expr = instr;
    return issue;
}

ºSrcFileIssue astExprRewriteGlyphsIntoInstrs(AstExpr* const expr, Ast const* const ast, Bool const colon_to_call) {
    ºSrcFileIssue issue = ·none(SrcFileIssue);
    if (expr->kind == ast_expr_lit_bracket || expr->kind == ast_expr_lit_braces)
        ·forEach(AstExpr, sub_expr, expr->of_exprs, {
            issue = astExprRewriteGlyphsIntoInstrs(sub_expr, ast, expr->kind == ast_expr_lit_bracket);
            if (issue.got)
                return issue;
        });

    else if (astExprIsFunc(expr))
        issue = astExprRewriteGlyphsIntoInstrs(&expr->of_exprs.at[2], ast, true);

    else if (expr->kind == ast_expr_form && expr->of_exprs.len != 0 && !astExprIsInstrOrTag(expr, true, true, true)) {
        Bool matched = false; // each desugarer checks for it and sets it, so they can be re-ordered
        ºUInt const idx_qmark = astExprFormIndexOfIdent(expr, strL("?", 1));
        ºUInt const idx_func = astExprFormIndexOfIdent(expr, strL("->", 2));
        ºUInt const idx_colon = astExprFormIndexOfIdent(expr, strL(":", 1));

        // check for `:` key-value-pair sugar { usually: inside, struct: literals, .. }
        if ((!matched) && idx_colon.got) {
            if (!(colon_to_call && idx_colon.it > 0 && idx_colon.it < expr->of_exprs.len - 1))
                matched = astExprRewriteFirstGlyphIntoInstr(expr, strL(":", 1), true, true, false, false, idx_qmark, idx_func, ast, &issue);
            else {
                matched = true;
                AstExpr lhs = astExprFormSub(expr, 0, idx_colon.it);
                AstExpr rhs = astExprFormSub(expr, 1 + idx_colon.it, expr->of_exprs.len);
                issue = astExprRewriteGlyphsIntoInstrs(&lhs, ast, true);
                if (!issue.got)
                    issue = astExprRewriteGlyphsIntoInstrs(&rhs, ast, true);
                if (issue.got)
                    return issue;
                lhs.anns.parensed = 1;
                rhs.anns.parensed = 1;
                AstExpr call = astExpr(expr->node_base.toks_idx, expr->node_base.toks_len, ast_expr_form, 2);
                call.anns = expr->anns;
                call.of_exprs.at[0] = lhs;
                call.of_exprs.at[1] = rhs;
                *expr = call;
            }
            if (issue.got)
                return issue;
        }
        // range sugars: foo..bar , foo...bar
        if (!matched) {
            matched = astExprRewriteFirstGlyphIntoInstr(expr, strL("...", 3), true, false, false, false, idx_qmark, idx_func, ast, &issue);
            if (!(matched || issue.got))
                matched = astExprRewriteFirstGlyphIntoInstr(expr, strL("..", 2), true, false, false, false, idx_qmark, idx_func, ast, &issue);
            if (issue.got)
                return issue;
        }
        // next, sanitize `foo[1].bar[2].bar[3]` into `((((foo[1]).bar)[2]).bar)[3]`
        // before traversing into sub-desugarings
        if (!matched) {
            AstExpr* const last = ·last(expr->of_exprs);
            if ((last->kind == ast_expr_lit_braces || last->kind == ast_expr_lit_bracket)
                && astExprsThronged(&expr->of_exprs.at[expr->of_exprs.len - 2], last, ast)) {
                matched = true;
                AstExpr instr = astExpr(expr->node_base.toks_idx, expr->node_base.toks_len, ast_expr_form, 3);
                instr.of_exprs.at[0] = astExprInstrOrTag(expr->node_base, str(last->kind == ast_expr_lit_bracket ? "ª" : "º"), false);
                instr.of_exprs.at[1] = *last;
                instr.of_exprs.at[2] = astExprFormSub(expr, 0, expr->of_exprs.len - 1);
                astExprFormNorm(&instr.of_exprs.at[2], ·none(AstExpr));
                *expr = instr;
                issue = astExprRewriteGlyphsIntoInstrs(expr, ast, true);
                if (issue.got)
                    return issue;
            }
        }
        for (UInt i = expr->of_exprs.len - 1; i > 0 && !matched; i -= 1) {
            AstExpr* const cur = &expr->of_exprs.at[i];
            Bool const is_last = (i == expr->of_exprs.len - 1);
            if (!astExprsThronged(&expr->of_exprs.at[i - 1], cur, ast))
                continue;
            if ((!is_last) && cur->kind == ast_expr_ident && cur->of_ident.len == 1 && cur->of_ident.at[0] == '.') {
                matched = true;
                AstExpr instr = astExpr(expr->node_base.toks_idx, expr->node_base.toks_len, ast_expr_form, 3);
                instr.of_exprs.at[0] = astExprInstrOrTag(expr->node_base, strL(".", 1), false);
                instr.of_exprs.at[1] = astExprFormSub(expr, 0, i);
                instr.of_exprs.at[2] = astExprFormSub(expr, i + 1, expr->of_exprs.len);
                astExprFormNorm(&instr.of_exprs.at[1], ·none(AstExpr));
                astExprFormNorm(&instr.of_exprs.at[2], ·none(AstExpr));
                *expr = instr;
                issue = astExprRewriteGlyphsIntoInstrs(expr, ast, true);
                if (issue.got)
                    return issue;
                if (expr->of_exprs.at[2].kind != ast_expr_ident)
                    return srcFileErr(issue_parse, str("illegal '.' right-hand-side expression"),
                                      ast->toks.at[expr->of_exprs.at[2].node_base.toks_idx]);
                expr->of_exprs.at[2] = astExprInstrOrTag(expr->of_exprs.at[2].node_base, expr->of_exprs.at[2].of_ident, true);
            }
        }
        // check for `? |` but any earlier `->` has prio
        if ((!matched) && idx_qmark.got && idx_func.got && idx_func.it < idx_qmark.it) {
            matched = astExprRewriteFirstGlyphIntoInstr(expr, strL("->", 2), false, true, true, false, ·none(UInt), ·none(UInt), ast, &issue);
            if (issue.got)
                return issue;
        }
        if ((!matched) && idx_qmark.got) {
            matched = true;
            AstExpr instr = astExpr(expr->node_base.toks_idx, expr->node_base.toks_len, ast_expr_form, 3);
            // @if-instr callee
            instr.of_exprs.at[0] = astExprInstrOrTag(expr->of_exprs.at[idx_qmark.it].node_base, strL("?", 1), false);
            // @if-instr cond
            instr.of_exprs.at[1] = astExprFormSub(expr, 0, idx_qmark.it);
            astExprFormNorm(&instr.of_exprs.at[1],
                            ·got(AstExpr, astExprInstrOrTag(expr->of_exprs.at[idx_qmark.it].node_base, strL("true", 4), true)));
            // @if-instr cases
            AstExpr const q_follow = astExprFormSub(expr, 1 + idx_qmark.it, expr->of_exprs.len);
            if (q_follow.of_exprs.len < 3)
                return srcFileErr(issue_parse, str("insufficient cases following '?'"), ast->toks.at[expr->node_base.toks_idx]);
            AstExprs const cases = astExprFormSplit(&q_follow, strL("|", 1), ·got(Str, strL("?", 1)));
            if (cases.len <= 1)
                return srcFileErr(issue_parse, str("insufficient cases following '?'"), ast->toks.at[expr->node_base.toks_idx]);
            instr.of_exprs.at[2] = astExpr(expr->of_exprs.at[1 + idx_qmark.it].node_base.toks_idx,
                                           expr->node_base.toks_len - (1 + idx_qmark.it), ast_expr_lit_bracket, cases.len);
            UInt count_arrows = 0;
            ·forEach(AstExpr, case_expr, cases, {
                if (case_expr->kind == ast_expr_form && case_expr->of_exprs.len == 0)
                    return srcFileErr(issue_parse, str("expected expression in case"), ast->toks.at[case_expr->node_base.toks_idx]);
                AstExpr² const arrow = astExprFormBreakOn(case_expr, strL("=>", 2), false, false,
                                                          astExprFormIndexOfIdent(case_expr, strL("?", 1)), ·none(UInt), ast, &issue);
                if (issue.got)
                    return issue;
                AstExpr expr_case = *case_expr;
                if (arrow.glyph == NULL)
                    astExprFormNorm(&expr_case, ·none(AstExpr));
                else {
                    count_arrows += 1;
                    expr_case = astExpr(case_expr->node_base.toks_idx, case_expr->node_base.toks_len, ast_expr_form, 3);
                    expr_case.of_exprs.at[0] = astExprInstrOrTag(arrow.glyph->node_base, strL("|", 1), false);
                    if ((!arrow.lhs_form.got) || arrow.lhs_form.it.of_exprs.len == 0)
                        return srcFileErr(issue_parse, str("expected expression before '=>'"), ast->toks.at[case_expr->node_base.toks_idx]);
                    expr_case.of_exprs.at[1] = arrow.lhs_form.it;
                    astExprFormNorm(&expr_case.of_exprs.at[1], ·none(AstExpr));
                    if ((!arrow.rhs_form.got) || arrow.rhs_form.it.of_exprs.len == 0)
                        return srcFileErr(issue_parse, str("expected expression after '=>'"), ast->toks.at[case_expr->node_base.toks_idx]);
                    expr_case.of_exprs.at[2] = arrow.rhs_form.it;
                    astExprFormNorm(&expr_case.of_exprs.at[2], ·none(AstExpr));
                }
                instr.of_exprs.at[2].of_exprs.at[iˇcase_expr] = expr_case;
            });
            if (count_arrows == 0 && cases.len == 2 && idx_qmark.it > 0) {
                AstExprs const* const new_cases = &instr.of_exprs.at[2].of_exprs;
                ·assert(new_cases->len == 2);
                AstExpr const expr_case_true = new_cases->at[0];
                AstExpr const expr_case_false = new_cases->at[1];

                new_cases->at[0] = astExpr(expr->of_exprs.at[idx_qmark.it].node_base.toks_idx,
                                           expr->of_exprs.at[idx_qmark.it].node_base.toks_len, ast_expr_form, 3);
                new_cases->at[0].of_exprs.at[0] = astExprInstrOrTag(expr->of_exprs.at[idx_qmark.it].node_base, strL("|", 1), false);
                new_cases->at[0].of_exprs.at[1] = astExprInstrOrTag(expr->of_exprs.at[idx_qmark.it].node_base, strL("true", 4), true);
                new_cases->at[0].of_exprs.at[2] = expr_case_true;

                new_cases->at[1] = astExpr(expr_case_false.node_base.toks_idx - 1, 1, ast_expr_form, 3);
                new_cases->at[1].of_exprs.at[0] =
                    astExprInstrOrTag((AstNodeBase) {.toks_idx = expr_case_false.node_base.toks_idx - 1, .toks_len = 1}, strL("|", 1), false);
                new_cases->at[1].of_exprs.at[1] = astExprInstrOrTag(
                    (AstNodeBase) {.toks_idx = expr_case_false.node_base.toks_idx - 1, .toks_len = 1}, strL("false", 5), true);
                new_cases->at[1].of_exprs.at[2] = expr_case_false;
            } else if (count_arrows != cases.len)
                return srcFileErr(issue_parse, str("some cases are lacking '=>'"), ast->toks.at[expr->node_base.toks_idx]);
            issue = astExprRewriteGlyphsIntoInstrs(&instr.of_exprs.at[1], ast, true);
            if (!issue.got)
                issue = astExprRewriteGlyphsIntoInstrs(&instr.of_exprs.at[2], ast, true);
            if (issue.got)
                return issue;
            *expr = instr;
        }
        // check for anon-func sugar `->`
        if ((!matched) && idx_func.got) {
            matched = astExprRewriteFirstGlyphIntoInstr(expr, strL("->", 2), false, true, true, false, ·none(UInt), ·none(UInt), ast, &issue);
            if (issue.got)
                return issue;
        }
        // check for logical operators
        if (!matched) {
            ºUInt const idx_and = astExprFormIndexOfIdent(expr, strL("&&", 2));
            ºUInt const idx_or = astExprFormIndexOfIdent(expr, strL("||", 2));
            if (idx_and.got && idx_or.got)
                return srcFileErr(issue_parse, str("same precedence for '&&' and '||', clarify intent with parens"),
                                  ast->toks.at[expr->node_base.toks_idx]);
            if (idx_and.got || idx_or.got) {
                matched = true;
                Str const op = strL(idx_and.got ? "&&" : "||", 2);
                issue = astExprRewriteOpIntoInstr(expr, op, true, str("logical"), ast);
                if (issue.got)
                    return issue;
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
            int const n = idx_eq.got + idx_neq.got + idx_geq.got + idx_leq.got + idx_gt.got + idx_lt.got;
            if (n > 1)
                return srcFileErr(issue_parse, str("mix of comparison operators, clarify intent with parens"),
                                  ast->toks.at[expr->node_base.toks_idx]);
            else if (n != 0) {
                matched = true;
                Str const op = (idx_gt.got || idx_lt.got) ? strL(idx_gt.got ? ">" : "<", 1)
                                                          : strL(idx_neq.got ? "/=" : idx_leq.got ? "<=" : idx_geq.got ? ">=" : "==", 2);
                issue = astExprRewriteOpIntoInstr(expr, op, false, str("comparison"), ast);
                if (issue.got)
                    return issue;
            }
        }
        // check for int arithmetic operators
        if (!matched) {
            ºUInt const idx_add = astExprFormIndexOfIdent(expr, strL("+", 1));
            ºUInt const idx_sub = astExprFormIndexOfIdent(expr, strL("-", 1));
            ºUInt const idx_mul = astExprFormIndexOfIdent(expr, strL("*", 1));
            ºUInt const idx_div = astExprFormIndexOfIdent(expr, strL("/", 1));
            ºUInt const idx_rem = astExprFormIndexOfIdent(expr, strL("\x25", 1));
            int const n = idx_add.got + idx_sub.got + idx_mul.got + idx_div.got + idx_rem.got;
            if (n > 1)
                return srcFileErr(issue_parse, str("mix of arithmetic operators, clarify intent with parens"),
                                  ast->toks.at[expr->node_base.toks_idx]);
            else if (n != 0) {
                matched = true;
                Str op = strL(idx_add.got ? "+" : idx_sub.got ? "-" : idx_mul.got ? "*" : idx_div.got ? "/" : "\x25", 1);
                issue = astExprRewriteOpIntoInstr(expr, op, idx_mul.got || idx_add.got, str("arithmetic"), ast);
                if (issue.got)
                    return issue;
            }
        }
        // check for `.` field-selector sugar (foo.bar.baz)
        if (!matched) {
            ºUInt const idx_dot = astExprFormIndexOfIdent(expr, strL(".", 1));
            matched = idx_dot.got;
            if (idx_dot.got)
                issue = astExprRewriteOpIntoInstr(expr, strL(".", 1), false, strL("", 0), ast);
            if (issue.got)
                return issue;
        }
        // nothing desugared, traverse normally into form
        if (!matched)
            ·forEach(AstExpr, sub_expr, expr->of_exprs, {
                issue = astExprRewriteGlyphsIntoInstrs(sub_expr, ast, true);
                if (issue.got)
                    return issue;
            });
    }
    return issue;
}

void astDefRewriteGlyphsIntoInstrs(AstDef* const def, Ast* const ast) {
    ·forEach(AstDef, sub_def, def->sub_defs, { astDefRewriteGlyphsIntoInstrs(sub_def, ast); });
    ºSrcFileIssue issue = astExprRewriteGlyphsIntoInstrs(&def->body, ast, true);
    if (issue.got)
        ·push(ast->issues, issue.it);
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
            printStr(uIntToStr(NULL, expr->of_lit_int, 1, 10));
        } break;

        case ast_expr_lit_str: {
            printStr(strQuot(NULL, expr->of_lit_str));
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

        case ast_expr_lit_bracket: // fall through to:
        case ast_expr_lit_braces: {
            U8 const sep1 = (expr->kind == ast_expr_lit_bracket ? '[' : '{');
            U8 const sep2 = (expr->kind == ast_expr_lit_bracket ? ']' : '}');
            if (expr->of_exprs.len == 0) {
                printChr(sep1);
                printChr(sep2);
            } else if (astExprsHaveNonAtoms(expr->of_exprs)) {
                printChr(sep1);
                printChr('\n');
                UInt const ind_next = 2 + ind;
                ·forEach(AstExpr, sub_expr, expr->of_exprs, {
                    for (UInt i = 0; i < ind_next; i += 1)
                        printChr(' ');
                    astPrintExpr(sub_expr, false, ind_next);
                    printStr(str(",\n"));
                });
                for (UInt i = 0; i < ind; i += 1)
                    printChr(' ');
                printChr(sep2);
            } else {
                printChr(sep1);
                ·forEach(AstExpr, sub_expr, expr->of_exprs, {
                    if (iˇsub_expr != 0)
                        printStr(str(", "));
                    astPrintExpr(sub_expr, false, ind);
                });
                printChr(sep2);
            }
        } break;

        default: {
            ·fail(str2(NULL, str("TODO: astPrintExpr for .kind of "), uIntToStr(NULL, expr->kind, 1, 10)));
        } break;
    }
    for (UInt i = 0; i < expr->anns.parensed; i += 1)
        printChr(')');
}




typedef struct CtxAsts {
    Asts asts;
    Strs src_file_paths;
    Bool issues;
} CtxAsts;

Bool astsIssues(CtxAsts* const ctx, Bool const update) {
    if ((!ctx->issues) && update)
        ·forEach(Ast, ast, ctx->asts, {
            if (ast->issues.len > 0)
                ctx->issues = true;
        });
    return ctx->issues;
}

AstExpr* astExprFindNode(Ast* const ast, AstExpr* const expr, UInt const pos_line, UInt const pos_col) {
    if (toksHavePos(astNodeToks(&expr->node_base, ast), pos_line, pos_col)) {
        if (expr->kind == ast_expr_form || expr->kind == ast_expr_lit_braces || expr->kind == ast_expr_lit_bracket)
            ·forEach(AstExpr, sub_expr, expr->of_exprs, {
                AstExpr* found = astExprFindNode(ast, sub_expr, pos_line, pos_col);
                if (found != NULL)
                    return found;
            });
        else if (ast->toks.at[expr->node_base.toks_idx].line_nr == pos_line) { // always true: (expr->node_base.toks_len == 1)
            UInt pos_col_start = tokPosCol(&ast->toks.at[expr->node_base.toks_idx]);
            UInt pos_col_end = pos_col_start + ast->toks.at[expr->node_base.toks_idx].str_len;
            if (pos_col >= pos_col_start && pos_col < pos_col_end)
                return expr;
        }
    }
    return NULL;
}

void astDefFindNode(Ast* const ast, AstDef* const def, UInt const pos_line, UInt const pos_col, AstDef** found_def, AstExpr** found_expr) {
    if (toksHavePos(astNodeToks(&def->node_base, ast), pos_line, pos_col)) {
        *found_def = def;
        Bool found_in_sub_def = false;
        ·forEach(AstDef, sub_def, def->sub_defs, {
            astDefFindNode(ast, sub_def, pos_line, pos_col, found_def, found_expr);
            if ((*found_def) == sub_def) {
                found_in_sub_def = true;
                break;
            }
        });
        if (!found_in_sub_def)
            *found_expr = astExprFindNode(ast, &def->body, pos_line, pos_col);
    }
}

void astFindNode(CtxAsts const* const ctx, Str const src_file_path, UInt const pos_line, UInt const pos_col, Ast** found_ast,
                 AstDef** found_def, AstExpr** found_expr) {
    *found_ast = NULL;
    *found_def = NULL;
    *found_expr = NULL;
    Ast* ast = NULL;
    for (UInt i = 0; i < ctx->src_file_paths.len && ast == NULL; i += 1)
        if (strEql(ctx->src_file_paths.at[i], src_file_path))
            ast = &ctx->asts.at[i];
    if (ast != NULL)
        ·forEach(AstDef, top_def, ast->top_defs, {
            astDefFindNode(ast, top_def, pos_line, pos_col, found_def, found_expr);
            if ((*found_def) != NULL)
                break;
        });
    if ((*found_def) != NULL)
        *found_ast = ast;
}
