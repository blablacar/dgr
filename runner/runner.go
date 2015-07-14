package runner

type Runner interface {
	Prepare(target string)

	Run(target string, command ...string)

	Release()
}
