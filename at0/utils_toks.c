#pragma once
#include "utils_std_mem.c"


const CStr tok_op_chars = "!#$&*+-;:./<=>?@\\^~|\x25";
const CStr tok_sep_chars = "[]{}(),:";


typedef enum TokenKind {
    tok_kind_nope,              // used for state machine inside `tokenize`, never produced to consumers
    tok_kind_comment,           // double-slash comments, until EOL
    tok_kind_ident,             // fallback for all otherwise-unmatched tokens
    tok_kind_lit_num_prefixed,  // any tokens starting with '0'-'9'
    tok_kind_lit_str_qdouble,   // "string-ish quote marks"
    tok_kind_lit_str_qsingle,   // 'char-ish quote marks'
    tok_kind_sep_bparen_open,   // (
    tok_kind_sep_bparen_close,  // )
    tok_kind_sep_bcurly_open,   // {
    tok_kind_sep_bcurly_close,  // }
    tok_kind_sep_bsquare_open,  // [
    tok_kind_sep_bsquare_close, // ]
    tok_kind_sep_comma,         // ,
} TokenKind;

typedef struct Token {
    TokenKind kind;
    UInt line_nr;
    UInt char_pos_line_start;
    UInt char_pos;
    UInt str_len;
    Str file_name;
} Token;
typedef ·SliceOf(Token) Tokens;
typedef ·SliceOf(Tokens) Tokenss;


typedef enum SrcFileIssueKind {
    issue_tok,
    issue_parse,
    issue_irhl,
    issue_irml,
    issue_irml_cps,
    issue_irml_sema,
    issue_irll,
} SrcFileIssueKind;

typedef struct SrcFileIssue {
    struct {
        UInt line_nr;
        UInt char_pos_line_start;
        UInt char_pos;
    } src_pos;
    Bool error;
    Str msg;
    SrcFileIssueKind kind;
} SrcFileIssue;
typedef ·ListOf(SrcFileIssue) SrcFileIssues;
typedef ·Maybe(SrcFileIssue) ºSrcFileIssue;

ºSrcFileIssue srcFileIssue(SrcFileIssueKind kind, Bool const error, Str const msg, Token const src_pos) {
    return ·got(SrcFileIssue,
                ((SrcFileIssue) {
                    .error = error,
                    .src_pos = {.line_nr = src_pos.line_nr, .char_pos_line_start = src_pos.char_pos_line_start, .char_pos = src_pos.char_pos},
                    .msg = msg}));
}
ºSrcFileIssue srcFileErr(SrcFileIssueKind kind, Str const msg, Token const src_pos) {
    return srcFileIssue(kind, true, msg, src_pos);
}


Bool tokIsOpeningBracket(TokenKind const tok_kind) {
    return tok_kind == tok_kind_sep_bcurly_open || tok_kind == tok_kind_sep_bparen_open || tok_kind == tok_kind_sep_bsquare_open;
}

Bool tokIsClosingBracket(TokenKind const tok_kind) {
    return tok_kind == tok_kind_sep_bcurly_close || tok_kind == tok_kind_sep_bparen_close || tok_kind == tok_kind_sep_bsquare_close;
}

Bool tokIsBracket(TokenKind const tok_kind) {
    return tokIsOpeningBracket(tok_kind) || tokIsClosingBracket(tok_kind);
}

UInt tokPosCol(Token const* const tok) {
    return tok->char_pos - tok->char_pos_line_start;
}

Str tokPosStr(MemHeap* mem_heap, Token const* const tok) {
    Str const line_nr = uIntToStr(mem_heap, 1 + tok->line_nr, 1, 10);
    return str3(mem_heap, tok->file_name, strL(":", 1), line_nr);
}

Str tokSrc(Token const* const tok, Str const full_src) {
    return strSub(full_src, tok->char_pos, tok->char_pos + tok->str_len);
}

Str toksSrc(Tokens const toks, Str const full_src) {
    Token* const tok_last = ·last(toks);
    return strSub(full_src, toks.at[0].char_pos, tok_last->char_pos + tok_last->str_len);
}

UInt toksCountUnnested(Tokens const toks, TokenKind const tok_kind) {
    ·assert(!tokIsBracket(tok_kind));
    UInt ret_num = 0;
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

Bool toksHavePos(Tokens const toks, UInt const pos_line, UInt const pos_col) {
    return (pos_line >= toks.at[0].line_nr && pos_line <= ·last(toks)->line_nr);
}

ºSrcFileIssue toksVerifyBrackets(Tokens const toks, Str const file_name) {
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
            return srcFileErr(issue_tok, str("unmatched closing parenthesis"), *tok);
        else if (level_bcurly < 0)
            return srcFileErr(issue_tok, str("unmatched closing curly brace"), *tok);
        else if (level_bsquare < 0)
            return srcFileErr(issue_tok, str("unmatched closing square bracket"), *tok);
    });
    if (level_bparen > 0)
        return srcFileErr(issue_tok, str("unmatched opening parenthesis"),
                          (Token) {.file_name = file_name, .line_nr = line_bparen, .kind = tok_kind_nope});
    else if (level_bcurly > 0)
        return srcFileErr(issue_tok, str("unmatched opening curly brace"),
                          (Token) {.file_name = file_name, .line_nr = line_bcurly, .kind = tok_kind_nope});
    else if (level_bsquare > 0)
        return srcFileErr(issue_tok, str("unmatched opening square bracket"),
                          (Token) {.file_name = file_name, .line_nr = line_bsquare, .kind = tok_kind_nope});
    return ·none(SrcFileIssue);
}

