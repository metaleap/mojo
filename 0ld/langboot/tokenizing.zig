usingnamespace @import("./_usingnamespace.zig");

pub const Token = struct {
    idx: usize,
    len: usize,
    line_nr: usize,
    line_idx: usize,
    kind: Kind,

    pub fn posCol(me: *Token) usize {
        return me.idx - me.line_idx;
    }

    pub fn str(toks: []Token, src: Str) Str {
        return src[toks[0].idx .. toks[toks.len - 1].idx + toks[toks.len - 1].len];
    }

    pub fn indentBasedChunks(toks: []Token) [][]Token {
        var cmp_pos_col: usize = toks[0].posCol();
        for (toks) |*tok| {
            if (tok.posCol() < cmp_pos_col)
                cmp_pos_col = tok.posCol();
        }
        const num_chunks: usize = count_chunks: {
            var count: usize = 0;
            for (toks) |*tok, i| {
                if (i == 0 or tok.posCol() <= cmp_pos_col)
                    count += 1;
            }
            break :count_chunks count;
        };

        var ret = alloc([]Token, num_chunks);
        {
            var start_from: ?usize = null;
            var next_idx: usize = 0;
            for (toks) |*tok, i| {
                if (i == 0 or tok.posCol() <= cmp_pos_col) {
                    if (start_from) |idx| {
                        ret[next_idx] = toks[idx..i];
                        next_idx += 1;
                    }
                    start_from = i;
                }
            }
            if (start_from) |idx| {
                ret[next_idx] = toks[idx..toks.len];
                next_idx += 1;
            }
            assert(next_idx == num_chunks);
        }
        return ret;
    }

    pub fn indexOfMatchingBracket(toks: []Token) ?usize {
        const this_open = toks[0].kind;
        const this_close = switch (this_open) {
            .sep_bcurly_open => Kind.sep_bcurly_close,
            .sep_bparen_open => Kind.sep_bparen_close,
            .sep_bsquare_open => Kind.sep_bsquare_close,
            else => fail("indexOfMatchingBracket bad arg: {}", .{this_open}),
        };
        var level: usize = 0;
        for (toks) |*tok, i| {
            if (tok.kind == this_open)
                level += 1
            else if (tok.kind == this_close) {
                level -= 1;
                if (level == 0)
                    return i;
            }
        }
        return null;
    }

    pub fn indexOfFirst(toks: []Token, kind: Kind) ?usize {
        for (toks) |*tok, i|
            if (tok.kind == kind)
                return i;
        return null;
    }

    pub fn indexOfLast(toks: []Token, kind: Kind) ?usize {
        var i: usize = toks.len;
        while (i > 0) {
            i -= 1;
            if (toks[i].kind == kind)
                return i;
        }
        return null;
    }

    pub fn countUnnested(full_src: Str, toks: []Token, kind: Kind) usize {
        var num: usize = 0;
        var level: usize = 0;
        for (toks) |*tok, i| {
            if (tok.kind == .sep_bcurly_open or tok.kind == .sep_bparen_open or tok.kind == .sep_bsquare_open)
                level += 1
            else if (tok.kind == .sep_bcurly_close or tok.kind == .sep_bparen_close or tok.kind == .sep_bsquare_close)
                level = if (level > 0) level - 1 else fail("brackets mismatch near:\n{}\n", .{str(toks, full_src)})
            else if (level == 0 and tok.kind == kind)
                num += 1;
        }
        return num;
    }

    pub fn split(full_src: Str, toks: []Token, break_on_kind: Kind) [][]Token {
        const max_num_elems: usize = 1 + countUnnested(full_src, toks, break_on_kind);
        var ret = alloc([]Token, max_num_elems);
        var ret_idx: usize = 0;

        var level: usize = 0;
        var start_from: usize = 0;
        for (toks) |*tok, i| {
            if (tok.kind == .sep_bcurly_open or tok.kind == .sep_bparen_open or tok.kind == .sep_bsquare_open)
                level += 1
            else if (tok.kind == .sep_bcurly_close or tok.kind == .sep_bparen_close or tok.kind == .sep_bsquare_close)
                level = if (level > 0) level - 1 else fail("brackets mismatch near:\n{}\n", .{str(toks, full_src)})
            else if (level == 0 and tok.kind == break_on_kind) {
                const sub_toks = toks[start_from..i];
                if (sub_toks.len != 0) {
                    ret[ret_idx] = sub_toks;
                    ret_idx += 1;
                }
                start_from = i + 1;
            }
        }
        const sub_toks = toks[start_from..];
        if (sub_toks.len != 0) {
            ret[ret_idx] = sub_toks;
            ret_idx += 1;
        }

        return ret[0..ret_idx];
    }

    pub const Kind = enum {
        none,

        comment,
        ident,

        lit_int,
        lit_str,

        sep_bparen_open,
        sep_bparen_close,
        sep_bcurly_open,
        sep_bcurly_close,
        sep_bsquare_open,
        sep_bsquare_close,
        sep_comma,
        sep_colon,
        sep_colon_eq,
        sep_semicolon,
    };
};

