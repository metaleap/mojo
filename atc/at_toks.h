#pragma once
#include "std.h"



typedef enum TokenKind {
    tok_kind_none = 0,
    tok_kind_comment = 1,
    tok_kind_ident = 2,
    tok_kind_lit_int = 3,
    tok_kind_lit_str_double = 4,
    tok_kind_lit_str_single = 5,
    tok_kind_sep_bparen_open = 6,
    tok_kind_sep_bparen_close = 7,
    tok_kind_sep_bcurly_open = 8,
    tok_kind_sep_bcurly_close = 9,
    tok_kind_sep_bsquare_open = 10,
    tok_kind_sep_bsquare_close = 11,
    tok_kind_sep_comma = 12,
} TokenKind;

typedef struct Token {
    Int byte_id;
    Int num_bytes;
    Int line_nr;
    Int line_start_byte_idx;
    TokenKind kind;
} Token;
typedef SliceOf(Token) Tokens;



Tokens tokenize(Str full_src, Bool keep_comment_toks) {
    TokenKind state = tok_kind_none;
    Uint i = 0;
    Uint cur_line_nr = 0;
    Uint cur_line_idx = 0;
    Uint toks_count = 0;
    Int tok_start = -1;
    Int tok_last = -1;

    Tokens toks = alloc(Token, full_src.len);
    if (full_src.len > 2 && full_src._[0] == '#' && full_src._[1] == '!') {
        state = tok_kind_comment;
        tok_start = 0;
        i = 2;
    }
    for (; i < full_src.len; i += 1) {
        U8 c = full_src._[i];

        if (c == '\n') {
            if (state == tok_kind_lit_str_double || state == tok_kind_lit_str_single) {
                panic("TODO %s", 123);
                // fail(uintToStr(uint64(1+cur_line_nr), 10, 1, Str("line-break in literal in line
                // ")), ":\n", full_src[tok_start:i]);
            }
            if (tok_start != -1 && tok_last == -1)
                tok_last = i - 1;
        } else {
        }
    }
    return toks;
}
