package atmoast

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
	. "github.com/metaleap/atmo/old/v2019"
)

func (me *AstFileChunk) At(byte0PosOffsetInSrcFile int) []IAstNode {
	return me.Ast.at(&me.Ast, byte0PosOffsetInSrcFile-me.offset.B)
}

func (me *AstFileChunk) Encloses(byte0PosOffsetInSrcFile int) bool {
	return me.Ast.Tokens.AreEnclosing(byte0PosOffsetInSrcFile - me.offset.B)
}

// PosOffsetLine implements `atmo.IErrPosOffsets`.
func (me *AstFileChunk) PosOffsetLine() int { return me.offset.Ln }

// PosOffsetByte implements `atmo.IErrPosOffsets`.
func (me *AstFileChunk) PosOffsetByte() int { return me.offset.B }

func (me *AstFileChunk) Errs() Errors {
	if me._errs == nil {
		me._errs = make(Errors, 0, len(me.errs.lexing)+1)
		if me._errs.Add(me.errs.lexing...); me.errs.parsing != nil {
			me._errs.Add(me.errs.parsing)
		}
	}
	return me._errs
}

func (me *AstFileChunk) HasErrors() bool {
	return me.errs.parsing != nil || len(me.errs.lexing) != 0
}

func (me *AstFile) HasDefs(name string, includeUnparsed bool) bool {
	if name[0] == '_' {
		name = name[1:]
	}
	for i := range me.TopLevel {
		if tld := &me.TopLevel[i]; (!tld.HasErrors()) && tld.Ast.Def.Orig != nil {
			if tld.Ast.Def.Orig.Name.Val == name || (includeUnparsed && tld.Ast.Def.NameIfErr == name) {
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

func (me *AstFile) Errors() Errors {
	if me._errs == nil {
		if me._errs = make(Errors, 0, 4); me.errs.loading != nil {
			me._errs = append(me._errs, me.errs.loading)
		}
		for i := range me.TopLevel {
			me._errs.Add(me.TopLevel[i].Errs()...)
		}
	}
	return me._errs
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
					if tok := &def.Tokens[t]; tok.Pos.Ln1 != lastline && tok.Kind != udevlex.TOKEN_COMMENT {
						lastline, sloc = tok.Pos.Ln1, sloc+1
					}
				}
			}
		}
	}
	return
}

func (me *AstFile) TopLevelChunkAt(pos0ByteOffset int) *AstFileChunk {
	ilast := len(me.TopLevel) - 1
	for i := range me.TopLevel {
		if pos0ByteOffset == me.TopLevel[i].offset.B || (i == ilast && pos0ByteOffset > me.TopLevel[i].offset.B) {
			return &me.TopLevel[i]
		} else if me.TopLevel[i].offset.B > pos0ByteOffset {
			if i == 0 {
				return &me.TopLevel[0]
			} else {
				return &me.TopLevel[i-1]
			}
		}
	}
	return nil
}

func (me *AstFileChunk) Id() string {
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

func (me AstFiles) ByFilePath(srcFilePath string) *AstFile {
	for _, f := range me {
		if f.SrcFilePath == srcFilePath {
			return f
		}
	}
	return nil
}

func (me AstFiles) TopLevelChunkByDefId(defId string) *AstFileChunk {
	for _, f := range me {
		for i := range f.TopLevel {
			if tlc := &f.TopLevel[i]; tlc.Id() == defId {
				return tlc
			}
		}
	}
	return nil
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
