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
typedef SliceOf(Tokens) Tokenss;


Bool tokIsOpeningBracket(TokenKind const tok_kind) {
    return tok_kind == tok_kind_sep_bcurly_open || tok_kind == tok_kind_sep_bparen_open || tok_kind == tok_kind_sep_bsquare_open;
}

Bool tokIsClosingBracket(TokenKind const tok_kind) {
    return tok_kind == tok_kind_sep_bcurly_close || tok_kind == tok_kind_sep_bparen_close || tok_kind == tok_kind_sep_bsquare_close;
}

Bool tokIsBracket(TokenKind const tok_kind) {
    return tokIsOpeningBracket(tok_kind) || tokIsClosingBracket(tok_kind);
}

Uint tokPosCol(Token const *const tok) {
    return tok->char_pos - tok->char_pos_line_start;
}

Bool tokCanThrong(Token const *const tok, Str const full_src) {
    return tok->kind == tok_kind_lit_int || tok->kind == tok_kind_lit_str_double || tok->kind == tok_kind_lit_str_single
           || (tok->kind == tok_kind_ident && full_src.at[tok->char_pos] != ':' && full_src.at[tok->char_pos] != '=');
}

Uint tokThrong(Tokens const toks, Uint const tok_idx, Str const full_src) {
    Uint ret_idx = tok_idx;
    if (tokCanThrong(&toks.at[ret_idx], full_src)) {
        for (Uint i = ret_idx + 1; i < toks.len; i += 1)
            if (toks.at[i].char_pos == toks.at[i - 1].char_pos + toks.at[i - 1].str_len && tokCanThrong(&toks.at[i], full_src))
                ret_idx = i;
            else
                break;
    }
    return ret_idx;
}

Str tokSrc(Token const *const tok, Str const full_src) {
    return strSub(full_src, tok->char_pos, tok->char_pos + tok->str_len);
}

Str toksSrc(Tokens const toks, Str const full_src) {
    Token *tok_last = &toks.at[toks.len - 1];
    return strSub(full_src, toks.at[0].char_pos, tok_last->char_pos + tok_last->str_len);
}

Uint toksCount(Tokens const toks, Str const ident, Str const full_src) {
    Uint ret_num = 0;
    for (Uint i = 0; i < toks.len; i += 1)
        if (toks.at[i].kind == tok_kind_ident && strEql(ident, tokSrc(&toks.at[i], full_src)))
            ret_num += 1;
    return ret_num;
}

Uint toksCountUnnested(Tokens const toks, TokenKind const tok_kind) {
    assert(!tokIsBracket(tok_kind));
    Uint ret_num = 0;
    Int level = 0;
    for (Uint i = 0; i < toks.len; i += 1) {
        TokenKind this_kind = toks.at[i].kind;
        if (this_kind == tok_kind && level == 0)
            ret_num += 1;
        else if (tokIsOpeningBracket(this_kind))
            level += 1;
        else if (tokIsClosingBracket(this_kind))
            level -= 1;
    }
    return ret_num;
}

