package runner
import "github.com/blablacar/cnt/types"

type Runner interface {
	Prepare(targetFullPath string, buildImage types.AciName)

	Run(targetFullPath string, imageName string, command ...string)

	Release(targetFullPath string, imageName string, noBuildImage bool)
}
