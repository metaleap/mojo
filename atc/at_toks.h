#pragma once
#include "std.h"


const String tok_op_chars = "!#$%&*+-;:./<=>?@\\^~|";
const String tok_sep_chars = "[]{}(),:";


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
    TokenKind kind;
    Uint line_nr;
    Uint line_start_byte_idx;
    Uint byte_idx;
    Uint num_bytes;
} Token;
typedef SliceOf(Token) Tokens;


Str tokSrc(Token *tok, Str full_src) {
    return strSub(full_src, tok->byte_idx, tok->byte_idx + tok->num_bytes);
}


Tokens tokenize(Str full_src, Bool keep_comment_toks) {
    Tokens toks = alloc(Token, full_src.len);
    toks.len = 0;

    TokenKind state = tok_kind_none;
    Uint i = 0;
    Uint cur_line_nr = 0;
    Uint cur_line_idx = 0;
    Int tok_idx_start = -1;
    Int tok_idx_last = -1;

    // shebang? #!
    if (full_src.len > 2 && full_src.at[0] == '#' && full_src.at[1] == '!') {
        state = tok_kind_comment;
        tok_idx_start = 0;
        i = 2;
    }

    for (; i < full_src.len; i += 1) {
        U8 c = full_src.at[i];
        if (c == '\n') {
            if (state == tok_kind_lit_str_double || state == tok_kind_lit_str_single)
                panic("line-break in literal in line %zu:\n%s", uintToStr(1 + cur_line_nr, 10),
                      strSub(full_src, tok_idx_start, i));
            if (tok_idx_start != -1 && tok_idx_last == -1)
                tok_idx_last = i - 1;
        } else {
            switch (state) {
            case tok_kind_lit_int:
                // fall-through to:
            case tok_kind_ident:
                if (c == ' ' || c == '\t' || c == '\"' || c == '\'' || strHasChar(tok_sep_chars, c)
                    || (strHasChar(tok_op_chars, c) && !strHasChar(tok_op_chars, full_src.at[i - 1]))
                    || (strHasChar(tok_op_chars, full_src.at[i - 1]) && !strHasChar(tok_op_chars, c))) {
                    i -= 1;
                    tok_idx_last = i;
                }
                break;
            case tok_kind_lit_str_double:
                if (c == '\"')
                    tok_idx_last = i;
                break;
            case tok_kind_lit_str_single:
                if (c == '\'')
                    tok_idx_last = i;
                break;
            case tok_kind_comment:
                break;
            case tok_kind_none:
                switch (c) {
                case '\"':
                    tok_idx_start = i;
                    state = tok_kind_lit_str_double;
                    break;
                case '\'':
                    tok_idx_start = i;
                    state = tok_kind_lit_str_single;
                    break;
                case '[':
                    tok_idx_last = i;
                    tok_idx_start = i;
                    state = tok_kind_sep_bsquare_open;
                    break;
                case ']':
                    tok_idx_last = i;
                    tok_idx_start = i;
                    state = tok_kind_sep_bsquare_close;
                    break;
                case '(':
                    tok_idx_last = i;
                    tok_idx_start = i;
                    state = tok_kind_sep_bparen_open;
                    break;
                case ')':
                    tok_idx_last = i;
                    tok_idx_start = i;
                    state = tok_kind_sep_bparen_close;
                    break;
                case '{':
                    tok_idx_last = i;
                    tok_idx_start = i;
                    state = tok_kind_sep_bcurly_open;
                    break;
                case '}':
                    tok_idx_last = i;
                    tok_idx_start = i;
                    state = tok_kind_sep_bcurly_close;
                    break;
                case ',':
                    tok_idx_last = i;
                    tok_idx_start = i;
                    state = tok_kind_sep_comma;
                    break;
                default:
                    if (c == '/' && i < full_src.len - 1 && full_src.at[i + 1] == '/') {
                        tok_idx_start = i;
                        state = tok_kind_comment;
                    } else if (c >= '0' && c <= '9') {
                        tok_idx_start = i;
                        state = tok_kind_lit_int;
                    } else if (c == ' ' || c == '\t') {
                        if (tok_idx_start != -1 && tok_idx_last == -1)
                            tok_idx_last = i - 1;
                    } else {
                        tok_idx_start = i;
                        state = tok_kind_ident;
                    }
                    break;
                }
                break;
            default:
                panic("unreachable");
            }
        }
        if (tok_idx_last != -1) {
            assert(state != tok_kind_none && tok_idx_start != -1, "unreachable");
            if (state != tok_kind_comment || keep_comment_toks) {
                toks.at[toks.len] = (Token) {
                    .kind = state,
                    .line_nr = cur_line_nr,
                    .line_start_byte_idx = cur_line_idx,
                    .byte_idx = (Uint)(tok_idx_start),
                    .num_bytes = (Uint)(1 + (tok_idx_last - tok_idx_start)),
                };
                toks.len += 1;
            }
            state = tok_kind_none;
            tok_idx_start = -1;
            tok_idx_last = -1;
        }
        if (c == '\n') {
            cur_line_nr += 1;
            cur_line_idx = i + 1;
        }
    }
    if (tok_idx_start != -1) {
        assert(state != tok_kind_none, "unreachable");
        if (state != tok_kind_comment || keep_comment_toks) {
            toks.at[toks.len] = (Token) {
                .kind = state,
                .line_nr = cur_line_nr,
                .line_start_byte_idx = cur_line_idx,
                .byte_idx = (Uint)(tok_idx_start),
                .num_bytes = i - tok_idx_start,
            };
            toks.len += 1;
        }
    }
    return toks;
}
