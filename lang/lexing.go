package atmolang

import (
	"bytes"
	"io"
	"time"

	"github.com/go-leap/dev/lex"
	"github.com/go-leap/std"
)

type AstFileTopLevelChunk struct {
	Src    []byte
	Offset struct {
		Line int
		Pos  int
	}
	srcDirty bool
	errs     struct {
		lexing  []*udevlex.Error
		parsing *Error
	}
	Ast AstTopLevel
}

func (me *AstFile) LexAndParseSrc(r io.Reader) {
	var src []byte
	if src, me.errs.loading = ustd.ReadAll(r, me.LastLoad.Size); me.errs.loading == nil {
		if me.LastLoad.Size = int64(len(src)); bytes.Equal(src, me.LastLoad.Src) {
			return
		}
		me.LastLoad.Time, me.LastLoad.Src = time.Now().UnixNano(), src
		me.populateTopLevelChunksFrom(src)
		for i := range me.TopLevel {
			if this := &me.TopLevel[i]; this.srcDirty {
				this.errs.parsing, this.Ast.Comments, this.Ast.Def = nil, nil, nil
				this.Ast.Tokens, this.errs.lexing = udevlex.Lex(&ustd.BytesReader{Data: this.Src},
					me.SrcFilePath, this.Offset.Line, this.Offset.Pos, me.LastLoad.TokCountInitialGuess)
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

	// stage ONE: go over all src bytes and gather `tlchunks`

	var newline, istoplevelfulllinecomment, inmultilinecomment, inlinecomment bool
	var curline, lastpos, lastln int
	var chlast byte
	me.LastLoad.TokCountInitialGuess = 0
	allemptysofar, il := true, len(src)-1
	for i, ch := range src {
		if allemptysofar && !(ch == '\n' || ch == ' ') {
			allemptysofar, lastpos, lastln = false, i, curline
		}
		if inmultilinecomment {
			if chlast == '*' && ch == '/' {
				inmultilinecomment = false
			}
		} else if (!inlinecomment) && chlast == '/' && ch == '*' {
			inmultilinecomment = true
		} else if (!inlinecomment) && chlast == '/' && ch == '/' {
			inlinecomment = true
		}

		if ch == '\n' {
			inlinecomment, istoplevelfulllinecomment, newline, curline =
				false, false, true, curline+1
		} else if newline {
			if newline = false; (!inmultilinecomment) && ch != ' ' {
				isntlast := i < il
				istoplevelfulllinecomment = isntlast && ch == '/' && src[i+1] == '/'
				// naive at first: every non-indented line begins its own new chunk. after the loop we merge comments directly prefixed to defs
				if lastpos != i {
					tlchunks = append(tlchunks, _topLevelChunk{src: src[lastpos:i], pos: lastpos, line: lastln})
				}
				lastpos, lastln = i, curline
			}
		} else if (!(istoplevelfulllinecomment || inmultilinecomment || inlinecomment)) && ch == ' ' && chlast != ' ' {
			me.LastLoad.TokCountInitialGuess++
		}
		chlast = ch
	}
	if me.LastLoad.NumLines = curline; lastpos < il {
		tlchunks = append(tlchunks, _topLevelChunk{src: src[lastpos:], pos: lastpos, line: lastln})
	}
	// fix naive tlchunks: stitch together what belongs together
	for i := len(tlchunks) - 1; i > 0; i-- {
		if tlchunks[i-1].line == tlchunks[i].line-1 && // belong together?
			len(tlchunks[i-1].src) >= 2 && tlchunks[i-1].src[0] == '/' && tlchunks[i-1].src[1] == '/' {
			tlchunks[i-1].src = append(tlchunks[i-1].src, tlchunks[i].src...)
			for j := i; j < len(tlchunks)-1; j++ {
				tlchunks[j] = tlchunks[j+1]
			}
			tlchunks = tlchunks[0 : len(tlchunks)-1]
		}
	}

	// stage TWO: compare gathered `tlchunks` to existing `AstFileTopLevelChunk`s in `me.TopLevel`,
	// dropping those that are gone, adding those that are new, and repositioning others as needed

	srcsame := make(map[int]int, len(tlchunks))
	for oldidx := range me.TopLevel {
		for newidx := range tlchunks {
			if bytes.Equal(me.TopLevel[oldidx].Src, tlchunks[newidx].src) {
				srcsame[newidx] = oldidx
				break
			}
		}
	}
	allthesame := len(srcsame) == len(me.TopLevel) && len(me.TopLevel) == len(tlchunks)
	if allthesame {
		for newidx, oldidx := range srcsame {
			if newidx != oldidx {
				allthesame = false
				break
			}
		}
	}
	if !allthesame {
		oldtlc := me.TopLevel
		me._toks, me.TopLevel = nil, make([]AstFileTopLevelChunk, len(tlchunks))
		for i := range tlchunks {
			if oldidx, ok := srcsame[i]; ok {
				me.TopLevel[i] = oldtlc[oldidx]
			} else {
				me.TopLevel[i].Src = tlchunks[i].src
				me.TopLevel[i].srcDirty = true
			}
			me.TopLevel[i].Offset.Line, me.TopLevel[i].Offset.Pos = tlchunks[i].line, tlchunks[i].pos
		}
	}
}
