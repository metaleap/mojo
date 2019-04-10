package odlang

import (
	"bytes"
	"io"
	"os"
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

type AstFile struct {
	TopLevel []AstFileTopLevelChunk
	errs     struct {
		loading error
	}
	LastLoad struct {
		Src                  []byte
		SrcHashSum1          uint64
		SrcHashSum2          uint64
		Time                 int64
		Size                 int64
		tokCountInitialGuess int
	}

	Options struct {
		ApplStyle ApplStyle
	}
	SrcFilePath string

	_toks udevlex.Tokens
	_errs []error
}

func (me *AstFile) Errs() []error {
	if me._errs == nil {
		if me._errs = make([]error, 0); me.errs.loading != nil {
			me._errs = append(me._errs, me.errs.loading)
		}
		for i := range me.TopLevel {
			for _, e := range me.TopLevel[i].errs.lexing {
				me._errs = append(me._errs, e)
			}
			if e := me.TopLevel[i].errs.parsing; e != nil {
				me._errs = append(me._errs, e)
			}
		}
	}
	return me._errs
}

func (me *AstFile) Tokens() udevlex.Tokens {
	if me._toks == nil {
		me._toks = make(udevlex.Tokens, 0, len(me.TopLevel)*16)
		for i := range me.TopLevel {
			me._toks = append(me._toks, me.TopLevel[i].Ast.Tokens...)
		}
	}
	return me._toks
}

func (me *AstFile) LexAndParseFile(onlyIfModifiedSinceLastLoad bool, stdinIfNoSrcFilePathSet bool) {
	if me.SrcFilePath != "" {
		if srcfileinfo, _ := os.Stat(me.SrcFilePath); srcfileinfo != nil {
			if me.LastLoad.Size = srcfileinfo.Size(); onlyIfModifiedSinceLastLoad &&
				me.errs.loading == nil && me.LastLoad.Time > srcfileinfo.ModTime().UnixNano() {
				return
			}
		}
	}

	var srcfile *os.File
	if me._errs, me.errs.loading = nil, nil; me.SrcFilePath != "" {
		if srcfile, me.errs.loading = os.Open(me.SrcFilePath); me.errs.loading == nil {
			defer srcfile.Close()
		}
	} else if stdinIfNoSrcFilePathSet {
		srcfile = os.Stdin
	}
	if me.errs.loading == nil && srcfile != nil {
		me.LexAndParseSrc(srcfile)
	}
}

func (me *AstFile) LexAndParseSrc(r io.Reader) {
	var src []byte
	if src, me.errs.loading = ustd.ReadAll(r, me.LastLoad.Size); me.errs.loading == nil {
		me.LastLoad.Size = int64(len(src))

		var sameoldhashes bool
		if me.LastLoad.SrcHashSum1, me.LastLoad.SrcHashSum2, sameoldhashes, _ =
			ustd.HashWyPlus(me.LastLoad.SrcHashSum1, me.LastLoad.SrcHashSum2, src); sameoldhashes /* && bytes.Equal(src, me.LastLoad.Src) */ {
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
	type tlc struct {
		src  []byte
		pos  int
		line int
	}

	il, tlchunks := len(src)-1, make([]tlc, 0, 32)
	var newline, iscomment, wascomment bool
	var curline int
	var lastpos, lastln int
	var chlast byte
	if len(src) > 0 {
		chlast = src[0]
	}

	if me.LastLoad.tokCountInitialGuess = 0; chlast == '\n' {
		curline = 1
	}
	for i, l := 1, len(src); i < l; i++ {
		ch := src[i]
		if ch == '\n' {
			wascomment, iscomment, newline, curline, me.LastLoad.tokCountInitialGuess = iscomment, false, true, curline+1, me.LastLoad.tokCountInitialGuess+1
		} else if newline {
			if newline = false; ch != ' ' {
				iscomment = ch == '/' && i < il && src[i+1] == '/'
				if (!(iscomment && wascomment)) &&
					!(ch != '/' && src[lastpos] == '/' && (src[lastpos+1] == '/') && (i < 2 || src[i-2] != '\n')) {
					tlchunks = append(tlchunks, tlc{src: src[lastpos:i], pos: lastpos, line: lastln})
					lastpos, lastln = i, curline
				}
			}
		} else if (!iscomment) && ch == ' ' && chlast != ' ' && chlast != '\n' {
			me.LastLoad.tokCountInitialGuess++
		}
		chlast = ch
	}
	if lastpos < il {
		tlchunks = append(tlchunks, tlc{src: src[lastpos:], pos: lastpos, line: lastln})
	}

	unchanged := make(map[int]int, len(tlchunks))
	for o := range me.TopLevel {
		for n := range tlchunks {
			if bytes.Equal(me.TopLevel[o].src, tlchunks[n].src) {
				unchanged[n] = o
				break
			}
		}
	}
	sameasbefore := len(unchanged) == len(me.TopLevel) && len(me.TopLevel) == len(tlchunks)
	if sameasbefore {
		for n, o := range unchanged {
			if n != o {
				sameasbefore = false
				break
			}
		}
	}
	if !sameasbefore {
		tlcold := me.TopLevel
		me._toks, me.TopLevel = nil, make([]AstFileTopLevelChunk, len(tlchunks))
		for i := range tlchunks {
			if o, ok := unchanged[i]; ok {
				me.TopLevel[i] = tlcold[o]
			} else {
				me.TopLevel[i].src = tlchunks[i].src
				me.TopLevel[i].dirty = true
			}
			me.TopLevel[i].offset.line, me.TopLevel[i].offset.pos = tlchunks[i].line, tlchunks[i].pos
		}
	}
}
