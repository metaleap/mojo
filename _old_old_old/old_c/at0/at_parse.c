#pragma once
#include "utils_std_mem.c"
#include "utils_toks.c"
#include "fs_io.c"
#include "at_ast.c"


#define asts_capacity 4


Ast astParse(Tokens const all_toks, Str const full_src, Str const src_file_path, SrcFileIssues* const gather_issues);
AstExpr astParseExpr(Tokens const, UInt const, Ast* const, Bool const, ºSrcFileIssue* const);
ºSrcFileIssue astParseDef(AstDef* const dst_def, Ast* const ast);

CtxAsts loadAndParseRootSourceFileAndImports(Str const root_src_file_path) {
    CtxAsts ctx = (CtxAsts) {
        .src_file_paths = ·sliceOf(Str, NULL, 0, asts_capacity),
        .asts = ·listOf(Ast, NULL, 0, asts_capacity),
        .issues = false,
    };
    ·push(ctx.src_file_paths, root_src_file_path);
    while (ctx.src_file_paths.len > ctx.asts.len) {
        SrcFileIssues issues = ·listOf(SrcFileIssue, NULL, 0, 16);

        Str const next_src_file_path = ctx.src_file_paths.at[ctx.asts.len];
        Str const full_src = readFile(next_src_file_path);
        if (full_src.at == NULL)
            ·append(issues, srcFileErr(issue_parse, str("error reading file"),
                                       (Token) {
                                           .kind = tok_kind_nope,
                                           .file_name = next_src_file_path,
                                       })
                                .it);

        Tokens const toks = tokenize(NULL, full_src, false, next_src_file_path, &issues);

        Ast ast = astParse(toks, full_src, next_src_file_path, &issues);
        ·append(ctx.asts, ast);
        ctx.issues |= (ast.issues.len > 0);

        for (UInt i = 0; i < ast.anns.incl_file_paths.len; i += 1) {
            Str const incl_file_path = relPathFromRelPath(next_src_file_path, ast.anns.incl_file_paths.at[i]);
            if (!strIn(incl_file_path, ctx.src_file_paths)) {
                if (ctx.src_file_paths.len == ctx.asts.cap)
                    ·fail(str("TODO: increase asts_capacity"));
                ·push(ctx.src_file_paths, incl_file_path);
            }
        }
    }
    return ctx;
}

Ast astParse(Tokens const all_toks, Str const full_src, Str const src_file_path, SrcFileIssues* const issues) {
    Tokenss const chunks = toksIndentBasedChunks(NULL, all_toks);

    Ast ret_ast = (Ast) {.src = full_src,
                         .issues = *issues,
                         .toks = all_toks,
                         .top_defs = ·sliceOf(AstDef, NULL, 0, chunks.len),
                         .anns = {.total_nr_of_def_toks = 0,
                                  .total_nr_of_tag_toks = 0,
                                  .incl_file_paths = ·sliceOf(Str, NULL, 0, asts_capacity),
                                  .src_file_path = src_file_path,
                                  .path_based_ident_prefix = ident(NULL, src_file_path)}};

    // guesstimate `total_nr_of_def_toks` and `total_nr_of_tag_toks`
    ·forEach(Token, tok, all_toks, {
        if (tok->kind == tok_kind_ident) {
            Str const tok_src = tokSrc(tok, full_src);
            if (strEql(tok_src, strL(":=", 2)) || strEql(tok_src, strL("->", 2)) || strEql(tok_src, strL("@->", 3)))
                ret_ast.anns.total_nr_of_def_toks += 1;
            else if (strEql(tok_src, strL("#", 1)))
                ret_ast.anns.total_nr_of_tag_toks += 1;
        }
    });

    UInt toks_idx = 0;
    ·forEach(Tokens, chunk_toks, chunks, {
        AstDef* const dst_def = &ret_ast.top_defs.at[ret_ast.top_defs.len];
        *dst_def = astDef(NULL, toks_idx, chunk_toks->len);
        ºSrcFileIssue issue = astParseDef(dst_def, &ret_ast);
        if (issue.got)
            ·push(ret_ast.issues, issue.it);
        else
            ret_ast.top_defs.len += 1;
        toks_idx += chunk_toks->len;
    });

    ·forEach(AstDef, top_def, ret_ast.top_defs, { astDefRewriteGlyphsIntoInstrs(top_def, &ret_ast); });

    ret_ast.anns.all_top_def_names = ·sliceOf(Str, NULL, ret_ast.top_defs.len, 0);
    ·forEach(AstDef, top_def, ret_ast.top_defs, { ret_ast.anns.all_top_def_names.at[iˇtop_def] = top_def->name; });

    *issues = ret_ast.issues;
    return ret_ast;
}

