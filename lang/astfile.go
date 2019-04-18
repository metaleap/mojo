package atmolang

import (
	"os"

	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

type AstFiles []AstFile

type AstFile struct {
	TopLevel []AstFileTopLevelChunk
	errs     struct {
		loading error
	}
	LastLoad struct {
		Src                  []byte
		Time                 int64
		Size                 int64
		TokCountInitialGuess int
		NumLines             int
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
			if me.LastLoad.Size = srcfileinfo.Size(); onlyIfModifiedSinceLastLoad && me.errs.loading == nil {
				if modtime := srcfileinfo.ModTime().UnixNano(); modtime > 0 && me.LastLoad.Time > modtime {
					return
				}
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

func (me *AstFile) CountTopLevelDefs() (total int, unexported int) {
	for i := range me.TopLevel {
		if ast := &me.TopLevel[i].Ast; ast.Def != nil {
			if total++; ast.DefIsUnexported {
				unexported++
			}
		}
	}
	return
}

func (me *AstFile) CountNetLinesOfCode() (sloc int) {
	var lastline int

	for i := range me.TopLevel {
		if def := me.TopLevel[i].Ast.Def; def != nil {
			for t := range def.Tokens {
				if tok := &def.Tokens[t]; tok.Meta.Line != lastline && tok.Kind() != udevlex.TOKEN_COMMENT {
					lastline, sloc = tok.Meta.Line, sloc+1
				}
			}
		}
	}
	return
}

func (me *AstFileTopLevelChunk) ID() string {
	if me._id == "" {
		me._id = ustr.Uint64s('-', me.id[:])
	}
	return me._id
}

func (me AstFiles) Index(srcFilePath string) int {
	for i := range me {
		if me[i].SrcFilePath == srcFilePath {
			return i
		}
	}
	return -1
}

func (me *AstFiles) RemoveAt(i int) { *me = append((*me)[:i], (*me)[i+1:]...) }
