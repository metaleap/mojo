package atmolang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
)

type AstFiles []*AstFile

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

type AstFileTopLevelChunk struct {
	Src     []byte
	SrcFile *AstFile
	Offset  struct {
		Line int
		Pos  int
	}
	id       [3]uint64
	_id      string
	_errs    []error
	srcDirty bool
	errs     struct {
		lexing  atmo.Errors
		parsing *atmo.Error
	}
	Ast AstTopLevel
}

func (me *AstFileTopLevelChunk) HasErrors() bool {
	return me.errs.parsing != nil || len(me.errs.lexing) > 0
}

func (me *AstFileTopLevelChunk) Errors() []error {
	if me._errs == nil {
		me._errs = make([]error, 0, len(me.errs.lexing)+1)
		for i := range me.errs.lexing {
			me._errs = append(me._errs, &me.errs.lexing[i])
		}
		if me.errs.parsing != nil {
			me._errs = append(me._errs, me.errs.parsing)
		}
	}
	return me._errs
}

func (me *AstFile) HasDefs(name string) bool {
	if name[0] == '_' {
		name = name[1:]
	}
	for i := range me.TopLevel {
		if tld := &me.TopLevel[i]; (!tld.HasErrors()) && tld.Ast.Def.Orig != nil {
			if tld.Ast.Def.Orig.Name.Val == name {
				return true
			}
		}
	}
	return false
}

func (me *AstFile) HasErrors() (r bool) {
	if r = me.errs.loading != nil; !r {
		for i := range me.TopLevel {
			if r = me.TopLevel[i].HasErrors(); r {
				break
			}
		}
	}
	return
}

func (me *AstFile) Errors() []error {
	if me._errs == nil {
		if me._errs = make([]error, 0, 4); me.errs.loading != nil {
			me._errs = append(me._errs, me.errs.loading)
		}
		for i := range me.TopLevel {
			me._errs = append(me._errs, me.TopLevel[i].Errors()...)
		}
	}
	return me._errs
}

func (me *AstFile) String() (r string) {
	for i := range me.TopLevel {
		if def := me.TopLevel[i].Ast.Def.Orig; def != nil {
			r += "\n" + def.Tokens.String() + "\n"
		}
	}
	return
}

func (me *AstFile) CountTopLevelDefs(onlyCountErrless bool) (total int, unexported int) {
	for i := range me.TopLevel {
		if tld := &me.TopLevel[i]; (!onlyCountErrless) || (!tld.HasErrors()) {
			if def := &tld.Ast.Def; def.Orig != nil {
				if total++; def.IsUnexported {
					unexported++
				}
			}
		}
	}
	return
}

func (me *AstFile) CountNetLinesOfCode(onlyCountErrless bool) (sloc int) {
	var lastline int

	for i := range me.TopLevel {
		if tld := &me.TopLevel[i]; (!onlyCountErrless) || (!tld.HasErrors()) {
			if def := tld.Ast.Def.Orig; def != nil {
				for t := range def.Tokens {
					if tok := &def.Tokens[t]; tok.Meta.Line != lastline && tok.Kind() != udevlex.TOKEN_COMMENT {
						lastline, sloc = tok.Meta.Line, sloc+1
					}
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

func (me AstFiles) Len() int          { return len(me) }
func (me AstFiles) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }
func (me AstFiles) Less(i int, j int) bool {
	if me[i].SrcFilePath == "" {
		return false
	}
	if me[j].SrcFilePath == "" {
		return true
	}
	return me[i].SrcFilePath < me[j].SrcFilePath
}

func (me AstFiles) Index(srcFilePath string) int {
	for i := range me {
		if me[i].SrcFilePath == srcFilePath {
			return i
		}
	}
	return -1
}

func (me *AstFiles) RemoveAt(idx int) {
	this := *me
	for i := idx; i < len(this)-1; i++ {
		this[i] = this[i+1]
	}
	this = this[:len(this)-1]
	*me = this
}
