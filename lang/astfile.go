package atmolang

import (
	"os"

	"github.com/go-leap/dev/lex"
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
		size                 int64
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
			if me.LastLoad.size = srcfileinfo.Size(); onlyIfModifiedSinceLastLoad && me.errs.loading == nil {
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

func (me AstFiles) Contains(srcFilePath string) bool {
	for i := range me {
		if me[i].SrcFilePath == srcFilePath {
			return true
		}
	}
	return false
}

func (me *AstFiles) RemoveAt(i int) { *me = append((*me)[:i], (*me)[i+1:]...) }
