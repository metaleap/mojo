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
	astNode
	topLevelChunks []astFileTokenChunk

	SrcFilePath string
	Errors      struct {
		Loading error
	}
	Nodes []IAstNode
}

func (me *AstFile) populateChunksFrom(src []byte) {
	type tlc struct {
		src    []byte
		offPos int
		offLn  int
	}
	tlchunks := make([]tlc, 0, 16)

	var newline bool
	var curline int
	var startfrom int
	for i := range src {
		if src[i] == '\n' {
			newline, curline = true, curline+1
		} else if i > 0 {
			if newline && i > 1 && src[i] != ' ' && src[i-2] == '\n' {
				tlchunks = append(tlchunks, tlc{src: src[startfrom:i]})
				startfrom = i
			}
			newline = false
		}
	}
	if startfrom < (len(src) - 1) {
		tlchunks = append(tlchunks, tlc{src: src[startfrom:]})
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
	tlcold := me.topLevelChunks[:]
	me.topLevelChunks = make([]astFileTokenChunk, len(tlchunks))
	for i := range tlchunks {
		if o, ok := unchanged[i]; ok {
			me.topLevelChunks[i] = tlcold[o]
		} else {
			me.topLevelChunks[i].src = tlchunks[i].src
			me.topLevelChunks[i].dirty = true
		}
		me.topLevelChunks[i].offset.line, me.topLevelChunks[i].offset.pos = tlchunks[i].offLn, tlchunks[i].offPos
	}
}

func (me *AstFile) errsReset() {
	me.Errors.Loading = nil
}

func (me *AstFile) Err() error {
	if me.Errors.Loading != nil {
		return me.Errors.Loading
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
