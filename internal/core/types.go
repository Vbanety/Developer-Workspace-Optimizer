package core

// Finding is the result of scanning a module's target path.
type Finding struct {
	Module    string
	Path      string
	SizeBytes int64
	ItemCount int
}

// CleanResult describes what happened when a module's Clean was invoked.
type CleanResult struct {
	Module     string
	Path       string
	FreedBytes int64
	DryRun     bool
	Skipped    bool
	SkipReason string
}
