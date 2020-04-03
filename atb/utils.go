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
	panic("reached unreachable, backtrace:")
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
