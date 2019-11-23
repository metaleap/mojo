# atmorepl
--
    import "github.com/metaleap/atmo/0ld/tooling/repl"

Package `atmorepl` provides the core functionality of the `atmo repl` command,
including an infinite readln loop via `Repl.Run`

## Usage

```go
var (
	Ux struct {
		AnimsEnabled    bool
		MoreLines       int
		MoreLinesPrompt []byte
		WelcomeMsgLines []string
		OldSchoolTty    bool
		WelcomeMsgShow  bool
	}
)
```

#### type Repl

```go
type Repl struct {
	Ctx             Ctx
	KnownDirectives directives
	IO              struct {
		Stdin           io.Reader
		Stdout          io.Writer
		Stderr          io.Writer
		MultiLineSuffix string
		TimeLastInput   time.Time
	}
}
```


#### func (*Repl) DInfo

```go
func (me *Repl) DInfo(what string) bool
```

#### func (*Repl) DIntro

```go
func (me *Repl) DIntro(string) bool
```

#### func (*Repl) DList

```go
func (me *Repl) DList(what string) bool
```

#### func (*Repl) DQuit

```go
func (me *Repl) DQuit(s string) bool
```

#### func (*Repl) DSrcs

```go
func (me *Repl) DSrcs(what string) bool
```

#### func (*Repl) QuitNonDirectiveInitiated

```go
func (me *Repl) QuitNonDirectiveInitiated(anim bool)
```

#### func (*Repl) Run

```go
func (me *Repl) Run(loadSessDirFauxKit bool, loadKitsByImpPaths ...string)
```
