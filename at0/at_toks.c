#pragma once
#include "metaleap.c"


const String tok_op_chars = "!#$&*+-;:./<=>?@\\^~|\x25";
const String tok_sep_chars = "[]{}(),:";


typedef enum TokenKind {
    tok_kind_nope = 0,               // used for state machine inside `tokenize`, never produced to consumers
    tok_kind_comment = 1,            // double-slash comments, until EOL
    tok_kind_ident = 2,              // fallback for all otherwise-unmatched tokens
    tok_kind_lit_num_prefixed = 3,   // any tokens starting with '0'-'9'
    tok_kind_lit_str_qdouble = 4,    // "string-ish quote marks"
    tok_kind_lit_str_qsingle = 5,    // 'char-ish quote marks'
    tok_kind_sep_bparen_open = 6,    // (
    tok_kind_sep_bparen_close = 7,   // )
    tok_kind_sep_bcurly_open = 8,    // {
    tok_kind_sep_bcurly_close = 9,   // }
    tok_kind_sep_bsquare_open = 10,  // [
    tok_kind_sep_bsquare_close = 11, // ]
    tok_kind_sep_comma = 12,         // ,
} TokenKind;

typedef struct Token {
    TokenKind kind;
    Uint line_nr;
    Uint char_pos_line_start;
    Uint char_pos;
    Uint str_len;
} Token;
typedef ·SliceOf(Token) Tokens;
typedef ·SliceOf(Tokens) Tokenss;


static Bool tokIsOpeningBracket(TokenKind const tok_kind) {
    return tok_kind == tok_kind_sep_bcurly_open || tok_kind == tok_kind_sep_bparen_open || tok_kind == tok_kind_sep_bsquare_open;
}

static Bool tokIsClosingBracket(TokenKind const tok_kind) {
    return tok_kind == tok_kind_sep_bcurly_close || tok_kind == tok_kind_sep_bparen_close || tok_kind == tok_kind_sep_bsquare_close;
}

static Bool tokIsBracket(TokenKind const tok_kind) {
    return tokIsOpeningBracket(tok_kind) || tokIsClosingBracket(tok_kind);
}

static Uint tokPosCol(Token const* const tok) {
    return tok->char_pos - tok->char_pos_line_start;
}

static Bool tokCanThrong(Token const* const tok, Str const full_src) {
    return tok->kind == tok_kind_lit_num_prefixed || tok->kind == tok_kind_lit_str_qdouble || tok->kind == tok_kind_lit_str_qsingle
           || (tok->kind == tok_kind_ident && full_src.at[tok->char_pos] != ':' && full_src.at[tok->char_pos] != '=');
}

static Uint tokThrong(Tokens const toks, Uint const tok_idx, Str const full_src) {
    Uint ret_idx = tok_idx;
    if (tokCanThrong(&toks.at[tok_idx], full_src)) {
        for (Uint i = tok_idx + 1; i < toks.len; i += 1)
            if (toks.at[i].char_pos == toks.at[i - 1].char_pos + toks.at[i - 1].str_len && tokCanThrong(&toks.at[i], full_src))
                ret_idx = i;
            else
                break;
    }
    return ret_idx;
}

static Str tokSrc(Token const* const tok, Str const full_src) {
    return strSub(full_src, tok->char_pos, tok->char_pos + tok->str_len);
}

static Str toksSrc(Tokens const toks, Str const full_src) {
    Token* const tok_last = &toks.at[toks.len - 1];
    return strSub(full_src, toks.at[0].char_pos, tok_last->char_pos + tok_last->str_len);
}

static Uint toksCountUnnested(Tokens const toks, TokenKind const tok_kind) {
    ·assert(!tokIsBracket(tok_kind));
    Uint ret_num = 0;
    Int level = 0;
    ·forEach(Token, tok, toks, {
        if (tok->kind == tok_kind && level == 0)
            ret_num += 1;
        else if (tokIsOpeningBracket(tok->kind))
            level += 1;
        else if (tokIsClosingBracket(tok->kind))
            level -= 1;
    });
    return ret_num;
}

