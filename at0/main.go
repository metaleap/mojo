package main

import (
	"bufio"
	"os"
)

var stdIn = bufio.NewScanner(os.Stdin)

func main() {
	for stdIn.Scan() {
		data := stdIn.Bytes()

		if _, err := os.Stdout.Write(data); err != nil {
			panic(err)
		}
	}
	if err := stdIn.Err(); err != nil {
		panic(err)
	}
}
