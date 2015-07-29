package builder
import (
	"os"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"strings"
	"fmt"
	"path/filepath"
	"github.com/blablacar/cnt/types"
	"github.com/blablacar/cnt/runner"
	"github.com/blablacar/cnt/utils"
	"github.com/blablacar/cnt/config"
	"github.com/blablacar/cnt/log"
	"github.com/appc/spec/schema"
	"bytes"
)

const (
	targetRootfs = "/target/rootfs"
	target = "/target"
	targetManifest = "/target/manifest"

	buildScript = `#!/bin/bash
set -x
set -e
export TARGET=$( dirname $0 )
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


if [ -f "$TARGET/portage.sh" ]; then
	$TARGET/portage.sh
fi

if [ -d "$TARGET/runlevels/build-late" ]; then
	execute_files "$TARGET/runlevels/build-late"
fi

`
portageInstall = `#!/bin/bash
set -x
set -e

ROOTFS=/build/root

mkdir -p /build/{root,db-pkg/.cache/{names,provides}}
mkdir ${ROOTFS}/{dev,etc}
mknod -m 622 ${ROOTFS}/dev/console c 5 1
mknod -m 666 ${ROOTFS}/dev/null c 1 3
mknod -m 666 ${ROOTFS}/dev/zero c 1 5
mknod -m 444 ${ROOTFS}/dev/random c 1 8
mknod -m 444 ${ROOTFS}/dev/urandom c 1 9
touch ${ROOTFS}/etc/ld.so.cache
mkdir ${ROOTFS}/lib64
cd ${ROOTFS}
ln -s lib64 lib

echo "alias cave-i='cave resume --resume-file ~/.cave-resume'" >> /etc/bash/bashrc
echo "alias cave-p='cave resolve --resume-file ~/.cave-resume'" >> /etc/bash/bashrc
echo "alias cave-chroot=\"cave resolve --resume-file ~/.cave-resume -mc -/b --chroot-path /build/root/ -0 '*/*::installed'\"" >> /etc/bash/bashrc



ln -sf /usr/portage/profiles/default/linux/amd64/13.0 ${TARGET}/etc/portage/make.profile
#emerge-webrsync
if [ %%ENTER%% -eq 1 ]; then
	ROOT=${ROOTFS} emerge -vp --config-root=${TARGET}/ %%INSTALL%%
	bash
else
	ROOT=${ROOTFS} emerge -v --config-root=${TARGET}/ %%INSTALL%%
fi
`

	makeConf = `
USE="-doc %%USE%%"
FEATURES="nodoc noinfo noman %%FEATURES%%"
`
)

type Cnt struct {
	path     string
	manifest CntManifest
	rkt		 *schema.ImageManifest
	args     BuildArgs
}
type CntBuild struct {
	Image types.AciName                `yaml:"image,omitempty"`
}

func (b *CntBuild) NoBuildImage() bool {
	return b.Image == ""
}

type CntManifest struct {
	ProjectName types.AciName       `yaml:"projectName,omitempty"`
	Version     string                  `yaml:"version,omitempty"`
	Build       CntBuild                `yaml:"build,omitempty"`
	Portage     struct {
					Use      string                `yaml:"use,omitempty"`
					Mask     string                `yaml:"mask,omitempty"`
					Install  string                `yaml:"install"`
					Features string                `yaml:"features,omitempty"`
					Keywords string				   `yaml:"keywords,omitempty"`
				}                       `yaml:"portage,omitempty"`
}

////////////////////////////////////////////

func OpenCnt(path string, args BuildArgs) (*Cnt, error) {
	cnt := new(Cnt)
	cnt.args = args

	if fullPath, err := filepath.Abs(path); err != nil {
		log.Get().Panic("Cannot get fullpath of project", err)
	} else {
		cnt.path = fullPath
	}

	if _, err := os.Stat(cnt.path + "/image-manifest.json"); os.IsNotExist(err) {
		log.Get().Debug(cnt.path, "/image-manifest.json does not exists")
		if _, err := os.Stat(cnt.path + "/cnt-config.yml"); os.IsNotExist(err) {
			log.Get().Debug(cnt.path, "/cnt-config.yml does not exists")
			return nil, &BuildError{"file not found : " + cnt.path + "/cnt-config.yml", err}
		}
	}

	return cnt, nil
}

func (cnt *Cnt) Clean() {
	if err := os.RemoveAll(cnt.path + "/target/"); err != nil {
		log.Get().Panic("Cannot clean " + cnt.manifest.ProjectName, err)
	}
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
		log.Get().Panic(err)
	}
	err = yaml.Unmarshal([]byte(source), &cnt.manifest)
	if err != nil {
		log.Get().Panic(err)
	}
}

func (cnt *Cnt) Install() {
	if _, err := os.Stat(cnt.path + "/target/image.aci"); os.IsNotExist(err) {
		if err := cnt.Build(); err != nil {
			log.Get().Panic("Cannot Install since build failed")
		}
	}
	utils.ExecCmd("rkt", "--insecure-skip-verify=true", "fetch", cnt.path + "/target/image.aci")
}

