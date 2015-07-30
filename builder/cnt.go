package builder
import (
	"os"
	"io/ioutil"
	"strings"
	"path/filepath"
	"github.com/blablacar/cnt/utils"
	"github.com/blablacar/cnt/config"
	"github.com/blablacar/cnt/log"
	"github.com/appc/spec/schema"
	"bytes"
	"github.com/blablacar/cnt/bats"
	"github.com/ghodss/yaml"
	"github.com/appc/spec/schema/types"
	"os/exec"
	"io"
)

const (
	buildScript = `#!/bin/bash
set -x
set -e
export TARGET=$( dirname $0 )
export ROOTFS=%%ROOTFS%%
export TERM=xterm

execute_files() {
  fdir=$1
  [ -d "$fdir" ] || return 0


  find "$fdir" -mindepth 1 -maxdepth 1 -type f -print0 |
  while read -r -d $'\0' file; do
      echo "Running : $file"
      [ -x "$file" ] && "$file"
  done
}

execute_files "$TARGET/runlevels/inherit-build-early"
execute_files "$TARGET/runlevels/build"
execute_files "$TARGET/runlevels/inherit-build-late"`
)

type Cnt struct {
	path     string
	target   string
	rootfs   string
	manifest CntManifest
	args     BuildArgs
}
type CntBuild struct {
	Image types.ACIdentifier                `json:"image"`
}

func (b *CntBuild) NoBuildImage() bool {
	return b.Image == ""
}

type CntManifest struct {
	Name    types.ACIdentifier            `json:"name"`
	Version string                        `json:"version"`
	From    string                      `json:"from"`
	Build   CntBuild                    `json:"build"`
	Aci     schema.ImageManifest        `json:"aci"`
}

func ShortName(name types.ACIdentifier) string {
	split := strings.Split(string(name), "/")
	return split[1]
}

////////////////////////////////////////////

func OpenCnt(path string, args BuildArgs) (*Cnt, error) {
	cnt := new(Cnt)
	cnt.args = args

	if fullPath, err := filepath.Abs(path); err != nil {
		log.Get().Panic("Cannot get fullpath of project", err)
	} else {
		cnt.path = fullPath
		cnt.target = cnt.path + "/target"
		cnt.rootfs = cnt.target + "/rootfs"
	}

	if _, err := os.Stat(cnt.path + "/cnt-manifest.yml"); os.IsNotExist(err) {
		log.Get().Debug(cnt.path, "/cnt-manifest.yml does not exists")
		return nil, &BuildError{"file not found : " + cnt.path + "/cnt-manifest.yml", err}
	}

	cnt.manifest.Aci = *utils.BasicManifest()
	cnt.readManifest(cnt.path + "/cnt-manifest.yml")

	return cnt, nil
}

func (cnt *Cnt) Clean() {
	log.Get().Info("Cleaning " + cnt.manifest.Aci.Name)
	if err := os.RemoveAll(cnt.target + "/"); err != nil {
		log.Get().Panic("Cannot clean " + cnt.manifest.Aci.Name, err)
	}
}

func (cnt *Cnt) Test() {
	log.Get().Info("Testing " + cnt.manifest.Aci.Name)
	if _, err := os.Stat(cnt.target + "/image.aci"); os.IsNotExist(err) {
		if err := cnt.Build(); err != nil {
			log.Get().Panic("Cannot Install since build failed")
		}
	}

	// BATS
	os.MkdirAll(cnt.target + "/test", 0777)
	bats.WriteBats(cnt.target + "/test")
	utils.ExecCmd("rkt", "--insecure-skip-verify=true", "run", cnt.target + "/image.aci") // TODO missing command override that will arrive in next RKT version
}

func (cnt *Cnt) Push() {
	if config.GetConfig().Push.Type == "" {
		log.Get().Panic("Can't push, push is not configured in cnt global configuration file")
	}

	cnt.readManifest(cnt.target + "/cnt-manifest.yml")

	val, _ := cnt.manifest.Aci.Labels.Get("version")
	utils.ExecCmd("curl", "-i",
		"-F", "r=releases",
		"-F", "hasPom=false",
		"-F", "e=aci",
		"-F", "g=com.blablacar.aci.linux.amd64",
		"-F", "p=aci",
		"-F", "v=" + val,
		"-F", "a=" + ShortName(cnt.manifest.Aci.Name),
		"-F", "file=@" + cnt.target + "/image.aci",
		"-u", config.GetConfig().Push.Username + ":" + config.GetConfig().Push.Password,
		config.GetConfig().Push.Url + "/service/local/artifact/maven/content")
}

