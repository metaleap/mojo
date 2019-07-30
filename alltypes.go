package atmo

import (
	"github.com/go-leap/dev/lex"
)

type Exist struct{}
type StringKeys map[string]Exist

type Error struct {
	tldOff IErrPosOffsets

	ref  *Error
	pos  *udevlex.Pos
	msg  string
	len  int
	cat  ErrorCategory
	code int
}

type Errors []*Error

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