void toksCheckBrackets(Tokens const toks) {
    Int level_bparen = 0, level_bsquare = 0, level_bcurly = 0;
    Int line_bparen = -1, line_bsquare = -1, line_bcurly = -1;
    for (Uint i = 0; i < toks.len; i += 1) {
        switch (toks.at[i].kind) {
        case tok_kind_sep_bcurly_open:
            level_bcurly += 1;
            if (line_bcurly == -1)
                line_bcurly = toks.at[i].line_nr;
            break;
        case tok_kind_sep_bparen_open:
            level_bparen += 1;
            if (line_bparen == -1)
                line_bparen = toks.at[i].line_nr;
            break;
        case tok_kind_sep_bsquare_open:
            level_bsquare += 1;
            if (line_bsquare == -1)
                line_bsquare = toks.at[i].line_nr;
            break;
        case tok_kind_sep_bcurly_close:
            level_bcurly -= 1;
            if (level_bcurly == 0)
                line_bcurly = -1;
            break;
        case tok_kind_sep_bparen_close:
            level_bparen -= 1;
            if (level_bparen == 0)
                line_bparen = -1;
            break;
        case tok_kind_sep_bsquare_close:
            level_bsquare -= 1;
            if (level_bsquare == 0)
                line_bsquare = -1;
            break;
        default: break;
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

Tokenss toksIndentBasedChunks(Tokens const toks) {
    assert(toks.len > 0);
    Uint cmp_pos_col = tokPosCol(&toks.at[0]);
    Int level = 0;
    for (Uint i = 0; i < toks.len; i += 1) {
        Token *tok = &toks.at[i];
        if (level == 0) {
            Uint pos_col = tokPosCol(tok);
            if (pos_col < cmp_pos_col)
                cmp_pos_col = pos_col;
        } else if (tokIsOpeningBracket(tok->kind))
            level += 1;
        else if (tokIsClosingBracket(tok->kind))
            level -= 1;
    }
    assert(level == 0);

    Uint num_chunks = 0;
    for (Uint i = 0; i < toks.len; i += 1) {
        Token *tok = &toks.at[i];
        if (level == 0) {
            if (i == 0 || tokPosCol(tok) <= cmp_pos_col)
                num_chunks += 1;
        } else if (tokIsOpeningBracket(tok->kind))
            level += 1;
        else if (tokIsClosingBracket(tok->kind))
            level -= 1;
    }
    assert(level == 0);

    Tokenss ret_chunks = alloc(Tokens, num_chunks);
    {
        ret_chunks.len = 0;
        Int start_from = -1;
        for (Uint i = 0; i < toks.len; i += 1) {
            Token *tok = &toks.at[i];
            if (i == 0 || (level == 0 && tokPosCol(tok) <= cmp_pos_col)) {
                if (start_from != -1)
                    append(ret_chunks, slice(Token, toks, start_from, i));
                start_from = i;
            }
            if (tokIsOpeningBracket(tok->kind))
                level += 1;
            else if (tokIsClosingBracket(tok->kind))
                level -= 1;
        }
        if (start_from != -1)
            append(ret_chunks, slice(Token, toks, start_from, toks.len));
        assert(ret_chunks.len == num_chunks);
    }
    return ret_chunks;
}

Int toksIndexOfIdent(Tokens const toks, Str const ident, Str const full_src) {
    for (Uint i = 0; i < toks.len; i += 1)
        if (toks.at[i].kind == tok_kind_ident && strEql(ident, tokSrc(&toks.at[i], full_src)))
            return i;
    return -1;
}

Int toksIndexOfMatchingBracket(Tokens const toks) {
    TokenKind tok_open_kind = toks.at[0].kind;
    TokenKind tok_close_kind = tok_kind_none;
    switch (tok_open_kind) {
    case tok_kind_sep_bcurly_open: tok_close_kind = tok_kind_sep_bcurly_close; break;
    case tok_kind_sep_bsquare_open: tok_close_kind = tok_kind_sep_bsquare_close; break;
    case tok_kind_sep_bparen_open: tok_close_kind = tok_kind_sep_bparen_close; break;
    default: break;
    }
    assert(tok_close_kind != tok_kind_none);

    Int level = 0;
    for (Uint i = 0; i < toks.len; i += 1)
        if (toks.at[i].kind == tok_open_kind)
            level += 1;
        else if (toks.at[i].kind == tok_close_kind) {
            level -= 1;
            if (level == 0)
                return i;
        }
    return -1;
}

Tokenss toksSplit(Tokens const toks, TokenKind const tok_kind) {
    assert(!tokIsBracket(tok_kind));
    Tokenss ret_sub_toks = alloc(Tokens, 1 + toksCountUnnested(toks, tok_kind));
    ret_sub_toks.len = 0;
    {
        Int level = 0;
        Uint start_from = 0;
        for (Uint i = 0; i < toks.len; i += 1) {
            Token *tok = &toks.at[i];
            if (tok->kind == tok_kind && level == 0) {
                append(ret_sub_toks, slice(Token, toks, start_from, i));
                start_from = i + 1;
            } else if (tokIsOpeningBracket(tok->kind))
                level += 1;
            else if (tokIsClosingBracket(tok->kind))
                level -= 1;
        }
        append(ret_sub_toks, slice(Token, toks, start_from, toks.len));
    }
    return ret_sub_toks;
}

Tokens tokenize(Str const full_src, Bool const keep_comment_toks) {
    Tokens toks = alloc(Token, full_src.len);
    toks.len = 0;

    TokenKind state = tok_kind_none;
    Uint cur_line_nr = 0;
    Uint cur_line_idx = 0;
    Int tok_idx_start = -1;
    Int tok_idx_last = -1;

    Uint i = 0;
    // shebang? #!
    if (full_src.len > 2 && full_src.at[0] == '#' && full_src.at[1] == '!') {
        state = tok_kind_comment;
        tok_idx_start = 0;
        i = 2;
    } // now we start:

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
            case tok_kind_comment: break;
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
            default: panic("unreachable");
            }
        }
        if (tok_idx_last != -1) {
            assert(state != tok_kind_none && tok_idx_start != -1);
            if (state != tok_kind_comment || keep_comment_toks)
                append(toks, ((Token) {
                                 .kind = state,
                                 .line_nr = cur_line_nr,
                                 .char_pos_line_start = cur_line_idx,
                                 .char_pos = (Uint)(tok_idx_start),
                                 .str_len = (Uint)(1 + (tok_idx_last - tok_idx_start)),
                             }));
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
        assert(state != tok_kind_none);
        if (state != tok_kind_comment || keep_comment_toks)
            append(toks, ((Token) {
                             .kind = state,
                             .line_nr = cur_line_nr,
                             .char_pos_line_start = cur_line_idx,
                             .char_pos = (Uint)(tok_idx_start),
                             .str_len = i - tok_idx_start,
                         }));
    }
    return toks;
}
