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
    Uint char_pos_line_start;
    Uint char_pos;
    Uint str_len;
} Token;
typedef SliceOf(Token) Tokens;


Bool tokIsOpeningBracket(TokenKind tok_kind) {
    return tok_kind == tok_kind_sep_bcurly_open || tok_kind == tok_kind_sep_bparen_open || tok_kind == tok_kind_sep_bsquare_open;
}

Bool tokIsClosingBracket(TokenKind tok_kind) {
    return tok_kind == tok_kind_sep_bcurly_close || tok_kind == tok_kind_sep_bparen_close || tok_kind == tok_kind_sep_bsquare_close;
}

Bool tokIsBracket(TokenKind tok_kind) {
    return tokIsOpeningBracket(tok_kind) || tokIsClosingBracket(tok_kind);
}

Uint tokPosCol(Token *tok) {
    return tok->char_pos - tok->char_pos_line_start;
}

Str tokSrc(Token *tok, Str full_src) {
    return strSub(full_src, tok->char_pos, tok->char_pos + tok->str_len);
}

Bool tokCanThrong(Token *tok, Str full_src) {
    return tok->kind == tok_kind_lit_int || tok->kind == tok_kind_lit_str_double || tok->kind == tok_kind_lit_str_single
           || (tok->kind == tok_kind_ident && full_src.at[tok->char_pos] != ':' && full_src.at[tok->char_pos] != '=');
}

Uint tokThrong(Tokens toks, Uint tok_idx, Str full_src) {
    if (tokCanThrong(&toks.at[tok_idx], full_src)) {
        for (Uint i = tok_idx + 1; i < toks.len; i += 1)
            if (toks.at[i].char_pos == toks.at[i - 1].char_pos + toks.at[i - 1].str_len && tokCanThrong(&toks.at[i], full_src))
                tok_idx = i;
            else
                break;
    }
    return tok_idx;
}

Uint toksCount(Tokens toks, Str ident, Str full_src) {
    Uint ret_num = 0;
    for (Uint i = 0; i < toks.len; i += 1)
        if (toks.at[i].kind == tok_kind_ident && strEql(ident, tokSrc(&toks.at[i], full_src)))
            ret_num += 1;
    return ret_num;
}

Uint toksCountUnnested(Tokens toks, TokenKind tok_kind) {
    assert(!tokIsBracket(tok_kind), "toksCountUnnested: caller mistake");
    Uint ret_num = 0;
    Int level = 0;
    for (Uint i = 0; i < toks.len; i += 1) {
        TokenKind this_kind = toks.at[i].kind;
        if (this_kind == tok_kind && level == 0)
            ret_num++;
        else if (tokIsOpeningBracket(this_kind))
            level += 1;
        else if (tokIsClosingBracket(this_kind))
            level -= 1;
    }
    return ret_num;
}

void toksCheckBrackets(Tokens toks) {
    Int level_bparen = 0, level_bsquare = 0, level_bcurly = 0;
    Int line_bparen = -1, line_bsquare = -1, line_bcurly = -1;
    for (Uint i = 0; i < toks.len; i += 1) {
        switch (toks.at[i].kind) {
        case tok_kind_sep_bcurly_open:
            level_bcurly++;
            if (line_bcurly == -1)
                line_bcurly = toks.at[i].line_nr;
            break;
        case tok_kind_sep_bparen_open:
            level_bparen++;
            if (line_bparen == -1)
                line_bparen = toks.at[i].line_nr;
            break;
        case tok_kind_sep_bsquare_open:
            level_bsquare++;
            if (line_bsquare == -1)
                line_bsquare = toks.at[i].line_nr;
            break;
        case tok_kind_sep_bcurly_close:
            level_bcurly--;
            if (level_bcurly == 0)
                line_bcurly = -1;
            break;
        case tok_kind_sep_bparen_close:
            level_bparen--;
            if (level_bparen == 0)
                line_bparen = -1;
            break;
        case tok_kind_sep_bsquare_close:
            level_bsquare--;
            if (level_bsquare == 0)
                line_bsquare = -1;
            break;
        default:
            break;
        }
        if (level_bparen < 0)
            panic("unmatched closing parenthesis in line %s", strZ(uintToStr(1 + toks.at[i].line_nr, 10)));
        else if (level_bcurly < 0)
            panic("unmatched closing curly brace in line %s", strZ(uintToStr(1 + toks.at[i].line_nr, 10)));
        else if (level_bsquare < 0)
            panic("unmatched closing square bracket in line %s", strZ(uintToStr(1 + toks.at[i].line_nr, 10)));
    }
    if (level_bparen > 0)
        panic("unmatched opening parenthesis in line %s", strZ(uintToStr(1 + line_bparen, 10)));
    else if (level_bcurly > 0)
        panic("unmatched opening curly brace in line %s", strZ(uintToStr(1 + line_bcurly, 10)));
    else if (level_bsquare > 0)
        panic("unmatched opening square bracket in line %s", strZ(uintToStr(1 + line_bsquare, 10)));
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
                panic("line-break in literal in line %s:\n%s", strZ(uintToStr(1 + cur_line_nr, 10)), strZ(strSub(full_src, tok_idx_start, i)));
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
                    .char_pos_line_start = cur_line_idx,
                    .char_pos = (Uint)(tok_idx_start),
                    .str_len = (Uint)(1 + (tok_idx_last - tok_idx_start)),
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
                .char_pos_line_start = cur_line_idx,
                .char_pos = (Uint)(tok_idx_start),
                .str_len = i - tok_idx_start,
            };
            toks.len += 1;
        }
    }
    return toks;
}
