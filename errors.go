package atmo

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

type IErrPosOffsets interface {
	PosOffsetLine() int
	PosOffsetByte() int
}

type ErrorCategory int

const (
	_ ErrorCategory = iota
	ErrCatBug
	ErrCatTodo
	ErrCatLexing
	ErrCatParsing
	ErrCatNaming
	ErrCatPreduce
	ErrCatSess
	ErrCatUnreachable
)

func (me ErrorCategory) String() string {
	switch me {
	case ErrCatBug:
		return "BUG"
	case ErrCatTodo:
		return "TODO"
	case ErrCatSess:
		return "session"
	case ErrCatLexing:
		return "lexical"
	case ErrCatParsing:
		return "syntax"
	case ErrCatNaming:
		return "naming"
	case ErrCatPreduce:
		return "preducing"
	case ErrCatUnreachable:
		return "unreachable"
	}
	return "other"
}

type Error struct {
	tldOff IErrPosOffsets

	ref  *Error
	pos  *udevlex.Pos
	msg  string
	len  int
	cat  ErrorCategory
	code int
}

func (me *Error) Cat() ErrorCategory {
	if me.ref != nil {
		return me.ref.Cat()
	}
	return me.cat
}

func (me *Error) Code() int {
	if me.ref != nil {
		return me.ref.Code()
	}
	return me.code
}

func (me *Error) CodeAndCat() string {
	return "Æ" + ustr.Int(me.Code()) + " #" + me.Cat().String()
}

func (me *Error) Len() int {
	if me.ref != nil {
		return me.ref.Len()
	}
	return me.len
}

func (me *Error) Msg() string {
	if me.ref != nil {
		return me.ref.Msg()
	}
	return me.msg
}

func (me *Error) Pos() *udevlex.Pos {
	if me.ref != nil {
		return me.ref.Pos()
	}
	if me.pos != nil && me.tldOff != nil {
		pos := *me.pos
		if me.tldOff != nil {
			pos.Ln1, pos.Off0 = pos.Ln1+me.tldOff.PosOffsetLine(), pos.Off0+me.tldOff.PosOffsetByte()
		}
		return &pos
	}
	return me.pos
}

func (me *Error) Error() string {
	if me.ref != nil {
		return me.ref.Error()
	}
	var pref string
	if p := me.Pos(); p != nil {
		pref = p.String() + ": "
	}
	return pref + "[" + me.CodeAndCat() + "] " + me.msg
}

func (me *Error) IsRef() bool { return me.ref != nil }

func (me *Error) UpdatePosOffsets(offsets IErrPosOffsets) {
	me.tldOff = offsets
}

func ErrAtPos(cat ErrorCategory, code int, pos *udevlex.Pos, length int, msg string) (err *Error) {
	err = &Error{msg: msg, len: length, cat: cat, code: code, pos: pos}
	return
}

func ErrAtTok(cat ErrorCategory, code int, tok *udevlex.Token, msg string) *Error {
	if tok == nil {
		return ErrAtPos(cat, code, nil, 1, msg)
	}
	return ErrAtPos(cat, code, &tok.Pos, len(tok.Lexeme), msg)
}

func ErrAtToks(cat ErrorCategory, code int, toks udevlex.Tokens, msg string) *Error {
	return ErrAtPos(cat, code, toks.Pos(), toks.Length(), msg)
}

func ErrLex(code int, pos *udevlex.Pos, msg string) *Error {
	return ErrAtPos(ErrCatLexing, code, pos, 1, msg)
}

func ErrNaming(code int, tok *udevlex.Token, msg string) *Error {
	return ErrAtTok(ErrCatNaming, code, tok, msg)
}

func ErrPreduce(code int, toks udevlex.Tokens, msg string) *Error {
	return ErrAtToks(ErrCatPreduce, code, toks, msg)
}

func ErrUnreach(code int, toks udevlex.Tokens, msg string) *Error {
	return ErrAtToks(ErrCatUnreachable, code, toks, msg)
}

func ErrSyn(code int, tok *udevlex.Token, msg string) *Error {
	return ErrAtTok(ErrCatParsing, code, tok, msg)
}

func ErrBug(code int, toks udevlex.Tokens, msg string) *Error {
	return ErrAtToks(ErrCatBug, code, toks, msg)
}

func ErrTodo(code int, toks udevlex.Tokens, msg string) *Error {
	return ErrAtToks(ErrCatTodo, code, toks, msg)
}

