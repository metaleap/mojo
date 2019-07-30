# atmosess
--
    import "github.com/metaleap/atmo/session"

Package `atmosess` allows implementations of live sessions operating on atmo
"kits" (libs/pkgs of src-files basically), such as REPLs or language servers. It
will also be the foundation of interpreter / transpiler / compiler runs once
such scenarios begin to materialize later on.

An `atmosess.Ctx` session builds on the dumber `atmolang` and `atmoil` packages
to manage "kits" and their source files, keeping track of their existence,
loading them and their imports when instructed or needed, catching up on
file-system modifications to incrementally refresh in-memory representations as
well as invalidating / revalidating affected dependants on-the-fly etc.

A one-per-kit virtual/pretend/faux "source-file", entirely in-memory but treated
as equal to the kit's real source-files and called the "scratchpad" is also part
of the offering, allowing entry of (either exprs to be eval'd-and-forgotten or)
defs to be added to / modified in / removed from it.

## Usage

```go
const (
	ErrSessInit_IoCacheDirCreationFailure = iota + 3100
	ErrSessInit_IoCacheDirDeletionFailure
	ErrSessInit_KitsDirsConflict
	ErrSessInit_KitsDirsNotSpecified
	ErrSessInit_KitsDirsNotFound
	ErrSessInit_KitsDirAutoNotFound
	ErrSessInit_IoFauxKitDirFailure
)
```

```go
const (
	ErrSessKits_IoReadDirFailure = iota + 3200
	ErrSessKits_ImportNotFound
)
```

#### func  CtxDefaultCacheDirPath

```go
func CtxDefaultCacheDirPath() string
```
CtxDefaultCacheDirPath returns the default used by `Ctx.Init` if
`Ctx.Dirs.Cache` was left empty. It returns a platform-specific dir path such as
`~/.cache/atmo`, `~/.config/atmo` etc. or in the worst case the current user's
home directory.

#### func  IsValidKitDirName

```go
func IsValidKitDirName(dirName string) bool
```

#### type Ctx

```go
type Ctx struct {
	Dirs struct {
		CacheData   string
		KitsStashes []string
	}
	Kits struct {
		All Kits
	}
	On struct {
		NewBackgroundMessages func(*Ctx)
		SomeKitsRefreshed     func(ctx *Ctx, hadFreshErrs bool)
	}
	Options struct {
		BgMsgs struct {
			IncludeLiveKitsErrs bool
		}
		FileModsCatchup struct {
			BurstLimit time.Duration
		}
		Scratchpad struct {
			FauxFileNameForErrorMessages string
		}
	}
}
```

Ctx fields must never be written to from the outside after the `Ctx.Init` call.

#### func (*Ctx) BackgroundMessages

```go
func (me *Ctx) BackgroundMessages(clear bool) (msgs []ctxBgMsg)
```

#### func (*Ctx) BackgroundMessagesCount

```go
func (me *Ctx) BackgroundMessagesCount() (count int)
```

#### func (*Ctx) CatchUpOnFileMods

```go
func (me *Ctx) CatchUpOnFileMods(ensureFilesMarkedAsChanged ...*atmolang.AstFile)
```

#### func (*Ctx) FauxKitsAdd

```go
func (me *Ctx) FauxKitsAdd(dirPath string) (is bool, err error)
```

#### func (*Ctx) FauxKitsHas

```go
func (me *Ctx) FauxKitsHas(dirPath string) bool
```

#### func (*Ctx) Init

```go
func (me *Ctx) Init(clearCacheDir bool, sessionFauxKitDir string) (kitImpPathIfFauxKitDirActualKit string, err *atmo.Error)
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

#### func (*Ctx) KitEnsureLoaded

```go
func (me *Ctx) KitEnsureLoaded(kit *Kit) (freshErrs atmo.Errors)
```

#### func (*Ctx) KitsCollectDependants

```go
func (me *Ctx) KitsCollectDependants(forceLoadAllKnownKits bool, defNames atmo.StringKeys, indirects bool) (dependantsDefIds map[string]*Kit)
```

#### func (*Ctx) KitsCollectReferences

```go
func (me *Ctx) KitsCollectReferences(forceLoadAllKnownKits bool, name string) map[*atmoil.IrDefTop][]atmoil.IExpr
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

#### func (*Ctx) Locked

```go
func (me *Ctx) Locked(do func())
```
Locked is never used by `atmosess` itself but a convenience helper for outside
callers that run parallel code-paths and thus need to serialize concurrent
accesses to their `Ctx`. Wrap any and all of your `Ctx` uses in a `func` passed
to `Locked` and concurrent accesses will queue up. Caution: calling `Locked`
again from inside such a wrapper `func` will deadlock.

#### func (*Ctx) Preduce

```go
func (me *Ctx) Preduce(kit *Kit, node atmoil.INode) (atmoil.IPreduced, atmo.Errors)
```

#### func (*Ctx) ScratchpadEntry

```go
func (me *Ctx) ScratchpadEntry(kit *Kit, maybeTopDefId string, src string) (ret atmoil.IPreduced, errs atmo.Errors)
```

#### func (*Ctx) WithInMemFileMod

```go
func (me *Ctx) WithInMemFileMod(srcFilePath string, altSrc string, do func()) (recoveredPanic interface{})
```

#### func (*Ctx) WithInMemFileMods

```go
func (me *Ctx) WithInMemFileMods(srcFilePathsAndAltSrcs map[string]string, do func()) (recoveredPanic interface{})
```

#### type IrDefRef

```go
type IrDefRef struct {
	*atmoil.IrDefTop
	Kit *Kit
}
```


#### type Kit

```go
type Kit struct {
	DirPath           string
	ImpPath           string
	WasEverToBeLoaded bool

	SrcFiles atmolang.AstFiles

	Errs struct {
		Stage1DirAccessDuringRefresh *atmo.Error
		Stage1BadImports             atmo.Errors
	}
}
```

Kit is a set of atmo source files residing in the same directory and being
interpreted or compiled all together as a unit.

#### func (*Kit) AstNodeAt

```go
func (me *Kit) AstNodeAt(srcFilePath string, pos0ByteOffset int) (topLevelChunk *atmolang.SrcTopChunk, theNodeAndItsAncestors []atmolang.IAstNode)
```

#### func (*Kit) Defs

```go
func (me *Kit) Defs(name string, includeUnparsedOnes bool) (defs atmoil.IrTopDefs)
```

#### func (*Kit) DoesImport

```go
func (me *Kit) DoesImport(kitImpPath string) bool
```

#### func (*Kit) Errors

```go
func (me *Kit) Errors(maybeErrsToSrcs map[*atmo.Error][]byte) (errs atmo.Errors)
```
Errors collects whatever issues exist in any of the `Kit`'s source files
(file-system errors, lexing/parsing errors, semantic errors etc).

#### func (*Kit) HasDefs

```go
func (me *Kit) HasDefs(name string) bool
```
HasDefs returns whether any of the `Kit`'s source files define `name`.

#### func (*Kit) Imports

```go
func (me *Kit) Imports() []string
```

#### func (*Kit) IrNodeOfAstNode

```go
func (me *Kit) IrNodeOfAstNode(defId string, origNode atmolang.IAstNode) (astDefTop *atmoil.IrDefTop, theNodeAndItsAncestors []atmoil.INode)
```

#### func (*Kit) ScratchpadClear

```go
func (me *Kit) ScratchpadClear()
```

#### func (*Kit) ScratchpadView

```go
func (me *Kit) ScratchpadView() []byte
```

#### func (*Kit) SelectNodes

```go
func (me *Kit) SelectNodes(tldOk func(*atmoil.IrDefTop) bool, nodeOk func([]atmoil.INode, atmoil.INode, []atmoil.INode) (ismatch bool, dontdescend bool, tlddone bool, alldone bool)) (matches map[atmoil.INode]*atmoil.IrDefTop)
```

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

#### func (Kits) DependantKitsOf

```go
func (me Kits) DependantKitsOf(kitImpPath string, indirects bool) (dependantKitImpPaths map[string]*Kit)
```

#### func (Kits) ImportersOf

```go
func (me Kits) ImportersOf(kitImpPath string) Kits
```

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

#### func (Kits) SrcFilePaths

```go
func (me Kits) SrcFilePaths() (srcFilePaths []string)
```

#### func (Kits) Swap

```go
func (me Kits) Swap(i int, j int)
```
Swap implements Go's standard `sort.Interface`.

#### func (Kits) Where

```go
func (me Kits) Where(check func(*Kit) bool) (kits Kits)
```
