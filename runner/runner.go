package runner

type Runner interface {
	Prepare(target string)

	Run(target string, imageName string, command ...string)

	Release(target string, imageName string, noBuildImage bool)
}