pub fn tokenize(src: Str, include_comments: bool) []Token {
    var cur_line_nr: usize = 0;
    var cur_line_idx: usize = 0;
    var tok_start: ?usize = null;
    var tok_last: ?usize = null;
    var state: Token.Kind = .none;
    var toks_count: usize = 0;
    var toks = alloc(Token, src.len);

    var i: usize = 0;
    while (i < src.len) : (i += 1) {
        const c = src[i];
        if (c == '\n') {
            if (state == .lit_str)
                fail("line-break in literal near:\n{}", .{src[tok_start.?..i]});
            if (tok_start != null and tok_last == null)
                tok_last = i - 1;
        } else switch (state) {
            .lit_int, .ident => if (c == ' ' or c == '\t' or c == '\"' or c == '\'' or c == '[' or c == ']' or c == '{' or c == '}' or c == '(' or c == ')' or c == ';' or c == ',' or c == ':') {
                i -= 1;
                tok_last = i;
            },
            .lit_str => if (c == '\"') {
                tok_last = i;
            },
            .comment => {},
            .none => switch (c) {
                '\"' => {
                    tok_start = i;
                    state = .lit_str;
                },
                '[' => {
                    tok_last = i;
                    tok_start = i;
                    state = .sep_bsquare_open;
                },
                ']' => {
                    tok_last = i;
                    tok_start = i;
                    state = .sep_bsquare_close;
                },
                '(' => {
                    tok_last = i;
                    tok_start = i;
                    state = .sep_bparen_open;
                },
                ')' => {
                    tok_last = i;
                    tok_start = i;
                    state = .sep_bparen_close;
                },
                '{' => {
                    tok_last = i;
                    tok_start = i;
                    state = .sep_bcurly_open;
                },
                '}' => {
                    tok_last = i;
                    tok_start = i;
                    state = .sep_bcurly_close;
                },
                ',' => {
                    tok_last = i;
                    tok_start = i;
                    state = .sep_comma;
                },
                ';' => {
                    tok_last = i;
                    tok_start = i;
                    state = .sep_semicolon;
                },
                ':' => if (i < src.len - 1 and src[i + 1] == '=') {
                    tok_last = i + 1;
                    tok_start = i;
                    state = .sep_colon_eq;
                    i += 1;
                } else {
                    tok_last = i;
                    tok_start = i;
                    state = .sep_colon;
                },
                else => if (c == '/' and i < src.len - 1 and src[i + 1] == '/') {
                    tok_start = i;
                    state = .comment;
                } else if (c >= '0' and c <= '9') {
                    tok_start = i;
                    state = .lit_int;
                } else if (c == ' ' or c == '\t') {
                    if (tok_start != null and tok_last == null)
                        tok_last = i - 1;
                } else {
                    tok_start = i;
                    state = .ident;
                },
            },
            else => unreachable,
        }
        if (tok_last != null) {
            if (state == .none or tok_start == null)
                unreachable;
            if (state != .comment or include_comments) {
                const tok_len = (tok_last.? - tok_start.?) + 1;
                toks[toks_count] = .{
                    .kind = state,
                    .line_nr = cur_line_nr,
                    .line_idx = cur_line_idx,
                    .idx = tok_start.?,
                    .len = tok_len,
                };
                toks_count += 1;
            }
            state = .none;
            tok_start = null;
            tok_last = null;
        }
        if (c == '\n') {
            cur_line_nr += 1;
            cur_line_idx = i + 1;
        }
    }
    if (tok_start != null)
        if (state == .none) unreachable else if (state != .comment or include_comments) {
            const tok_len = i - tok_start.?;
            toks[toks_count] = .{
                .kind = state,
                .idx = tok_start.?,
                .len = tok_len,
                .line_nr = cur_line_nr,
                .line_idx = cur_line_idx,
            };
            toks_count += 1;
        };
    return toks[0..toks_count];
}
