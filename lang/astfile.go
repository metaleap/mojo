package odlang

import (
	"bytes"

	"github.com/go-leap/dev/lex"
)

type astFileTokenChunk struct {
	src    []byte
	toks   udevlex.Tokens
	node   IAstNode
	dirty  bool
	offset struct {
		line int
		pos  int
	}
	errs struct {
		lexing  []*udevlex.Error
		parsing []*Error
	}
}

type AstFile struct {
	topLevelChunks []astFileTokenChunk
	errs           struct {
		loading error
	}

	SrcFilePath string
	Nodes       []IAstNode

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

	var newline bool
	var curline int
	var lastpos, lastln int
	for i := range src {
		if src[i] == '\n' {
			newline, curline = true, curline+1
		} else if i > 0 {
			if newline && i > 1 && src[i] != ' ' && src[i-2] == '\n' {
				tlchunks = append(tlchunks, tlc{src: src[lastpos:i], pos: lastpos, line: lastln})
				lastpos, lastln = i, curline
			}
			newline = false
		}
	}
	if lastpos < (len(src) - 1) {
		tlchunks = append(tlchunks, tlc{src: src[lastpos:], pos: lastpos, line: lastln})
	}

	unchanged := make(map[int]int, 8)
	for o := range me.topLevelChunks {
		for n := range tlchunks {
			if bytes.Equal(me.topLevelChunks[o].src, tlchunks[n].src) {
				unchanged[n] = o
				break
			}
		}
	}
	sameasbefore := len(unchanged) == len(me.topLevelChunks) && len(unchanged) == len(tlchunks)
	if sameasbefore {
		for k, v := range unchanged {
			if k != v {
				sameasbefore = false
				break
			}
		}
	}
	if !sameasbefore {
		tlcold := me.topLevelChunks[:]
		me._src, me.topLevelChunks = nil, make([]astFileTokenChunk, len(tlchunks))
		for i := range tlchunks {
			if o, ok := unchanged[i]; ok {
				me.topLevelChunks[i] = tlcold[o]
			} else {
				me.topLevelChunks[i].src = tlchunks[i].src
				me.topLevelChunks[i].dirty = true
			}
			me.topLevelChunks[i].offset.line, me.topLevelChunks[i].offset.pos = tlchunks[i].line, tlchunks[i].pos
		}
	}
}

func (me *AstFile) Err() error {
	if me.errs.loading != nil {
		return me.errs.loading
	}
	for i := range me.topLevelChunks {
		for _, e := range me.topLevelChunks[i].errs.lexing {
			return e
		}
		for _, e := range me.topLevelChunks[i].errs.parsing {
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
		for i := range me.topLevelChunks {
			for _, e := range me.topLevelChunks[i].errs.lexing {
				me._errs = append(me._errs, e)
			}
			for _, e := range me.topLevelChunks[i].errs.parsing {
				me._errs = append(me._errs, e)
			}
		}
	}
	return me._errs
}

func (me *AstFile) Src() udevlex.Tokens {
	if me._src == nil {
		me._src = make(udevlex.Tokens, 0, len(me.topLevelChunks)*16)
		for i := range me.topLevelChunks {
			me._src = append(me._src, me.topLevelChunks[i].toks...)
		}
	}
	return me._src
}
