package builder
import (
	"os"
	"io/ioutil"
	"log"
	"gopkg.in/yaml.v2"
	"strings"
	"fmt"
	"path/filepath"
	"github.com/blablacar/cnt/types"
	"github.com/blablacar/cnt/runner"
	"github.com/blablacar/cnt/utils"
	"github.com/blablacar/cnt/config"
)

const (
	targetRootfs = "/target/rootfs"
	target = "/target"
	targetManifest = "/target/manifest"

	buildScript = `#!/bin/bash
set -x
set -e
export TARGET=/target
export ROOTFS=%%ROOTFS%%

execute_files() {
  fdir=$1
  [ -d "$fdir" ] || return

  find "$fdir" -mindepth 1 -maxdepth 1 -type f -print0 |
  while read -r -d $'\0' file; do
      echo "$file"
      [ -x "$file" ] && "$file"
  done
}


export TERM=xterm
source /etc/profile && env-update

if [ -d "$TARGET/runlevels/build-early" ]; then
	execute_files "$TARGET/runlevels/build-early"
fi


if [ -f "$BUILD_PATH/portage.sh" ]; then
	$BUILD_PATH/portage.sh
fi

if [ -d "$TARGET/runlevels/build-late" ]; then
	execute_files "$TARGET/runlevels/build-late"
fi

`
portageInstall = `#!/bin/bash
set -x
set -e

ln -sf /usr/portage/profiles/default/linux/amd64/13.0 ${TARGET}/etc/portage/make.profile
#emerge --sync
ROOT=${ROOTFS} emerge -v --config-root=${TARGET}/ %%INSTALL%%

`

//	buildScript = `#!/bin/bash -x
//	echo $1
//	BUILD_PATH=/builds/${1}
//#export TERM=vt100
//source /etc/profile && env-update
//ln -sf /usr/portage/profiles/default/linux/amd64/13.0 ${BUILD_PATH}/etc/portage/make.profile
//emerge-webrsync
//
//	if [ -f "$BUILD_PATH/install.sh" ]; then
//		$BUILD_PATH/install.sh
//	fi
//
//ROOT=/builds/${1}/rootfs emerge -v --config-root=${BUILD_PATH}/ %%INSTALL%%
///builds/${1}/install-portage.sh
//`

	makeConf = `
USE="-doc static static-libs %%USE%%"
FEATURES="nodoc noinfo noman %%FEATURES%%"
`
)

type Cnt struct {
	path     string
	manifest CntManifest
	args     BuildArgs
}
type CntBuild struct {
	image string                `yaml:"image,omitempty"`
}

func (b *CntBuild) NoBuildImage() bool {
	return b.image == ""
}

type CntManifest struct {
	ProjectName types.ProjectName       `yaml:"projectName,omitempty"`
	Version     string                  `yaml:"version,omitempty"`
	Build       CntBuild                `yaml:"build,omitempty"`
	Portage     struct {
					Use      string                `yaml:"use,omitempty"`
					Mask     string                `yaml:"mask,omitempty"`
					Install  string                `yaml:"install"`
					Features string                `yaml:"features,omitempty"`
					Target   string          	   `yaml:"target,omitempty"`
				}                       `yaml:"portage,omitempty"`
}


////////////////////////////////////////////

func OpenCnt(path string, args BuildArgs) (*Cnt, error) {
	cnt := new(Cnt)
	path, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	cnt.path = path
	cnt.args = args
	return cnt, nil
}


func (cnt *Cnt) Push() {
	cnt.readManifest("/target/cnt-manifest.yml")
	fmt.Printf("%#v\n\n", cnt)
	utils.ExecCmd("curl", "-i",
		"-F", "r=releases",
		"-F", "hasPom=false",
		"-F", "e=aci",
		"-F", "g=com.blablacar.aci.linux.amd64",
		"-F", "p=aci",
		"-F", "v=" + cnt.manifest.Version,
		"-F", "a=" + cnt.manifest.ProjectName.ShortName(),
		"-F", "file=@" + cnt.path + "/target/image.aci",
		"-u", config.GetConfig().Push.Username + ":" + config.GetConfig().Push.Password,
		config.GetConfig().Push.Url + "/service/local/artifact/maven/content")
}

func (cnt *Cnt) writeCntManifest() {
	d, _ := yaml.Marshal(&cnt.manifest)
	ioutil.WriteFile(cnt.path + "/target/cnt-manifest.yml", []byte(d), 0777)
}

func (cnt *Cnt) readManifest(path string) {
	source, err := ioutil.ReadFile(cnt.path + path)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(source), &cnt.manifest)
	if err != nil {
		panic(err)
	}
}

func (cnt *Cnt) Build(runner runner.Runner) {
	println("Building cnt")
	cnt.readManifest("/cnt-manifest.yml")

	log.Println("building ACI")

	os.Mkdir(cnt.path + target, 0777)

	cnt.runlevelBuildSetup()
	cnt.copyRunlevelsBuild()
	cnt.copyRunlevelsPrestart()
	cnt.copyAttributes()
	cnt.copyRootfs()
	cnt.copyConfd()
	cnt.copyInstallAndCreatePacker()

	cnt.writeBuildScript()
	cnt.writePortage()
	cnt.writeRktManifest()
	cnt.writeCntManifest() // TODO move that, here because we update the version number to generated version

	cnt.runPacker()
	cnt.runPortage(runner)

	cnt.tarAci()
	//	ExecCmd("chown " + os.Getenv("SUDO_USER") + ": " + target + "/*") //TODO chown
}

