#pragma once
#include "utils_json.c"
#include "utils_std_mem.c"
#include "utils_toks.c"
#include "fs_io.c"
#include "at_ast.c"
#include "at_parse.c"
#include "ir_hl.c"
#include "ir_ml.c"


#define lsp_loop_stdio_buf_size (8 * 1024 * 1024)




JsonValue lspNewNotify(Str const method, JsonValue* const params) {
    JsonValue ret_msg = jsonNewObj(NULL, 3);
    ·push(ret_msg.of.obj.keys, strL("jsonrpc", 7));
    ·push(ret_msg.of.obj.vals, jsonNewStr(strL("2.0", 3)));
    ·push(ret_msg.of.obj.keys, strL("method", 6));
    ·push(ret_msg.of.obj.vals, jsonNewStr(method));
    ·push(ret_msg.of.obj.keys, strL("params", 6));
    ·push(ret_msg.of.obj.vals, ((JsonValue) {.kind = json_object, .of = {.obj = params->of.obj}}));
    return ret_msg;
}

JsonValue lspNewResponse(I64 const req_id, JsonValue* const result, JsonValue* const error) {
    JsonValue ret_msg = jsonNewObj(NULL, 3);
    ·push(ret_msg.of.obj.keys, strL("jsonrpc", 7));
    ·push(ret_msg.of.obj.vals, jsonNewStr(strL("2.0", 3)));
    ·push(ret_msg.of.obj.keys, strL("id", 2));
    ·push(ret_msg.of.obj.vals, jsonNewNum(req_id));

    if (error != NULL) {
        ·push(ret_msg.of.obj.keys, strL("error", 5));
        ·push(ret_msg.of.obj.vals, *error);
    } else if (result != NULL) {
        ·push(ret_msg.of.obj.keys, strL("result", 6));
        ·push(ret_msg.of.obj.vals, *result);
    } else
        ·fail(str("BUG: lspNewResponse with no error and no result"));

    return ret_msg;
}

JsonValue lspNewRequest(Str const method, JsonValue* const params) {
    static I64 next_req_id = 0;
    next_req_id += 1;

    JsonValue ret_msg = jsonNewObj(NULL, 4);
    ·push(ret_msg.of.obj.keys, strL("jsonrpc", 7));
    ·push(ret_msg.of.obj.vals, jsonNewStr(strL("2.0", 3)));
    ·push(ret_msg.of.obj.keys, strL("id", 2));
    ·push(ret_msg.of.obj.vals, jsonNewNum(next_req_id));
    ·push(ret_msg.of.obj.keys, strL("method", 6));
    ·push(ret_msg.of.obj.vals, jsonNewStr(method));
    ·push(ret_msg.of.obj.keys, strL("params", 6));
    ·push(ret_msg.of.obj.vals, ((JsonValue) {.kind = json_object, .of = {.obj = params->of.obj}}));
    return ret_msg;
}

JsonValue lspNewMarkupContent(Str const content, Bool const markdown) {
    JsonValue ret_json = jsonNewObj(NULL, 2);
    ·push(ret_json.of.obj.keys, strL("kind", 4));
    ·push(ret_json.of.obj.vals, jsonNewStr(markdown ? str("markdown") : str("plaintext")));
    ·push(ret_json.of.obj.keys, strL("value", 5));
    ·push(ret_json.of.obj.vals, jsonNewStr(content));
    return ret_json;
}

JsonValue lspNewPosition(UInt const pos_line, UInt const pos_col) {
    JsonValue ret_json = jsonNewObj(NULL, 2);
    ·push(ret_json.of.obj.keys, strL("line", 4));
    ·push(ret_json.of.obj.vals, jsonNewNum(pos_line));
    ·push(ret_json.of.obj.keys, strL("character", 9));
    ·push(ret_json.of.obj.vals, jsonNewNum(pos_col));
    return ret_json;
}

JsonValue lspNewRange(UInt const pos_line, UInt const pos_col) {
    JsonValue ret_json = jsonNewObj(NULL, 2);
    JsonValue pos = lspNewPosition(pos_line, pos_col);
    ·push(ret_json.of.obj.keys, strL("start", 5));
    ·push(ret_json.of.obj.vals, pos);
    ·push(ret_json.of.obj.keys, strL("end", 3));
    ·push(ret_json.of.obj.vals, pos);
    return ret_json;
}

