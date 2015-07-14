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
echo $1
export TARGET=/target

#export TERM=vt100
source /etc/profile && env-update
ln -sf /usr/portage/profiles/default/linux/amd64/13.0 ${TARGET}/etc/portage/make.profile

emerge --sync

if [ -f "$BUILD_PATH/install.sh" ]; then
	$BUILD_PATH/install.sh
fi

ROOT=${TARGET}/rootfs emerge -v --config-root=${TARGET}/ %%INSTALL%%
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

type PortageTarget int

type CntManifest struct {
	ProjectName types.ProjectName       `yaml:"projectName,omitempty"`
	Version     string                  `yaml:"version,omitempty"`
	Build		struct {
					BaseImage	string                `yaml:"baseImage,omitempty"`
				 }						`yaml:"build,omitempty"`
	Portage     struct {
					Use      	string                `yaml:"use,omitempty"`
					Mask     	string                `yaml:"mask,omitempty"`
					Install  	string                `yaml:"install"`
					Features 	string                `yaml:"features,omitempty"`
					Target		string		  `yaml:"target,omitempty"`
				}                       `yaml:"portage,omitempty"`
}

func (cnt *Cnt) Push() {
	cnt.readManifest("/target/cnt-manifest.yml")
	fmt.Printf("%#v\n\n", cnt)
	utils.ExecCmd("curl", "-i",
		"-F","r=releases",
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

	cnt.runBuildSetup()
	cnt.copyPrestart()
	cnt.copyAttributes()
	cnt.copyRootfs()
	cnt.copyConfd()
	cnt.copyInstallAndCreatePacker()

	cnt.writePortage()
	cnt.writeRktManifest()
	cnt.writeCntManifest() // TODO move that, here because we update the version number to generated version

	cnt.runPacker()
	cnt.runPortage(runner)

	cnt.tarAci()
}

func (cnt *Cnt) runBuildSetup() {
//	files, err := utils.ExecCmdGetOutput("find", "~", "-mindepth", "1", "-maxdepth", "1", "-type", "f")
		if _, err := os.Stat(cnt.path + "/setup"); os.IsNotExist(err) {
		return
	}
	if err := utils.ExecCmd(cnt.path + "/setup"); err != nil {
		panic(err)
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

func (cnt *Cnt) copyPrestart() {
	if err := os.MkdirAll(cnt.path + targetRootfs + "/etc/prestart/late-prestart.d", 0755); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(cnt.path + targetRootfs + "/etc/prestart/early-prestart.d", 0755); err != nil {
		panic(err)
	}
	utils.CopyDir(cnt.path + "/prestart/early", cnt.path + targetRootfs + "/etc/prestart/early-prestart.d")
	utils.CopyDir(cnt.path + "/prestart/late", cnt.path + targetRootfs + "/etc/prestart/late-prestart.d")
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

func (cnt *Cnt) writePortage() {
	if _, err := os.Stat(cnt.path + "/install-portage.sh"); os.IsNotExist(err) {
		return
	}
	targetFull := cnt.path + target

	fmt.Printf("---- %#v\n\n", cnt.manifest)

	build := strings.Replace(buildScript, "%%INSTALL%%", cnt.manifest.Portage.Install, 1)
	ioutil.WriteFile(targetFull + "/build.sh", []byte(build), 0777)
	utils.CopyFile(cnt.path + "/install-portage.sh", targetFull + "/install-portage.sh")

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
	runner.Run(cnt.path + target, "/target/build.sh")
	runner.Release()
}