ºSrcFileIssue astParseDef(AstDef* const dst_def, Ast* const ast) {
    ºSrcFileIssue issue = ·none(SrcFileIssue);

    Tokens const toks = astNodeToks(&dst_def->node_base, ast);
    ºUInt const idx_tok_def = toksIndexOfIdent(toks, str(":="), ast->src);
    if ((!idx_tok_def.got) || idx_tok_def.it == 0 || idx_tok_def.it == toks.len - 1)
        return srcFileErr(issue_parse, str("expected '<head_expr> := <body_expr>'"), ast->toks.at[dst_def->node_base.toks_idx]);

    Tokenss const def_body_chunks = toksIndentBasedChunks(NULL, ·slice(Token, toks, idx_tok_def.it + 1, toks.len));
    dst_def->sub_defs = ·sliceOf(AstDef, NULL, 0, def_body_chunks.len - 1);
    UInt all_toks_idx = dst_def->node_base.toks_idx + idx_tok_def.it + 1;
    dst_def->body = astParseExpr(def_body_chunks.at[0], all_toks_idx, ast, false, &issue);
    if (issue.got)
        return issue;

    AstExpr head = astParseExpr(·slice(Token, toks, 0, idx_tok_def.it), dst_def->node_base.toks_idx, ast, false, &issue);
    if (issue.got)
        return issue;
    dst_def->anns.head_node_base = head.node_base;
    switch (head.kind) {
        case ast_expr_ident: {
            dst_def->name = head.of_ident;
        } break;
        case ast_expr_form: {
            if (head.of_exprs.at[0].kind != ast_expr_ident)
                return srcFileErr(issue_parse, str("unsupported def header form"), ast->toks.at[head.node_base.toks_idx]);
            dst_def->name = head.of_exprs.at[0].of_ident;

            AstExpr fn = astExpr(dst_def->node_base.toks_idx, dst_def->node_base.toks_len, ast_expr_form, 3);
            fn.of_exprs.at[0] = astExprInstrOrTag(head.node_base, strL("->", 2), false);
            fn.of_exprs.at[2] = dst_def->body;
            fn.of_exprs.at[1] =
                astExpr(head.of_exprs.at[1].node_base.toks_idx, head.node_base.toks_len - 1, ast_expr_lit_bracket, head.of_exprs.len - 1);
            dst_def->anns.param_names = ·sliceOf(Str, NULL, fn.of_exprs.at[1].of_exprs.len, 0);
            for (UInt i = 1; i < head.of_exprs.len; i += 1)
                if (head.of_exprs.at[i].kind != ast_expr_ident)
                    return srcFileErr(issue_parse, str("unsupported def header form"), ast->toks.at[head.node_base.toks_idx]);
                else {
                    dst_def->anns.param_names.at[i - 1] = head.of_exprs.at[i].of_ident;
                    fn.of_exprs.at[1].of_exprs.at[i - 1] =
                        astExprInstrOrTag(head.of_exprs.at[i].node_base, head.of_exprs.at[i].of_ident, true);
                }
            dst_def->body = fn;
        } break;
        default: return srcFileErr(issue_parse, str("unsupported def header form"), ast->toks.at[head.node_base.toks_idx]);
    }
    dst_def->anns.qname =
        (dst_def->anns.parent_def != NULL) ? str3(NULL, dst_def->anns.parent_def->anns.qname, strL("-", 1), dst_def->name) : dst_def->name;

    ·forEach(Tokens, chunk_toks, def_body_chunks, {
        if (iˇchunk_toks != 0) {
            AstDef* const sub_def = &dst_def->sub_defs.at[dst_def->sub_defs.len];
            dst_def->sub_defs.len += 1;
            *sub_def = astDef(dst_def, all_toks_idx, chunk_toks->len);
            issue = astParseDef(sub_def, ast);
            if (issue.got)
                return issue;
        }
        all_toks_idx += chunk_toks->len;
    });
    return issue;
}

