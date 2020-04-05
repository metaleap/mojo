usingnamespace @import("./_usingnamespace.zig");

pub fn parse(all_toks: []Token, full_src: Str) Ast {
    const chunks = Token.indentBasedChunks(all_toks);
    var ret_ast = Ast{ .src = full_src, .toks = all_toks, .defs = alloc(AstDef, chunks.len * 2) };
    var toks_idx: usize = 0;
    for (chunks) |chunk_toks, i| {
        ret_ast.defs[i].base.toks_idx = toks_idx;
        ret_ast.defs[i].base.toks_len = chunk_toks.len;
        toks_idx += chunk_toks.len;
    }
    for (ret_ast.defs[0..chunks.len]) |*top_def, i| {
        top_def.is_top_def = true;
        parseDef(full_src, all_toks, top_def);
    }

    var num_str_lits: usize = 0;
    for (ret_ast.defs[0..chunks.len]) |*top_def| {
        var str_lit_and_new_name_pairs = alloc([2]Str, 8);
        const num = top_def.rewriteStrLits(str_lit_and_new_name_pairs, 0);
        for (str_lit_and_new_name_pairs[0..num]) |*str_lit_and_new_name| {
            ret_ast.defs[chunks.len + num_str_lits] = .{
                .base = AstNode.zero,
                .head = AstExpr{ .base = AstNode.zero, .kind = .{ .ident = str_lit_and_new_name[1] } },
                .body = AstExpr{ .base = AstNode.zero, .kind = .{ .lit_str = str_lit_and_new_name[0] } },
                .defs = &[_]AstDef{},
                .is_top_def = true,
            };
            num_str_lits += 1;
        }
    }

    ret_ast.defs = ret_ast.defs[0 .. chunks.len + num_str_lits];
    return ret_ast;
}

fn parseDef(full_src: Str, all_toks: []Token, dst_def: *AstDef) void {
    const toks = dst_def.base.toks(all_toks);
    const tok_idx_colon_eq = Token.indexOfFirst(toks, .sep_colon_eq) orelse
        fail("expected ':=', near:\n{}", .{dst_def.base.str(all_toks, full_src)});
    if (tok_idx_colon_eq == toks.len - 1)
        fail("expected def body, near:\n{}", .{dst_def.base.str(all_toks, full_src)});

    const chunks = Token.indentBasedChunks(toks[tok_idx_colon_eq + 1 ..]);
    dst_def.head = parseExpr(full_src, all_toks, toks[0..tok_idx_colon_eq], dst_def.base.toks_idx);
    if (dst_def.head.kind != .ident)
        fail("currently only supporting single-ident def heads, violated near:\n{}", .{dst_def.base.str(all_toks, full_src)});
    dst_def.defs = alloc(AstDef, chunks.len - 1);
    var toks_idx: usize = dst_def.base.toks_idx + tok_idx_colon_eq + 1;
    for (chunks) |chunk_toks, i| {
        if (i == 0)
            dst_def.body = parseExpr(full_src, all_toks, chunk_toks, toks_idx)
        else {
            const sub_def = &dst_def.defs[i - 1];
            sub_def.base.toks_idx = toks_idx;
            sub_def.base.toks_len = chunk_toks.len;
            sub_def.is_top_def = false;
            parseDef(full_src, all_toks, sub_def);
        }
        toks_idx += chunk_toks.len;
    }
}

