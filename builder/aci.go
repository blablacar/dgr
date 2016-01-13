package builder

import (
	"github.com/blablacar/cnt/cnt"
	"github.com/blablacar/cnt/spec"
	"github.com/blablacar/cnt/utils"
	"github.com/ghodss/yaml"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
	"path/filepath"
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

type Aci struct {
	fields          data.Fields
	path            string
	target          string
	rootfs          string
	podName         *spec.ACFullname
	manifest        spec.AciManifest
	args            BuildArgs
	FullyResolveDep bool
}

func NewAciWithManifest(path string, args BuildArgs, manifest spec.AciManifest, checked *chan bool) (*Aci, error) {
	if manifest.NameAndVersion == "" {
		logs.WithField("path", path).Fatal("name is mandatory in manifest")
	}

	fields := data.WithField("aci", manifest.NameAndVersion.String())
	logs.WithF(fields).WithFields(data.Fields{"args": args, "path": path, "manifest": manifest}).Debug("New aci")

	fullPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errs.WithEF(err, fields, "Cannot get fullpath of project")
	}

	target := fullPath + PATH_TARGET
	if cnt.Home.Config.TargetWorkDir != "" {
		currentAbsDir, err := filepath.Abs(cnt.Home.Config.TargetWorkDir + "/" + manifest.NameAndVersion.ShortName())
		if err != nil {
			return nil, errs.WithEF(err, fields.WithField("path", path), "Invalid target path")
		}
		target = currentAbsDir
	}

	aci := &Aci{
		fields:          fields,
		args:            args,
		path:            fullPath,
		manifest:        manifest,
		target:          target,
		rootfs:          target + PATH_ROOTFS,
		FullyResolveDep: true,
	}

	go aci.checkLatestVersions(checked)
	return aci, nil
}

func NewAci(path string, args BuildArgs) (*Aci, error) {
	manifest, err := readAciManifest(path + PATH_CNT_MANIFEST)
	if err != nil {
		return nil, errs.WithEF(err, data.WithField("path", path+PATH_CNT_MANIFEST), "Cannot read manifest")
	}
	return NewAciWithManifest(path, args, *manifest, nil)
}

//////////////////////////////////////////////////////////////////

func readAciManifest(manifestPath string) (*spec.AciManifest, error) {
	manifest := spec.AciManifest{Aci: spec.AciDefinition{}}

	source, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(source), &manifest)
	if err != nil {
		return nil, errs.WithE(err, "Cannot unmarshall manifest")
	}

	return &manifest, nil
}

func (aci *Aci) tarAci(zip bool) {
	target := PATH_IMAGE_ACI[1:]
	if zip {
		target = PATH_IMAGE_ACI_ZIP[1:]
	}
	dir, _ := os.Getwd()
	logs.WithField("path", aci.target).Debug("chdir")
	os.Chdir(aci.target)
	utils.Tar(zip, target, PATH_MANIFEST[1:], PATH_ROOTFS[1:])
	logs.WithField("path", dir).Debug("chdir")
	os.Chdir(dir)
}

func (aci *Aci) checkLatestVersions(checked *chan bool) {
	if aci.manifest.From != "" && aci.manifest.From.Version() != "" {
		version, _ := aci.manifest.From.LatestVersion()
		logs.WithField("version", aci.manifest.From.Name()+":"+version).Debug("Discovered from latest verion")
		if version != "" && utils.Version(aci.manifest.From.Version()).LessThan(utils.Version(version)) {
			logs.WithF(aci.fields.WithField("version", aci.manifest.From.Name()+":"+version)).Warn("Newer from version")
		}
	}
	for _, dep := range aci.manifest.Aci.Dependencies {
		if dep.Version() == "" {
			continue
		}
		version, _ := dep.LatestVersion()
		if version != "" && utils.Version(dep.Version()).LessThan(utils.Version(version)) {
			logs.WithF(aci.fields.WithField("version", dep.Name()+":"+version)).Warn("Newer dependency version")
		}
	}
	if checked != nil {
		*checked <- true
	}
}
