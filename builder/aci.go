package builder

import (
	log "github.com/Sirupsen/logrus"
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/cnt/spec"
	"github.com/blablacar/cnt/utils"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const EXEC_FILES = `
execute_files() {
  fdir=$1
  [ -d "$fdir" ] || return 0

  for file in $fdir/*; do
    [ -e "$file" ] && {
     	[ -x "$file" ] || /cnt/bin/busybox chmod +x "$file"
     	echo -e "\e[1m\e[32mRunning script -> $file\e[0m"
     	$file
    }
  done
}`

const BUILD_SCRIPT = `#!/cnt/bin/busybox sh
set -x
set -e
export TARGET=$( dirname $0 )
export ROOTFS=%%ROOTFS%%
export TERM=xterm

` + EXEC_FILES + `

execute_files "$ROOTFS/cnt/runlevels/inherit-build-early"
execute_files "$TARGET/runlevels/build"
`

const BUILD_SCRIPT_LATE = `#!/cnt/bin/busybox sh
set -x
set -e
export TARGET=$( dirname $0 )
export ROOTFS=%%ROOTFS%%
export TERM=xterm

` + EXEC_FILES + `

execute_files "$TARGET/runlevels/build-late"
execute_files "$ROOTFS/cnt/runlevels/inherit-build-late"
`

const PRESTART = `#!/cnt/bin/busybox sh
set -x
set -e

BASEDIR=${0%/*}
CNT_PATH=/cnt

` + EXEC_FILES + `

execute_files ${CNT_PATH}/runlevels/prestart-early

if [ -d ${CNT_PATH}/attributes ]; then
	echo "$CONFD_OVERRIDE"
    ${BASEDIR}/attributes-merger -i ${CNT_PATH}/attributes -e CONFD_OVERRIDE
    export CONFD_DATA=$(cat attributes.json)
fi
${BASEDIR}/confd -onetime -config-file=${CNT_PATH}/prestart/confd.toml

execute_files ${CNT_PATH}/runlevels/prestart-late
`
const PATH_BIN = "/bin"
const PATH_TESTS = "/tests"
const PATH_INSTALLED = "/installed"
const PATH_MANIFEST = "/manifest"
const PATH_IMAGE_ACI = "/image.aci"
const PATH_IMAGE_ACI_ZIP = "/image-zip.aci"
const PATH_ROOTFS = "/rootfs"
const PATH_TARGET = "/target"
const PATH_CNT = "/cnt"
const PATH_CNT_MANIFEST = "/cnt-manifest.yml"
const PATH_RUNLEVELS = "/runlevels"
const PATH_PRESTART_EARLY = "/prestart-early"
const PATH_PRESTART_LATE = "/prestart-late"
const PATH_INHERIT_BUILD_LATE = "/inherit-build-late"
const PATH_INHERIT_BUILD_EARLY = "/inherit-build-early"
const PATH_ATTRIBUTES = "/attributes"
const PATH_FILES = "/files"
const PATH_CONFD = "/confd"
const PATH_BUILD_LATE = "/build-late"
const PATH_BUILD_SETUP = "/build-setup"
const PATH_BUILD = "/build"
const PATH_CONFDOTD = "/conf.d"
const PATH_TEMPLATES = "/templates"

type Img struct {
	path     string
	target   string
	rootfs   string
	PodName  *spec.ACFullname
	manifest spec.AciManifest
	args     BuildArgs
}

func Version(nameAndVersion string) string {
	split := strings.Split(nameAndVersion, ":")
	if len(split) == 1 {
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

////////////////////////////////////////////

func NewAciWithManifest(path string, args BuildArgs, manifest spec.AciManifest, checked *chan bool) (*Img, error) {
	log.Debug("New aci", path, args, manifest)
	cnt, err := PrepAci(path, args)
	if err != nil {
		return nil, err
	}
	cnt.manifest = manifest

	go cnt.checkLatestVersions(checked)

	return cnt, nil
}

func NewAci(path string, args BuildArgs) (*Img, error) {
	manifest, err := readManifest(path + PATH_CNT_MANIFEST)
	if err != nil {
		log.Debug(path, PATH_CNT_MANIFEST+" does not exists")
		return nil, err
	}
	return NewAciWithManifest(path, args, *manifest, nil)
}

func PrepAci(aciPath string, args BuildArgs) (*Img, error) {
	cnt := new(Img)
	cnt.args = args

	if fullPath, err := filepath.Abs(aciPath); err != nil {
		panic("Cannot get fullpath of project" + err.Error())
	} else {
		cnt.path = fullPath
		cnt.target = cnt.path + PATH_TARGET
		if args.TargetPath != "" {
			currentAbsDir, err := filepath.Abs(args.TargetPath + "/" + cnt.manifest.NameAndVersion.ShortName())
			if err != nil {
				panic("invalid target path")
			}
			cnt.target = currentAbsDir
		}
		cnt.rootfs = cnt.target + PATH_ROOTFS
	}
	return cnt, nil
}

//////////////////////////////////////////////////////////////////

func readManifest(manifestPath string) (*spec.AciManifest, error) {
	manifest := spec.AciManifest{Aci: spec.AciDefinition{}}

	source, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(source), &manifest)
	if err != nil {
		panic(err)
	}

	return &manifest, nil
}

func (cnt *Img) tarAci(zip bool) {
	target := PATH_IMAGE_ACI[1:]
	if zip {
		target = PATH_IMAGE_ACI_ZIP[1:]
	}
	dir, _ := os.Getwd()
	log.Debug("chdir to", cnt.target)
	os.Chdir(cnt.target)
	utils.Tar(zip, target, PATH_MANIFEST[1:], PATH_ROOTFS[1:])
	log.Debug("chdir to", dir)
	os.Chdir(dir)
}

func (cnt *Img) checkLatestVersions(checked *chan bool) {
	if cnt.manifest.From != "" {
		version, _ := cnt.manifest.From.LatestVersion()
		log.Debug("latest version of from : " + cnt.manifest.NameAndVersion.Name() + ":" + version)
		if version != "" && utils.Version(cnt.manifest.From.Version()).LessThan(utils.Version(version)) {
			log.Warn("---------------------------------")
			log.Warn("From has newer version : " + cnt.manifest.From.Name() + ":" + version)
			log.Warn("---------------------------------")
		}
	}
	for _, dep := range cnt.manifest.Aci.Dependencies {
		version, _ := dep.LatestVersion()
		if version != "" && utils.Version(dep.Version()).LessThan(utils.Version(version)) {
			log.Warn("---------------------------------")
			log.Warn("Newer dependency version : " + dep.Name() + ":" + version)
			log.Warn("---------------------------------")
		}
	}
	if checked != nil {
		*checked <- true
	}
}
