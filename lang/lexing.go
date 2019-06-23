package atmolang

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/go-leap/dev/lex"
	"github.com/go-leap/std"
	"github.com/metaleap/atmo"
)

func (me *AstFile) LexAndParseFile(onlyIfModifiedSinceLastLoad bool, stdinIfNoSrcFile bool, noChangesDetected *bool) (freshErrs []error) {
	if me.Options.TmpAltSrc != nil {
		me.LastLoad.Time, me.LastLoad.FileSize = 0, 0
	} else if me.SrcFilePath != "" {
		if srcfileinfo, _ := os.Stat(me.SrcFilePath); srcfileinfo != nil {
			if onlyIfModifiedSinceLastLoad && me.errs.loading == nil && srcfileinfo.Size() == me.LastLoad.FileSize {
				if modtime := srcfileinfo.ModTime().UnixNano(); modtime > 0 && me.LastLoad.Time > modtime {
					if noChangesDetected != nil {
						*noChangesDetected = true
					}
					return
				}
			}
			me.LastLoad.FileSize = srcfileinfo.Size()
		}
	}

	var reader io.Reader
	if me._errs, me.errs.loading = nil, nil; me.Options.TmpAltSrc != nil {
		reader = &ustd.BytesReader{Data: me.Options.TmpAltSrc}
	} else if me.SrcFilePath != "" {
		var srcfile *os.File
		if srcfile, me.errs.loading = os.Open(me.SrcFilePath); me.errs.loading == nil {
			reader = srcfile
			defer srcfile.Close()
		} else {
			freshErrs = append(freshErrs, me.errs.loading)
		}
	} else if stdinIfNoSrcFile {
		reader = os.Stdin
	}
	if me.errs.loading == nil && reader != nil {
		freshErrs = append(freshErrs, me.LexAndParseSrc(reader, noChangesDetected)...)
	}
	return
}

func LexAndParseExpr(fauxSrcFileNameForErrs string, src []byte) (IAstExpr, *atmo.Error) {
	toks, errs := udevlex.Lex(&ustd.BytesReader{Data: src}, fauxSrcFileNameForErrs, 64)
	for _, e := range errs {
		return nil, atmo.ErrLex(&e.Pos, e.Msg)
	}

	if expr, err := (&ctxTldParse{}).parseExpr(toks); err != nil { // need this..
		return nil, err // ..explicit branch because else a `nil`..
	} else { // ..`*Error` would turn into a non-nil `error`..
		return expr, nil // ..interface with inner `nil` value *sigh!*
	}
}

func (me *AstFile) LexAndParseSrc(r io.Reader, noChangesDetected *bool) (freshErrs []error) {
	var src []byte
	if src, me.errs.loading = ustd.ReadAll(r, me.LastLoad.FileSize); me.errs.loading != nil {
		freshErrs = append(freshErrs, me.errs.loading)
	} else {
		if bytes.Equal(src, me.LastLoad.Src) {
			if noChangesDetected != nil {
				*noChangesDetected = true
			}
			return
		}
		me.LastLoad.Time, me.LastLoad.Src = time.Now().UnixNano(), src
		if me.populateTopLevelChunksFrom(src) {
			if noChangesDetected != nil {
				*noChangesDetected = true
			}
			return
		}

		for i := range me.TopLevel {
			if this := &me.TopLevel[i]; this.srcDirty {
				this.srcDirty, this.errs.parsing, this.errs.lexing, this.Ast.Def.Orig, this.Ast.comments.Leading, this.Ast.comments.Trailing =
					false, nil, nil, nil, nil, nil
				toks, errs := udevlex.Lex(&ustd.BytesReader{Data: this.Src},
					me.SrcFilePath, 64)
				if this.Ast.Tokens = toks; len(errs) > 0 {
					for _, e := range errs {
						this.errs.lexing.AddLex(&e.Pos, e.Msg)
						freshErrs = append(freshErrs, e)
					}
				} else {
					freshErrs = append(freshErrs, me.parse(this)...)
				}
			}
		}
	}
	return
}

