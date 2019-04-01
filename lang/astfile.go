package odlang

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-leap/dev/lex"
)

type AstFileTopLevelChunk struct {
	src    []byte
	toks   udevlex.Tokens
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
	lastLoadTime int64

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
	tlchunks := make([]tlc, 0, 16)

	var newline, iscomment, wascomment bool
	var curline int
	var lastpos, lastln int
	for i := range src {
		if src[i] == '\n' {
			wascomment, iscomment, newline, curline = iscomment, false, true, curline+1
		} else if i > 0 {
			if newline {
				if src[i] != ' ' {
					iscomment = src[i] == '/' && i < (len(src)-1) && src[i+1] == '/'
					if (!(iscomment && wascomment)) &&
						!(src[i] != '/' && src[lastpos] == '/' && (src[lastpos+1] == '/') && (i < 2 || src[i-2] != '\n')) {
						tlchunks = append(tlchunks, tlc{src: src[lastpos:i], pos: lastpos, line: lastln})
						lastpos, lastln = i, curline
					}
				}
				newline = false
			}
		}
	}
	if lastpos < (len(src) - 1) {
		tlchunks = append(tlchunks, tlc{src: src[lastpos:], pos: lastpos, line: lastln})
	}

	unchanged := make(map[int]int, 8)
	for o := range me.TopLevel {
		for n := range tlchunks {
			if bytes.Equal(me.TopLevel[o].src, tlchunks[n].src) {
				unchanged[n] = o
				break
			}
		}
	}
	sameasbefore := len(unchanged) == len(me.TopLevel) && len(unchanged) == len(tlchunks)
	if sameasbefore {
		for k, v := range unchanged {
			if k != v {
				sameasbefore = false
				break
			}
		}
	}
	if !sameasbefore {
		tlcold := me.TopLevel[:]
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
			me._src = append(me._src, me.TopLevel[i].toks...)
			me._src = append(me._src, udevlex.Token{Meta: udevlex.TokenMeta{Orig: "...\n|\n|\n|\n|\n"}})
		}
	}
	return me._src
}

func (me *AstFile) LexAndParseFile(onlyIfModifiedSinceLastLoad bool, stdinIfNoSrcFilePathSet bool) {
	if me.SrcFilePath != "" && onlyIfModifiedSinceLastLoad && me.errs.loading == nil {
		if file, e := os.Stat(me.SrcFilePath); e == nil && me.lastLoadTime > file.ModTime().UnixNano() {
			return
		}
	}

	var src *os.File
	if me._errs, me.errs.loading = nil, nil; me.SrcFilePath != "" {
		if src, me.errs.loading = os.Open(me.SrcFilePath); me.errs.loading == nil {
			defer src.Close()
		}
	} else if stdinIfNoSrcFilePathSet {
		src = os.Stdin
	}
	if me.errs.loading == nil && src != nil {
		me.LexAndParseSrc(src)
	}
}

func (me *AstFile) LexAndParseSrc(r io.Reader) {
	var src []byte
	if src, me.errs.loading = ioutil.ReadAll(r); me.errs.loading == nil {
		me.lastLoadTime = time.Now().UnixNano()
		me.populateChunksFrom(src)
		for i := range me.TopLevel {
			if this := &me.TopLevel[i]; this.dirty {
				this.toks, this.errs.lexing =
					udevlex.Lex(me.SrcFilePath, bytes.NewReader(this.src), this.offset.line, this.offset.pos, len(this.src)/6)
				if len(this.errs.lexing) == 0 {
					me.parse(this)
				}
			}
		}
	}
}
