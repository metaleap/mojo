package main

import (
	"io/ioutil"
	"os"
	"runtime"
	"runtime/debug"
)

func main() {
	runtime.LockOSThread()
	debug.SetGCPercent(-1)

	// name_main := Str("main")
	input_src_file_bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	println(string(input_src_file_bytes))
}
