package atmo

import (
	"text/scanner"

	"github.com/go-leap/dev/lex"
)

type ErrorCategory int

const (
	_ ErrorCategory = iota
	ErrCatTodo
	ErrCatLexing
	ErrCatParsing
	ErrCatNaming
	ErrCatSubst
)

type Error struct {
	ref *Error
	msg string
	pos scanner.Position
	len int
	cat ErrorCategory
}

// At ensures that `Error` shares an interface with `udevlex.Error`.
func (me *Error) At() *scanner.Position {
	if me.ref != nil {
		return me.ref.At()
	}
	return &me.pos
}

func (me *Error) Error() (msg string) {
	if me.ref != nil {
		return me.ref.Error()
	}
	msg = me.pos.String() + ": "
	switch me.cat {
	case ErrCatTodo:
		msg += "[──TODO──] not yet implemented: "
	case ErrCatLexing:
		msg += "[lexical] "
	case ErrCatParsing:
		msg += "[syntax] "
	case ErrCatNaming:
		msg += "[naming] "
	case ErrCatSubst:
		msg += "[substantiation] "
	default:
		msg += "[other] "
	}
	msg += me.msg
	return
}

func (me *Error) IsRef() bool { return me.ref != nil }

func (me *Errors) Add(errs Errors) (anyAdded bool) {
	if anyAdded = len(errs) > 0; anyAdded {
		*me = append(*me, errs...)
	}
	return
}

func (me *Errors) AddVia(v interface{}, errs Errors) interface{} { me.Add(errs); return v }

func ErrAt(cat ErrorCategory, pos *scanner.Position, length int, msg string) *Error {
	return &Error{msg: msg, pos: *pos, len: length, cat: cat}
}

func ErrLex(pos *scanner.Position, msg string) *Error {
	return ErrAt(ErrCatLexing, pos, 1, msg)
}

func ErrNaming(tok *udevlex.Token, msg string) *Error {
	return ErrAt(ErrCatNaming, &tok.Meta.Position, len(tok.Meta.Orig), msg)
}

func ErrSubst(tok *udevlex.Token, msg string) *Error {
	return ErrAt(ErrCatSubst, &tok.Meta.Position, len(tok.Meta.Orig), msg)
}

func ErrSyn(tok *udevlex.Token, msg string) *Error {
	return ErrAt(ErrCatParsing, &tok.Meta.Position, len(tok.Meta.Orig), msg)
}

func ErrTodo(tok *udevlex.Token, msg string) *Error {
	return ErrAt(ErrCatTodo, &tok.Meta.Position, len(tok.Meta.Orig), msg)
}

func ErrRef(err *Error) *Error {
	if err.ref != nil {
		return err
	}
	return &Error{ref: err}
}

type Errors []Error

func (me *Errors) AddAt(cat ErrorCategory, pos *scanner.Position, length int, msg string) {
	*me = append(*me, Error{msg: msg, pos: *pos, len: length, cat: cat})
}

func (me *Errors) AddLex(pos *scanner.Position, msg string) {
	me.AddAt(ErrCatLexing, pos, 1, msg)
}

func (me *Errors) AddSyn(tok *udevlex.Token, msg string) {
	me.AddFrom(ErrCatParsing, tok, msg)
}

func (me *Errors) AddTodo(tok *udevlex.Token, msg string) {
	me.AddFrom(ErrCatTodo, tok, msg)
}

func (me *Errors) AddNaming(tok *udevlex.Token, msg string) {
	me.AddFrom(ErrCatNaming, tok, msg)
}

func (me *Errors) AddSubst(tok *udevlex.Token, msg string) {
	me.AddFrom(ErrCatSubst, tok, msg)
}

func (me *Errors) AddFrom(cat ErrorCategory, tok *udevlex.Token, msg string) {
	me.AddAt(cat, &tok.Meta.Position, len(tok.Meta.Orig), msg)
}

func (me Errors) Errors() (s []string) {
	s = make([]string, len(me))
	for i := range me {
		s[i] = me[i].Error()
	}
	return
}

func (me Errors) Len() int          { return len(me) }
func (me Errors) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }
func (me Errors) Less(i int, j int) bool {
	ei, ej := &me[i], &me[j]
	if ei.pos.Filename == ej.pos.Filename {
		if ei.pos.Offset == ej.pos.Offset {
			return ei.msg < ej.msg
		}
		return ei.pos.Offset < ej.pos.Offset
	}
	return ei.pos.Filename < ej.pos.Filename
}
