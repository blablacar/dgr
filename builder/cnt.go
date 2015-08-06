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
	"github.com/appc/spec/discovery"
	"github.com/appc/spec/aci"
	"strconv"
	"runtime"
	"regexp"
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

  for file in $fdir/*; do
    if [ -x "$file" ]; then
      $file
    else
      echo "$file is not exectuable"
    fi
  done
}

execute_files "$TARGET/runlevels/inherit-build-early"
execute_files "$TARGET/runlevels/build"
execute_files "$TARGET/runlevels/inherit-build-late"`
)

const MANIFEST = "cnt-manifest.yml"
const RUNLEVELS = "runlevels"
const RUNLEVELS_PRESTART = RUNLEVELS + "/prestart-early"
const RUNLEVELS_LATESTART =  RUNLEVELS + "/prestart-late"
const RUNLEVELS_BUILD =  RUNLEVELS + "/build"
const RUNLEVELS_BUILD_SETUP =  RUNLEVELS + "/build-setup"
const RUNLEVELS_BUILD_INHERIT_EARLY =  RUNLEVELS + "/inherit-build-early"
const RUNLEVELS_BUILD_INHERIT_LATE = RUNLEVELS + "/inherit-build-late"
const CONFD = "confd"
const CONFD_TEMPLATE = CONFD + "/conf.d"
const CONFD_CONFIG = CONFD + "/templates"
const ATTRIBUTES = "attributes"
const FILES_PATH = "files"

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
	NameAndVersion string            `json:"name"`
	From  string                      `json:"from"`
	Build CntBuild                    `json:"build"`
	Aci   schema.ImageManifest        `json:"aci"`
}

func Version(nameAndVersion string) string {
	split := strings.Split(nameAndVersion, ":")
	if (len(split) == 1) {
		return ""
	}
	return split[1]
}

func ShortNameId(name types.ACIdentifier) string {
	return strings.Split(string(name), "/")[1]
}

func ShortName(nameAndVersion string) string {
	return strings.Split(Name(nameAndVersion), "/")[1]
}

func Name(nameAndVersion string) string {
	return strings.Split(nameAndVersion, ":")[0]
}

func getCallerName() string {
	pc, _, _, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name()
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

	notInit,_ := regexp.MatchString("/commands.discoverAndRunInitType/",getCallerName())
	if _, err := os.Stat(cnt.path + "/" + MANIFEST); notInit  && os.IsNotExist(err)  {
		log.Get().Debug(cnt.path, "/"+ MANIFEST +" does not exists")
		return nil, &BuildError{"file not found : " + cnt.path +  "/"+ MANIFEST, err}
	}

	if notInit {
		cnt.manifest.Aci = *utils.BasicManifest()
		cnt.readManifest(cnt.path + "/"+ MANIFEST)
	}

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

func (cnt *Cnt) Init() {
	log.Get().Info("Setting up files three")
	uid := os.Getenv("SUDO_UID")
	uidInt,err := strconv.Atoi(uid)
	gid := os.Getenv("SUDO_GID")
	gidInt ,err := strconv.Atoi(gid)

	if err != nil {
		log.Get().Panic(err)
	}
	folderList := []string{
		RUNLEVELS,
		RUNLEVELS_PRESTART,
		RUNLEVELS_LATESTART,
		RUNLEVELS_BUILD,
		RUNLEVELS_BUILD_SETUP,
		RUNLEVELS_BUILD_INHERIT_EARLY,
		RUNLEVELS_BUILD_INHERIT_LATE,
		CONFD,
		CONFD_TEMPLATE,
		CONFD_CONFIG,
		ATTRIBUTES,
		FILES_PATH,
	}
	for _,folder := range folderList  {
		fpath := cnt.path + "/" +folder
		os.MkdirAll(fpath,0777 )
		os.Lchown(fpath,uidInt,gidInt)
	}
}

func (cnt *Cnt) Push() {
	cnt.checkBuilt()
	if config.GetConfig().Push.Type == "" {
		log.Get().Panic("Can't push, push is not configured in cnt global configuration file")
	}

	im := extractManifestFromAci(cnt.target + "/image.aci")
	val, _ := im.Labels.Get("version")
	utils.ExecCmd("curl", "-i",
		"-F", "r=releases",
		"-F", "hasPom=false",
		"-F", "e=aci",
		"-F", "g=com.blablacar.aci.linux.amd64",
		"-F", "p=aci",
		"-F", "v=" + val,
		"-F", "a=" + ShortNameId(im.Name),
		"-F", "file=@" + cnt.target + "/image.aci",
		"-u", config.GetConfig().Push.Username + ":" + config.GetConfig().Push.Password,
		config.GetConfig().Push.Url + "/service/local/artifact/maven/content")
}

func (cnt *Cnt) Install() {
	cnt.checkBuilt()
	utils.ExecCmd("rkt", "--insecure-skip-verify=true", "fetch", cnt.target + "/image.aci")
}

func (cnt *Cnt) Build() error {
	log.Get().Info("Building Image : ", cnt.manifest.Aci.Name)

	os.MkdirAll(cnt.rootfs, 0777)

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

	cnt.tarAci()
	//	ExecCmd("chown " + os.Getenv("SUDO_USER") + ": " + target + "/*") //TODO chown
	return nil
}

//////////////////////////////////////////////////////////////////

func (cnt *Cnt) writeCntManifest() {
	utils.CopyFile(cnt.path +  "/"+ MANIFEST, cnt.target + "/"+ MANIFEST)
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

	cnt.manifest.Aci.Name.Set(Name(cnt.manifest.NameAndVersion))
	changeVersion(&cnt.manifest.Aci.Labels, Version(cnt.manifest.NameAndVersion))

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

func (cnt *Cnt) checkBuilt() {
	if _, err := os.Stat(cnt.target + "/image.aci"); os.IsNotExist(err) {
		if err := cnt.Build(); err != nil {
			log.Get().Panic("Cannot Install since build failed")
		}
	}
}

func (cnt *Cnt) runBuild() {
	if res, err := utils.IsDirEmpty(cnt.target + RUNLEVELS_BUILD); res || err != nil {
		return
	}
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
		cmd := []string{"run", "--name=" + ShortName(cnt.manifest.NameAndVersion), "-v", cnt.target + ":/target", imgId, "/target/build.sh"}
		utils.ExecCmd("docker", "rm", ShortName(cnt.manifest.NameAndVersion))
		if err := utils.ExecCmd("docker", cmd...); err != nil {
			panic(err)
		}

		//
		log.Get().Info("Release Docker");
		if cnt.manifest.Build.NoBuildImage() {
			os.RemoveAll(cnt.rootfs)
			os.Mkdir(cnt.rootfs, 0777)

			if err := utils.ExecCmd("docker", "export", "-o", cnt.target + "/dockerfs.tar", ShortName(cnt.manifest.NameAndVersion)); err != nil {
				panic(err)
			}

			utils.ExecCmd("tar", "xpf", cnt.target + "/dockerfs.tar", "-C", cnt.rootfs)
		}
		if err := utils.ExecCmd("docker", "rm", ShortName(cnt.manifest.NameAndVersion)); err != nil {
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

		app, err := discovery.NewAppFromString(cnt.manifest.From)
		if app.Labels["os"] == "" {
			app.Labels["os"] = "linux"
		}
		if app.Labels["arch"] == "" {
			app.Labels["arch"] = "amd64"
		}

		endpoint, _, err := discovery.DiscoverEndpoints(*app, false)
		if err != nil {
			panic(err)
		}

		url := endpoint.ACIEndpoints[0].ACI

		aciPath := config.GetConfig().AciPath + "/" + cnt.manifest.From
		if _, err := os.Stat(aciPath + "/image.aci"); cnt.args.ForceUpdate || os.IsNotExist(err) {
			if err := os.MkdirAll(aciPath, 0755); err != nil {
				log.Get().Panic(err)
			}
			utils.ExecCmd("wget", "-O", aciPath + "/image.aci", url)
		} else {
			log.Get().Info("Image " + cnt.manifest.From + " Already exists locally, will not be downloaded")
		}

		utils.ExecCmd("tar", "xpf", aciPath + "/image.aci", "-C", cnt.target)

		//		utils.ExecCmd("rkt", "--insecure-skip-verify=true", "fetch", cnt.manifest.From)
		//		utils.ExecCmd("rkt", "image", "export", "--overwrite", cnt.manifest.From, cnt.target + "/from.aci")
		//		utils.ExecCmd("tar", "xf", cnt.target + "/from.aci", "-C", cnt.target)
		//		os.Remove(cnt.target + "/from.aci")
	}
}


func (cnt *Cnt) copyRunlevelsBuild() {
	if err := os.MkdirAll(cnt.target + RUNLEVELS, 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path + RUNLEVELS, cnt.target + RUNLEVELS)
}

func (cnt *Cnt) runlevelBuildSetup() {
	files, err := ioutil.ReadDir(cnt.path + RUNLEVELS_BUILD_SETUP) // already sorted by name
	if err != nil {
		return
	}

	os.Setenv("TARGET", cnt.target)
	for _, f := range files {
		if !f.IsDir() {
			log.Get().Info("Running Build setup level : ", f.Name())
			if err := utils.ExecCmd(cnt.path + RUNLEVELS_BUILD_SETUP + f.Name()); err != nil {
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

	if _, err := os.Stat(cnt.path + RUNLEVELS_BUILD_INHERIT_EARLY); err == nil {
		args = append(args, RUNLEVELS_BUILD_INHERIT_EARLY)
	}
	if _, err := os.Stat(cnt.path + RUNLEVELS_BUILD_INHERIT_LATE); err == nil {
		args = append(args, RUNLEVELS_BUILD_INHERIT_LATE)
	}

	utils.Tar(cnt.args.Zip, "image.aci", args...)
	log.Get().Debug("chdir to", dir)
	os.Chdir(dir);
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
	utils.CopyDir(cnt.path + RUNLEVELS_PRESTART, cnt.rootfs + "/etc/prestart/early-prestart.d")
	utils.CopyDir(cnt.path + RUNLEVELS_LATESTART, cnt.rootfs + "/etc/prestart/late-prestart.d")
}

func (cnt *Cnt) copyConfd() {
	if err := os.MkdirAll(cnt.rootfs + "/etc/prestart/", 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path + CONFD_CONFIG, cnt.rootfs + "/etc/prestart/conf.d")
	utils.CopyDir(cnt.path + CONFD_TEMPLATE, cnt.rootfs + "/etc/prestart/templates")
}

func (cnt *Cnt) copyFiles() {
	utils.CopyDir(cnt.path + FILES_PATH, cnt.rootfs)
}

func (cnt *Cnt) copyAttributes() {
	if err := os.MkdirAll(cnt.rootfs + "/etc/prestart/attributes/" + ShortNameId(cnt.manifest.Aci.Name), 0755); err != nil {
		log.Get().Panic(err)
	}
	utils.CopyDir(cnt.path + ATTRIBUTES, cnt.rootfs + "/etc/prestart/attributes/" + ShortNameId(cnt.manifest.Aci.Name))
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


func extractManifestFromAci(aciPath string) schema.ImageManifest {
	input, err := os.Open(aciPath)
	if err != nil {
		log.Get().Panic("cat-manifest: Cannot open %s: %v", aciPath, err)
	}
	defer input.Close()

	tr, err := aci.NewCompressedTarReader(input)
	if err != nil {
		log.Get().Panic("cat-manifest: Cannot open tar %s: %v", aciPath, err)
	}


	im := schema.ImageManifest{}

	Tar:
	for {
		hdr, err := tr.Next()
		switch err {
		case io.EOF:
			break Tar
		case nil:
			if filepath.Clean(hdr.Name) == aci.ManifestFile {
				bytes, err := ioutil.ReadAll(tr)
				if err != nil {
					log.Get().Panic(err)
				}

				err = im.UnmarshalJSON(bytes)
				if err != nil {
					log.Get().Panic(err)
				}
				return im
			}
		default:
			log.Get().Panic("error reading tarball: %v", err)
		}
	}
	log.Get().Panic("Cannot found manifest if aci");
	return im
}