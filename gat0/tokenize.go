package main

import (
	"fmt"
	"strings"
)

var tokenizer = Tokenizer{
	strLitDelimChars:  "'\"`",
	sepChars:          ",:;.(){}[]",
	opChars:           "^!$%&/=\\?*+~-<>|@",
	lineCommentPrefix: "//",
}

type Tokenizer struct {
	strLitDelimChars  string
	braceChars        string
	sepChars          string
	opChars           string
	lineCommentPrefix string
}

type Tokens []Token

type Token struct {
	kind       TokenKind
	idx        int
	idxLnStart int
	lnNr0      int
	src        string
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

func (me *Tokenizer) tokenize(src string) (toks Tokens) {
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
		case strings.HasPrefix(src[idx:], me.lineCommentPrefix): // start of comment?
			tokdone(idx - 1)
			cur.idx, cur.idxLnStart, cur.lnNr0, cur.kind = idx, idxln, lnnr, tokKindComment
		case c >= '0' && c <= '9': // start of number?
			if cur.kind != tokKindIdentName {
				tokdone(idx - 1)
				cur.idx, cur.idxLnStart, cur.lnNr0, cur.kind = idx, idxln, lnnr, tokKindNumLit
			}
		case strings.IndexByte(me.strLitDelimChars, c) >= 0: // start of string?
			tokdone(idx - 1)
			cur.idx, cur.idxLnStart, cur.lnNr0, cur.kind = idx, idxln, lnnr, tokKindStrLit
		case strings.IndexByte(me.sepChars, c) >= 0: // a sep?
			tokdone(idx - 1)
			cur.idx, cur.idxLnStart, cur.lnNr0, cur.kind = idx, idxln, lnnr, tokKindSep
			tokdone(idx)
		case strings.IndexByte(me.opChars, c) >= 0: // start of opish ident?
			tokdone(idx - 1)
			cur.idx, cur.idxLnStart, cur.lnNr0, cur.kind = idx, idxln, lnnr, tokKindIdentOp
		case c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\v' || c == '\b':
			tokdone(idx - 1)
		default:
			if cur.kind == 0 {
				cur.idx, cur.idxLnStart, cur.lnNr0, cur.kind = idx, idxln, lnnr, tokKindIdentName
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

func (me *Token) col0() int {
	return me.idx - me.idxLnStart
}

func (me *Token) String() string {
	return fmt.Sprintf("L%d C%d '%s'>>>>%s<<<<", me.lnNr0+1, me.col0()+1, me.kind.String(), me.src)
}

func (me Tokens) String(origSrc string) string {
	last := &me[len(me)-1]
	return origSrc[me[0].idx : last.idx+len(last.src)]
}

func (me Tokens) indentLevelChunks(toks Tokens, col0 int) (ret []Tokens) {
	var idxchunk int
	for i := range toks {
		if i > 0 && toks[i].col0() == col0 && toks[i].kind != tokKindComment {
			ret = append(ret, toks[idxchunk:i])
			idxchunk = i
		}
	}
	if tail := toks[idxchunk:]; len(tail) > 0 {
		ret = append(ret, tail)
	}
	return
}
