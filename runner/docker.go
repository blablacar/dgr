package runner
import (
	"fmt"
//	"github.com/mitchellh/packer/builder/docker"
)


type DockerRunner struct {
}

func (r DockerRunner) Prepare() {
	fmt.Printf("Prepare Docker");
}

func (r DockerRunner) Run(toExec string) {
//	docker.StepRun{}
	fmt.Printf("Run Docker");
}

func (r DockerRunner) Release() {
	fmt.Printf("Release Docker");
}
