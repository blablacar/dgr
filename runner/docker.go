package runner
import (
	"fmt"
//	"github.com/mitchellh/packer/builder/docker"
	"github.com/blablacar/cnt/utils"
	"bytes"
	"os/exec"
	"os"
	"io"
	"strings"
)


type DockerRunner struct {
	imageId string
}

func (r *DockerRunner) Prepare(target string) {
	fmt.Printf("Prepare Docker\n");
 	first := exec.Command("bash", "-c", "cd " + target + "/rootfs" + " && tar cf - .")
	second := exec.Command("docker", "import", "-", "")

	reader, writer := io.Pipe()
	first.Stdout = writer
	second.Stdin = reader

	var buff bytes.Buffer
	second.Stdout = &buff

	first.Start()
	second.Start()
	first.Wait()
	writer.Close()
	second.Wait()

	r.imageId = strings.TrimSpace(buff.String())
}

//	id, err := utils.ExecCmdGetOutput("bash", "-c", "cd " + target + "/rootfs" + " && tar cf - . | docker import - 'test4242'")
//
//////	id, err := r.Import(target + "/rootfs", "")
//	if err != nil {
//		panic(err)
//	}


//#blablabaseExists=$(docker images | grep blablabase | wc -l)
//#if [ ${blablabaseExists} -ne 1 ]; then
//#  curl ${GTOO_URL} | bzcat | docker import - 'blablabase'
//#fi
//#
//#docker build -t blablabuild .


func (r *DockerRunner) Run(target string, imageName string, command ...string) {
	fmt.Printf("Run Docker\n");
	cmd := []string {"run", "--name=" + imageName, "-v", target + ":/target", r.imageId}
	cmd = append(cmd, command...)
	if err := utils.ExecCmd("docker", cmd...); err != nil {
		panic(err)
	}
}

func (r *DockerRunner) Release(target string, imageName string, noBuildImage bool) {
	fmt.Printf("Release Docker");
	if noBuildImage {
		os.RemoveAll(target + "/rootfs")
		os.Mkdir(target + "/rootfs", 0777)

		if err := utils.ExecCmd("docker", "export", "-o", target + "/dockerfs.tar", imageName); err != nil {
			panic(err)
		}

		utils.ExecCmd("tar", "xpf", target + "/dockerfs.tar", "-C", target + "/rootfs")
	}
	if err := utils.ExecCmd("docker", "rm", imageName); err != nil {
		panic(err)
	}
	if err := utils.ExecCmd("docker", "rmi", r.imageId); err != nil {
		panic(err)
	}
}
