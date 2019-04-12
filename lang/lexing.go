package atemlang

import (
	"bytes"
	"io"
	"time"

	"github.com/go-leap/dev/lex"
	"github.com/go-leap/std"
)

type AstFileTopLevelChunk struct {
	src    []byte
	offset struct {
		line int
		pos  int
	}
	dirty bool
	errs  struct {
		lexing  []*udevlex.Error
		parsing *Error
	}
	Ast AstTopLevel
}

func (me *AstFile) LexAndParseSrc(r io.Reader) {
	var src []byte
	if src, me.errs.loading = ustd.ReadAll(r, me.LastLoad.size); me.errs.loading == nil {
		if me.LastLoad.size = int64(len(src)); bytes.Equal(src, me.LastLoad.Src) {
			return
		}
		me.LastLoad.Time, me.LastLoad.Src = time.Now().UnixNano(), src
		me.populateTopLevelChunksFrom(src)
		for i := range me.TopLevel {
			if this := &me.TopLevel[i]; this.dirty {
				this.Ast.Tokens, this.errs.lexing = udevlex.Lex(&ustd.BytesReader{Data: this.src},
					me.SrcFilePath, this.offset.line, this.offset.pos, me.LastLoad.tokCountInitialGuess)
				if len(this.errs.lexing) == 0 {
					me.parse(this)
				}
			}
		}
	}
}

func (me *AstFile) populateTopLevelChunksFrom(src []byte) {
	type _topLevelChunk struct {
		src  []byte
		pos  int
		line int
	}
	tlchunks := make([]_topLevelChunk, 0, 32)

	// stage ONE: iterate all src bytes and gather `tlchunks`

	var newline, isfulllinecomment, wasfulllinecomment, inmultilinecomment bool
	var curline, lastpos, lastln int
	var chlast byte
	if len(src) > 0 {
		chlast = src[0]
	}
	if me.LastLoad.tokCountInitialGuess = 0; chlast == '\n' {
		curline = 1
	}
	il := len(src) - 1
	for i, l := 1, len(src); i < l; i++ {
		ch := src[i]
		if inmultilinecomment {
			if chlast == '*' && ch == '/' {
				inmultilinecomment = false
			}
		} else if (!isfulllinecomment) && chlast == '/' && ch == '*' {
			inmultilinecomment = true
		}

		if ch == '\n' {
			wasfulllinecomment, isfulllinecomment, newline, curline, me.LastLoad.tokCountInitialGuess = isfulllinecomment, false, true, curline+1, me.LastLoad.tokCountInitialGuess+1
		} else if newline {
			if newline = false; (!inmultilinecomment) && ch != ' ' {
				isntlast := i < il
				isfulllinecomment = ch == '/' && isntlast && src[i+1] == '/'
				if (!(isfulllinecomment && wasfulllinecomment)) &&
					!(ch != '/' && src[lastpos] == '/' && (src[lastpos+1] == '/') && (i < 2 || src[i-2] != '\n')) {
					tlchunks = append(tlchunks, _topLevelChunk{src: src[lastpos:i], pos: lastpos, line: lastln})
					lastpos, lastln = i, curline
				}
			}
		} else if (!isfulllinecomment) && (!inmultilinecomment) && ch == ' ' && chlast != ' ' && chlast != '\n' {
			me.LastLoad.tokCountInitialGuess++
		}
		chlast = ch
	}
	if lastpos < il {
		tlchunks = append(tlchunks, _topLevelChunk{src: src[lastpos:], pos: lastpos, line: lastln})
	}

	// stage TWO: compare `tlchunks` to existing `AstFileTopLevelChunk`s in `me.TopLevel`,
	// dropping those that are gone, adding those that are new, and marking those that have changed

	unchanged := make(map[int]int, len(tlchunks))
	for o := range me.TopLevel {
		for n := range tlchunks {
			if bytes.Equal(me.TopLevel[o].src, tlchunks[n].src) {
				unchanged[n] = o
				break
			}
		}
	}
	allthesame := len(unchanged) == len(me.TopLevel) && len(me.TopLevel) == len(tlchunks)
	if allthesame {
		for n, o := range unchanged {
			if n != o {
				allthesame = false
				break
			}
		}
	}
	if !allthesame {
		oldtlc := me.TopLevel
		me._toks, me.TopLevel = nil, make([]AstFileTopLevelChunk, len(tlchunks))
		for i := range tlchunks {
			if o, ok := unchanged[i]; ok {
				me.TopLevel[i] = oldtlc[o]
			} else {
				me.TopLevel[i].src = tlchunks[i].src
				me.TopLevel[i].dirty = true
			}
			me.TopLevel[i].offset.line, me.TopLevel[i].offset.pos = tlchunks[i].line, tlchunks[i].pos
		}
	}
}
