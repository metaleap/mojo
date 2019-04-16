# atem
--
    import "github.com/metaleap/atem"


## Usage

```go
const (
	EnvVarLibDirs = "ATEM_LIBSDIRS"
	NameAutoLib   = "ever"
	SrcFileExt    = ".at"
)
```

```go
var LibWatchInterval = 1 * time.Second
```

#### type Ctx

```go
type Ctx struct {
	ClearCacheDir bool
	Dirs          struct {
		Cur   string
		Cache string
		Libs  []string
	}
}
```


#### func (*Ctx) Dispose

```go
func (me *Ctx) Dispose()
```

#### func (*Ctx) Init

```go
func (me *Ctx) Init(dirCur string) (err error)
```

#### func (*Ctx) KnownLibPaths

```go
func (me *Ctx) KnownLibPaths() (libPaths []string)
```

#### func (*Ctx) KnownLibs

```go
func (me *Ctx) KnownLibs() (known []Lib)
```

#### func (*Ctx) Lib

```go
func (me *Ctx) Lib(libPath string) (lib *Lib)
```

#### func (*Ctx) LibEver

```go
func (me *Ctx) LibEver() (lib *Lib)
```

#### func (*Ctx) LibReachable

```go
func (me *Ctx) LibReachable(lib *Lib) (reachable bool)
```

#### func (*Ctx) ReadEvalPrint

```go
func (me *Ctx) ReadEvalPrint(in string) (out fmt.Stringer, err error)
```

#### type Lib

```go
type Lib struct {
	LibPath string
	DirPath string
	Errors  struct {
		Reload error
	}
	SrcFiles atemlang.AstFiles
}
```


#### func (*Lib) Err

```go
func (me *Lib) Err() error
```

#### func (*Lib) Error

```go
func (me *Lib) Error() (errMsg string)
```

#### func (*Lib) Errs

```go
func (me *Lib) Errs() (errs []error)
```

#### func (*Lib) IsEverLib

```go
func (me *Lib) IsEverLib() bool
```
