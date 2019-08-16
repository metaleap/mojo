package atmosess

const (
	_ = 3100 + iota
	ErrSessInit_IoCacheDirCreationFailure
	ErrSessInit_IoCacheDirDeletionFailure
	ErrSessInit_KitsDirsConflict
	ErrSessInit_KitsDirsNotSpecified
	ErrSessInit_KitsDirsNotFound
	ErrSessInit_KitsDirAutoNotFound
	ErrSessInit_IoFauxKitDirFailure
)
const (
	_ = 3200 + iota
	ErrSessKits_IoReadDirFailure
	ErrSessKits_ImportNotFound
)
const (
	_ = 3300 + iota
	ErrNames_NotFound
	ErrNames_Ambiguous
)