type srcTopLevelChunk struct {
	src   []byte
	offb0 int
	line0 int
}

func (me *AstFile) topLevelChunksFrom(src []byte) (tlChunks []srcTopLevelChunk) {
	// isnewline := true
	// var last byte
	// for i, cur := range src {

	// }

	return
}

func (me *AstFile) populateTopLevelChunksFrom(src []byte) (allTheSame bool) {
	type topLevelChunk struct {
		src  []byte
		pos  int
		line int
	}
	tlchunks := make([]topLevelChunk, 0, 32)

	// stage ONE: go over all src bytes and gather `tlchunks`

	var newline, inmultilinecomment, inlinecomment bool
	var curline, lastpos, lastln int
	var chlast byte
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
			inlinecomment, newline, curline = false, true, curline+1
		} else if newline {
			if newline = false; (!inmultilinecomment) && ch != ' ' {
				// naive at first: every non-indented line begins its own new chunk. after the loop we merge comments directly prefixed to defs
				if lastpos != i {
					tlchunks = append(tlchunks, topLevelChunk{src: src[lastpos:i], pos: lastpos, line: lastln})
				}
				lastpos, lastln = i, curline
			}
		}
		chlast = ch
	}
	if me.LastLoad.NumLines = curline; lastpos < il {
		tlchunks = append(tlchunks, topLevelChunk{src: src[lastpos:], pos: lastpos, line: lastln})
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

	for newidx := range tlchunks { // trim-right \n bytes really helps with not reloading more than necessary
		var numn int
		for j := len(tlchunks[newidx].src) - 1; j > 0; j-- {
			if tlchunks[newidx].src[j] == '\n' {
				numn++
			} else {
				break
			}
		}
		tlchunks[newidx].src = tlchunks[newidx].src[:len(tlchunks[newidx].src)-numn]
	}

	// stage TWO: compare gathered `tlchunks` to existing `AstFileTopLevelChunk`s in `me.TopLevel`,
	// dropping those that are gone, adding those that are new, and repositioning others as needed

	srcsame := make(map[int]int, len(tlchunks))
	for oldidx := range me.TopLevel {
		for newidx := range tlchunks {
			if stale, fresh := &me.TopLevel[oldidx], &tlchunks[newidx]; bytes.Equal(stale.Src, fresh.src) {
				srcsame[newidx] = oldidx
				break
			}
		}
	}
	allTheSame = len(srcsame) == len(me.TopLevel) && len(me.TopLevel) == len(tlchunks)
	if allTheSame {
		for newidx, oldidx := range srcsame {
			if newidx != oldidx {
				allTheSame = false
				break
			}
		}
	}
	if !allTheSame {
		oldtlc := me.TopLevel
		me._toks, me.TopLevel = nil, make([]SrcTopChunk, len(tlchunks))
		for newidx := range tlchunks {
			if oldidx, ok := srcsame[newidx]; ok {
				me.TopLevel[newidx] = oldtlc[oldidx]
			} else {
				me.TopLevel[newidx].srcDirty, me.TopLevel[newidx]._id, me.TopLevel[newidx]._errs, me.TopLevel[newidx].Src =
					true, "", nil, tlchunks[newidx].src
				me.TopLevel[newidx].id[1], me.TopLevel[newidx].id[2], _ = ustd.HashTwo(0, 0, me.TopLevel[newidx].Src)
				me.TopLevel[newidx].id[0] = uint64(len(me.TopLevel[newidx].Src))
			}
			me.TopLevel[newidx].SrcFile = me
		}
	}
	for newidx := range tlchunks {
		me.TopLevel[newidx].offset.Ln, me.TopLevel[newidx].offset.B =
			tlchunks[newidx].line, tlchunks[newidx].pos
	}
	return
}
