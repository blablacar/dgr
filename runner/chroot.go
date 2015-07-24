package runner
import (
	"github.com/blablacar/cnt/utils"
	"os"
	"log"
	"github.com/appc/spec/discovery"
	"os/exec"
	"runtime"
	"github.com/blablacar/cnt/types"
	"strings"
)

type ChrootRunner struct {
	tmpBuildPath string
	newRoot      string
}

func (r *ChrootRunner) Prepare(target string, buildImage types.AciName) {
	homeAci := utils.UserHomeOrFatal() + "/.config/cnt/aci"
	os.MkdirAll(homeAci, 0777)

	r.newRoot = target + "/rootfs"

	if buildImage != "" {
		r.newRoot = homeAci + "/" + buildImage.String() + "/rootfs"
		if _, err := os.Stat(r.newRoot); os.IsNotExist(err) {
			url := r.findBuildAciUrl(buildImage.String())
			aciPath := homeAci + "/" + buildImage.ShortName() + ".aci"
			utils.ExecCmd("wget", "-O", aciPath, url)
			r.umount(r.newRoot)
			os.RemoveAll(r.newRoot)
			os.MkdirAll(r.newRoot, 0777)
			utils.ExecCmd("tar", "xpf", aciPath, "-C", homeAci + "/" + buildImage.String())
			os.Remove(aciPath)
		}
	}

	r.mountChroot(r.newRoot)
	utils.CopyFile("/etc/resolv.conf", r.newRoot + "/etc/resolv.conf")
}

func (r *ChrootRunner) Run(targetFullPath string, imageName string, command ...string) {
	log.Println("Running command ", strings.Join(command, " "))
	r.tmpBuildPath = "/builds/" + utils.RandSeq(10)
	os.MkdirAll(r.newRoot + r.tmpBuildPath, 0755)
	defer os.Remove(r.tmpBuildPath)
	utils.ExecCmd("mount", "-o", "bind", targetFullPath, r.newRoot + r.tmpBuildPath)
	defer utils.ExecCmd("umount", r.newRoot + r.tmpBuildPath)    //TODO use signal capture to cleanup in case of CTRL + C

	if err := utils.ExecCmd("chroot", r.newRoot, r.tmpBuildPath + "/build.sh"); err != nil {
		panic(err)
	}
}

func (r *ChrootRunner) Release(target string, imageName string, noBuildImage bool) {
	log.Println("Unmounting elements in rootfs ", r.newRoot)
	r.umount(r.newRoot)

}

//////////////////////////////////////////////

func (r *ChrootRunner) umount(newRoot string) {
	utils.ExecCmd("umount", newRoot + "/dev/pts")
	utils.ExecCmd("umount", newRoot + "/dev/shm")
	utils.ExecCmd("umount", newRoot + "/dev/")
	utils.ExecCmd("umount", newRoot + "/sys/")
	utils.ExecCmd("umount", newRoot + "/proc/")
}

func (r *ChrootRunner) findBuildAciUrl(buildImage string) string {
	app, err := discovery.NewAppFromString(buildImage)
	if app.Labels["os"] == "" {
		app.Labels["os"] = runtime.GOOS
	}
	if app.Labels["arch"] == "" {
		app.Labels["arch"] = runtime.GOARCH
	}

	endpoint, _, err := discovery.DiscoverEndpoints(*app, false)
	if err != nil {
		panic(err)
	}

	return endpoint.ACIEndpoints[0].ACI
}

func (r *ChrootRunner) isMounted(path string) bool {
	err := utils.ExecCmd("bash", "-c", "grep '" + path + "' /proc/mounts")
	if _, ok := err.(*exec.ExitError); ok {
		return false;
		//		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
		//			log.Printf("Exit Status: %d", status.ExitStatus())
		//		}
	}
	return true
}

func (r *ChrootRunner) mountChroot(newRoot string) {
	log.Println("Mounting elements in rootfs ", newRoot)
	if !r.isMounted(newRoot + "/proc") {
		utils.ExecCmd("mount", "-t", "proc", "proc", newRoot + "/proc")
	}
	if !r.isMounted(newRoot + "/sys") {
		utils.ExecCmd("mount", "-t", "sysfs", "sys", newRoot + "/sys")
	}
	if !r.isMounted(newRoot + "/dev") {
		utils.ExecCmd("mount", "-o", "bind", "/dev", newRoot + "/dev")
	}
	if !r.isMounted(newRoot + "/dev/shm") {
		utils.ExecCmd("mount", "-o", "bind", "/dev/shm", newRoot + "/dev/shm")
	}
	if !r.isMounted(newRoot + "/dev/pts") {
		utils.ExecCmd("mount", "-t", "devpts", "devpts", newRoot + "/dev/pts")
	}
}
