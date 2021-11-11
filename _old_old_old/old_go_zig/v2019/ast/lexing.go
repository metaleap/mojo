package atmoast

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/go-leap/dev/lex"
	"github.com/go-leap/std"
	. "github.com/metaleap/atmo/old/v2019"
)

func init() {
	udevlex.SepsGroupers, udevlex.SepsOthers, udevlex.RestrictedWhitespace, udevlex.ScannerStringDelims, udevlex.ScannerStringDelimNoEsc =
		"([{}])", ",", true, "\"'", '\''
}

func (me *AstFile) LexAndParseFile(onlyIfModifiedSinceLastLoad bool, stdinIfNoSrcFile bool, noChangesDetected *bool) (freshErrs Errors) {
	if me.Options.TmpAltSrc != nil || me.SrcFilePath == "" {
		me.LastLoad.Time, me.LastLoad.FileSize = 0, 0
	} else if me.SrcFilePath != "" {
		if srcfileinfo, _ := os.Stat(me.SrcFilePath); srcfileinfo != nil {
			if onlyIfModifiedSinceLastLoad && me.errs.loading == nil && srcfileinfo.Size() == me.LastLoad.FileSize {
				if me.LastLoad.Time > srcfileinfo.ModTime().UnixNano() {
					if noChangesDetected != nil {
						*noChangesDetected = true
					}
					return
				}
			}
			me.LastLoad.FileSize = srcfileinfo.Size()
		}
		me.LastLoad.Time = time.Now().UnixNano() // we havent even begun but for future modtime comparison this is indeed the proper moment
	}

	var reader io.Reader
	if me._errs, me.errs.loading = nil, nil; me.Options.TmpAltSrc != nil {
		reader = &ustd.BytesReader{Data: me.Options.TmpAltSrc}
	} else if me.SrcFilePath != "" {
		var srcfile *os.File
		var err error
		if srcfile, err = os.Open(me.SrcFilePath); err == nil {
			reader = srcfile
			defer srcfile.Close()
		} else {
			me.errs.loading = freshErrs.AddLex(ErrLexing_IoFileOpenFailure, ErrFauxPos(me.SrcFilePath), err.Error())
		}
	} else if stdinIfNoSrcFile {
		reader = os.Stdin
	}
	if me.errs.loading == nil && reader != nil {
		freshErrs = append(freshErrs, me.LexAndParseSrc(reader, noChangesDetected)...)
	}
	return
}

func LexAndGuess(fauxSrcFileNameForErrs string, src []byte) (guessIsDef bool, guessIsExpr bool, lexedToks udevlex.Tokens, err *Error) {
	if len(src) != 0 {
		indentguess := ' '
		if src[0] == '\t' {
			indentguess = '\t'
		} else {
			for i := 1; i < len(src); i++ {
				if src[i-1] == '\n' {
					if src[i] == ' ' {
						break
					} else if src[i] == '\t' {
						indentguess = '\t'
						break
					}
				}
			}
		}

		toks, errs := udevlex.Lex(src, fauxSrcFileNameForErrs, 32, indentguess)
		for _, e := range errs {
			return false, false, nil, ErrLex(ErrLexing_Tokenization, &e.Pos, e.Msg)
		}
		lexedToks = toks
		idxdecl, idxcomma := toks.Index(KnownIdentDecl, false), toks.Index(",", false)
		if idxdecl < 0 {
			guessIsExpr = true
		} else if idxcomma < 0 {
			guessIsDef = true
		} else if idxcomma < idxdecl {
			guessIsExpr = true
		} else {
			guessIsDef = true
		}
	}
	return
}

func LexAndParseDefOrExpr(def bool, toks udevlex.Tokens) (ret IAstNode, err *Error) {
	ctx := ctxTldParse{bracketsHalfIdx: len(udevlex.SepsGroupers) / 2}
	if def {
		ctx.curTopDef = &AstDef{IsTopLevel: true}
		if err = ctx.parseDef(toks, ctx.curTopDef); err == nil {
			ret = ctx.curTopDef
		}
	} else {
		ret, err = ctx.parseExpr(toks)
	}
	return
}

