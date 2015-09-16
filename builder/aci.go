package builder

import (
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/spec"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	execFiles = `
execute_files() {
  fdir=$1
  [ -d "$fdir" ] || return 0

  for file in $fdir/*; do
    if [ -x "$file" ]; then
      $file
    else
      echo -e "\e[31m$file is not exectuable\e[0m"
    fi
  done
}
	`

	BUILD_SCRIPT = `#!/bin/bash
set -x
set -e
export TARGET=$( dirname $0 )
export ROOTFS=%%ROOTFS%%
export TERM=xterm

` + execFiles + `

execute_files "$TARGET/runlevels/inherit-build-early"
execute_files "$TARGET/runlevels/build"
`
)

const (
	BUILD_SCRIPT_LATE = `#!/bin/bash
set -x
set -e
export TARGET=$( dirname $0 )
export ROOTFS=%%ROOTFS%%
export TERM=xterm

` + execFiles + `

execute_files "$TARGET/runlevels/build-late"
execute_files "$TARGET/runlevels/inherit-build-late"
`
)

const IMG_MANIFEST = "/cnt-manifest.yml"
const RUNLEVELS = "/runlevels"
const RUNLEVELS_PRESTART = RUNLEVELS + "/prestart-early"
const RUNLEVELS_LATESTART = RUNLEVELS + "/prestart-late"
const RUNLEVELS_BUILD = RUNLEVELS + "/build"
const RUNLEVELS_BUILD_SETUP = RUNLEVELS + "/build-setup"
const RUNLEVELS_BUILD_LATE = RUNLEVELS + "/build-late"
const RUNLEVELS_BUILD_INHERIT_EARLY = RUNLEVELS + "/inherit-build-early"
const RUNLEVELS_BUILD_INHERIT_LATE = RUNLEVELS + "/inherit-build-late"
const CONFD = "/confd"
const CONFD_TEMPLATE = CONFD + "/templates"
const CONFD_CONFIG = CONFD + "/conf.d"
const ATTRIBUTES = "/attributes"
const FILES_PATH = "/files"

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

func NewAciWithManifest(path string, args BuildArgs, manifest spec.AciManifest) (*Img, error) {
	log.Get().Debug("New aci", path, args, manifest)
	cnt, err := PrepAci(path, args)
	if err != nil {
		return nil, err
	}
	cnt.manifest = manifest
	return cnt, nil
}

func NewAci(path string, args BuildArgs) (*Img, error) {
	manifest, err := readManifest(path + IMG_MANIFEST)
	if err != nil {
		log.Get().Debug(path, IMG_MANIFEST+" does not exists")
		return nil, err
	}
	return NewAciWithManifest(path, args, *manifest)
}

func PrepAci(aciPath string, args BuildArgs) (*Img, error) {
	cnt := new(Img)
	cnt.args = args

	if fullPath, err := filepath.Abs(aciPath); err != nil {
		log.Get().Panic("Cannot get fullpath of project", err)
	} else {
		cnt.path = fullPath
		cnt.target = cnt.path + "/target"
		if args.Path != "" {
			currentAbsDir, err := filepath.Abs(args.Path)
			if err != nil {
				log.Get().Panic("invalid target path")
			}
			cnt.target = currentAbsDir
		}
		cnt.rootfs = cnt.target + "/rootfs"
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
		log.Get().Panic(err)
	}

	return &manifest, nil
}

func (i *Img) checkBuilt() {
	if _, err := os.Stat(i.target + "/image.aci"); os.IsNotExist(err) {
		if err := i.Build(); err != nil {
			log.Get().Panic("Cannot Install since build failed")
		}
	}
}
