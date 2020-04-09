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
	counter uint64
	stdout  = os.Stdout
)

func write(s Str) {
	if _, err := stdout.Write(s); err != nil {
		panic(err)
	}
}

func assert(b bool) {
	if !b {
		fail("assertion failure")
	}
}

func strConcat(strs []Str) Str {
	str_len := 0
	for _, str := range strs {
		str_len += len(str)
	}
	ret_str := allocˇu8(str_len)
	idx := 0
	for _, str := range strs {
		for i, c := range str {
			ret_str[idx+i] = c
		}
		idx += len(str)
	}
	return ret_str
}

func trimAt(str Str, char byte) Str {
	for i := range str {
		if str[i] == char {
			return str[0:i]
		}
	}
	return str
}

func unreachable() {
	fail("reached unreachable")
}

func fail(msg_parts ...Any) {
	for i := 0; i < len(msg_parts); i++ {
		switch msg_part := msg_parts[i].(type) {
		case Str:
			print(string(msg_part))
		case AstExprIdent:
			print(string(msg_part))
		case AstExprLitStr:
			print(string(msg_part))
		case LLExprIdentLocal:
			print(string(msg_part))
		case LLExprIdentGlobal:
			print(string(msg_part))
		case LLExprLitStr:
			print(string(msg_part))
		case string: // same-looking as the default case, but need it explicitly for expected output at least in go1.14
			print(msg_part)
		default:
			print(msg_part)
		}
	}
	print("\n\n")
	panic("\n__________\nBACKTRACE:")
}

func strEql(one Str, two Str) bool {
	if len(one) == len(two) {
		for i := range one {
			if one[i] != two[i] {
				return false
			}
		}
		return true
	}
	return false
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

/*
	why have the below instead of direct explicit `make()` calls everywhere?
	later want to move from OS heap allocs to a pre-alloc'd fixed-size buffer.
*/

func allocˇu8(len int) Str                      { return make(Str, len) }
func allocˇStr(len int) []Str                   { return make([]Str, len) }
func allocˇToken(len int) []Token               { return make([]Token, len) }
func allocˇTokens(len int) [][]Token            { return make([][]Token, len) }
func allocˇAstDef(len int) []AstDef             { return make([]AstDef, len) }
func allocˇAstExpr(len int) []AstExpr           { return make([]AstExpr, len) }
func allocˇAstExprPtr(len int) []*AstExpr       { return make([]*AstExpr, len) }
func allocˇAstNameRef(len int) []AstNameRef     { return make([]AstNameRef, len) }
func allocˇStrNamed(len int) []StrNamed         { return make([]StrNamed, len) }
func allocˇLLGlobal(len int) []LLGlobal         { return make([]LLGlobal, len) }
func allocˇLLFunc(len int) []LLFunc             { return make([]LLFunc, len) }
func allocˇLLFuncParam(len int) []LLFuncParam   { return make([]LLFuncParam, len) }
func allocˇLLBasicBlock(len int) []LLBasicBlock { return make([]LLBasicBlock, len) }
func allocˇLLInstr(len int) []LLInstr           { return make([]LLInstr, len) }
func allocˇLLType(len int) []LLType             { return make([]LLType, len) }
func allocˇLLExprTyped(len int) []LLExprTyped   { return make([]LLExprTyped, len) }
func allocˇLLPhiPred(len int) []LLPhiPred       { return make([]LLPhiPred, len) }
