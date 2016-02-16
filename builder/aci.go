package builder

import (
	"github.com/appc/spec/schema"
	"github.com/blablacar/dgr/dgr"
	"github.com/blablacar/dgr/spec"
	"github.com/blablacar/dgr/utils"
	"github.com/ghodss/yaml"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

const SH_FUNCTIONS = `
execute_files() {
  fdir=$1
  [ -d "$fdir" ] || return 0

  for file in "$fdir"/*; do
    [ -e "$file" ] && {
     	[ -x "$file" ] || chmod +x "$file"
		isLevelEnabled 4 && echo -e "\e[1m\e[32mRunning script -> $file\e[0m"
     	"$file"
    }
  done
}

levelFromString() {
	case ` + "`echo ${1} | awk '{print toupper($0)}'`" + ` in
		"FATAL") echo 0; return 0 ;;
		"PANIC") echo 1; return 0 ;;
		"ERROR") echo 2; return 0 ;;
		"WARN"|"WARNING") echo 3; return 0 ;;
		"INFO") echo 4; return 0 ;;
		"DEBUG") echo 5; return 0 ;;
		"TRACE") echo 6; return 0 ;;
		*) echo 4 ;;
	esac
}

isLevelEnabled() {
	l=$(levelFromString $1)

	if [ $l -le $log_level ]; then
		return 0
	fi
	return 1
}

export log_level=$(levelFromString ${LOG_LEVEL:-INFO})
`

const BUILD_SCRIPT = `#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

export TARGET=$( dirname $0 )
export ROOTFS=%%ROOTFS%%
export TERM=xterm

execute_files "$ROOTFS/dgr/runlevels/inherit-build-early"
execute_files "$TARGET/runlevels/build"
`

const BUILD_SCRIPT_LATE = `#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x


export TARGET=$( dirname $0 )
export ROOTFS=%%ROOTFS%%
export TERM=xterm

execute_files "$TARGET/runlevels/build-late"
execute_files "$ROOTFS/dgr/runlevels/inherit-build-late"
`

const PRESTART = `#!/dgr/bin/busybox sh
set -e
. /dgr/bin/functions.sh
isLevelEnabled "debug" && set -x

BASEDIR=${0%/*}
dgr_PATH=/dgr

execute_files ${dgr_PATH}/runlevels/prestart-early

if [ -z ${LOG_LEVEL} ]; then
	${BASEDIR}/templater -o TEMPLATER_OVERRIDE -t / /dgr
else
	${BASEDIR}/templater -o TEMPLATER_OVERRIDE -L "${LOG_LEVEL}" -t / /dgr
fi

#if [ -d ${dgr_PATH}/attributes ]; then
#	echo "$CONFD_OVERRIDE"
#    ${BASEDIR}/attributes-merger -i ${dgr_PATH}/attributes -e CONFD_OVERRIDE
#    export CONFD_DATA=$(cat attributes.json)
#fi
#${BASEDIR}/confd -onetime -config-file=${dgr_PATH}/prestart/confd.toml

execute_files ${dgr_PATH}/runlevels/prestart-late
`
const BUILD_SETUP = `#!/bin/sh
set -e
. "${TARGET}/rootfs/dgr/bin/functions.sh"
isLevelEnabled "debug" && set -x

execute_files "${BASEDIR}/runlevels/build-setup"
`

const PATH_BIN = "/bin"
const PATH_TESTS = "/tests"
const PATH_INSTALLED = "/installed"
const PATH_MANIFEST = "/manifest"
const PATH_IMAGE_ACI = "/image.aci"
const PATH_IMAGE_ACI_ZIP = "/image-zip.aci"
const PATH_ROOTFS = "/rootfs"
const PATH_TARGET = "/target"
const PATH_DGR = "/dgr"
const PATH_ACI_MANIFEST = "/aci-manifest.yml"
const PATH_RUNLEVELS = "/runlevels"
const PATH_PRESTART_EARLY = "/prestart-early"
const PATH_PRESTART_LATE = "/prestart-late"
const PATH_INHERIT_BUILD_LATE = "/inherit-build-late"
const PATH_INHERIT_BUILD_EARLY = "/inherit-build-early"
const PATH_ATTRIBUTES = "/attributes"
const PATH_FILES = "/files"
const PATH_BUILD_LATE = "/build-late"
const PATH_BUILD_SETUP = "/build-setup"
const PATH_BUILD = "/build"
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

func NewAciWithManifest(path string, args BuildArgs, manifest spec.AciManifest) (*Aci, error) {
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
	if dgr.Home.Config.TargetWorkDir != "" {
		currentAbsDir, err := filepath.Abs(dgr.Home.Config.TargetWorkDir + "/" + manifest.NameAndVersion.ShortName())
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

	aci.checkCompatibilityVersions()
	aci.checkLatestVersions()
	return aci, nil
}

func NewAci(path string, args BuildArgs) (*Aci, error) {
	manifest, err := readAciManifest(path + PATH_ACI_MANIFEST)
	if err != nil {
		manifest2, err2 := readAciManifest(path + "/cnt-manifest.yml")
		if err2 != nil {
			return nil, errs.WithEF(err, data.WithField("path", path+PATH_ACI_MANIFEST).WithField("err2", err2), "Cannot read manifest")
		}
		logs.WithField("old", "cnt-manifest.yml").WithField("new", "aci-manifest.yml").Warn("You are using the old aci configuration file")
		manifest = manifest2
	}
	return NewAciWithManifest(path, args, *manifest)
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

func (aci *Aci) checkCompatibilityVersions() {
	froms, err := aci.manifest.GetFroms()
	if err != nil {
		logs.WithEF(err, aci.fields).Fatal("Invalid from")
	}
	for _, from := range froms {
		if from == "" {
			continue
		}

		fromFields := aci.fields.WithField("dependency", from.String())

		err := utils.ExecCmd("rkt", "--insecure-options=image", "fetch", from.String())
		if err != nil {
			logs.WithEF(err, fromFields).Fatal("Cannot fetch from")
		}
		out, err := utils.ExecCmdGetOutput("rkt", "image", "cat-manifest", from.String())
		if err != nil {
			logs.WithEF(err, fromFields).Fatal("Cannot find dependency")
		}

		version, ok := loadManifest(out).Annotations.Get("dgr-version")
		if !ok {
			version, ok = loadManifest(out).Annotations.Get("cnt-version")
		}
		var val int
		if ok {
			val, err = strconv.Atoi(version)
			if err != nil {
				logs.WithEF(err, fromFields).WithField("version", version).Fatal("Failed to parse dgr-version from manifest")
			}
		}
		if !ok || val < 51 {
			logs.WithF(aci.fields).
				WithField("from", from).
				WithField("require", ">=51").
				Error("from aci was not build with a compatible version of dgr")
		}
	}

	for _, dep := range aci.manifest.Aci.Dependencies {
		depFields := aci.fields.WithField("dependency", dep.String())
		err := utils.ExecCmd("rkt", "--insecure-options=image", "fetch", dep.String())
		if err != nil {
			logs.WithEF(err, depFields).Fatal("Cannot fetch dependency")
		}
		out, err := utils.ExecCmdGetOutput("rkt", "image", "cat-manifest", dep.String())
		if err != nil {
			logs.WithEF(err, depFields).Fatal("Cannot find dependency")
		}

		version, ok := loadManifest(out).Annotations.Get("dgr-version")
		if !ok {
			version, ok = loadManifest(out).Annotations.Get("cnt-version")
		}
		var val int
		if ok {
			val, err = strconv.Atoi(version)
			if err != nil {
				logs.WithEF(err, depFields).WithField("version", version).Fatal("Failed to parse dgr-version from manifest")
			}
		}
		if !ok || val < 51 {
			logs.WithF(aci.fields).
				WithField("dependency", dep).
				WithField("require", ">=51").
				Error("dependency aci was not build with a compatible version of dgr")
		}
	}
}

func loadManifest(content string) schema.ImageManifest {
	im := schema.ImageManifest{}
	err := im.UnmarshalJSON([]byte(content))
	if err != nil {
		logs.WithE(err).WithField("content", content).Fatal("Failed to read manifest content")
	}
	return im
}

func (aci *Aci) checkLatestVersions() {
	froms, err := aci.manifest.GetFroms()
	if err != nil {
		logs.WithEF(err, aci.fields).Fatal("Invalid from")
	}
	for _, from := range froms {
		if from == "" {
			continue
		}

		version, _ := from.LatestVersion()
		logs.WithField("version", from.Name()+":"+version).Debug("Discovered from latest verion")
		if version != "" && utils.Version(from.Version()).LessThan(utils.Version(version)) {
			logs.WithField("newer", from.Name()+":"+version).
				WithField("current", from.String()).
				Warn("Newer 'from' version")
		}
	}
	for _, dep := range aci.manifest.Aci.Dependencies {
		if dep.Version() == "" {
			continue
		}
		version, _ := dep.LatestVersion()
		if version != "" && utils.Version(dep.Version()).LessThan(utils.Version(version)) {
			logs.WithField("newer", dep.Name()+":"+version).
				WithField("current", dep.String()).
				Warn("Newer 'dependency' version")
		}
	}
}
