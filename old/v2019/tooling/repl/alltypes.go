// Package `atmorepl` provides the core functionality of the `atmo repl`
// command, including an infinite readln loop via `Repl.Run`
package atmorepl

import (
	"io"
	"time"

	. "github.com/metaleap/atmo/0ld/session"
)

type directives []directive

type directive struct {
	Desc   string
	Help   []string
	Run    func(string) bool
	Hidden bool
}

type Repl struct {
	Ctx             Ctx
	KnownDirectives directives
	IO              struct {
		Stdin           io.Reader
		Stdout          io.Writer
		Stderr          io.Writer
		MultiLineSuffix string
		TimeLastInput   time.Time

		write              func(string, int)
		writeLns, printLns func(...string)
	}

	// current mutable state during a `Run` loop
	run struct {
		quit                       bool
		indent                     int
		multiLnInputHadLeadingTabs bool
	}
}