JsonValue lspNewUri(Str const src_file_path) {
    return jsonNewStr(str2(NULL, strL("file://", 7), src_file_path));
}

JsonValue lspNewLocation(Str const src_file_path, UInt const pos_line, UInt const pos_col) {
    JsonValue ret_json = jsonNewObj(NULL, 2);
    ·push(ret_json.of.obj.keys, strL("uri", 3));
    ·push(ret_json.of.obj.vals, lspNewUri(src_file_path));
    ·push(ret_json.of.obj.keys, strL("range", 5));
    ·push(ret_json.of.obj.vals, lspNewRange(pos_line, pos_col));
    return ret_json;
}



typedef struct CtxSrcFileSess {
    CtxAsts* parsed;
    IrHlProg* prog_hl;
    IrMlProg* prog_ml;
} CtxSrcFileSess;

CtxSrcFileSess loadFromSrcFile(Str const src_file_path) {
    CtxSrcFileSess ret_ctx = (CtxSrcFileSess) {.prog_hl = NULL, .prog_ml = NULL, .parsed = ·new(CtxAsts, NULL)};
    *ret_ctx.parsed = loadAndParseRootSourceFileAndImports(src_file_path);
    if (!ret_ctx.parsed->issues) {
        IrHlProg* tmp_hl = ·new(IrHlProg, NULL);
        *tmp_hl = irhlProgFrom(ret_ctx.parsed->asts);
        if (!astsIssues(ret_ctx.parsed, true)) {
            ret_ctx.prog_hl = tmp_hl;
            IrMlProg* tmp_ml = ·new(IrMlProg, NULL);
            *tmp_ml = irmlProgFrom(ret_ctx.prog_hl, ret_ctx.parsed->asts.at[0].anns.path_based_ident_prefix);
            if (!astsIssues(ret_ctx.parsed, true))
                ret_ctx.prog_ml = tmp_ml;
        }
    }
    return ret_ctx;
}

typedef struct CtxLsp {
    CtxSrcFileSess sess;

    ºI64 req_id;
    Str method;
    JsonValue* params;
    JsonValue* result;
    JsonValue* error;

    struct {
        Str path;
        struct {
            UInt line;
            UInt col; // not unicode glyphs during stage0 so col = character
        } pos;
        struct {
            Ast* ast;
            AstDef* ast_def;
            AstExpr* ast_expr;
            AstNodeBase* ast_node;
            Token* tok;
            IrHlDef* hl_def;
            IrHlExpr* hl_expr;
        } cur;
    } src_file;
} CtxLsp;

void lookupFromCurSrcFilePos(CtxLsp* const ctx) {
    if (ctx->sess.parsed != NULL) {
        astFindNode(ctx->sess.parsed, ctx->src_file.path, ctx->src_file.pos.line, ctx->src_file.pos.col, &ctx->src_file.cur.ast,
                    &ctx->src_file.cur.ast_def, &ctx->src_file.cur.ast_expr);
        if (ctx->src_file.cur.ast != NULL && ctx->src_file.cur.ast_def != NULL && ctx->sess.prog_hl != NULL) {
            irhlFind(ctx->sess.prog_hl, ctx->src_file.cur.ast, ctx->src_file.cur.ast_def, ctx->src_file.cur.ast_expr,
                     &ctx->src_file.cur.hl_def, &ctx->src_file.cur.hl_expr);
            if (ctx->src_file.cur.hl_expr != NULL)
                irhlRefTarget(ctx->sess.prog_hl, ctx->src_file.cur.hl_def, ctx->src_file.cur.hl_expr, &ctx->src_file.cur.hl_def,
                              &ctx->src_file.cur.hl_expr, &ctx->src_file.cur.ast, &ctx->src_file.cur.ast_node);
            if (ctx->src_file.cur.ast_node == NULL)
                ctx->src_file.cur.ast_node =
                    (ctx->src_file.cur.ast_expr == NULL) ? &ctx->src_file.cur.ast_def->node_base : &ctx->src_file.cur.ast_expr->node_base;
            ctx->src_file.cur.tok = &ctx->src_file.cur.ast->toks.at[ctx->src_file.cur.ast_node->toks_idx];
        }
    }
}

