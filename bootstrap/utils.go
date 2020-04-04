package main

import (
	"os"
)

type (
	Any = interface{}
	Str = []byte
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

// func uintToStr(n uint64)
