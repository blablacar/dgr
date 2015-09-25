package builder

import (
	"errors"
	"github.com/appc/spec/discovery"
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/spec"
	"github.com/blablacar/cnt/utils"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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

${BASEDIR}/attributes-merger -i ${CNT_PATH}/attributes -e CONFD_OVERRIDE
export CONFD_DATA=$(cat attributes.json)
${BASEDIR}/confd -onetime -config-file=${CNT_PATH}/prestart/confd.toml

execute_files ${CNT_PATH}/runlevels/prestart-late
`
const PATH_BIN = "/bin"
const PATH_TEST = "/tests"
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

func NewAciWithManifest(path string, args BuildArgs, manifest spec.AciManifest) (*Img, error) {
	log.Get().Debug("New aci", path, args, manifest)
	cnt, err := PrepAci(path, args)
	if err != nil {
		return nil, err
	}
	cnt.manifest = manifest

	cnt.checkLatestVersions()

	return cnt, nil
}

func NewAci(path string, args BuildArgs) (*Img, error) {
	manifest, err := readManifest(path + PATH_CNT_MANIFEST)
	if err != nil {
		log.Get().Debug(path, PATH_CNT_MANIFEST+" does not exists")
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
		cnt.target = cnt.path + PATH_TARGET
		if args.TargetPath != "" {
			currentAbsDir, err := filepath.Abs(args.TargetPath + "/" + cnt.manifest.NameAndVersion.ShortName())
			if err != nil {
				log.Get().Panic("invalid target path")
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
		log.Get().Panic(err)
	}

	return &manifest, nil
}

func (cnt *Img) tarAci(zip bool) {
	target := PATH_IMAGE_ACI[1:]
	if zip {
		target = PATH_IMAGE_ACI_ZIP[1:]
	}
	dir, _ := os.Getwd()
	log.Get().Debug("chdir to", cnt.target)
	os.Chdir(cnt.target)
	utils.Tar(zip, target, PATH_MANIFEST[1:], PATH_ROOTFS[1:])
	log.Get().Debug("chdir to", dir)
	os.Chdir(dir)
}

func (cnt *Img) checkLatestVersions() {
	if cnt.manifest.From != "" {
		version := getLatestVersion(cnt.manifest.From)
		log.Get().Debug("latest version of from : " + cnt.manifest.NameAndVersion.Name() + ":" + version)
		if version != "" && utils.Version(cnt.manifest.From.Version()).LessThan(utils.Version(version)) {
			log.Get().Warn("---------------------------------")
			log.Get().Warn("From has newer version : " + cnt.manifest.From.Name() + ":" + version)
			log.Get().Warn("---------------------------------")
		}
	}
	for _, dep := range cnt.manifest.Aci.Dependencies {
		version := getLatestVersion(dep)
		if version != "" && utils.Version(cnt.manifest.From.Version()).LessThan(utils.Version(version)) {
			log.Get().Warn("---------------------------------")
			log.Get().Warn("Newer dependency version : " + dep.Name() + ":" + version)
			log.Get().Warn("---------------------------------")
		}
	}
}

func getLatestVersion(name spec.ACFullname) string {
	app, err := discovery.NewAppFromString(name.Name() + ":latest")
	if app.Labels["os"] == "" {
		app.Labels["os"] = "linux"
	}
	if app.Labels["arch"] == "" {
		app.Labels["arch"] = "amd64"
	}

	endpoint, _, err := discovery.DiscoverEndpoints(*app, false)
	if err != nil {
		return ""
	}

	r, _ := regexp.Compile(`^(\d+\.)?(\d+\.)?(\*|\d+)$`)

	url := getRedirectForLatest(endpoint.ACIEndpoints[0].ACI)
	log.Get().Debug("latest version url is ", url)

	for _, part := range strings.Split(url, "/") {
		if r.Match([]byte(part)) {
			return part
		}
	}
	return ""
}

func getRedirectForLatest(url string) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	transport := http.DefaultTransport
	//	if insecureSkipVerify {
	//		transport = &http.Transport{
	//			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//		}
	//	}
	client := &http.Client{Transport: transport}
	myurl := ""
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		myurl = req.URL.Path
		return errors.New("do not want to get the file")
	}
	_, err2 := client.Do(req)
	if err2 != nil {
		if myurl != "" {
			return myurl
		}
		return ""
	}
	return myurl
}
