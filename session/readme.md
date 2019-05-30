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

#### type AstDefRef

```go
type AstDefRef struct {
	*atmolang_irfun.AstDefTop
	KitImpPath string
}
```


#### type Ctx

```go
type Ctx struct {
	Dirs struct {
		Cache string
		Kits  []string
	}
	Kits struct {
		All                                  Kits
		AlwaysEnsureLoadedAsSoonAsDiscovered bool
		RecurringBackgroundWatch             struct {
			ShouldNow func() bool
		}
	}
}
```

Ctx fields must never be written to from the outside after the `Ctx.Init` call.

#### func (*Ctx) BackgroundMessages

```go
func (me *Ctx) BackgroundMessages(clear bool) (msgs []CtxBgMsg)
```

#### func (*Ctx) BackgroundMessagesCount

```go
func (me *Ctx) BackgroundMessagesCount() (count int)
```

#### func (*Ctx) CatchUp

```go
func (me *Ctx) CatchUp(checkForFileModsNow bool)
```

#### func (*Ctx) Dispose

```go
func (me *Ctx) Dispose()
```
Dispose is called when done with the `Ctx`. There may be tickers to halt, etc.

#### func (*Ctx) Eval

```go
func (me *Ctx) Eval(kit *Kit, src string) (str string, errs atmo.Errors)
```

#### func (*Ctx) FauxKitsAdd

```go
func (me *Ctx) FauxKitsAdd(dirPath string) (err error)
```

#### func (*Ctx) FauxKitsHas

```go
func (me *Ctx) FauxKitsHas(dirPath string) (isSessionDirFauxKit bool)
```

#### func (*Ctx) Init

```go
func (me *Ctx) Init(clearCacheDir bool, sessionFauxKitDir string) (err error)
```
Init validates the `Ctx.Dirs` fields currently set, then builds up its `Kits`
reflective of the structures found in the various `me.Dirs.Kits` search paths
and from now on in sync with live modifications to those.

#### func (*Ctx) KitByDirPath

```go
func (me *Ctx) KitByDirPath(dirPath string, tryToAddToFauxKits bool) (kit *Kit)
```

#### func (*Ctx) KitByImpPath

```go
func (me *Ctx) KitByImpPath(impPath string) *Kit
```

#### func (*Ctx) KitDefFacts

```go
func (me *Ctx) KitDefFacts(kit *Kit, def *atmolang_irfun.AstDefTop) ValFacts
```

#### func (*Ctx) KitEnsureLoaded

```go
func (me *Ctx) KitEnsureLoaded(kit *Kit)
```

#### func (*Ctx) KitsCollectReferencers

```go
func (me *Ctx) KitsCollectReferencers(forceLoadAllKnownKits bool, defNames atmo.StringsUnorderedButUnique, indirects bool) (referencerDefIds map[string]*Kit)
```

#### func (*Ctx) KitsCollectReferences

```go
func (me *Ctx) KitsCollectReferences(forceLoadAllKnownKits bool, name string) map[*atmolang_irfun.AstDefTop][]atmolang_irfun.IAstNode
```

#### func (*Ctx) KitsEnsureLoaded

```go
func (me *Ctx) KitsEnsureLoaded(plusSessDirFauxKits bool, kitImpPaths ...string)
```

#### func (*Ctx) KitsReloadModifiedsUnlessAlreadyWatching

```go
func (me *Ctx) KitsReloadModifiedsUnlessAlreadyWatching()
```
KitsReloadModifiedsUnlessAlreadyWatching returns -1 if file-watching is enabled,
otherwise it scans all currently-known kits-dirs for modifications and refreshes
the `Ctx`'s internal represenation of `Kits` if any were noted.

#### func (*Ctx) KnownKitImpPaths

```go
func (me *Ctx) KnownKitImpPaths() (kitImpPaths []string)
```
KnownKitImpPaths returns all the import-paths of all currently known `Kit`s.

#### type CtxBgMsg

```go
type CtxBgMsg struct {
	Issue bool
	Time  time.Time
	Lines []string
}
```


#### type IValFact

```go
type IValFact interface {
	Errs() atmo.Errors
	String() string
}
```


#### type Kit

```go
type Kit struct {
	DirPath           string
	ImpPath           string
	WasEverToBeLoaded bool
	Imports           []string

	SrcFiles atmolang.AstFiles

	Errs struct {
		Stage0DirAccessDuringRefresh error
		Stage0BadImports             []error
		Stage1BadNames               atmo.Errors
	}
}
```

Kit is a pile of atmo source files residing in the same directory and being
interpreted or compiled all together as a unit.

#### func (*Kit) AstNodeAt

```go
func (me *Kit) AstNodeAt(srcFilePath string, pos0ByteOffset int) (topLevelChunk *atmolang.AstFileTopLevelChunk, theNodeAndItsAncestors []atmolang.IAstNode)
```

#### func (*Kit) AstNodeIrFunFor

```go
func (me *Kit) AstNodeIrFunFor(defId string, origNode atmolang.IAstNode) (theNodeAndItsAncestors []atmolang_irfun.IAstNode)
```

#### func (*Kit) Defs

```go
func (me *Kit) Defs(name string) (defs atmolang_irfun.AstTopDefs)
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

#### type Kits

```go
type Kits []*Kit
```


#### func (Kits) ByDirPath

```go
func (me Kits) ByDirPath(kitDirPath string) *Kit
```

#### func (Kits) ByImpPath

```go
func (me Kits) ByImpPath(kitImpPath string) *Kit
```
ByImpPath finds the `Kit` in `Kits` with the given import-path.

#### func (Kits) IndexDirPath

```go
func (me Kits) IndexDirPath(dirPath string) int
```

#### func (Kits) IndexImpPath

```go
func (me Kits) IndexImpPath(impPath string) int
```

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

#### type ValFacts

```go
type ValFacts struct {
}
```


#### func (*ValFacts) Errs

```go
func (me *ValFacts) Errs() atmo.Errors
```

#### func (ValFacts) String

```go
func (me ValFacts) String() string
```