func ErrSess(code int, maybePath string, msg string) *Error {
	return ErrAtPos(ErrCatSess, code, ErrFauxPos(maybePath), 1, msg)
}

func ErrFrom(cat ErrorCategory, code int, maybeSrcFilePath string, err error) *Error {
	if err != nil {
		return ErrAtPos(cat, code, ErrFauxPos(maybeSrcFilePath), 1, err.Error())
	}
	return nil
}

func ErrFauxPos(maybeSrcFilePath string) (pos *udevlex.Pos) {
	if maybeSrcFilePath != "" {
		pos = &udevlex.Pos{Ln1: 1, Col1: 1, FilePath: maybeSrcFilePath}
	}
	return
}

func ErrRef(err *Error) *Error {
	if err.ref != nil {
		return err
	}
	return &Error{ref: err}
}

type Errors []*Error

func (me Errors) UpdatePosOffsets(offsets IErrPosOffsets) {
	for i := range me {
		me[i].UpdatePosOffsets(offsets)
	}
}

func (me *Errors) Add(errs ...*Error) (anyAdded bool) {
	if anyAdded = len(errs) > 0; anyAdded {
		*me = append(*me, errs...)
	}
	return
}

func (me *Errors) AddVia(v interface{}, errs Errors) interface{} { me.Add(errs...); return v }

func (me *Errors) AddAt(cat ErrorCategory, code int, pos *udevlex.Pos, length int, msg string) *Error {
	err := &Error{msg: msg, len: length, cat: cat, code: code, pos: pos}
	*me = append(*me, err)
	return err
}

func (me *Errors) AddLex(code int, pos *udevlex.Pos, msg string) *Error {
	return me.AddAt(ErrCatLexing, code, pos, 1, msg)
}

func (me *Errors) AddSyn(code int, toks udevlex.Tokens, msg string) *Error {
	return me.AddFromToks(ErrCatParsing, code, toks, msg)
}

func (me *Errors) AddBug(code int, toks udevlex.Tokens, msg string) *Error {
	return me.AddFromToks(ErrCatBug, code, toks, msg)
}

func (me *Errors) AddTodo(code int, toks udevlex.Tokens, msg string) *Error {
	return me.AddFromToks(ErrCatTodo, code, toks, msg)
}

func (me *Errors) AddSess(code int, maybePath string, msg string) *Error {
	return me.AddAt(ErrCatSess, code, ErrFauxPos(maybePath), 1, msg)
}

func (me *Errors) AddNaming(code int, tok *udevlex.Token, msg string) *Error {
	return me.AddFromTok(ErrCatNaming, code, tok, msg)
}

func (me *Errors) AddPreduce(code int, toks udevlex.Tokens, msg string) *Error {
	return me.AddFromToks(ErrCatPreduce, code, toks, msg)
}

func (me *Errors) AddUnreach(code int, toks udevlex.Tokens, msg string) *Error {
	return me.AddFromToks(ErrCatUnreachable, code, toks, msg)
}

func (me *Errors) AddFromTok(cat ErrorCategory, code int, tok *udevlex.Token, msg string) *Error {
	if tok == nil {
		return me.AddAt(cat, code, nil, 1, msg)
	}
	return me.AddAt(cat, code, &tok.Pos, len(tok.Lexeme), msg)
}

func (me *Errors) AddFromToks(cat ErrorCategory, code int, toks udevlex.Tokens, msg string) *Error {
	return me.AddAt(cat, code, toks.Pos(), toks.Length(), msg)
}

func (me Errors) Strings() (s []string) {
	s = make([]string, len(me))
	for i := range me {
		s[i] = me[i].Error()
	}
	return
}

func (me Errors) Refs() (refs Errors) {
	refs = make(Errors, len(me))
	for i, e := range me {
		refs[i] = ErrRef(e)
	}
	return
}

// Len implements `sort.Interface`.
func (me Errors) Len() int { return len(me) }

// Swap implements `sort.Interface`.
func (me Errors) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }

// Less implements `sort.Interface`.
func (me Errors) Less(i int, j int) bool {
	ei, ej := me[i], me[j]
	pei, pej := ei.Pos(), ej.Pos()
	if pej == nil {
		return false
	}
	if pei == nil {
		return true
	}
	if pei.FilePath == pej.FilePath {
		if pei.Off0 == pej.Off0 {
			return ei.msg < ej.msg
		}
		return pei.Off0 < pej.Off0
	}
	return pei.FilePath < pej.FilePath
}
