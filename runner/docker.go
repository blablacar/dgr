package runner

//
//import (
//	"bytes"
//	"github.com/blablacar/cnt/utils"
//	"io"
//	"os"
//	"os/exec"
//	"strings"
//)
//
//type DockerRunner struct {
//	imageId string
//}
//
//func (r *DockerRunner) Prepare(target string) error {
//	log.Get().Debug("Preparing docker")
//	first := exec.Command("bash", "-c", "cd "+target+"/rootfs"+" && tar cf - .")
//	second := exec.Command("docker", "import", "-", "")
//
//	reader, writer := io.Pipe()
//	first.Stdout = writer
//	second.Stdin = reader
//
//	var buff bytes.Buffer
//	second.Stdout = &buff
//
//	first.Start()
//	second.Start()
//	if err := first.Wait(); err != nil {
//		return err
//	}
//	writer.Close()
//	second.Wait()
//
//	r.imageId = strings.TrimSpace(buff.String())
//	return nil
//}
//
//func (r *DockerRunner) Run(target string, imageName string, command ...string) {
//	log.Get().Debug("Run Docker")
//	cmd := []string{"run", "--name=" + imageName, "-v", target + ":/target", r.imageId, "/target/build.sh"}
//	utils.ExecCmd("docker", "rm", ShortName(cnt.manifest.Name))
//	if err := utils.ExecCmd("docker", cmd...); err != nil {
//		panic(err)
//	}
//}
//
//func (r *DockerRunner) Release(target string, imageName string, noBuildImage bool) {
//	log.Get().Debug("Release Docker")
//	if noBuildImage {
//		os.RemoveAll(target + "/rootfs")
//		os.Mkdir(target+"/rootfs", 0777)
//
//		if err := utils.ExecCmd("docker", "export", "-o", target+"/dockerfs.tar", imageName); err != nil {
//			return err
//		}
//
//		utils.ExecCmd("tar", "xpf", target+"/dockerfs.tar", "-C", target+"/rootfs")
//	}
//	if err := utils.ExecCmd("docker", "rm", imageName); err != nil {
//		return err
//	}
//	if err := utils.ExecCmd("docker", "rmi", r.imageId); err != nil {
//		return err
//	}
//}
