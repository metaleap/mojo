# atmosess
--
    import "github.com/metaleap/atmo/session"


## Usage

```go
var KitsWatchInterval time.Duration
```
KitsWatchInterval is the default file-watching interval that is picked up by
each `Ctx.Init` call and then used throughout the `Ctx`'s life time.

#### func  CtxDefaultCacheDirPath

```go
func CtxDefaultCacheDirPath() string
```
CtxDefaultCacheDirPath returns the default used by `Ctx.Init` if
`Ctx.Dirs.Cache` was left empty. It returns a platform-specific dir path such as
`~/.cache/atmo`, `~/.config/atmo` etc. or in the worst case the current user's
home directory.

#### type Ctx

```go
type Ctx struct {
	Dirs struct {
		Cache string
		Kits  []string
	}

	Kits struct {
		RecurringBackgroundWatch struct {
			ShouldNow func() bool
		}
	}
}
```


#### func (*Ctx) AddFauxKit

```go
func (me *Ctx) AddFauxKit(dirPath string) (err error)
```

#### func (*Ctx) BackgroundMessages

```go
func (me *Ctx) BackgroundMessages(clear bool) (msgs []CtxBgMsg)
```

#### func (*Ctx) BackgroundMessagesCount

```go
func (me *Ctx) BackgroundMessagesCount() (count int)
```

#### func (*Ctx) Dispose

```go
func (me *Ctx) Dispose()
```
Dispose is called when done with the `Ctx`. There may be tickers to halt, etc.

#### func (*Ctx) Eval

```go
func (me *Ctx) Eval(kit *Kit, src string) (str string, errs []error)
```

#### func (*Ctx) Init

```go
func (me *Ctx) Init(clearCacheDir bool, sessionDir string) (err error)
```
Init validates the `Ctx.Dirs` fields currently set, then builds up its `Kits`
reflective of the structures found in the various `me.Dirs.Kits` search paths
and from now on in sync with live modifications to those.

#### func (*Ctx) KitEnsureLoaded

```go
func (me *Ctx) KitEnsureLoaded(kit *Kit)
```
KitEnsureLoaded forces (re)loading the `kit` only if it never was. (Primarily
for interactive load-on-demand scenarios like REPLs or editor language
servers.))

#### func (*Ctx) KitIsSessionDirFauxKit

```go
func (me *Ctx) KitIsSessionDirFauxKit(kit *Kit) bool
```

#### func (*Ctx) KitsEnsureLoaded

```go
func (me *Ctx) KitsEnsureLoaded(plusSessDirFauxKits bool, kitImpPaths ...string)
```

#### func (*Ctx) KnownKitImpPaths

```go
func (me *Ctx) KnownKitImpPaths() (kitImpPaths []string)
```
KnownKitImpPaths returns all the import-paths of all currently known `Kit`s.

#### func (*Ctx) ReloadModifiedKitsUnlessAlreadyWatching

```go
func (me *Ctx) ReloadModifiedKitsUnlessAlreadyWatching() (numFileSystemModsNoticedAndActedUpon int)
```
ReloadModifiedKitsUnlessAlreadyWatching returns -1 if file-watching is enabled,
otherwise it scans all currently-known kits-dirs for modifications and refreshes
the `Ctx`'s internal represenation of `Kits` if any were noted.

#### func (*Ctx) WithKit

```go
func (me *Ctx) WithKit(impPath string, do func(*Kit))
```
WithKit runs `do` with the specified `Kit` if it exists, else with `nil`. The
`Kit` must not be written to. While `do` runs, the `Kit` is blocked for updates
triggered by file modifications etc.

#### func (*Ctx) WithKnownKits

```go
func (me *Ctx) WithKnownKits(do func(Kits))
```
WithKnownKits runs `do` with all currently-known (loaded or not) `Kit`s passed
to it. The `Kits` slice or its contents must not be written to. While `do` runs,
the slice is blocked for updates triggered by file modifications etc.

#### func (*Ctx) WithKnownKitsWhere

```go
func (me *Ctx) WithKnownKitsWhere(where func(*Kit) bool, do func(Kits))
```
WithKnownKitsWhere works like `WithKnownKits` but with pre-filtering via
`where`.

#### type CtxBgMsg

```go
type CtxBgMsg struct {
	Issue bool
	Time  time.Time
	Lines []string
}
```


#### type Kit

```go
type Kit struct {
	DirPath           string
	ImpPath           string
	WasEverToBeLoaded bool
	Imports           []string
}
```

Kit is a pile of atmo source files residing in the same directory and being
interpreted or compiled all together as a unit.

#### func (*Kit) Defs

```go
func (me *Kit) Defs(name string, resolveNakedAliases bool) (defs atmolang_irfun.AstTopDefs)
```

#### func (*Kit) Errors

```go
func (me *Kit) Errors() (errs []error)
```
Errors collects whatever issues exist in any of the `Kit`'s source files
(file-system errors, lexing/parsing errors, semantic errors etc).

#### func (*Kit) HasDefs

```go
func (me *Kit) HasDefs(name string) bool
```
HasDefs returns whether any of the `Kit`'s source files define `name`.

#### func (*Kit) SrcFiles

```go
func (me *Kit) SrcFiles() atmolang.AstFiles
```
SrcFiles returns all source files belonging to the `Kit`. The slice or its
contents must not be written to.

#### type Kits

```go
type Kits []*Kit
```


#### func (Kits) ByImpPath

```go
func (me Kits) ByImpPath(kitImpPath string) *Kit
```
ByImpPath finds the `Kit` in `Kits` with the given import-path.

#### func (Kits) Len

```go
func (me Kits) Len() int
```
Len implements Go's standard `sort.Interface`.

#### func (Kits) Less

```go
func (me Kits) Less(i int, j int) bool
```
Less implements Go's standard `sort.Interface`.

#### func (Kits) Swap

```go
func (me Kits) Swap(i int, j int)
```
Swap implements Go's standard `sort.Interface`.

#### func (Kits) Where

```go
func (me Kits) Where(check func(*Kit) bool) (kits Kits)
```
