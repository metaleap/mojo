package atmo

import (
	"github.com/go-leap/dev/lex"
)

type IErrPosOffsets interface {
	Id() string
	PosOffsetLine() int
	PosOffsetByte() int
}

type ErrorCategory int

const (
	_ ErrorCategory = iota
	ErrCatTodo
	ErrCatLexing
	ErrCatParsing
	ErrCatNaming
	ErrCatSubst
	ErrCatUnreachable
)

func (me ErrorCategory) String() string {
	switch me {
	case ErrCatTodo:
		return "TODO"
	case ErrCatLexing:
		return "lexical"
	case ErrCatParsing:
		return "syntax"
	case ErrCatNaming:
		return "naming"
	case ErrCatSubst:
		return "substantiation"
	case ErrCatUnreachable:
		return "unreachable"
	}
	return "other"
}

type Error struct {
	tldOff IErrPosOffsets

	ref *Error
	msg string
	pos udevlex.Pos
	len int
	cat ErrorCategory
}

func (me *Error) Cat() ErrorCategory {
	if me.ref != nil {
		return me.ref.Cat()
	}
	return me.cat
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
	pos := me.pos
	if me.tldOff != nil {
		pos.Ln1, pos.Off0 = pos.Ln1+me.tldOff.PosOffsetLine(), pos.Off0+me.tldOff.PosOffsetByte()
	}
	return &pos
}

func (me *Error) Error() string {
	if me.ref != nil {
		return me.ref.Error()
	}
	return me.Pos().String() + ": [" + me.cat.String() + "] " + me.msg
}

func (me *Error) IsRef() bool { return me.ref != nil }

func (me *Error) UpdatePosOffsets(offsets IErrPosOffsets) {
	me.tldOff = offsets
}

func ErrAtPos(cat ErrorCategory, pos *udevlex.Pos, length int, msg string) (err *Error) {
	err = &Error{msg: msg, len: length, cat: cat}
	if pos != nil {
		err.pos = *pos
	}
	return
}

func ErrAtTok(cat ErrorCategory, tok *udevlex.Token, msg string) *Error {
	if tok == nil {
		return ErrAtPos(cat, nil, 0, msg)
	}
	return ErrAtPos(cat, &tok.Pos, len(tok.Lexeme), msg)
}

func ErrAtToks(cat ErrorCategory, toks udevlex.Tokens, msg string) *Error {
	return ErrAtPos(cat, toks.Pos(), toks.Length(), msg)
}

func ErrLex(pos *udevlex.Pos, msg string) *Error {
	return ErrAtPos(ErrCatLexing, pos, 1, msg)
}

func ErrNaming(tok *udevlex.Token, msg string) *Error {
	return ErrAtTok(ErrCatNaming, tok, msg)
}

func ErrSubst(toks udevlex.Tokens, msg string) *Error {
	return ErrAtToks(ErrCatSubst, toks, msg)
}

func ErrUnreach(toks udevlex.Tokens, msg string) *Error {
	return ErrAtToks(ErrCatUnreachable, toks, msg)
}

func ErrSyn(tok *udevlex.Token, msg string) *Error {
	return ErrAtTok(ErrCatParsing, tok, msg)
}

func ErrTodo(toks udevlex.Tokens, msg string) *Error {
	return ErrAtToks(ErrCatTodo, toks, msg)
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

func (me *Errors) Add(errs Errors) (anyAdded bool) {
	if anyAdded = len(errs) > 0; anyAdded {
		*me = append(*me, errs...)
	}
	return
}

func (me *Errors) AddVia(v interface{}, errs Errors) interface{} { me.Add(errs); return v }

func (me *Errors) AddAt(cat ErrorCategory, pos *udevlex.Pos, length int, msg string) *Error {
	err := &Error{msg: msg, len: length, cat: cat}
	if pos != nil {
		err.pos = *pos
	}
	*me = append(*me, err)
	return err
}

func (me *Errors) AddLex(pos *udevlex.Pos, msg string) {
	me.AddAt(ErrCatLexing, pos, 1, msg)
}

func (me *Errors) AddSyn(toks udevlex.Tokens, msg string) {
	me.AddFromToks(ErrCatParsing, toks, msg)
}

func (me *Errors) AddTodo(toks udevlex.Tokens, msg string) {
	me.AddFromToks(ErrCatTodo, toks, msg)
}

func (me *Errors) AddNaming(tok *udevlex.Token, msg string) {
	me.AddFromTok(ErrCatNaming, tok, msg)
}

func (me *Errors) AddSubst(toks udevlex.Tokens, msg string) {
	me.AddFromToks(ErrCatSubst, toks, msg)
}

func (me *Errors) AddUnreach(toks udevlex.Tokens, msg string) {
	me.AddFromToks(ErrCatUnreachable, toks, msg)
}

func (me *Errors) AddFromTok(cat ErrorCategory, tok *udevlex.Token, msg string) *Error {
	if tok == nil {
		return me.AddAt(cat, nil, 0, msg)
	}
	return me.AddAt(cat, &tok.Pos, len(tok.Lexeme), msg)
}

func (me *Errors) AddFromToks(cat ErrorCategory, toks udevlex.Tokens, msg string) *Error {
	return me.AddAt(cat, toks.Pos(), toks.Length(), msg)
}

func (me Errors) Errors() (errs []error) {
	errs = make([]error, len(me))
	for i, e := range me {
		errs[i] = e
	}
	return
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
	if pei.FilePath == pej.FilePath {
		if pei.Off0 == pej.Off0 {
			return ei.msg < ej.msg
		}
		return pei.Off0 < pej.Off0
	}
	return pei.FilePath < pej.FilePath
}