AstExpr astParseExprLitInt(UInt const all_toks_idx, Ast const* const ast, Token const* const tok, ºSrcFileIssue* const issue) {
    AstExpr ret_expr = astExpr(all_toks_idx, 1, ast_expr_lit_int, 0);
    ºU64 const maybe = uInt64Parse(tokSrc(tok, ast->src));
    if (!maybe.got)
        *issue = srcFileErr(issue_parse, str("malformed or not-yet-supported integer literal"), *tok);
    else
        ret_expr.of_lit_int = maybe.it;
    return ret_expr;
}

AstExpr astParseExprLitStr(UInt const all_toks_idx, Ast const* const ast, Token const* const tok, U8 const quote_char,
                           ºSrcFileIssue* const issue) {
    AstExpr ret_expr = astExpr(all_toks_idx + 1, 1, ast_expr_lit_str, 0);
    Str const lit_src = tokSrc(tok, ast->src);
    ret_expr.of_lit_str = strParse(NULL, lit_src);
    if (ret_expr.of_lit_str.at == NULL)
        *issue = srcFileErr(issue_parse, str("malformed or not-yet-supported string literal"), *tok);
    return ret_expr;
}

AstExprs astParseExprsDelimited(Tokens const toks, UInt const all_toks_idx, TokenKind const tok_kind_sep, Ast* const ast,
                                ºSrcFileIssue* const issue) {
    if (toks.len == 0)
        return ·len0(AstExpr);
    Tokenss const per_elem_toks = toksSplit(NULL, toks, tok_kind_sep);
    AstExprs ret_exprs = ·sliceOf(AstExpr, NULL, 0, per_elem_toks.len);
    UInt toks_idx = all_toks_idx;
    ·forEach(Tokens, this_elem_toks, per_elem_toks, {
        if (this_elem_toks->len == 0)
            toks_idx += 1; // 1 for eaten delimiter
        else {
            ·push(ret_exprs, astParseExpr(*this_elem_toks, toks_idx, ast, false, issue));
            if (issue->got)
                break;
            toks_idx += (1 + this_elem_toks->len); // 1 for eaten delimiter
        }
    });
    return ret_exprs;
}