func (me *AstFile) LexAndParseSrc(r io.Reader, noChangesDetected *bool) (freshErrs Errors) {
	if src, err := ustd.ReadAll(r, me.LastLoad.FileSize); err != nil {
		me.errs.loading = freshErrs.AddLex(ErrLexing_IoFileReadFailure, ErrFauxPos(me.SrcFilePath), err.Error())
	} else {
		if bytes.Equal(src, me.LastLoad.Src) {
			if noChangesDetected != nil {
				*noChangesDetected = true
			}
			return
		}
		me.LastLoad.Src = src
		if me.topChunksCompare(me.topLevelChunksGatherFrom(src)) {
			if noChangesDetected != nil {
				*noChangesDetected = true
			}
			return
		}

		for i := range me.TopLevel {
			if this := &me.TopLevel[i]; this.srcDirty {
				this.srcDirty, this.errs.parsing, this.errs.lexing, this.Ast.Def.Orig, this.Ast.comments.Leading, this.Ast.comments.Trailing =
					false, nil, nil, nil, nil, nil

				indent, inderr := ' ', false
				if this.preLex.numLinesTabIndented != 0 {
					if this.preLex.numLinesSpaceIndented == 0 {
						indent = '\t'
					} else {
						inderr, freshErrs = true, append(freshErrs,
							this.errs.lexing.AddAt(ErrCatLexing, ErrLexing_IndentationInconsistent,
								&udevlex.Pos{FilePath: this.SrcFile.SrcFilePath, Col1: 1, Ln1: this.offset.Ln},
								len(this.Src),
								"inconsistent indentation in this top-level block: either replace leading tabs with spaces or vice-versa"))
					}
				}
				if !inderr {
					toks, errs := udevlex.Lex(this.Src, me.SrcFilePath, 64, indent)
					if this.Ast.Tokens = toks; len(errs) != 0 {
						for _, e := range errs {
							freshErrs = append(freshErrs,
								this.errs.lexing.AddLex(ErrLexing_Tokenization, &e.Pos, e.Msg))
						}
					} else {
						freshErrs.Add(me.parse(this)...)
					}
				}
			}
		}
	}
	return
}

func (me *AstFile) topLevelChunksGatherFrom(src []byte) (tlChunks []preLexTopLevelChunk) {
	tlChunks = make([]preLexTopLevelChunk, 0, 32)

	var newline bool
	var curline, lastchunkedat, lastchunkedln, numtabs, numspaces int
	var insth, insthn []byte
	allemptysofar, il := true, len(src)-1
	instr1, instr2, incml, incsl, instr1n := []byte("\""), []byte("'"), []byte("*/"), []byte("\n"), []byte("\\\"")
	for i, ch := range src {
		isnl := (ch == '\n')
		if isnl {
			curline++
		}
		if allemptysofar && !(isnl || ch == ' ' || ch == '\t') {
			allemptysofar, lastchunkedat, lastchunkedln = false, i, curline
		}

		if insth != nil {
			if i1 := i + 1; bytes.Equal(src[i-len(insth)+1:i1], insth) && (insthn == nil || !bytes.Equal(src[i-len(insthn)+1:i1], insthn)) {
				insth, insthn = nil, nil
			} else {
				continue
			}
		} else if ch == '"' {
			insth, insthn = instr1, instr1n
		} else if ch == '\'' {
			insth = instr2
		} else if ch == '/' && i < il {
			if chnext := src[i+1]; chnext == '*' {
				insth = incml
			} else if chnext == '/' {
				insth = incsl
			}
		}
		if ch == '\n' {
			newline = true
		} else if newline {
			newline = false
			if ch == '\t' {
				numtabs++
			} else if ch == ' ' {
				numspaces++
			} else {
				// naive at first: every non-indented line begins its own new chunk. after the loop we merge comments directly prefixed to defs
				if lastchunkedat != i {
					tlChunks = append(tlChunks, preLexTopLevelChunk{src: src[lastchunkedat:i], pos: lastchunkedat, line: lastchunkedln, numLinesSpaceIndented: numspaces, numLinesTabIndented: numtabs})
					numspaces, numtabs = 0, 0
				}
				lastchunkedat, lastchunkedln = i, curline
			}
		}
	}
	if me.LastLoad.NumLines = curline; lastchunkedat < il {
		tlChunks = append(tlChunks, preLexTopLevelChunk{src: src[lastchunkedat:], pos: lastchunkedat, line: lastchunkedln, numLinesSpaceIndented: numspaces, numLinesTabIndented: numtabs})
	}

	mergewithprior := func(i int) {
		tlChunks[i-1].numLinesSpaceIndented += tlChunks[i].numLinesSpaceIndented
		tlChunks[i-1].numLinesTabIndented += tlChunks[i].numLinesTabIndented
		tlChunks[i-1].src = append(tlChunks[i-1].src, tlChunks[i].src...)
		for j := i; j < len(tlChunks)-1; j++ {
			tlChunks[j] = tlChunks[j+1]
		}
		tlChunks = tlChunks[0 : len(tlChunks)-1]
	}

	// fix naive tlChunks: stitch multiple top-level full-line-comments (each a sep chunk for now) together
	for i := len(tlChunks) - 1; i > 0; i-- {
		if tlChunks[i-1].line == tlChunks[i].line-1 &&
			tlChunks[i-1].beginsWithLineComment() {
			mergewithprior(i)
		}
	}

	// fix naive tlChunks: chunks begun from top-level full-line comments with subsequent code line indented join the prior top-def
	for i := len(tlChunks) - 1; i > 0; i-- {
		if rest := tlChunks[i].fromFirstNonCommentLine(); len(rest) > 0 &&
			(rest[0] == ' ' || rest[0] == '\t') {
			mergewithprior(i)
		}
	}

	for newidx := range tlChunks { // trim-right \n bytes really helps with not reloading more than necessary
		var numn int
		for j := len(tlChunks[newidx].src) - 1; j > 0; j-- {
			if tlChunks[newidx].src[j] == '\n' {
				numn++
			} else {
				break
			}
		}
		tlChunks[newidx].src = tlChunks[newidx].src[:len(tlChunks[newidx].src)-numn]
	}
	return
}

