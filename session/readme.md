# atmosess
--
    import "github.com/metaleap/atmo/session"


## Usage

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
		All Kits
	}
	On struct {
		NewBackgroundMessages func()
		SomeKitsRefreshed     func(hadFreshErrs bool)
	}
	Options struct {
		BgMsgs struct {
			IncludeKitsErrs bool
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

#### func (*Ctx) CatchUpOnFileMods

```go
func (me *Ctx) CatchUpOnFileMods(ensureFilesMarkedAsChanged ...*atmolang.AstFile)
```

#### func (*Ctx) Eval

```go
func (me *Ctx) Eval(kit *Kit, src string) (str string, errs atmo.Errors)
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

#### func (*Ctx) KitEnsureLoaded

```go
func (me *Ctx) KitEnsureLoaded(kit *Kit)
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

#### func (*Ctx) KitsReloadModifiedsUnlessAlreadyWatching

```go
func (me *Ctx) KitsReloadModifiedsUnlessAlreadyWatching()
```

#### func (*Ctx) KnownKitImpPaths

```go
func (me *Ctx) KnownKitImpPaths() (kitImpPaths []string)
```
KnownKitImpPaths returns all the import-paths of all currently known `Kit`s.

#### func (*Ctx) WithInMemFileMod

```go
func (me *Ctx) WithInMemFileMod(srcFilePath string, altSrc string, do func()) (recoveredPanic interface{})
```

#### func (*Ctx) WithInMemFileMods

```go
func (me *Ctx) WithInMemFileMods(srcFilePathsAndAltSrcs map[string]string, do func()) (recoveredPanic interface{})
```

#### type CtxBgMsg

```go
type CtxBgMsg struct {
	Issue bool
	Time  time.Time
	Lines []string
}
```


#### type IrDefRef

```go
type IrDefRef struct {
	*atmoil.IrDefTop
	KitImpPath string
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
	}
}
```

Kit is a pile of atmo source files residing in the same directory and being
interpreted or compiled all together as a unit.

#### func (*Kit) AstNodeAt

```go
func (me *Kit) AstNodeAt(srcFilePath string, pos0ByteOffset int) (topLevelChunk *atmolang.SrcTopChunk, theNodeAndItsAncestors []atmolang.IAstNode)
```

#### func (*Kit) Defs

```go
func (me *Kit) Defs(name string) (defs atmoil.IrTopDefs)
```

#### func (*Kit) Errors

```go
func (me *Kit) Errors(maybeErrsToSrcs map[error][]byte) (errs []error)
```
Errors collects whatever issues exist in any of the `Kit`'s source files
(file-system errors, lexing/parsing errors, semantic errors etc).

#### func (*Kit) HasDefs

```go
func (me *Kit) HasDefs(name string) bool
```
HasDefs returns whether any of the `Kit`'s source files define `name`.

#### func (*Kit) IrNodeOfAstNode

```go
func (me *Kit) IrNodeOfAstNode(defId string, origNode atmolang.IAstNode) (astDefTop *atmoil.IrDefTop, theNodeAndItsAncestors []atmoil.INode)
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
