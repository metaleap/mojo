package main

import (
	"strings"
)

var tokenizer = Tokenizer{
	strLitDelimChars: "'\"`",
	sepChars:         ",:;.(){}[]",
}

type Tokenizer struct {
	strLitDelimChars string
	braceChars       string
	sepChars         string
	// opChars          string
}

type Token struct {
	kind TokenKind
	idx  int
	src  string
}
type TokenKind int

const (
	_ TokenKind = iota
	tokKindIdent
	tokKindSep
	tokKindStrLit
	tokKindNumLit
	tokKindComment
)

func (me *Tokenizer) tokenize(src string) (toks []Token) {
	var cur Token
	tokdone := func(idxLastChar int) {
		if cur.kind != 0 {
			cur.src = src[cur.idx : idxLastChar+1]
			toks = append(toks, cur)
		}
		cur = Token{}
	}
	for idx := 0; idx < len(src); idx++ {
		c := src[idx]
		switch {
		case cur.kind == tokKindStrLit:
			if c == src[cur.idx] && src[idx-1] != '\\' {
				tokdone(idx)
			}
		case cur.kind == tokKindNumLit:
			if !(c >= '0' && c <= '9') {
				tokdone(idx - 1)
				idx-- // fall-through kinda
			}
		case cur.kind == tokKindComment:
			if c == '\n' {
				tokdone(idx - 1)
			}
		case strings.IndexByte(me.strLitDelimChars, c) >= 0:
			tokdone(idx - 1)
			cur.idx, cur.kind = idx, tokKindStrLit
		case strings.IndexByte(me.sepChars, c) >= 0:
			tokdone(idx - 1)
			cur.idx, cur.kind = idx, tokKindSep
			tokdone(idx)
		case c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\v' || c == '\b':
			tokdone(idx - 1)
		default:
		}
	}
	tokdone(len(src) - 1)
	return
}