func (cnt *Cnt) copyRunlevelsBuild() {
	if err := os.MkdirAll(cnt.path + target + "/runlevels", 0755); err != nil {
		panic(err)
	}
	utils.CopyDir(cnt.path + "/runlevels/build-early", cnt.path + target + "/runlevels/build-early")
	utils.CopyDir(cnt.path + "/runlevels/build-late", cnt.path + target + "/runlevels/build-late")
}

func (cnt *Cnt) runlevelBuildSetup() {
	files, err := ioutil.ReadDir(cnt.path + "/runlevels/build-setup") // already sorted
	if err != nil {
		return
	}

	os.Setenv("TARGET", cnt.path + target)
	for _, f := range files {
		if !f.IsDir() {
			if err := utils.ExecCmd(cnt.path + "/runlevels/build-setup/" + f.Name()); err != nil {
				panic(err)
			}
		}
	}
}

func (cnt *Cnt) tarAci() {
	dir, _ := os.Getwd();
	log.Println("chdir to", cnt.path, target)
	os.Chdir(cnt.path + target);

	utils.Tar(cnt.args.Zip, "image.aci", "manifest", "rootfs/")

	log.Println("chdir to", dir)
	os.Chdir(dir);
}

func (cnt *Cnt) runPacker() {
	if _, err := os.Stat(cnt.path + target + "/packer.json"); os.IsNotExist(err) {
		return
	}

	dir, _ := os.Getwd();
	os.Chdir(cnt.path + target);
	defer os.Chdir(dir);
	utils.ExecCmd("packer", "build", "packer.json");

	if err := os.Chdir(cnt.path + target + "/rootfs"); err != nil {
		panic(err)
	}
	utils.ExecCmd("tar", "xf", "../rootfs.tar")
}

func (cnt *Cnt) copyInstallAndCreatePacker() {
	if _, err := os.Stat(cnt.path + "/install.sh"); err == nil {
		utils.CopyFile(cnt.path + "/install.sh", target + "/install.sh")
		utils.WritePackerFiles(cnt.path + target)
	}
}

func (cnt *Cnt) copyRunlevelsPrestart() {
	if err := os.MkdirAll(cnt.path + targetRootfs + "/etc/prestart/late-prestart.d", 0755); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(cnt.path + targetRootfs + "/etc/prestart/early-prestart.d", 0755); err != nil {
		panic(err)
	}
	utils.CopyDir(cnt.path + "/runlevels/prestart-early", cnt.path + targetRootfs + "/etc/prestart/early-prestart.d")
	utils.CopyDir(cnt.path + "/runlevels/prestart-late", cnt.path + targetRootfs + "/etc/prestart/late-prestart.d")
}

func (cnt *Cnt) copyConfd() {
	if err := os.MkdirAll(cnt.path + targetRootfs + "/etc/prestart/", 0755); err != nil {
		panic(err)
	}
	utils.CopyDir(cnt.path + "/confd/conf.d", cnt.path + targetRootfs + "/etc/prestart/conf.d")
	utils.CopyDir(cnt.path + "/confd/templates", cnt.path + targetRootfs + "/etc/prestart/templates")
}

func (cnt *Cnt) copyRootfs() {
	utils.CopyDir(cnt.path + "/rootfs", cnt.path + targetRootfs)
}

func (cnt *Cnt) copyAttributes() {
	if err := os.MkdirAll(cnt.path + targetRootfs + "/etc/prestart/attributes/" + cnt.manifest.ProjectName.ShortName(), 0755); err != nil {
		panic(err)
	}
	utils.CopyDir(cnt.path + "/attributes", cnt.path + targetRootfs + "/etc/prestart/attributes/" + cnt.manifest.ProjectName.ShortName())
}

func (cnt *Cnt) writeBuildScript() {
	targetFull := cnt.path + target
	rootfs := "/target/rootfs"
	if cnt.manifest.Build.NoBuildImage() {
		rootfs = ""
	}
	build := strings.Replace(buildScript, "%%ROOTFS%%", rootfs, 1)
	ioutil.WriteFile(targetFull + "/build.sh", []byte(build), 0777)
}

func (cnt *Cnt) writePortage() {
	if cnt.manifest.Portage.Install == "" {
		return
	}
	targetFull := cnt.path + target

	portage := strings.Replace(portageInstall, "%%INSTALL%%", cnt.manifest.Portage.Install, 1)
	ioutil.WriteFile(targetFull + "/portage.sh", []byte(portage), 0777)

	os.MkdirAll(targetFull + "/etc/portage", 0755)
	res := strings.Replace(makeConf, "%%USE%%", cnt.manifest.Portage.Use, 1)
	res = strings.Replace(res, "%%FEATURES%%", cnt.manifest.Portage.Features, 1)

	ioutil.WriteFile(targetFull + "/etc/portage/make.conf", []byte(res), 0777)
}

func (cnt *Cnt) writeRktManifest() {
	im := utils.BasicManifest()
	if cnt.manifest.Version == "" {
		cnt.manifest.Version = utils.GenerateVersion()
	}
	utils.WriteImageManifest(im, cnt.path + targetManifest, cnt.manifest.ProjectName, cnt.manifest.Version)
}

func (cnt *Cnt) runPortage(runner runner.Runner) {
	if _, err := os.Stat(cnt.path + "/target/build.sh"); os.IsNotExist(err) {
		return
	}

	runner.Prepare(cnt.path + target)
	runner.Run(cnt.path + target, cnt.manifest.ProjectName.ShortName(), "/target/build.sh")
	runner.Release(cnt.path + target, cnt.manifest.ProjectName.ShortName(), cnt.manifest.Build.NoBuildImage())
}
