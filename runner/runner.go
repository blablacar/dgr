package runner

type Runner interface {
	Prepare()

	Run()

	Release()
}
