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
	prepared string
}

func (r *DockerRunner) Prepare(target string) {
	fmt.Printf("Prepare Docker");
 	first := exec.Command("bash", "-c", "cd " + target + "/rootfs" + " && tar cf - .")
	second := exec.Command("docker", "import", "-", "test4242")

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

	r.prepared = strings.TrimSpace(buff.String())
}

//	id, err := utils.ExecCmdGetOutput("bash", "-c", "cd " + target + "/rootfs" + " && tar cf - . | docker import - 'test4242'")
//
//////	id, err := r.Import(target + "/rootfs", "")
//	if err != nil {
//		panic(err)
//	}

func (d *DockerRunner) Import(path string, repo string) (string, error) {
	var stdout bytes.Buffer
	cmd := exec.Command("docker", "import", "-", repo)
	cmd.Stdout = &stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	// There should be only one artifact of the Docker builder
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := cmd.Start(); err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		io.Copy(stdin, file)
	}()

	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("Error importing container: %s", err)
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

//#blablabaseExists=$(docker images | grep blablabase | wc -l)
//#if [ ${blablabaseExists} -ne 1 ]; then
//#  curl ${GTOO_URL} | bzcat | docker import - 'blablabase'
//#fi
//#
//#docker build -t blablabuild .


func (r *DockerRunner) Run(target string, command ...string) {
	cmd := []string {"run", "-i", "--rm", "-v", target + ":/target", r.prepared}
	cmd = append(cmd, command...)
	if err := utils.ExecCmd("docker", cmd...); err != nil {
		panic(err)
	}
}

func (r *DockerRunner) Release() {
	fmt.Printf("Release Docker");
}
