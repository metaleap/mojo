package atmo

import (
	"text/scanner"

	"github.com/go-leap/dev/lex"
)

type ErrorCategory int

const (
	_ ErrorCategory = iota
	ErrCatLexing
	ErrCatParsing
)

type Error struct {
	Msg string
	Pos scanner.Position
	Len int
	Cat ErrorCategory
}

func (me *Error) Error() (msg string) {
	msg = me.Pos.String() + ": "
	switch me.Cat {
	case ErrCatLexing:
		msg += "[lexical] "
	case ErrCatParsing:
		msg += "[syntax] "
	default:
		msg += "[other] "
	}
	msg += me.Msg
	return
}

func ErrAt(cat ErrorCategory, pos *scanner.Position, length int, msg string) *Error {
	return &Error{Msg: msg, Pos: *pos, Len: length, Cat: cat}
}

func ErrLex(pos *scanner.Position, msg string) *Error {
	return ErrAt(ErrCatLexing, pos, 1, msg)
}

func ErrSyn(tok *udevlex.Token, msg string) *Error {
	return ErrAt(ErrCatParsing, &tok.Meta.Position, len(tok.Meta.Orig), msg)
}
