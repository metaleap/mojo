package main

type TokenKind int

const (
	tok_kind_none TokenKind = iota

	tok_kind_comment
	tok_kind_ident

	tok_kind_lit_int
	tok_kind_lit_str

	tok_kind_sep_bparen_open
	tok_kind_sep_bparen_close
	tok_kind_sep_bcurly_open
	tok_kind_sep_bcurly_close
	tok_kind_sep_bsquare_open
	tok_kind_sep_bsquare_close
	tok_kind_sep_comma
	tok_kind_sep_colon
	tok_kind_sep_def
)

type Token struct {
	idx      int
	len      int
	line_nr  int
	line_idx int
	kind     TokenKind
}

func tokenize(full_src Str) []Token {
	i, cur_line_nr, cur_line_idx, toks_count := 0, 0, 0, 0
	tok_start, tok_last := -1, -1
	var state TokenKind = tok_kind_none
	toks := allocˇToken(len(full_src))
	for i = 0; i < len(full_src); i++ {
		c := full_src[i]
		if c == '\n' {
			if state == tok_kind_lit_str {
				fail("line-break in literal near:\n", full_src[tok_start:i])
			}
			if tok_start != -1 && tok_last == -1 {
				tok_last = i - 1
			}
		} else {
			switch state {
			case tok_kind_lit_int, tok_kind_ident:
				if c == ' ' || c == '\t' || c == '"' || c == '\'' || c == '[' || c == ']' || c == '{' || c == '}' || c == '(' || c == ')' || c == ',' || c == ':' {
					i--
					tok_last = i
				}
			case tok_kind_lit_str:
				if c == '"' {
					tok_last = i
				}
			case tok_kind_none:
				switch c {
				case '"':
					tok_start = i
					state = tok_kind_lit_str
				case '[':
					tok_last = i
					tok_start = i
					state = tok_kind_sep_bsquare_open
				case ']':
					tok_last = i
					tok_start = i
					state = tok_kind_sep_bsquare_close
				case '(':
					tok_last = i
					tok_start = i
					state = tok_kind_sep_bparen_open
				case ')':
					tok_last = i
					tok_start = i
					state = tok_kind_sep_bparen_close
				case '{':
					tok_last = i
					tok_start = i
					state = tok_kind_sep_bcurly_open
				case '}':
					tok_last = i
					tok_start = i
					state = tok_kind_sep_bcurly_close
				case ',':
					tok_last = i
					tok_start = i
					state = tok_kind_sep_comma
				case ':':
					if i < len(full_src)-1 && full_src[i+1] == '=' {
						tok_last = i + 1
						tok_start = i
						state = tok_kind_sep_def
						i++
					} else {
						tok_last = i
						tok_start = i
						state = tok_kind_sep_colon
					}
				default:
					if c == '/' && i < len(full_src)-1 && full_src[i+1] == '/' {
						tok_start = i
						state = tok_kind_comment
					} else if c >= '0' && c <= '9' {
						tok_start = i
						state = tok_kind_lit_int
					} else if c == ' ' || c == '\t' {
						if tok_start != -1 && tok_last == -1 {
							tok_last = i - 1
						}
					} else {
						tok_start = i
						state = tok_kind_ident
					}
				}
			case tok_kind_comment:
			default:
				unreachable()
			}
		}
		if tok_last != -1 {
			if state == tok_kind_none || tok_start == -1 {
				unreachable()
			}
			{
				tok_len := (tok_last - tok_start) + 1
				toks[toks_count] = Token{
					kind:     state,
					line_nr:  cur_line_nr,
					line_idx: cur_line_idx,
					idx:      tok_start,
					len:      tok_len,
				}
				toks_count++
			}
			state = tok_kind_none
			tok_start = -1
			tok_last = -1
		}
		if c == '\n' {
			cur_line_nr++
			cur_line_idx = i + 1
		}
	}
	if tok_start != -1 {
		if state == tok_kind_none {
			unreachable()
		} else {
			tok_len := i - tok_start
			toks[toks_count] = Token{
				kind:     state,
				idx:      tok_start,
				len:      tok_len,
				line_nr:  cur_line_nr,
				line_idx: cur_line_idx,
			}
			toks_count++
		}
	}
	return toks[0:toks_count]
}

func tokPosCol(tok *Token) int { return tok.idx - tok.line_idx }

func tokIsOpening(tok_kind TokenKind) bool {
	return tok_kind == tok_kind_sep_bcurly_open || tok_kind == tok_kind_sep_bparen_open || tok_kind == tok_kind_sep_bsquare_open
}

func tokIsClosing(tok_kind TokenKind) bool {
	return tok_kind == tok_kind_sep_bcurly_close || tok_kind == tok_kind_sep_bparen_close || tok_kind == tok_kind_sep_bsquare_close
}

func toksCountUnnested(toks []Token, full_src Str, tok_kind TokenKind) int {
	assert(!(tokIsOpening(tok_kind) || tokIsClosing(tok_kind)))

	ret_num := 0
	level := 0
	for i := range toks {
		tok := &toks[i]
		if tok.kind == tok_kind && level == 0 {
			ret_num++
		} else if tokIsOpening(tok.kind) {
			level++
		} else if tokIsClosing(tok.kind) {
			level--
		}
	}
	return ret_num
}