AstExpr astParseExpr(Tokens const expr_toks, UInt const all_toks_idx, Ast* const ast, Bool const known_whole_form_throng,
                     ºSrcFileIssue* const issue) {
    ·assert(expr_toks.len != 0);
    AstExprs ret_acc = ·sliceOf(AstExpr, NULL, 0, expr_toks.len);
    Bool const whole_form_throng =
        known_whole_form_throng || ((expr_toks.len > 1) && (tokThrong(expr_toks, 0, ast->src) == expr_toks.len - 1));

    for (UInt i = 0; i < expr_toks.len && !issue->got; i += 1) {
        UInt const idx_throng_end = whole_form_throng ? i : tokThrong(expr_toks, i, ast->src);
        if (idx_throng_end > i) {
            ·push(ret_acc, astParseExpr(·slice(Token, expr_toks, i, idx_throng_end + 1), all_toks_idx + i, ast, true, issue));
            if (issue->got)
                break;
            i = idx_throng_end; // loop header will increment
            continue;
        }
        switch (expr_toks.at[i].kind) {

            case tok_kind_comment: {
                ·fail(str("unreachable"));
            } break;

            case tok_kind_lit_num_prefixed: {
                ·push(ret_acc, astParseExprLitInt(all_toks_idx + 1, ast, &expr_toks.at[i], issue));
            } break;

            case tok_kind_lit_str_qdouble: {
                ·push(ret_acc, astParseExprLitStr(all_toks_idx + 1, ast, &expr_toks.at[i], '\"', issue));
            } break;

            case tok_kind_lit_str_qsingle: {
                AstExpr expr_lit = astParseExprLitStr(all_toks_idx + 1, ast, &expr_toks.at[i], '\"', issue);
                if ((!issue->got) && expr_lit.of_lit_str.len != 1)
                    *issue = srcFileErr(issue_parse, str("currently only supporting single-byte char literals"), expr_toks.at[i]);
                if (issue->got)
                    break;
                expr_lit.kind = ast_expr_lit_int;
                expr_lit.of_lit_int = expr_lit.of_lit_str.at[0];
                ·push(ret_acc, expr_lit);
            } break;

            case tok_kind_sep_bcurly_open:  // fall through to:
            case tok_kind_sep_bsquare_open: // fall through to:
            case tok_kind_sep_bparen_open: {
                TokenKind const tok_kind = expr_toks.at[i].kind;
                ºUInt idx_closing = toksIndexOfMatchingBracket(·slice(Token, expr_toks, i, expr_toks.len));
                ·assert(idx_closing.got); // the other-case will have been caught already by toksCheckBrackets
                idx_closing.it += i;
                if (tok_kind == tok_kind_sep_bparen_open) {
                    Tokens const toks_inside_parens = ·slice(Token, expr_toks, i + 1, idx_closing.it);
                    if (toks_inside_parens.len == 0) {
                        AstExpr const expr_ident = astExprIdent(all_toks_idx + i, 2, strL("()", 2));
                        ·push(ret_acc, expr_ident);
                    } else {
                        AstExpr expr_inside_parens = astParseExpr(toks_inside_parens, all_toks_idx + i + 1, ast, false, issue);
                        if (issue->got)
                            break;
                        expr_inside_parens.anns.parensed += (expr_inside_parens.anns.parensed < 255) ? 1 : 0;
                        // still want the parens toks captured in node base:
                        expr_inside_parens.node_base.toks_idx = all_toks_idx + i;
                        expr_inside_parens.node_base.toks_len += 2;
                        ·push(ret_acc, expr_inside_parens);
                    }
                } else { // no parens: either square brackets or curly braces
                    AstExprs const exprs_inside = astParseExprsDelimited(·slice(Token, expr_toks, i + 1, idx_closing.it),
                                                                         all_toks_idx + i + 1, tok_kind_sep_comma, ast, issue);
                    if (issue->got)
                        break;
                    Bool const is_braces = (tok_kind == tok_kind_sep_bcurly_open);
                    Bool const is_bracket = (tok_kind == tok_kind_sep_bsquare_open);
                    ·assert(is_braces || is_bracket); // always true right now obviously, but for future overpaced refactorers..
                    AstExpr expr_brac =
                        astExpr(all_toks_idx + i, 1 + (idx_closing.it - i), is_bracket ? ast_expr_lit_bracket : ast_expr_lit_braces, 0);
                    expr_brac.of_exprs = exprs_inside;
                    ·push(ret_acc, expr_brac);
                }
                i = idx_closing.it;
            } break;

            case tok_kind_ident: {
                AstExpr const expr_ident = astExprIdent(all_toks_idx + i, 1, tokSrc(&expr_toks.at[i], ast->src));
                ·push(ret_acc, expr_ident);
            } break;

            default: {
                *issue = srcFileErr(issue_parse, str("unrecognized token"), expr_toks.at[i]);
            } break;
        }
    }

    if (ret_acc.len == 1 && !issue->got)
        return ret_acc.at[0];

    AstExpr ret_expr = astExpr(all_toks_idx, expr_toks.len, ast_expr_form, 0);
    if (issue->got)
        return ret_expr;
    ·assert(ret_acc.len != 0);
    ret_expr.of_exprs = ret_acc;
    ret_expr.anns.toks_throng = whole_form_throng;

    Str const incl_file_path = astExprIsIncl(&ret_expr);
    if (incl_file_path.at != NULL) {
        if (!strIn(incl_file_path, ast->anns.incl_file_paths)) {
            if (ast->anns.incl_file_paths.len == asts_capacity)
                ·fail(str("TODO: increase asts_capacity"));
            ·push(ast->anns.incl_file_paths, incl_file_path);
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