Tokenss toksIndentBasedChunks(MemHeap* mem_heap, Tokens const toks) {
    if (toks.len == 0)
        return ·sliceOf(Tokens, mem_heap, 0, 0);
    UInt cmp_pos_col = tokPosCol(&toks.at[0]);
    Int level = 0;
    ·forEach(Token, tok, toks, {
        if (level == 0) {
            UInt const pos_col = tokPosCol(tok);
            if (pos_col < cmp_pos_col)
                cmp_pos_col = pos_col;
        }
        if (tokIsOpeningBracket(tok->kind))
            level += 1;
        else if (tokIsClosingBracket(tok->kind))
            level -= 1;
    });
    ·assert(level == 0);

    UInt num_chunks = 0;
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

    Tokenss ret_chunks = ·sliceOf(Tokens, mem_heap, 0, num_chunks);
    {
        Int start_from = -1;
        ·forEach(Token, tok, toks, {
            if (iˇtok == 0 || (level == 0 && tokPosCol(tok) <= cmp_pos_col)) {
                if (start_from != -1)
                    ·push(ret_chunks, ·slice(Token, toks, start_from, iˇtok));
                start_from = iˇtok;
            }
            if (tokIsOpeningBracket(tok->kind))
                level += 1;
            else if (tokIsClosingBracket(tok->kind))
                level -= 1;
        });
        if (start_from != -1)
            ·push(ret_chunks, ·slice(Token, toks, start_from, toks.len));
        ·assert(ret_chunks.len == num_chunks);
    }
    return ret_chunks;
}

ºUInt toksIndexOfIdent(Tokens const toks, Str const ident, Str const full_src) {
    ·forEach(Token, tok, toks, {
        if (tok->kind == tok_kind_ident && strEql(ident, tokSrc(tok, full_src)))
            return ·got(UInt, iˇtok);
    });
    return ·none(UInt);
}

ºUInt toksIndexOfMatchingBracket(Tokens const toks) {
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
                return ·got(UInt, iˇtok);
        }
    });
    return ·none(UInt);
}

Bool tokCanThrong(Token const* const tok, Str const full_src) {
    return tok->kind == tok_kind_lit_num_prefixed || tok->kind == tok_kind_lit_str_qdouble || tok->kind == tok_kind_lit_str_qsingle
           || (tok->kind == tok_kind_ident && full_src.at[tok->char_pos] != ':' && (full_src.at[tok->char_pos] != '=' || tok->str_len > 1));
}

UInt tokThrong(Tokens const toks, UInt const tok_idx, Str const full_src) {
    UInt ret_idx = tok_idx;

    if (tokIsOpeningBracket(toks.at[tok_idx].kind)) {
        ºUInt idx_close = toksIndexOfMatchingBracket(·slice(Token, toks, tok_idx, toks.len));
        ·assert(idx_close.got);
        idx_close.it += tok_idx;
        UInt idx_next = idx_close.it + 1;
        if (idx_next < toks.len && toks.at[idx_next].char_pos == toks.at[idx_close.it].char_pos + 1
            && (tokCanThrong(&toks.at[idx_next], full_src) || tokIsOpeningBracket(toks.at[idx_next].kind))) {
            UInt const idx_throng = tokThrong(toks, idx_next, full_src);
            return idx_throng;
        }
    } else if (tokCanThrong(&toks.at[tok_idx], full_src)) {
        Bool ok = true;
        for (UInt i = tok_idx + 1; ok && i < toks.len; i += 1) {
            Token* const tok = &toks.at[i];
            ok = (tok->char_pos == (toks.at[i - 1].char_pos + toks.at[i - 1].str_len));
            if (ok) {
                ok = tokCanThrong(tok, full_src);
                if (!ok && tokIsOpeningBracket(tok->kind)) {
                    ºUInt idx_closing_bracket = toksIndexOfMatchingBracket(·slice(Token, toks, i, toks.len));
                    ·assert(idx_closing_bracket.got);
                    i += idx_closing_bracket.it;
                    ok = true;
                }
                if (ok)
                    ret_idx = i;
            }
        }
    }
    return ret_idx;
}

