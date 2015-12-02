package builder

type BuildArgs struct {
	Clean      bool
	Test       bool
	NoTestFail bool
	Path       string
}

type BuildError struct {
	Message string
	Err     error
}

func (e *BuildError) Error() string { return e.Message + " " + e.Err.Error() }
