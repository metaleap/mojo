# atemrepl
--
    import "github.com/metaleap/atem/tooling/repl"


## Usage

#### type Repl

```go
type Repl struct {
	Ctx             atem.Ctx
	KnownDirectives directives
	IO              struct {
		Stdin           io.Reader
		Stdout          io.Writer
		Stderr          io.Writer
		MultiLineSuffix string
	}
}
```


#### func (*Repl) DInfo

```go
func (me *Repl) DInfo(what string) bool
```

#### func (*Repl) DList

```go
func (me *Repl) DList(what string) bool
```

#### func (*Repl) DQuit

```go
func (me *Repl) DQuit(string) bool
```

#### func (*Repl) DWelcomeMsg

```go
func (me *Repl) DWelcomeMsg(string) bool
```

#### func (*Repl) Run

```go
func (me *Repl) Run(showWelcomeMsg bool)
```