Tokenss toksSplit(MemHeap* mem_heap, Tokens const toks, TokenKind const tok_kind) {
    ·assert(!tokIsBracket(tok_kind));
    if (toks.len == 0)
        return ·len0(Tokens);
    UInt capacity = 1 + toksCountUnnested(toks, tok_kind);
    Tokenss ret_sub_toks = ·sliceOf(Tokens, mem_heap, 0, capacity);
    {
        Int level = 0;
        UInt start_from = 0;
        ·forEach(Token, tok, toks, {
            if (tok->kind == tok_kind && level == 0) {
                ·push(ret_sub_toks, ·slice(Token, toks, start_from, iˇtok));
                start_from = iˇtok + 1;
            } else if (tokIsOpeningBracket(tok->kind))
                level += 1;
            else if (tokIsClosingBracket(tok->kind))
                level -= 1;
        });
        ·push(ret_sub_toks, ·slice(Token, toks, start_from, toks.len));
    }
    return ret_sub_toks;
}

Tokens tokenize(MemHeap* mem_heap, Str const full_src, Bool const keep_comment_toks, Str const file_name,
                SrcFileIssues* const gather_issues) {
    SrcFileIssues issues;
    if (gather_issues != NULL)
        issues = *gather_issues;
    Tokens toks = ·sliceOf(Token, mem_heap, 0, full_src.len);

    TokenKind state = tok_kind_nope;
    UInt cur_line_nr = 0;
    UInt cur_line_idx = 0;
    Int tok_idx_start = -1;
    Int tok_idx_last = -1;

    UInt i = 0;
    // shebang? #!
    if (full_src.len > 2 && full_src.at[0] == '#' && full_src.at[1] == '!') {
        state = tok_kind_comment;
        tok_idx_start = 0;
        i = 2;
    } // now we start:

    for (; i < full_src.len; i += 1) {
        U8 const c = full_src.at[i];
        if (c == '\n') {
            if (state == tok_kind_lit_str_qdouble || state == tok_kind_lit_str_qsingle) {
                if (gather_issues != NULL)
                    ·append(issues, srcFileErr(issue_tok, str("line-break in literal"),
                                               (Token) {.file_name = file_name,
                                                        .kind = state,
                                                        .str_len = i - tok_idx_start,
                                                        .line_nr = cur_line_nr,
                                                        .char_pos_line_start = cur_line_idx,
                                                        .char_pos = tok_idx_start})
                                        .it);
                toks.len = 0;
                break;
            }
            if (tok_idx_start != -1 && tok_idx_last == -1)
                tok_idx_last = i - 1;
        } else {
            switch (state) {
                case tok_kind_lit_num_prefixed:
                case tok_kind_ident:
                    if (c == ' ' || c == '\t' || c == '\"' || c == '\'' || cStrHasChar(tok_sep_chars, c)
                        || (cStrHasChar(tok_op_chars, c) && !cStrHasChar(tok_op_chars, full_src.at[i - 1]))
                        || (cStrHasChar(tok_op_chars, full_src.at[i - 1]) && !cStrHasChar(tok_op_chars, c))) {
                        i -= 1;
                        tok_idx_last = i;
                    }
                    break;
                case tok_kind_lit_str_qdouble:
                    if (c == '\"' && full_src.at[i - 1] != '\\')
                        tok_idx_last = i;
                    break;
                case tok_kind_lit_str_qsingle:
                    if (c == '\'' && full_src.at[i - 1] != '\\')
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
        Bool reset_line_nr = false;
        if (tok_idx_last != -1) {
            ·assert(state != tok_kind_nope && tok_idx_start != -1);
            if (state != tok_kind_comment || keep_comment_toks)
                ·push(toks, ((Token) {
                                .kind = state,
                                .line_nr = cur_line_nr,
                                .char_pos_line_start = cur_line_idx,
                                .char_pos = (UInt)(tok_idx_start),
                                .str_len = (UInt)(1 + (tok_idx_last - tok_idx_start)),
                                .file_name = file_name,
                            }));
            state = tok_kind_nope;
            tok_idx_start = -1;
            tok_idx_last = -1;
        }
        if (c == '\n') {
            cur_line_nr = reset_line_nr ? 0 : cur_line_nr + 1;
            cur_line_idx = i + 1;
        }
    }
    if (toks.at != NULL && tok_idx_start != -1) {
        ·assert(state != tok_kind_nope);
        if (state != tok_kind_comment || keep_comment_toks)
            ·push(toks, ((Token) {
                            .kind = state,
                            .line_nr = cur_line_nr,
                            .char_pos_line_start = cur_line_idx,
                            .char_pos = (UInt)(tok_idx_start),
                            .str_len = i - tok_idx_start,
                            .file_name = file_name,
                        }));
    }
    ºSrcFileIssue mismatch = toksVerifyBrackets(toks, file_name);
    if (mismatch.got) {
        toks.len = 0;
        if (gather_issues != NULL)
            ·append(issues, mismatch.it);
    }
    if (gather_issues != NULL)
        *gather_issues = issues;
    return toks;
}
