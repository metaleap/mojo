package main

import (
	"fmt"
	"strings"
)

var tokenizer = Tokenizer{
	strLitDelimChars: "'\"`",
	sepChars:         ",:;.(){}[]",
	opChars:          "^!$%&/=\\?*+~-<>|",
}

type Tokenizer struct {
	strLitDelimChars string
	braceChars       string
	sepChars         string
	opChars          string
}

type Token struct {
	kind  TokenKind
	idx   int
	idxLn int
	lnNr  int
	src   string
}
type TokenKind int

const (
	_ TokenKind = iota
	tokKindIdentName
	tokKindIdentOp
	tokKindSep
	tokKindStrLit
	tokKindNumLit
	tokKindComment
)

func (me *Tokenizer) tokenize(src string) (toks []Token) {
	var cur Token
	var idxln, lnnr int
	tokdone := func(idxLastChar int) {
		if cur.kind != 0 && idxLastChar >= cur.idx {
			cur.src = src[cur.idx : idxLastChar+1]
			toks = append(toks, cur)
		}
		cur = Token{}
	}
	for idx := 0; idx < len(src); idx++ {
		c := src[idx]
		switch {
		case cur.kind == tokKindComment: // already in comment?
			if c == '\n' {
				tokdone(idx - 1)
			}
		case cur.kind == tokKindStrLit: // already in string?
			if c == src[cur.idx] && src[idx-1] != '\\' {
				tokdone(idx)
			}
		case cur.kind == tokKindNumLit: // already in number?
			if !(c >= '0' && c <= '9') {
				tokdone(idx - 1)
				idx-- // revisit cur char
				continue
			}
		case cur.kind == tokKindIdentOp: // already in opish ident?
			if strings.IndexByte(me.opChars, c) < 0 {
				tokdone(idx - 1)
				idx-- // revisit cur char
				continue
			}
		case c >= '0' && c <= '9': // start of number?
			tokdone(idx - 1)
			cur.idx, cur.idxLn, cur.lnNr, cur.kind = idx, idxln, lnnr, tokKindNumLit
		case strings.IndexByte(me.strLitDelimChars, c) >= 0: // start of string?
			tokdone(idx - 1)
			cur.idx, cur.idxLn, cur.lnNr, cur.kind = idx, idxln, lnnr, tokKindStrLit
		case strings.IndexByte(me.sepChars, c) >= 0: // a sep?
			tokdone(idx - 1)
			cur.idx, cur.idxLn, cur.lnNr, cur.kind = idx, idxln, lnnr, tokKindSep
			tokdone(idx)
		case strings.IndexByte(me.opChars, c) >= 0: // start of opish ident?
			tokdone(idx - 1)
			cur.idx, cur.idxLn, cur.lnNr, cur.kind = idx, idxln, lnnr, tokKindIdentOp
		case c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\v' || c == '\b':
			tokdone(idx - 1)
		default:
			if cur.kind == 0 {
				cur.idx, cur.idxLn, cur.lnNr, cur.kind = idx, idxln, lnnr, tokKindIdentName
			}
		}
		if c == '\n' {
			idxln, lnnr = idx+1, lnnr+1
		}
	}
	tokdone(len(src) - 1)
	return
}

func (me TokenKind) String() string {
	switch me {
	case tokKindIdentName:
		return "IdentName"
	case tokKindIdentOp:
		return "IdentOp"
	case tokKindSep:
		return "Sep"
	case tokKindStrLit:
		return "StrLit"
	case tokKindNumLit:
		return "NumLit"
	case tokKindComment:
		return "Comment"
	}
	return ""
}

func (me *Token) String() string {
	return fmt.Sprintf("L%d C%d '%s'>>>>%s<<<<", me.lnNr+1, (me.idx-me.idxLn)+1, me.kind.String(), me.src)
}
