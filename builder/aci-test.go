package builder

import (
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/cnt/dist"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/spec"
	"github.com/blablacar/cnt/utils"
	"io/ioutil"
	"os"
	"strings"
)

const BATS_ACI = "aci.blablacar.com/aci-bats:1"
const PATH_RESULT = "/result"
const TEST_INIT_SCRIPT = `#!/bin/bash
set -x

%%COMMAND%% &

waitsuccess="0"
if [ -f "/tests/wait.sh" ]; then
	chmod +x /tests/wait.sh
	i="0"
	while [ $i -lt 60 ]; do
	  /tests/wait.sh
	  if [ $? == 0 ]; then
	  	waitsuccess="1"
	  	break;
	  fi
	  i=$[$i+1]
	  sleep 1
	done
fi

if [ wait_success == "0" ]; then
	echo "1\n" > /result/wait.sh
	echo "WAIT FAILED"
	exit 1
fi

/test.sh
`

func (cnt *Img) Test() {
	cnt.Install()
	log.Get().Info("Testing " + cnt.manifest.NameAndVersion)
	cnt.importAciBats()

	testAci, err := cnt.prepareTestAci()
	if err != nil {
		log.Get().Panic(err)
	}
	testAci.Build()

	os.MkdirAll(cnt.target+PATH_TEST+PATH_TARGET+PATH_RESULT, 0777)

	if err := utils.ExecCmd("rkt",
		"--insecure-skip-verify=true",
		"run",
		"--local",
		"--volume=result,kind=host,source="+cnt.target+PATH_TEST+PATH_TARGET+PATH_RESULT,
		cnt.target+PATH_TEST+PATH_TARGET+"/image.aci"); err != nil {
		// rkt+systemd cannot exit with fail status yet
		log.Get().Panic(err)
	}

	cnt.checkResult()
}

func (cnt *Img) checkResult() {
	files, err := ioutil.ReadDir(cnt.target + PATH_TEST + PATH_TARGET + PATH_RESULT)
	if err != nil {
		log.Get().Panic("Cannot read test result directory", err)
	}
	for _, f := range files {
		content, err := ioutil.ReadFile(cnt.target + PATH_TEST + PATH_TARGET + PATH_RESULT + "/" + f.Name())
		if err != nil {
			log.Get().Panic("Cannot read result file", f.Name(), err)
		}
		if string(content) != "0\n" {
			log.Get().Error("Failed test file : ", f.Name())
			os.Exit(2)
		}
	}
}

func (cnt *Img) importAciBats() {
	if err := utils.ExecCmd("bash", "-c", "rkt image list --fields name --no-legend | grep -q "+BATS_ACI); err != nil {
		content, _ := dist.Asset("dist/bindata/aci-bats.aci")
		if err := ioutil.WriteFile("/tmp/aci-bats.aci", content, 0644); err != nil {
			log.Get().Panic(err)
		}
		utils.ExecCmd("rkt", "--insecure-skip-verify=true", "fetch", "/tmp/aci-bats.aci")
		os.Remove("/tmp/aci-bats.aci")
	}
}

func (cnt *Img) prepareTestAci() (*Img, error) {
	files, err := ioutil.ReadDir(cnt.path + PATH_TEST)
	if err != nil {
		return nil, err
	}

	utils.CopyDir(cnt.path+PATH_TEST+PATH_FILES, cnt.target+PATH_TEST+PATH_FILES)
	utils.CopyDir(cnt.path+PATH_TEST+PATH_ATTRIBUTES, cnt.target+PATH_TEST+PATH_ATTRIBUTES)
	utils.CopyDir(cnt.path+PATH_TEST+PATH_CONFD, cnt.target+PATH_TEST+PATH_CONFD)
	utils.CopyDir(cnt.path+PATH_TEST+PATH_RUNLEVELS, cnt.target+PATH_TEST+PATH_RUNLEVELS)

	os.MkdirAll(cnt.target+PATH_TEST+PATH_FILES+PATH_TEST, 0777)
	for _, f := range files {
		if !f.IsDir() {
			if err := utils.CopyFile(cnt.path+PATH_TEST+"/"+f.Name(), cnt.target+PATH_TEST+PATH_FILES+PATH_TEST+"/"+f.Name()); err != nil {
				log.Get().Panic(err)
			}
		}
	}

	ExecScript := strings.Replace(TEST_INIT_SCRIPT, "%%COMMAND%%", "'"+strings.Join(cnt.manifest.Aci.App.Exec, "' '")+"'", 1)

	ioutil.WriteFile(cnt.target+PATH_TEST+PATH_FILES+"/init.sh", []byte(ExecScript), 0777)

	fullname, err := spec.NewACFullName(cnt.manifest.NameAndVersion.Name() + "_test:" + cnt.manifest.NameAndVersion.Version())
	if err != nil {
		log.Get().Panic(err)
	}

	resultMountName, _ := types.NewACName("result")
	testAci, err := NewAciWithManifest(cnt.target+PATH_TEST, cnt.args, spec.AciManifest{
		Aci: spec.AciDefinition{
			App: &spec.CntApp{
				Exec:        []string{"/init.sh"},
				MountPoints: []types.MountPoint{{Path: PATH_RESULT, Name: *resultMountName}},
			},
			Dependencies: []spec.ACFullname{BATS_ACI, cnt.manifest.NameAndVersion},
		},
		NameAndVersion: *fullname,
	})
	if err != nil {
		log.Get().Panic("Cannot build test aci", err)
	}
	return testAci, nil
}