JsonValue lspHandleRequest_Definitions(CtxLsp* const ctx) {
    JsonValue ret_json = jsonNewArr(NULL, 1);
    lookupFromCurSrcFilePos(ctx);
    if (ctx->src_file.cur.tok != NULL)
        ·push(ret_json.of.arr, lspNewLocation(ctx->src_file.path, ctx->src_file.cur.tok->line_nr, tokPosCol(ctx->src_file.cur.tok)));
    return ret_json;
}

JsonValue lspHandleRequest_Hover(CtxLsp* const ctx) {
    JsonValue ret_json = jsonNewObj(NULL, 1);
    ·push(ret_json.of.obj.keys, str("contents"));
    ·push(ret_json.of.obj.vals, lspNewMarkupContent(str("Hello **World**!"), true));
    return ret_json;
}

JsonValue lspHandleRequest_Initialize(CtxLsp* const ctx) {
    JsonValue server_capabilities = jsonNewObj(NULL, 3);
    JsonValue bool_true = jsonNewBool(true);

    ·push(server_capabilities.of.obj.keys, str("textDocumentSync"));
    ·push(server_capabilities.of.obj.vals, jsonNewObj(NULL, 2));
    ·push(server_capabilities.of.obj.vals.at[0].of.obj.keys, str("openClose"));
    ·push(server_capabilities.of.obj.vals.at[0].of.obj.vals, bool_true);
    ·push(server_capabilities.of.obj.vals.at[0].of.obj.keys, str("save"));
    ·push(server_capabilities.of.obj.vals.at[0].of.obj.vals, bool_true);

    ·push(server_capabilities.of.obj.keys, str("definitionProvider"));
    ·push(server_capabilities.of.obj.vals, bool_true);

    ·push(server_capabilities.of.obj.keys, str("hoverProvider"));
    ·push(server_capabilities.of.obj.vals, bool_true);

    JsonValue initialize_result = jsonNewObj(NULL, 1);
    ·push(initialize_result.of.obj.keys, str("capabilities"));
    ·push(initialize_result.of.obj.vals, server_capabilities);
    return initialize_result;
}

JsonValues lspHandleNotify_DidSave(CtxLsp* const ctx) {
    JsonValues ret_jsons = ·len0(JsonValue);
    return ret_jsons;
}

JsonValues lspHandleNotify_DidOpen(CtxLsp* const ctx) {
    JsonValues ret_jsons = ·len0(JsonValue);
    return ret_jsons;
}

JsonValue lspHandleRequest(CtxLsp* const ctx) {
    JsonValue ret_json = (JsonValue) {.kind = json_invalid};

    Str const method = ctx->method;
    if (strEql(method, str("initialize")))
        ret_json = lspHandleRequest_Initialize(ctx);
    else if (strEql(method, str("textDocument/hover")))
        ret_json = lspHandleRequest_Hover(ctx);
    else if (strEql(method, str("textDocument/definition")))
        ret_json = lspHandleRequest_Definitions(ctx);

    return ret_json;
}

JsonValues lspHandleResponse(CtxLsp* const ctx) {
    JsonValues ret_jsons = ·len0(JsonValue);
    return ret_jsons;
}

JsonValues lspHandleNotify(CtxLsp* const ctx) {
    JsonValues ret_jsons = ·len0(JsonValue);

    Str const method = ctx->method;
    if (strEql(method, str("textDocument/didOpen")))
        ret_jsons = lspHandleNotify_DidOpen(ctx);
    else if (strEql(method, str("textDocument/didSave")))
        ret_jsons = lspHandleNotify_DidSave(ctx);

    return ret_jsons;
}

