package atmosess

const (
	_                                     = iota
	ErrSessInit_IoCacheDirCreationFailure = iota + 3100
	ErrSessInit_IoCacheDirDeletionFailure
	ErrSessInit_KitsDirsConflict
	ErrSessInit_KitsDirsNotSpecified
	ErrSessInit_KitsDirsNotFound
	ErrSessInit_KitsDirAutoNotFound
	ErrSessInit_IoFauxKitDirFailure
)

const (
	_                            = iota
	ErrSessKits_IoReadDirFailure = iota + 3200
	ErrSessKits_ImportNotFound
)