static void toksCheckBrackets(Tokens const toks) {
    Int level_bparen = 0, level_bsquare = 0, level_bcurly = 0;
    Int line_bparen = -1, line_bsquare = -1, line_bcurly = -1;
    ·forEach(Token, tok, toks, {
        switch (tok->kind) {
            case tok_kind_sep_bcurly_open:
                level_bcurly += 1;
                if (line_bcurly == -1)
                    line_bcurly = tok->line_nr;
                break;
            case tok_kind_sep_bparen_open:
                level_bparen += 1;
                if (line_bparen == -1)
                    line_bparen = tok->line_nr;
                break;
            case tok_kind_sep_bsquare_open:
                level_bsquare += 1;
                if (line_bsquare == -1)
                    line_bsquare = tok->line_nr;
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
            ·fail(str2(str("unmatched closing parenthesis in line "), uintToStr(1 + tok->line_nr, 1, 10)));
        else if (level_bcurly < 0)
            ·fail(str2(str("unmatched closing curly brace in line "), uintToStr(1 + tok->line_nr, 1, 10)));
        else if (level_bsquare < 0)
            ·fail(str2(str("unmatched closing square bracket in line "), uintToStr(1 + tok->line_nr, 1, 10)));
    });
    if (level_bparen > 0)
        ·fail(str2(str("unmatched opening parenthesis in line "), uintToStr(1 + line_bparen, 1, 10)));
    else if (level_bcurly > 0)
        ·fail(str2(str("unmatched opening curly brace in line "), uintToStr(1 + line_bcurly, 1, 10)));
    else if (level_bsquare > 0)
        ·fail(str2(str("unmatched opening square bracket in line "), uintToStr(1 + line_bsquare, 1, 10)));
}

static Tokenss toksIndentBasedChunks(Tokens const toks) {
    ·assert(toks.len > 0);
    Uint cmp_pos_col = tokPosCol(&toks.at[0]);
    Int level = 0;
    ·forEach(Token, tok, toks, {
        if (level == 0) {
            Uint const pos_col = tokPosCol(tok);
            if (pos_col < cmp_pos_col)
                cmp_pos_col = pos_col;
        }
        if (tokIsOpeningBracket(tok->kind))
            level += 1;
        else if (tokIsClosingBracket(tok->kind))
            level -= 1;
    });
    ·assert(level == 0);

    Uint num_chunks = 0;
    ·forEach(Token, tok, toks, {
        if (level == 0) {
            if (iˇtok == 0 || tokPosCol(tok) <= cmp_pos_col)
                num_chunks += 1;
        }
        if (tokIsOpeningBracket(tok->kind))
            level += 1;
        else if (tokIsClosingBracket(tok->kind))
            level -= 1;
    });
    ·assert(level == 0);

    Tokenss ret_chunks = ·make(Tokens, 0, num_chunks);
    {
        Int start_from = -1;
        ·forEach(Token, tok, toks, {
            if (iˇtok == 0 || (level == 0 && tokPosCol(tok) <= cmp_pos_col)) {
                if (start_from != -1)
                    ·append(ret_chunks, ·slice(Token, toks, start_from, iˇtok));
                start_from = iˇtok;
            }
            if (tokIsOpeningBracket(tok->kind))
                level += 1;
            else if (tokIsClosingBracket(tok->kind))
                level -= 1;
        });
        if (start_from != -1)
            ·append(ret_chunks, ·slice(Token, toks, start_from, toks.len));
        ·assert(ret_chunks.len == num_chunks);
    }
    return ret_chunks;
}

static ºUint toksIndexOfIdent(Tokens const toks, Str const ident, Str const full_src) {
    ·forEach(Token, tok, toks, {
        if (tok->kind == tok_kind_ident && strEql(ident, tokSrc(tok, full_src)))
            return ·ok(Uint, iˇtok);
    });
    return ·none(Uint);
}

static ºUint toksIndexOfMatchingBracket(Tokens const toks) {
    TokenKind const tok_open_kind = toks.at[0].kind;
    TokenKind tok_close_kind;
    switch (tok_open_kind) {
        case tok_kind_sep_bcurly_open: tok_close_kind = tok_kind_sep_bcurly_close; break;
        case tok_kind_sep_bsquare_open: tok_close_kind = tok_kind_sep_bsquare_close; break;
        case tok_kind_sep_bparen_open: tok_close_kind = tok_kind_sep_bparen_close; break;
        default: ·fail(str("toksIndexOfMatchingBracket: caller mistake")); break;
    }

    Int level = 0;
    ·forEach(Token, tok, toks, {
        if (tok->kind == tok_open_kind)
            level += 1;
        else if (tok->kind == tok_close_kind) {
            level -= 1;
            if (level == 0)
                return ·ok(Uint, iˇtok);
        }
    });
    return ·none(Uint);
}

static Tokenss toksSplit(Tokens const toks, TokenKind const tok_kind) {
    ·assert(!tokIsBracket(tok_kind));
    if (toks.len == 0)
        return (Tokenss) {.len = 0, .at = NULL};
    Uint capacity = 1 + toksCountUnnested(toks, tok_kind);
    Tokenss ret_sub_toks = ·make(Tokens, 0, capacity);
    {
        Int level = 0;
        Uint start_from = 0;
        ·forEach(Token, tok, toks, {
            if (tok->kind == tok_kind && level == 0) {
                ·append(ret_sub_toks, ·slice(Token, toks, start_from, iˇtok));
                start_from = iˇtok + 1;
            } else if (tokIsOpeningBracket(tok->kind))
                level += 1;
            else if (tokIsClosingBracket(tok->kind))
                level -= 1;
        });
        ·append(ret_sub_toks, ·slice(Token, toks, start_from, toks.len));
    }
    return ret_sub_toks;
}

static Tokens tokenize(Str const full_src, Bool const keep_comment_toks) {
    Tokens toks = ·make(Token, 0, full_src.len);

    TokenKind state = tok_kind_nope;
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
        U8 const c = full_src.at[i];
        if (c == '\n') {
            if (state == tok_kind_lit_str_qdouble || state == tok_kind_lit_str_qsingle)
                ·fail(str4(str("line-break in literal in line "), uintToStr(1 + cur_line_nr, 1, 10), str(":\n"),
                           strSub(full_src, tok_idx_start, i)));
            if (tok_idx_start != -1 && tok_idx_last == -1)
                tok_idx_last = i - 1;
        } else {
            switch (state) {
                case tok_kind_lit_num_prefixed:
                case tok_kind_ident:
                    if (c == ' ' || c == '\t' || c == '\"' || c == '\'' || strHasChar(tok_sep_chars, c)
                        || (strHasChar(tok_op_chars, c) && !strHasChar(tok_op_chars, full_src.at[i - 1]))
                        || (strHasChar(tok_op_chars, full_src.at[i - 1]) && !strHasChar(tok_op_chars, c))) {
                        i -= 1;
                        tok_idx_last = i;
                    }
                    break;
                case tok_kind_lit_str_qdouble:
                    if (c == '\"')
                        tok_idx_last = i;
                    break;
                case tok_kind_lit_str_qsingle:
                    if (c == '\'')
                        tok_idx_last = i;
                    break;
                case tok_kind_comment: break;
                case tok_kind_nope:
                    switch (c) {
                        case '\"':
                            tok_idx_start = i;
                            state = tok_kind_lit_str_qdouble;
                            break;
                        case '\'':
                            tok_idx_start = i;
                            state = tok_kind_lit_str_qsingle;
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
                                // begin comment tok
                                tok_idx_start = i;
                                state = tok_kind_comment;
                            } else if (c >= '0' && c <= '9') {
                                // begin num-ish tok
                                tok_idx_start = i;
                                state = tok_kind_lit_num_prefixed;
                            } else if (c == ' ' || c == '\t') {
                                // white-space not in string: end of running token
                                if (tok_idx_start != -1 && tok_idx_last == -1)
                                    tok_idx_last = i - 1;
                            } else {
                                // fallback is ident for otherwise-unrecognized toks
                                tok_idx_start = i;
                                state = tok_kind_ident;
                            }
                            break;
                    }
                    break;
                default: {
                    ·fail(str("unreachable"));
                } break;
            }
        }
        if (tok_idx_last != -1) {
            ·assert(state != tok_kind_nope && tok_idx_start != -1);
            if (state != tok_kind_comment || keep_comment_toks)
                ·append(toks, ((Token) {
                                  .kind = state,
                                  .line_nr = cur_line_nr,
                                  .char_pos_line_start = cur_line_idx,
                                  .char_pos = (Uint)(tok_idx_start),
                                  .str_len = (Uint)(1 + (tok_idx_last - tok_idx_start)),
                              }));
            state = tok_kind_nope;
            tok_idx_start = -1;
            tok_idx_last = -1;
        }
        if (c == '\n') {
            cur_line_nr += 1;
            cur_line_idx = i + 1;
        }
    }
    if (tok_idx_start != -1) {
        ·assert(state != tok_kind_nope);
        if (state != tok_kind_comment || keep_comment_toks)
            ·append(toks, ((Token) {
                              .kind = state,
                              .line_nr = cur_line_nr,
                              .char_pos_line_start = cur_line_idx,
                              .char_pos = (Uint)(tok_idx_start),
                              .str_len = i - tok_idx_start,
                          }));
    }
    return toks;
}
