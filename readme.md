# atmo
--
    import "github.com/metaleap/atmo"

For a rough idea, imagine a console screenshot instead of this paste ...

    ~/c/atmo> atmo repl

    ┌─────────────────────────────────────────
    │:intro
    └─────────────────────────────────────────

    This is a read-eval-print loop (repl).

    — repl commands start with `:`, any other
    inputs are eval'd as atmo expressions

    — in case of the latter, a line ending in ,,,
    introduces or concludes a multi-line input

    — to see --flags, quit and run `atmo help`

    ┌─────────────────────────────────────────
    │:
    └─────────────────────────────────────────

    Unknown command `:` — try:

        :list ‹kit›
        :info ‹kit› [‹def›]
        :srcs ‹kit› ‹def›
        :quit
        :intro

    (For usage details on an arg-ful
    command, invoke it without args.)

    ┌─────────────────────────────────────────
    │:l
    └─────────────────────────────────────────

    Input `` insufficient for command `:list`.

    Usage:

    :list ‹kit/import/path› ── list defs in the specified kit
    :list _                 ── list all currently known kits

    ┌─────────────────────────────────────────
    │:l _
    └─────────────────────────────────────────

    LIST of kits from current search paths:
    ─── /home/_/c/atmo/kits

    Found 3 kits:
    ├── [×] omni
    ├── [×] ·home·_·c·atmo
    ├── [_] omni/tmp

    Legend: [_] = unloaded, [×] = loaded or load attempted
    (To see kit details, use `:info ‹kit›`.)

    ┌─────────────────────────────────────────
    │:l ·
    └─────────────────────────────────────────

    LIST of defs in kit:    `·home·_·c·atmo`
               found in:    /home/_/c/atmo

    red.at: 10 top-level defs
    ├── it ─── (line 2)
    ├── ever ─── (line 4)
    ├── usefloat ─── (line 6)
    ├── dafl ─── (line 8)
    ├── daFl ─── (line 10)
    ├── leFl ─── (line 12)
    ├── fn ─── (line 14)
    ├── fVal ─── (line 16)
    ├── fval ─── (line 18)
    ├── test ─── (line 20)

    Total: 10 defs in 1 `*.at` source file

    (To see more details, try also:
    `:info ·` or `:info · ‹def›`.)

    ┌─────────────────────────────────────────
    │:i ·
    └─────────────────────────────────────────

    INFO summary on kit:    `·home·_·c·atmo`
               found in:    /home/_/c/atmo

    1 source file in kit `·`:
    ├── red.at
        20 lines (10 sloc), 10 top-level defs, 10 exported
    Total:
        20 lines (10 sloc), 10 top-level defs, 10 exported
        (Counts exclude failed-to-parse defs, if any.)

    (To see kit defs, use `:list ·`.)

    ┌─────────────────────────────────────────
    │:s · it
    └─────────────────────────────────────────

    1 def named `it` found in kit `·home·_·c·atmo`:

    ├── /home/_/c/atmo/red.at

    x it :=
        x

    ├── internal representation:

    it x :=
        x

    ┌─────────────────────────────────────────
    │:i · it
    └─────────────────────────────────────────

    1 def named `it` found in kit `·home·_·c·atmo`:

    ‹x: ()› » x

    ┌─────────────────────────────────────────
    │ :quit
    └─────────────────────────────────────────

    ~/c/atmo>

## Usage

```go
const (
	EnvVarKitsDirs = "ATMO_KITS_DIRS"
	NameAutoKit    = "omni"
	SrcFileExt     = ".at"
)
```

```go
var Options struct {
	// sorted errors, kits, source files, defs etc:
	// should be enabled for consistency in user-facing tools such as REPLs or language servers.
	// could remain off for mere script runners, transpilers etc.
	Sorts bool
}
```

#### func  SortMaybe

```go
func SortMaybe(s sort.Interface)
```

#### type Error

```go
type Error struct {
}
```


#### func  ErrAt

```go
func ErrAt(cat ErrorCategory, pos *scanner.Position, length int, msg string) *Error
```

#### func  ErrLex

```go
func ErrLex(pos *scanner.Position, msg string) *Error
```

#### func  ErrNaming

```go
func ErrNaming(tok *udevlex.Token, msg string) *Error
```

#### func  ErrRef

```go
func ErrRef(err *Error) *Error
```

#### func  ErrSubst

```go
func ErrSubst(tok *udevlex.Token, msg string) *Error
```

#### func  ErrSyn

```go
func ErrSyn(tok *udevlex.Token, msg string) *Error
```

#### func  ErrTodo

```go
func ErrTodo(tok *udevlex.Token, msg string) *Error
```

#### func (*Error) At

```go
func (me *Error) At() *scanner.Position
```
At ensures that `Error` shares an interface with `udevlex.Error`.

#### func (*Error) Error

```go
func (me *Error) Error() (msg string)
```

#### func (*Error) IsRef

```go
func (me *Error) IsRef() bool
```

#### type ErrorCategory

```go
type ErrorCategory int
```


```go
const (
	ErrCatTodo ErrorCategory
	ErrCatLexing
	ErrCatParsing
	ErrCatNaming
	ErrCatSubst
)
```

#### type Errors

```go
type Errors []*Error
```


#### func (*Errors) Add

```go
func (me *Errors) Add(errs Errors) (anyAdded bool)
```

#### func (*Errors) AddAt

```go
func (me *Errors) AddAt(cat ErrorCategory, pos *scanner.Position, length int, msg string)
```

#### func (*Errors) AddFrom

```go
func (me *Errors) AddFrom(cat ErrorCategory, tok *udevlex.Token, msg string)
```

#### func (*Errors) AddLex

```go
func (me *Errors) AddLex(pos *scanner.Position, msg string)
```

#### func (*Errors) AddNaming

```go
func (me *Errors) AddNaming(tok *udevlex.Token, msg string)
```

#### func (*Errors) AddSubst

```go
func (me *Errors) AddSubst(tok *udevlex.Token, msg string)
```

#### func (*Errors) AddSyn

```go
func (me *Errors) AddSyn(tok *udevlex.Token, msg string)
```

#### func (*Errors) AddTodo

```go
func (me *Errors) AddTodo(tok *udevlex.Token, msg string)
```

#### func (*Errors) AddVia

```go
func (me *Errors) AddVia(v interface{}, errs Errors) interface{}
```

#### func (Errors) Errors

```go
func (me Errors) Errors() (errs []error)
```

#### func (Errors) Len

```go
func (me Errors) Len() int
```

#### func (Errors) Less

```go
func (me Errors) Less(i int, j int) bool
```

#### func (Errors) Refs

```go
func (me Errors) Refs() (refs Errors)
```

#### func (Errors) Strings

```go
func (me Errors) Strings() (s []string)
```

#### func (Errors) Swap

```go
func (me Errors) Swap(i int, j int)
```
