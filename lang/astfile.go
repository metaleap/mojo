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
	AstTopLevel
}

type AstFile struct {
	TopLevel []AstFileTopLevelChunk
	errs     struct {
		loading error
	}
	lastLoad struct {
		time                 int64
		size                 int64
		tokCountInitialGuess int
	}

	Options struct {
		ApplStyle ApplStyle
	}
	SrcFilePath string

	_src  udevlex.Tokens
	_errs []error
}

func (me *AstFile) populateChunksFrom(src []byte) {
	type tlc struct {
		src  []byte
		pos  int
		line int
	}
	il, tlchunks, chlast := len(src)-1, make([]tlc, 0, 32), src[0]

	var newline, iscomment, wascomment bool
	var curline int
	var lastpos, lastln int

	if me.lastLoad.tokCountInitialGuess = 0; chlast == '\n' {
		curline = 1
	}
	for i, l := 1, len(src); i < l; i++ {
		ch := src[i]
		if ch == '\n' {
			wascomment, iscomment, newline, curline, me.lastLoad.tokCountInitialGuess = iscomment, false, true, curline+1, me.lastLoad.tokCountInitialGuess+1
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
			me.lastLoad.tokCountInitialGuess++
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
		me._src, me.TopLevel = nil, make([]AstFileTopLevelChunk, len(tlchunks))
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

func (me *AstFile) Err() error {
	if me.errs.loading != nil {
		return me.errs.loading
	}
	for i := range me.TopLevel {
		for _, e := range me.TopLevel[i].errs.lexing {
			return e
		}
		if e := me.TopLevel[i].errs.parsing; e != nil {
			return e
		}
	}
	return nil
}

func (me *AstFile) Errs() []error {
	if me._errs == nil {
		if me._errs = make([]error, 0, 0); me.errs.loading != nil {
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

func (me *AstFile) Src() udevlex.Tokens {
	if me._src == nil {
		me._src = make(udevlex.Tokens, 0, len(me.TopLevel)*16)
		for i := range me.TopLevel {
			me._src = append(me._src, me.TopLevel[i].Tokens...)
			me._src = append(me._src, udevlex.Token{Meta: udevlex.TokenMeta{Orig: "...\n|\n|\n|\n|\n"}})
		}
	}
	return me._src
}

func (me *AstFile) LexAndParseFile(onlyIfModifiedSinceLastLoad bool, stdinIfNoSrcFilePathSet bool) {
	if me.SrcFilePath != "" {
		if srcfileinfo, _ := os.Stat(me.SrcFilePath); srcfileinfo != nil {
			if me.lastLoad.size = srcfileinfo.Size(); onlyIfModifiedSinceLastLoad &&
				me.errs.loading == nil && me.lastLoad.time > srcfileinfo.ModTime().UnixNano() {
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
	if src, me.errs.loading = ustd.ReadAll(r, me.lastLoad.size); me.errs.loading == nil {
		if me.lastLoad.time = time.Now().UnixNano(); len(src) > 0 {
			me.populateChunksFrom(src)
		}
		for i := range me.TopLevel {
			if this := &me.TopLevel[i]; this.dirty {
				this.Tokens, this.errs.lexing = udevlex.Lex(&ustd.BytesReader{Data: this.src},
					me.SrcFilePath, this.offset.line, this.offset.pos, me.lastLoad.tokCountInitialGuess)
				if len(this.errs.lexing) == 0 {
					me.parse(this)
				}
			}
		}
	}
}
