package builder

type BuildArgs struct {
	Zip bool
}

type BuildError struct {
	Message   string
	Err  error
}

func (e *BuildError) Error() string { return e.Message + " " + e.Err.Error() }
