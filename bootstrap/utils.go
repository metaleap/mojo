package main

import (
	"os"
)

type (
	Any      = interface{}
	Str      = []byte
	StrNamed struct {
		name  Str
		value Str
	}
)

var (
	stdout = os.Stdout
)

func write(s Str) {
	if _, err := stdout.Write(s); err != nil {
		panic(err)
	}
}

func assert(b bool) {
	if !b {
		fail("assertion failure, backtrace:")
	}
}

func unreachable() {
	fail("reached unreachable, backtrace:")
}

func fail(msg_parts ...Any) {
	for i := 0; i < len(msg_parts)-1; i++ {
		switch msg_part := msg_parts[i].(type) {
		case Str:
			print(string(msg_part))
		default:
			print(msg_part)
		}
	}
	panic(msg_parts[len(msg_parts)-1])
}

func uintFromStr(str Str) uint64 {
	assert(len(str) > 0)
	var mult uint64 = 1
	var ret uint64
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] < '0' || str[i] > '9' {
			fail("malformed uint literal: ", str)
		}
		ret += mult * uint64(str[i]-48)
		mult *= 10
	}
	return ret
}

func uintToStr(integer uint64, base uint64, min_len uint64, prefix Str) Str {
	n := integer
	var num_digits uint64 = 1
	for n >= base {
		num_digits++
		n /= base
	}
	padding := (num_digits < min_len)
	if padding {
		num_digits = min_len
	}
	ret_str := allocˇu8(len(prefix) + int(num_digits))
	for i := range ret_str {
		ret_str[i] = '0'
	}
	for i := range prefix { // aka copy() but see comments in main.go, we stay low-level for a reason
		ret_str[i] = prefix[i]
	}
	// 123 / 10     123 % 10            12 / 10     12 % 10
	// =12          =3                  =1          =2
	idx := len(ret_str) - 1
	n = integer
	for keep_going := true; keep_going; {
		if n < base {
			ret_str[idx] = byte(48 + n)
			keep_going = false
		} else {
			ret_str[idx] = byte(48 + (n % base))
			n /= base
		}
		if base > 10 && ret_str[idx] > '9' {
			ret_str[idx] += 7
		}
		if keep_going {
			idx--
		}
	}
	return ret_str
}

func allocˇu8(len int) Str              { return make(Str, len) }
func allocˇToken(len int) Tokens        { return make(Tokens, len) }
func allocˇTokens(len int) []Tokens     { return make([]Tokens, len) }
func allocˇAstDef(len int) []AstDef     { return make([]AstDef, len) }
func allocˇAstExpr(len int) []AstExpr   { return make([]AstExpr, len) }
func allocˇStrNamed(len int) []StrNamed { return make([]StrNamed, len) }