func (cnt *Cnt) writeCntManifest() {
	d, _ := yaml.Marshal(&cnt.manifest)
	ioutil.WriteFile(cnt.path + "/target/cnt-manifest.yml", []byte(d), 0777)
}

func (cnt *Cnt) readManifest(manifestPath string) {
	source, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		log.Get().Panic(err)
	}
	err = yaml.Unmarshal([]byte(source), &cnt.manifest)
	if err != nil {
		log.Get().Panic(err)
	}

	cnt.manifest.Aci.Name.Set(string(cnt.manifest.Name))
	changeVersion(&cnt.manifest.Aci.Labels, cnt.manifest.Version)

	log.Get().Trace("Cnt manifest : ", cnt.manifest.Aci.Name, cnt.manifest, cnt.manifest.Aci.App)
}

func changeVersion(labels *types.Labels, version string) {
	labelMap := labels.ToMap()
	labelMap["version"] = version
	if newLabels, err := types.LabelsFromMap(labelMap); err != nil {
		log.Get().Panic(err)
	} else {
		*labels = newLabels
	}
}

func (cnt *Cnt) Install() {
	if _, err := os.Stat(cnt.target + "/image.aci"); os.IsNotExist(err) {
		if err := cnt.Build(); err != nil {
			log.Get().Panic("Cannot Install since build failed")
		}
	}
	utils.ExecCmd("rkt", "--insecure-skip-verify=true", "fetch", cnt.target + "/image.aci")
}

func (cnt *Cnt) Build() error {
	log.Get().Info("Building Image : ", cnt.manifest.Aci.Name)

	os.Mkdir(cnt.target, 0777)

	cnt.processFrom()

	cnt.runlevelBuildSetup()
	cnt.copyRunlevelsBuild()
	cnt.copyRunlevelsPrestart()
	cnt.copyAttributes()
	cnt.copyFiles()
	cnt.copyConfd()
	cnt.copyInstallAndCreatePacker()

	cnt.writeBuildScript()
	cnt.writeRktManifest()
	cnt.writeCntManifest() // TODO move that, here because we update the version number to generated version

	cnt.runBuild()
	cnt.runPacker()

	cnt.tarAci()
	//	ExecCmd("chown " + os.Getenv("SUDO_USER") + ": " + target + "/*") //TODO chown
	return nil
}

func (cnt *Cnt) runBuild() {
	if err := utils.ExecCmd("systemd-nspawn", "--version"); err == nil {
		log.Get().Info("Run with systemd-nspawn")
		if err := utils.ExecCmd("systemd-nspawn", "--directory=" + cnt.rootfs, "--capability=all",
			"--bind=" + cnt.target + "/:/target", "--share-system", "target/build.sh"); err != nil {
			log.Get().Panic("Build step did not succeed", err)
		}
	} else {
		log.Get().Info("Run with docker")

		//
		log.Get().Info("Prepare Docker");
		first := exec.Command("bash", "-c", "cd " + cnt.rootfs + " && tar cf - .")
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
		imgId := strings.TrimSpace(buff.String())

		//
		log.Get().Info("Run Docker\n");
		cmd := []string{"run", "--name=" + ShortName(cnt.manifest.Name), "-v", cnt.target + ":/target", imgId, "/target/build.sh"}
		utils.ExecCmd("docker", "rm", ShortName(cnt.manifest.Name))
		if err := utils.ExecCmd("docker", cmd...); err != nil {
			panic(err)
		}

		//
		log.Get().Info("Release Docker");
		if cnt.manifest.Build.NoBuildImage() {
			os.RemoveAll(cnt.rootfs)
			os.Mkdir(cnt.rootfs, 0777)

			if err := utils.ExecCmd("docker", "export", "-o", cnt.target + "/dockerfs.tar", ShortName(cnt.manifest.Name)); err != nil {
				panic(err)
			}

			utils.ExecCmd("tar", "xpf", cnt.target + "/dockerfs.tar", "-C", cnt.rootfs)
		}
		if err := utils.ExecCmd("docker", "rm", ShortName(cnt.manifest.Name)); err != nil {
			panic(err)
		}
		if err := utils.ExecCmd("docker", "rmi", imgId); err != nil {
			panic(err)
		}

	}
}

func (cnt *Cnt) processFrom() {
	if cnt.manifest.From != "" {
		log.Get().Info("Prepare rootfs from " + cnt.manifest.From)
		utils.ExecCmd("rkt", "--insecure-skip-verify=true", "fetch", cnt.manifest.From)
		utils.ExecCmd("rkt", "image", "export", "--overwrite", cnt.manifest.From, cnt.target + "/from.aci")
		utils.ExecCmd("tar", "xf", cnt.target + "/from.aci", "-C", cnt.target)
		os.Remove(cnt.target + "/from.aci")
	}
}