func toksCheckBrackets(toks []Token) {
	level_bparen, level_bsquare, level_bcurly := 0, 0, 0
	for i := range toks {
		switch toks[i].kind {
		case tok_kind_sep_bcurly_open:
			level_bcurly++
		case tok_kind_sep_bparen_open:
			level_bparen++
		case tok_kind_sep_bsquare_open:
			level_bsquare++
		case tok_kind_sep_bcurly_close:
			level_bcurly--
		case tok_kind_sep_bparen_close:
			level_bparen--
		case tok_kind_sep_bsquare_close:
			level_bsquare--
		}
		if level_bparen < 0 {
			fail("surplus closing parenthesis in line ", uintToStr(uint64(1+toks[i].line_nr), 10, 1, nil))
		} else if level_bcurly < 0 {
			fail("surplus closing curly brace in line ", uintToStr(uint64(1+toks[i].line_nr), 10, 1, nil))
		} else if level_bsquare < 0 {
			fail("surplus closing square bracket in line ", uintToStr(uint64(1+toks[i].line_nr), 10, 1, nil))
		}
	}
	if level_bparen > 0 {
		fail("missing closing parenthesis")
	} else if level_bcurly > 0 {
		fail("missing closing curly brace")
	} else if level_bsquare > 0 {
		fail("missing closing square bracket")
	}
}

func toksIndentBasedChunks(toks []Token) [][]Token {
	cmp_pos_col := tokPosCol(&toks[0])
	level := 0
	for i := range toks {
		tok := &toks[i]
		if level == 0 {
			if pos_col := tokPosCol(tok); pos_col < cmp_pos_col {
				cmp_pos_col = pos_col
			}
		}
		if tokIsOpening(tok.kind) {
			level++
		} else if tokIsClosing(tok.kind) {
			level--
		}
	}
	assert(level == 0)

	var num_chunks int
	for i := range toks {
		if level == 0 {
			if i == 0 || tokPosCol(&toks[i]) <= cmp_pos_col {
				num_chunks++
			}
		}
		if tokIsOpening(toks[i].kind) {
			level++
		} else if tokIsClosing(toks[i].kind) {
			level--
		}
	}
	assert(level == 0)

	ret := allocˇTokens(num_chunks)
	{
		start_from, next_idx := -1, 0
		for i := range toks {
			tok := &toks[i]
			if i == 0 || (level == 0 && tokPosCol(tok) <= cmp_pos_col) {
				if start_from != -1 {
					ret[next_idx] = toks[start_from:i]
					next_idx++
				}
				start_from = i
			}
			if tokIsOpening(tok.kind) {
				level++
			} else if tokIsClosing(tok.kind) {
				level--
			}
		}
		if start_from != -1 {
			ret[next_idx] = toks[start_from:]
			next_idx++
		}
		assert(next_idx == num_chunks)
	}
	return ret
}

func toksIndexOfKind(toks []Token, kind TokenKind) int {
	for i := range toks {
		if toks[i].kind == kind {
			return i
		}
	}
	return -1
}

func toksIndexOfMatchingBracket(toks []Token) int {
	tok_open := toks[0].kind
	tok_close := tok_kind_none
	switch tok_open {
	case tok_kind_sep_bcurly_open:
		tok_close = tok_kind_sep_bcurly_close
	case tok_kind_sep_bparen_open:
		tok_close = tok_kind_sep_bparen_close
	case tok_kind_sep_bsquare_open:
		tok_close = tok_kind_sep_bsquare_close
	}
	assert(tok_close != tok_kind_none)
	level := 0
	for i := range toks {
		switch toks[i].kind {
		case tok_open:
			level++
		case tok_close:
			level--
			if level == 0 {
				return i
			}
		}
	}
	return -1
}

func toksSplit(toks []Token, full_src Str, tok_kind TokenKind) [][]Token {
	assert(!(tokIsOpening(tok_kind) || tokIsClosing(tok_kind)))

	ret_toks := allocˇTokens(1 + toksCountUnnested(toks, full_src, tok_kind))
	ret_idx := 0
	{
		level := 0
		start_from := 0
		for i := range toks {
			tok := &toks[i]
			if tok.kind == tok_kind && level == 0 {
				sub_toks := toks[start_from:i]
				if len(sub_toks) != 0 {
					ret_toks[ret_idx] = sub_toks
					ret_idx++
				}
				start_from = i + 1
			} else if tokIsOpening(tok.kind) {
				level++
			} else if tokIsClosing(tok.kind) {
				level--
			}
		}
		sub_toks := toks[start_from:]
		if len(sub_toks) != 0 {
			ret_toks[ret_idx] = sub_toks
			ret_idx++
		}
	}
	return ret_toks[0:ret_idx]
}

func toksSrcStr(toks []Token, full_src Str) Str {
	first, last := &toks[0], &toks[len(toks)-1]
	return full_src[first.idx : last.idx+last.len]
}
