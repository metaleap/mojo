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
	ErrCatReducing
)

type Error struct {
	Msg string
	Pos scanner.Position
	Len int
	Cat ErrorCategory
}

// At ensures that `Error` shares an interface with `udevlex.Error`.
func (me *Error) At() *scanner.Position {
	return &me.Pos
}

func (me *Error) Error() (msg string) {
	msg = me.Pos.String() + ": "
	switch me.Cat {
	case ErrCatTodo:
		msg += "[──TODO──] not yet implemented: "
	case ErrCatLexing:
		msg += "[lexical] "
	case ErrCatParsing:
		msg += "[syntax] "
	case ErrCatNaming:
		msg += "[naming] "
	case ErrCatReducing:
		msg += "[validation] "
	default:
		msg += "[other] "
	}
	msg += me.Msg
	return
}

func (me *Errors) Add(errs Errors) (anyAdded bool) {
	if anyAdded = len(errs) > 0; anyAdded {
		*me = append(*me, errs...)
	}
	return
}

func (me *Errors) AddVia(v interface{}, errs Errors) interface{} { me.Add(errs); return v }

func ErrAt(cat ErrorCategory, pos *scanner.Position, length int, msg string) *Error {
	return &Error{Msg: msg, Pos: *pos, Len: length, Cat: cat}
}

func ErrLex(pos *scanner.Position, msg string) *Error {
	return ErrAt(ErrCatLexing, pos, 1, msg)
}

func ErrNaming(tok *udevlex.Token, msg string) *Error {
	return ErrAt(ErrCatNaming, &tok.Meta.Position, len(tok.Meta.Orig), msg)
}

func ErrReducing(tok *udevlex.Token, msg string) *Error {
	return ErrAt(ErrCatReducing, &tok.Meta.Position, len(tok.Meta.Orig), msg)
}

func ErrSyn(tok *udevlex.Token, msg string) *Error {
	return ErrAt(ErrCatParsing, &tok.Meta.Position, len(tok.Meta.Orig), msg)
}

func ErrTodo(tok *udevlex.Token, msg string) *Error {
	return ErrAt(ErrCatTodo, &tok.Meta.Position, len(tok.Meta.Orig), msg)
}

type Errors []Error

func (me *Errors) AddAt(cat ErrorCategory, pos *scanner.Position, length int, msg string) {
	*me = append(*me, Error{Msg: msg, Pos: *pos, Len: length, Cat: cat})
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

func (me *Errors) AddReducing(tok *udevlex.Token, msg string) {
	me.AddFrom(ErrCatReducing, tok, msg)
}

func (me *Errors) AddFrom(cat ErrorCategory, tok *udevlex.Token, msg string) {
	me.AddAt(cat, &tok.Meta.Position, len(tok.Meta.Orig), msg)
}

func (me Errors) Len() int          { return len(me) }
func (me Errors) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }
func (me Errors) Less(i int, j int) bool {
	ei, ej := &me[i], &me[j]
	if ei.Pos.Filename == ej.Pos.Filename {
		if ei.Pos.Offset == ej.Pos.Offset {
			return ei.Msg < ej.Msg
		}
		return ei.Pos.Offset < ej.Pos.Offset
	}
	return ei.Pos.Filename < ej.Pos.Filename
}