fn parseExpr(full_src: Str, all_toks: []Token, toks: []Token, all_toks_idx: usize) AstExpr {
    var acc_ret = alloc(AstExpr, toks.len);
    var acc_len: usize = 0;
    var cur_pos: usize = 0;
    while (cur_pos < toks.len) : (cur_pos += 1) {
        const tok = &toks[cur_pos];
        switch (tok.kind) {
            .ident => {
                const tok_str = Token.str(toks[cur_pos .. cur_pos + 1], full_src);
                acc_ret[acc_len] = AstExpr{
                    .base = AstNode.from(all_toks_idx + cur_pos, 1),
                    .kind = .{ .ident = tok_str },
                };
            },
            .lit_int => {
                const full_tok = Token.str(toks[cur_pos .. cur_pos + 1], full_src);
                acc_ret[acc_len] = AstExpr{
                    .base = AstNode.from(all_toks_idx + cur_pos, 1),
                    .kind = .{ .lit_int = parseExprLitInt(full_tok) },
                };
            },
            .lit_str => {
                const full_tok = Token.str(toks[cur_pos .. cur_pos + 1], full_src);
                acc_ret[acc_len] = AstExpr{
                    .base = AstNode.from(all_toks_idx + cur_pos, 1),
                    .kind = .{ .lit_str = parseExprLitStr(full_tok) },
                };
            },
            .sep_bparen_open => {
                var idx_close = cur_pos + (Token.indexOfMatchingBracket(toks[cur_pos..]) orelse
                    fail("non-matching closing bracket near:\n{}", .{Token.str(toks, full_src)}));
                acc_ret[acc_len] = parseExpr(full_src, all_toks, toks[cur_pos + 1 .. idx_close], all_toks_idx + cur_pos + 1);
                cur_pos = idx_close; // loop header will increment
            },
            .sep_bsquare_open => {
                var idx_close = cur_pos + (Token.indexOfMatchingBracket(toks[cur_pos..]) orelse
                    fail("non-matching closing bracket near:\n{}", .{Token.str(toks, full_src)}));
                acc_ret[acc_len] = AstExpr{
                    .base = AstNode.from(all_toks_idx + cur_pos, 1 + (idx_close - cur_pos)),
                    .kind = .{ .lit_arr = parseExprLitArr(full_src, all_toks, toks[cur_pos + 1 .. idx_close], all_toks_idx + cur_pos + 1) },
                };
                cur_pos = idx_close; // loop header will increment
            },
            .sep_bcurly_open => {
                var idx_close = cur_pos + (Token.indexOfMatchingBracket(toks[cur_pos..]) orelse
                    fail("non-matching closing bracket near:\n{}", .{Token.str(toks, full_src)}));
                acc_ret[acc_len] = AstExpr{
                    .base = AstNode.from(all_toks_idx + cur_pos, 1 + (idx_close - cur_pos)),
                    .kind = .{ .lit_obj = parseExprLitObj(full_src, all_toks, toks[cur_pos + 1 .. idx_close], all_toks_idx + cur_pos + 1) },
                };
                cur_pos = idx_close; // loop header will increment
            },
            else => fail("Unexpected token kind '{}' near:\n{}", .{ tok.kind, Token.str(toks, full_src) }),
        }
        acc_len += 1;
    }

    assert(acc_len != 0);
    var ret = if (acc_len == 1)
        acc_ret[0]
    else
        AstExpr{ .kind = .{ .form = acc_ret[0..acc_len] }, .base = AstNode.from(all_toks_idx, toks.len) };
    if (ret.toBuiltin(all_toks, full_src)) |builtin|
        ret.kind = .{ .builtin = builtin };
    return ret;
}

pub fn parseExprLitInt(full_tok_src: Str) u128 {
    var mult: u128 = 1;
    var ret_int: u128 = 0;
    var i = full_tok_src.len;
    assert(i > 0);
    while (i > 0) {
        i -= 1;
        if (full_tok_src[i] < '0' or full_tok_src[i] > '9')
            fail("malformed integer literal: {}", .{full_tok_src});
        ret_int += mult * @as(u128, full_tok_src[i] - 48);
        mult *= 10;
    }
    return ret_int;
}

fn parseExprLitStr(full_tok_src: Str) Str {
    assert(full_tok_src.len >= 2 and full_tok_src[0] == '"' and full_tok_src[full_tok_src.len - 1] == '"');
    var ret_str = alloc(u8, full_tok_src.len - 2);
    var ret_len: usize = 0;
    var i: usize = 1;
    while (i < (full_tok_src.len - 1)) : (i += 1) {
        if (full_tok_src[i] != '\\')
            ret_str[ret_len] = full_tok_src[i]
        else {
            const int10str = full_tok_src[i + 1 .. i + 4];
            i += 3;
            const uint = parseExprLitInt(int10str);
            assert(uint < 256);
            ret_str[ret_len] = @intCast(u8, uint);
        }
        ret_len += 1;
    }
    return ret_str[0..ret_len];
}

fn parseExprLitArr(full_src: Str, all_toks: []Token, toks: []Token, all_toks_idx: usize) AstExpr.LitArr {
    var per_elem_toks = Token.split(full_src, toks, .sep_comma);
    var ret_arr = AstExpr.LitArr{ .elems = alloc(AstExpr, per_elem_toks.len) };
    var toks_idx = all_toks_idx;
    for (per_elem_toks) |this_elem_toks, i| {
        ret_arr.elems[i] = parseExpr(full_src, all_toks, this_elem_toks, toks_idx);
        toks_idx += this_elem_toks.len;
    }
    return ret_arr;
}

fn parseExprLitObj(full_src: Str, all_toks: []Token, toks: []Token, all_toks_idx: usize) AstExpr.LitObj {
    var per_elem_toks = Token.split(full_src, toks, .sep_comma);
    var ret_obj = AstExpr.LitObj{ .elems = alloc(AstExpr.LitObj.Pair, per_elem_toks.len) };
    var toks_idx = all_toks_idx;
    for (per_elem_toks) |this_elem_toks, i| {
        const tok_idx_colon = Token.indexOfFirst(this_elem_toks, .sep_colon) orelse
            fail("expected ':', near:\n{}", .{Token.str(this_elem_toks, full_src)});
        if (tok_idx_colon == 0 or tok_idx_colon == this_elem_toks.len - 1)
            fail("expr expected both before and after ':', near:\n{}", .{Token.str(this_elem_toks, full_src)});
        ret_obj.elems[i].lhs = parseExpr(full_src, all_toks, this_elem_toks[0..tok_idx_colon], toks_idx);
        ret_obj.elems[i].rhs = parseExpr(full_src, all_toks, this_elem_toks[tok_idx_colon + 1 ..], toks_idx + tok_idx_colon + 1);
        toks_idx += this_elem_toks.len;
    }
    return ret_obj;
}
