// Package `atmosess` allows implementations of live sessions operating on
// atmo "kits" (libs/pkgs of src-files basically), such as REPLs or language
// servers. It will also be the foundation of interpreter / transpiler /
// compiler runs once such scenarios begin to materialize later on.
//
// An `atmosess.Ctx` session builds on the dumber `atmolang` and `atmoil`
// packages to manage "kits" and their source files, keeping track of their
// existence, loading them and their imports when instructed or needed,
// catching up on file-system modifications to incrementally refresh in-memory
// representations as well as invalidating / revalidating affected dependants
// on-the-fly etc.
//
// A one-per-kit virtual/pretend/faux "source-file", entirely in-memory but
// treated as equal to the kit's real source-files and called the "scratchpad"
// is also part of the offering, allowing entry of (either exprs to be
// eval'd-and-forgotten or) defs to be added to / modified in / removed from it.
package atmosess

import (
	"os"
	"sync"
	"time"

	. "github.com/metaleap/atmo"
	. "github.com/metaleap/atmo/ast"
	. "github.com/metaleap/atmo/il"
)

type ctxPreducing struct {
	dbgIndent  int
	curSessCtx *Ctx
	curNode    struct {
		owningTopDef *IrDef
		owningKit    *Kit
	}
	envStack []interface{}
}

// Ctx fields must never be written to from the outside after the `Ctx.Init` call.
type Ctx struct {
	Dirs struct {
		fauxKits    []string
		CacheData   string
		KitsStashes []string
	}
	Kits struct {
		All                Kits
		reprocessingNeeded bool
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
	state struct {
		bgMsgs        []ctxBgMsg
		fileModsWatch struct {
			lastCatchup                   time.Time
			latest                        []map[string]os.FileInfo
			collectFileModsForNextCatchup func([]string, []string) int
		}
		notUsedInternallyButAvailableForOutsideCallersConvenience sync.Mutex
	}
}

type ctxBgMsg = struct {
	Issue bool
	Time  time.Time
	Lines []string
}

type Kits []*Kit

// Kit is a set of atmo source files residing in the same directory and
// being interpreted or compiled all together as a unit.
type Kit struct {
	DirPath           string
	ImpPath           string
	WasEverToBeLoaded bool

	imports      []string
	topLevelDefs IrDefs
	SrcFiles     AstFiles
	state        struct {
		defsGoneIdsNames map[string]string
		defsBornIdsNames map[string]string
	}
	lookups struct {
		tlDefsByID      map[string]*IrDef
		tlDefIDsByName  map[string][]string
		namesInScopeOwn AnnNamesInScope
		namesInScopeExt AnnNamesInScope
		namesInScopeAll AnnNamesInScope
	}
	Errs struct {
		Stage1DirAccessDuringRefresh *Error
		Stage1BadImports             Errors
	}
}

type IrDefRef struct {
	*IrDef
	Kit *Kit
}