JsonValues lspHandleJson(CtxLsp* ctx, Str const json_src) {
    JsonValues ret_jsons = ·len0(JsonValue);
    JsonValue ret_json = (JsonValue) {.kind = json_invalid};

    JsonValue const json_value = jsonParse(NULL, tokenize(NULL, json_src, false, str("<stdin>"), NULL), json_src);
    if (json_value.kind != json_object)
        fprintf(stderr, "\nbad json:\n\t>>>>>%s<<<<<\n", json_src.at);
    else {
        // fprintf(stderr, "\nincoming json:\n\t>>>>>%s<<<<<\n", json_src.at);

        ctx->req_id = jsonObjValNum(&json_value, str("id"));
        ctx->method = jsonObjValStr(&json_value, str("method"));
        ctx->params = jsonObjVal(&json_value, str("params"), json_invalid);
        ctx->result = jsonObjVal(&json_value, str("result"), json_invalid);
        ctx->error = jsonObjVal(&json_value, str("error"), json_invalid);

        Bool const msg_method_got = (ctx->method.len != 0);
        Bool const is_request = ctx->req_id.got && msg_method_got;
        Bool const is_notify = msg_method_got && !ctx->req_id.got;
        Bool const is_response = ctx->req_id.got && (!msg_method_got) && (ctx->result != NULL || ctx->error != NULL);

        if (ctx->params != NULL) {
            //{"textDocument":{"uri":"file:///home/foo/bar.baz"},"position":{"line":45,"character":27}}
            JsonValue* const td = jsonObjVal(ctx->params, strL("textDocument", 12), json_object);
            JsonValue* pos = jsonObjVal(ctx->params, strL("position", 8), json_object);
            if (pos != NULL) {
                ºI64 const pos_ln = jsonObjValNum(pos, strL("line", 4));
                ºI64 const pos_col = jsonObjValNum(pos, strL("character", 9));
                if (pos_ln.got && pos_col.got) {
                    ctx->src_file.pos.line = pos_ln.it;
                    ctx->src_file.pos.col = pos_col.it;
                }
            }
            if (td != NULL) {
                ctx->src_file.path = jsonObjValStr(td, strL("uri", 3));
                if (ctx->src_file.path.len != 0) {
                    if (strPref(ctx->src_file.path, strL("file://", 7)))
                        ctx->src_file.path = strSub(ctx->src_file.path, 7, ctx->src_file.path.len);
                    if (pos == NULL) {
                        pos = jsonObjVal(td, strL("position", 8), json_object);
                        if (pos != NULL) {
                            ºI64 const pos_ln = jsonObjValNum(pos, strL("line", 4));
                            ºI64 const pos_col = jsonObjValNum(pos, strL("character", 9));
                            if (pos_ln.got && pos_col.got) {
                                ctx->src_file.pos.line = pos_ln.it;
                                ctx->src_file.pos.col = pos_col.it;
                            }
                        }
                    }
                    ctx->sess = loadFromSrcFile(ctx->src_file.path);
                }
            }
        }

        if (is_response)
            ret_jsons = lspHandleResponse(ctx);
        else if (is_notify)
            ret_jsons = lspHandleNotify(ctx);
        else if (is_request) {
            ret_json = lspHandleRequest(ctx);
            if (ret_json.kind != json_invalid) {
                Bool const is_err = (ret_json.kind == json_object) && jsonObjValNum(&ret_json, strL("code", 4)).got
                                    && (jsonObjValStr(&ret_json, strL("message", 7)).at != NULL);
                ret_json = lspNewResponse(ctx->req_id.it, (is_err ? NULL : &ret_json), (is_err ? &ret_json : NULL));
            }
        }

        {
            JsonValues tmp = ·sliceOf(JsonValue, NULL, 0,
                                      ((ret_json.kind == json_invalid) ? 0 : 1) + ret_jsons.len
                                          + ((ctx->sess.parsed == NULL) ? 0 : ctx->sess.parsed->src_file_paths.len));
            ·forEach(JsonValue, the_json_value, ret_jsons, { ·push(tmp, *the_json_value); });
            ret_jsons = tmp;
        }
        if (ret_json.kind != json_invalid)
            ·push(ret_jsons, ret_json);

        if (ctx->sess.parsed != NULL) // prep diags refresh
            for (UInt i = 0; i < ctx->sess.parsed->src_file_paths.len; i += 1) {
                Ast* const ast = &ctx->sess.parsed->asts.at[i];

                JsonValue diags = jsonNewObj(NULL, 2);
                ·push(diags.of.obj.keys, strL("uri", 3));
                ·push(diags.of.obj.vals, lspNewUri(ast->anns.src_file_path));
                ·push(diags.of.obj.keys, strL("diagnostics", 11));
                ·push(diags.of.obj.vals, jsonNewArr(NULL, ast->issues.len));
                JsonValue* issues = ·last(diags.of.obj.vals);
                ·forEach(SrcFileIssue, issue, ast->issues, {
                    JsonValue diag = jsonNewObj(NULL, 3);
                    ·push(diag.of.obj.keys, strL("message", 7));
                    ·push(diag.of.obj.vals, jsonNewStr(issue->msg));
                    ·push(diag.of.obj.keys, strL("severity", 8));
                    ·push(diag.of.obj.vals, jsonNewNum(issue->error ? 1 : 2));
                    ·push(diag.of.obj.keys, strL("range", 5));
                    ·push(diag.of.obj.vals,
                          lspNewRange(issue->src_pos.line_nr, issue->src_pos.char_pos - issue->src_pos.char_pos_line_start));
                    ·push(issues->of.arr, diag);
                });

                ·push(ret_jsons, lspNewNotify(str("textDocument/publishDiagnostics"), &diags));
            }
    }
    return ret_jsons;
}