func (me *preLexTopLevelChunk) beginsWithLineComment() bool {
	return len(me.src) >= 2 && me.src[0] == '/' && me.src[1] == '/'
}

func (me *preLexTopLevelChunk) fromFirstNonCommentLine() []byte {
	if !me.beginsWithLineComment() {
		return me.src
	}
	for l, i := len(me.src)-1, 3; i <= l; i++ {
		if me.src[i-1] == '\n' && me.src[i] != '/' && (i == l || me.src[i+1] != '/') {
			return me.src[i:]
		}
	}
	return nil
}

func (me *AstFile) topChunksCompare(tlChunks []preLexTopLevelChunk) (allTheSame bool) {
	// compare gathered `tlChunks` to existing `AstFileTopLevelChunk`s in `me.TopLevel`,
	// dropping those that are gone, adding those that are new, and repositioning others as needed
	srcsame := make(map[int]int, len(tlChunks))
	for oldidx := range me.TopLevel {
		for newidx := range tlChunks {
			if stale, fresh := &me.TopLevel[oldidx], &tlChunks[newidx]; bytes.Equal(stale.Src, fresh.src) {
				srcsame[newidx] = oldidx
				break
			}
		}
	}
	allTheSame = (len(srcsame) == len(me.TopLevel)) && (len(me.TopLevel) == len(tlChunks))
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
		me._toks, me.TopLevel = nil, make([]AstFileChunk, len(tlChunks))
		for newidx := range tlChunks {
			if oldidx, ok := srcsame[newidx]; ok {
				me.TopLevel[newidx] = oldtlc[oldidx]
			} else {
				me.TopLevel[newidx].srcDirty, me.TopLevel[newidx]._id, me.TopLevel[newidx]._errs, me.TopLevel[newidx].Src =
					true, "", nil, tlChunks[newidx].src
				me.TopLevel[newidx].id[1], me.TopLevel[newidx].id[2], _ = ustd.HashTwo(0, 0, me.TopLevel[newidx].Src)
				me.TopLevel[newidx].id[0] = uint64(len(me.TopLevel[newidx].Src))
			}
			me.TopLevel[newidx].SrcFile = me
		}
	}
	for newidx := range tlChunks {
		me.TopLevel[newidx].offset.Ln, me.TopLevel[newidx].offset.B, me.TopLevel[newidx].preLex.numLinesSpaceIndented, me.TopLevel[newidx].preLex.numLinesTabIndented =
			tlChunks[newidx].line, tlChunks[newidx].pos, tlChunks[newidx].numLinesSpaceIndented, tlChunks[newidx].numLinesTabIndented
	}

	return
}