func (cnt *Cnt) copyRunlevelsBuild() {
	if err := os.MkdirAll(cnt.target + "/runlevels", 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path + "/runlevels", cnt.target + "/runlevels")
}

func (cnt *Cnt) runlevelBuildSetup() {
	files, err := ioutil.ReadDir(cnt.path + "/runlevels/build-setup") // already sorted by name
	if err != nil {
		return
	}

	os.Setenv("TARGET", cnt.target)
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
	log.Get().Debug("chdir to", cnt.target)
	os.Chdir(cnt.target);

	args := []string{"manifest", "rootfs/"}

	if _, err := os.Stat(cnt.path + "/runlevels/inherit-build-early"); err == nil {
		args = append(args, "runlevels/inherit-build-early")
	}
	if _, err := os.Stat(cnt.path + "/runlevels/inherit-build-late"); err == nil {
		args = append(args, "runlevels/inherit-build-late")
	}

	utils.Tar(cnt.args.Zip, "image.aci", args...)
	log.Get().Debug("chdir to", dir)
	os.Chdir(dir);
}

func (cnt *Cnt) runPacker() {
	if _, err := os.Stat(cnt.target + "/packer.json"); os.IsNotExist(err) {
		return
	}

	dir, _ := os.Getwd();
	os.Chdir(cnt.target);
	defer os.Chdir(dir);
	utils.ExecCmd("packer", "build", "packer.json");

	if err := os.Chdir(cnt.target + "/rootfs"); err != nil {
		log.Get().Panic(err)
	}
	utils.ExecCmd("tar", "xf", "../rootfs.tar")
}

func (cnt *Cnt) copyInstallAndCreatePacker() {
	if _, err := os.Stat(cnt.path + "/install.sh"); err == nil {
		utils.CopyFile(cnt.path + "/install.sh", cnt.target + "/install.sh")
		sum, _ := utils.ChecksumFile(cnt.target + "/install.sh")
		lastSum, err := ioutil.ReadFile(cnt.target + "/install.sh.SUM")
		if err != nil || !bytes.Equal(lastSum, sum) {
			utils.WritePackerFiles(cnt.target)
			ioutil.WriteFile(cnt.target + "/install.sh.SUM", sum, 0755)
			return
		}
	}
	utils.RemovePackerFiles(cnt.target)
}

func (cnt *Cnt) copyRunlevelsPrestart() {
	if err := os.MkdirAll(cnt.rootfs + "/etc/prestart/late-prestart.d", 0755); err != nil {
		log.Get().Panic(err)
	}
	if err := os.MkdirAll(cnt.rootfs + "/etc/prestart/early-prestart.d", 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path + "/runlevels/prestart-early", cnt.rootfs + "/etc/prestart/early-prestart.d")
	utils.CopyDir(cnt.path + "/runlevels/prestart-late", cnt.rootfs + "/etc/prestart/late-prestart.d")
}

func (cnt *Cnt) copyConfd() {
	if err := os.MkdirAll(cnt.rootfs + "/etc/prestart/", 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path + "/confd/conf.d", cnt.rootfs + "/etc/prestart/conf.d")
	utils.CopyDir(cnt.path + "/confd/templates", cnt.rootfs + "/etc/prestart/templates")
}

func (cnt *Cnt) copyFiles() {
	utils.CopyDir(cnt.path + "/files", cnt.rootfs)
}

func (cnt *Cnt) copyAttributes() {
	if err := os.MkdirAll(cnt.rootfs + "/etc/prestart/attributes/" + ShortName(cnt.manifest.Aci.Name), 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path + "/attributes", cnt.rootfs + "/etc/prestart/attributes/" + ShortName(cnt.manifest.Aci.Name))
}

func (cnt *Cnt) writeBuildScript() {
	rootfs := "${TARGET}/rootfs"
	if cnt.manifest.Build.NoBuildImage() {
		rootfs = ""
	}
	build := strings.Replace(buildScript, "%%ROOTFS%%", rootfs, 1)
	ioutil.WriteFile(cnt.target + "/build.sh", []byte(build), 0777)
}

func (cnt *Cnt) writeRktManifest() {
	log.Get().Debug("Writing aci manifest")
	if val, _ := cnt.manifest.Aci.Labels.Get("version"); val == "" {
		changeVersion(&cnt.manifest.Aci.Labels, utils.GenerateVersion())
	}
	version, _ := cnt.manifest.Aci.Labels.Get("version")
	utils.WriteImageManifest(&cnt.manifest.Aci, cnt.target + "/manifest", cnt.manifest.Aci.Name, version)
}
