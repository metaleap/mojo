// Package `atmo` is the foundational shared package imported by any and
// every package belonging to atmo. In addition to some general-purpose utils
// and firm ground-rules (src-file extension, env-var name etc.) to abide by,
// it offers a common `Error` type and supporting types. The idea being that
// individual packages own their own error _codes_ but all use `atmo.Error`
// and the `atmo.ErrorCategory` enum.
package atmo

import (
	"github.com/go-leap/dev/lex"
)

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

type Exist struct{}
type StringKeys map[string]Exist
