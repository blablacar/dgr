package builder
import (
	"os"
	"io/ioutil"
	"strings"
	"path/filepath"
	"github.com/blablacar/cnt/utils"
	"github.com/blablacar/cnt/log"
	"github.com/appc/spec/schema"
	"github.com/ghodss/yaml"
	"github.com/appc/spec/schema/types"
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
      echo -e "\e[31m$file is not exectuable\e[0m"
    fi
  done
}

execute_files "$TARGET/runlevels/inherit-build-early"
execute_files "faili $TARGET/runlevels/build"
execute_files "$TARGET/runlevels/inherit-build-late"`
)

const MANIFEST = "cnt-manifest.yml"
const RUNLEVELS = "/runlevels"
const RUNLEVELS_PRESTART = RUNLEVELS + "/prestart-early"
const RUNLEVELS_LATESTART =  RUNLEVELS + "/prestart-late"
const RUNLEVELS_BUILD =  RUNLEVELS + "/build"
const RUNLEVELS_BUILD_SETUP =  RUNLEVELS + "/build-setup"
const RUNLEVELS_BUILD_INHERIT_EARLY =  RUNLEVELS + "/inherit-build-early"
const RUNLEVELS_BUILD_INHERIT_LATE = RUNLEVELS + "/inherit-build-late"
const CONFD = "/confd"
const CONFD_TEMPLATE = CONFD + "/templates"
const CONFD_CONFIG = CONFD + "/conf.d"
const ATTRIBUTES = "/attributes"
const FILES_PATH = "/files"

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

////////////////////////////////////////////

func OpenCnt(path string, args BuildArgs) (*Cnt, error) {
	cnt ,_ :=PrepCnt(path, args)

	if _, err := os.Stat(cnt.path + "/" + MANIFEST); os.IsNotExist(err)  {
		log.Get().Debug(cnt.path, "/"+ MANIFEST +" does not exists")
		return nil, &BuildError{"file not found : " + cnt.path +  "/"+ MANIFEST, err}
	}

	cnt.manifest.Aci = *utils.BasicManifest()
	cnt.readManifest(cnt.path + "/"+ MANIFEST)

	return cnt, nil
}

func PrepCnt(path string, args BuildArgs)(*Cnt, error){
	cnt := new(Cnt)
	cnt.args = args

	if fullPath, err := filepath.Abs(path); err != nil {
		log.Get().Panic("Cannot get fullpath of project", err)
	} else {
		cnt.path = fullPath
		cnt.target = cnt.path + "/target"
		cnt.rootfs = cnt.target + "/rootfs"
	}
	return cnt, nil
}

//////////////////////////////////////////////////////////////////

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