func (cnt *Cnt) Build() error {
	log.Get().Info("Building Image : ", cnt.manifest.ProjectName)
//	cnt.readManifest("/cnt-manifest.yml")

	cnt.rkt = utils.ReadManifest(cnt.path + "/image-manifest.json")
	cnt.manifest.ProjectName = types.AciName(string(cnt.rkt.Name))

	os.Mkdir(cnt.path + target, 0777)

	cnt.runlevelBuildSetup()
	cnt.copyRunlevelsBuild()
	cnt.copyRunlevelsPrestart()
	cnt.copyAttributes()
	cnt.copyRootfs()
	cnt.copyConfd()
	cnt.copyInstallAndCreatePacker()

	cnt.writeBuildScript()
	cnt.writeRktManifest()
	cnt.writeCntManifest() // TODO move that, here because we update the version number to generated version

	cnt.runPacker()

	cnt.tarAci()
	//	ExecCmd("chown " + os.Getenv("SUDO_USER") + ": " + target + "/*") //TODO chown
	return nil
}



func (cnt *Cnt) copyRunlevelsBuild() {
	if err := os.MkdirAll(cnt.path + target + "/runlevels", 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path + "/runlevels/build-early", cnt.path + target + "/runlevels/build-early")
	utils.CopyDir(cnt.path + "/runlevels/build-late", cnt.path + target + "/runlevels/build-late")
}

func (cnt *Cnt) runlevelBuildSetup() {
	files, err := ioutil.ReadDir(cnt.path + "/runlevels/build-setup") // already sorted by name
	if err != nil {
		return
	}

	os.Setenv("TARGET", cnt.path + target)
	for _, f := range files {
		if !f.IsDir() {
			log.Get().Info("Running Build setup level : ", f.Name())
			if err := utils.ExecCmd(cnt.path + "/runlevels/build-setup/" + f.Name()); err != nil {
				log.Get().Panic(err)
			}
		}
	}
}

func (cnt *Cnt) tarAci() {
	dir, _ := os.Getwd();
	log.Get().Debug("chdir to", cnt.path, target)
	os.Chdir(cnt.path + target);

	utils.Tar(cnt.args.Zip, "image.aci", "manifest", "rootfs/")
	log.Get().Debug("chdir to", dir)
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
		log.Get().Panic(err)
	}
	utils.ExecCmd("tar", "xf", "../rootfs.tar")
}

func (cnt *Cnt) copyInstallAndCreatePacker() {
	if _, err := os.Stat(cnt.path + "/install.sh"); err == nil {
		utils.CopyFile(cnt.path + "/install.sh", cnt.path + target + "/install.sh")
		sum, _ := utils.ChecksumFile(cnt.path + target + "/install.sh")
		lastSum, err := ioutil.ReadFile(cnt.path + target + "/install.sh.SUM")
		if err != nil || !bytes.Equal(lastSum, sum) {
			utils.WritePackerFiles(cnt.path + target)
			ioutil.WriteFile(cnt.path + target + "/install.sh.SUM", sum, 0755)
			return
		}
	}
	utils.RemovePackerFiles(cnt.path + target)
}

func (cnt *Cnt) copyRunlevelsPrestart() {
	if err := os.MkdirAll(cnt.path + targetRootfs + "/etc/prestart/late-prestart.d", 0755); err != nil {
		log.Get().Panic(err)
	}
	if err := os.MkdirAll(cnt.path + targetRootfs + "/etc/prestart/early-prestart.d", 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path + "/runlevels/prestart-early", cnt.path + targetRootfs + "/etc/prestart/early-prestart.d")
	utils.CopyDir(cnt.path + "/runlevels/prestart-late", cnt.path + targetRootfs + "/etc/prestart/late-prestart.d")
}

func (cnt *Cnt) copyConfd() {
	if err := os.MkdirAll(cnt.path + targetRootfs + "/etc/prestart/", 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path + "/confd/conf.d", cnt.path + targetRootfs + "/etc/prestart/conf.d")
	utils.CopyDir(cnt.path + "/confd/templates", cnt.path + targetRootfs + "/etc/prestart/templates")
}

func (cnt *Cnt) copyRootfs() {
	utils.CopyDir(cnt.path + "/files", cnt.path + targetRootfs)
}

func (cnt *Cnt) copyAttributes() {
	if err := os.MkdirAll(cnt.path + targetRootfs + "/etc/prestart/attributes/" + cnt.manifest.ProjectName.ShortName(), 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path + "/attributes", cnt.path + targetRootfs + "/etc/prestart/attributes/" + cnt.manifest.ProjectName.ShortName())
}

func (cnt *Cnt) writeBuildScript() {
	targetFull := cnt.path + target
	rootfs := "${TARGET}/rootfs"
	if cnt.manifest.Build.NoBuildImage() {
		rootfs = ""
	}
	build := strings.Replace(buildScript, "%%ROOTFS%%", rootfs, 1)
	ioutil.WriteFile(targetFull + "/build.sh", []byte(build), 0777)
}

func (cnt *Cnt) writeRktManifest() {
	log.Get().Debug("Writing aci manifest")
	im := cnt.rkt
	if cnt.manifest.Version == "" {
		cnt.manifest.Version = utils.GenerateVersion()
	}
	utils.WriteImageManifest(im, cnt.path + targetManifest, cnt.manifest.ProjectName, cnt.manifest.Version)
}

func (cnt *Cnt) runPortage(runner runner.Runner) {
	if _, err := os.Stat(cnt.path + "/target/build.sh"); os.IsNotExist(err) {
		return
	}

	runner.Prepare(cnt.path + target, cnt.manifest.Build.Image)
	runner.Run(cnt.path + target, cnt.manifest.ProjectName.ShortName(), "/target/build.sh")
	runner.Release(cnt.path + target, cnt.manifest.ProjectName.ShortName(), cnt.manifest.Build.NoBuildImage())
}
