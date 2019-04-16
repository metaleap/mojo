package atmolang

import (
	"text/scanner"

	"github.com/go-leap/dev/lex"
)

type ErrorCategory int

const (
	_ ErrorCategory = iota
	ErrCatSyntax
)

type Error struct {
	Msg string
	Pos scanner.Position
	Len int
	Cat ErrorCategory
}

func errAt(cat ErrorCategory, tok *udevlex.Token, length int, msg string) *Error {
	return &Error{Msg: msg, Pos: tok.Meta.Position, Len: length, Cat: cat}
}

func errSyntax(tok *udevlex.Token, msg string) *Error {
	return errAt(ErrCatSyntax, tok, len(tok.Meta.Orig), msg)
}

func (me *Error) Error() (msg string) {
	msg = me.Pos.String() + ": "
	switch me.Cat {
	case ErrCatSyntax:
		msg += "[syntax] "
	default:
		msg += "[other] "
	}
	msg += me.Msg
	return
}