void lspMainLoop() {
    static U8 stdin_buf[lsp_loop_stdio_buf_size];
    static U8 stdout_buf[lsp_loop_stdio_buf_size];

    fprintf(stderr, "LSP listening..\n");
    while (true) {
        mem_bss.pos = 0;
        Str headers = readStdinUntilSuffix(strL("\r\n\r\n", 4));
        ºUInt idx_clen = strIndexOf(headers, str("Content-Length: "));
        if (!idx_clen.got)
            fprintf(stderr, "\nbad headers 1:\n\t%s\n", headers.at);
        else {
            idx_clen.it += 16;
            headers = strSub(headers, idx_clen.it, headers.len);
            idx_clen = strIndexOf(headers, strL("\r\n", 2));
            if (!idx_clen.got)
                fprintf(stderr, "\nbad headers 2:\n\t%s\n", headers.at);
            else {
                headers = strSub(headers, 0, idx_clen.it);
                ºU64 const clen = uInt64Parse(headers);
                if (!clen.got)
                    fprintf(stderr, "\nbad headers 3:\n\t%s\n", headers.at);
                else {
                    if (clen.it >= lsp_loop_stdio_buf_size)
                        ·fail(str("TODO: increase lsp_loop_stdio_buf_size due to input"));
                    stdin_buf[clen.it] = 0;
                    if (clen.it == 0)
                        fprintf(stderr, "\ncLen of 0:\n\t%s\n", headers.at);
                    else if (clen.it != fread(&stdin_buf[0], 1, clen.it, stdin))
                        fprintf(stderr, "\nread less than %lu:\n\t%s\n", clen.it, headers.at);
                    else {
                        CtxLsp ctx = (CtxLsp) {.sess = {.parsed = NULL, .prog_hl = NULL, .prog_ml = NULL},
                                               .req_id = ·none(I64),
                                               .method = ·len0(U8),
                                               .params = NULL,
                                               .result = NULL,
                                               .error = NULL,
                                               .src_file = {.path = ·len0(U8),
                                                            .pos = {.line = 0, .col = 0},
                                                            .cur = {.ast = NULL,
                                                                    .ast_def = NULL,
                                                                    .ast_expr = NULL,
                                                                    .ast_node = NULL,
                                                                    .hl_def = NULL,
                                                                    .hl_expr = NULL,
                                                                    .tok = NULL}}};
                        JsonValues const resp_msgs = lspHandleJson(&ctx, (Str) {.at = &stdin_buf[0], .len = clen.it});
                        ·forEach(JsonValue, resp_json, resp_msgs, {
                            Str const buf = jsonWrite((Str) {.at = &stdout_buf[0], .len = 0}, lsp_loop_stdio_buf_size, resp_json);
                            // fprintf(stderr, "\njson sent %zu:\n\t>>>>>%s<<<<<\n", buf.len, buf.at);
                            fprintf(stdout, "Content-Length: %zu\r\n\r\n%s", buf.len, &stdout_buf[0]);
                            fflush(stdout);
                        });
                    }
                }
            }
        }
    }
}
