package main

type (
	Any = interface{}
	Str = []byte
)

func assert(b bool) {
	if !b {
		panic("assertion failure: indicates newly introduced bug")
	}
}

func unreachable() {
	panic("reached unreachable")
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
		case string: // same-looking as the default case, but need it explicitly
			print(msg_part)
		default:
			print(msg_part)
		}
	}
	print("\n\n")
	panic("\n__________\nBACKTRACE:")
}

func strConcat(strs []Str) Str {
	str_len := 0
	for _, str := range strs {
		str_len += len(str)
	}
	ret_str := ªbyte(str_len)
	idx := 0
	for _, str := range strs {
		for i, c := range str {
			ret_str[idx+i] = c
		}
		idx += len(str)
	}
	return ret_str
}

func strEq(one Str, two string) bool {
	return strEql(one, Str(two))
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

func uintFromStr(str Str) (uint64, bool) {
	assert(len(str) > 0)
	var mult uint64 = 1
	var ret uint64
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] < '0' || str[i] > '9' {
			return 0, false
		}
		ret += mult * uint64(str[i]-48)
		mult *= 10
	}
	return ret, true
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
	ret_str := ªbyte(len(prefix) + int(num_digits))
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

func ªbool(len int) []bool             { return make([]bool, len) }
func ªint(len int) []int               { return make([]int, len) }
func ªbyte(len int) Str                { return make(Str, len) }
func ªStr(len int) []Str               { return make([]Str, len) }
func ªAny(len int) []Any               { return make([]Any, len) }
func ªToken(len int) []Token           { return make([]Token, len) }
func ªTokens(len int) [][]Token        { return make([][]Token, len) }
func ªAstDef(len int) []AstDef         { return make([]AstDef, len) }
func ªAstExpr(len int) []AstExpr       { return make([]AstExpr, len) }
func ªAstExprPtr(len int) []*AstExpr   { return make([]*AstExpr, len) }
func ªAstNameRef(len int) []AstNameRef { return make([]AstNameRef, len) }
func ªIrLLExpr(len int) []IrLLExpr     { return make([]IrLLExpr, len) }
func ªIrHLDef(len int) []IrHLDef       { return make([]IrHLDef, len) }
func ªIrHLExpr(len int) []IrHLExpr     { return make([]IrHLExpr, len) }
