package builder

import (
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/dist"
	"github.com/blablacar/dgr/spec"
	"github.com/blablacar/dgr/utils"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
	"strings"
)

const BATS_ACI = "aci.blablacar.com/aci-bats:5"
const PATH_RESULT = "/result"
const STATUS_SUFFIX = "_status"
const TEST_INIT_SCRIPT = `#!/bin/bash
set -x

[ -d %%CWD%% ] && {
	cd %%CWD%%
}

%%COMMAND%% &
cd /
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

func (dgr *Aci) Test() {
	dgr.Install()
	logs.WithF(dgr.fields).Info("Testing")

	if _, err := os.Stat(dgr.path + PATH_TESTS); err != nil {
		if dgr.args.NoTestFail {
			panic("Test directory does not exists but tests are mandatory")
		}
		logs.WithF(dgr.fields).Warn("Tests directory does not exists")
		return
	}

	dgr.importAciBats()

	testAci, err := dgr.prepareTestAci()
	if err != nil {
		panic(err)
	}
	testAci.Build()

	os.MkdirAll(dgr.target+PATH_TESTS+PATH_TARGET+PATH_RESULT, 0777)

	if err := utils.ExecCmd("rkt",
		"--insecure-options=image",
		"run",
		"--net=host",
		"--mds-register=false",
		"--no-overlay=true",
		"--volume=result,kind=host,source="+dgr.target+PATH_TESTS+PATH_TARGET+PATH_RESULT,
		dgr.target+PATH_TESTS+PATH_TARGET+"/image.aci"); err != nil {
		// rkt+systemd cannot exit with fail status yet
		panic(err)
	}

	dgr.checkResult()
}

func (dgr *Aci) checkResult() {
	files, err := ioutil.ReadDir(dgr.target + PATH_TESTS + PATH_TARGET + PATH_RESULT)
	if err != nil {
		panic("Cannot read test result directory" + err.Error())
	}
	testFound := false
	for _, f := range files {
		fullPath := dgr.target + PATH_TESTS + PATH_TARGET + PATH_RESULT + "/" + f.Name()
		content, err := ioutil.ReadFile(fullPath)
		if err != nil {
			panic("Cannot read result file" + f.Name() + err.Error())
		}
		if !strings.HasSuffix(f.Name(), STATUS_SUFFIX) {
			if testFound == false && string(content) != "1..0\n" {
				testFound = true
			}
			continue
		}
		if string(content) != "0\n" {
			logs.WithF(dgr.fields).WithField("file", f.Name()).Error("Failed test")
			os.Exit(2)
		}
	}

	if dgr.args.NoTestFail && !testFound {
		panic("No tests found")
	}
}

func (dgr *Aci) importAciBats() {
	if err := utils.ExecCmd("bash", "-c", "rkt image list --fields name --no-legend | grep -q "+BATS_ACI); err != nil {
		content, _ := dist.Asset("dist/bindata/aci-bats.aci")
		if err := ioutil.WriteFile("/tmp/aci-bats.aci", content, 0644); err != nil {
			panic(err)
		}
		utils.ExecCmd("rkt", "--insecure-options=image", "fetch", "/tmp/aci-bats.aci")
		os.Remove("/tmp/aci-bats.aci")
	}
}

func (dgr *Aci) prepareTestAci() (*Aci, error) {
	files, err := ioutil.ReadDir(dgr.path + PATH_TESTS)
	if err != nil {
		return nil, err
	}

	utils.CopyDir(dgr.path+PATH_TESTS+PATH_FILES, dgr.target+PATH_TESTS+PATH_FILES)
	utils.CopyDir(dgr.path+PATH_TESTS+PATH_ATTRIBUTES, dgr.target+PATH_TESTS+PATH_ATTRIBUTES)
	utils.CopyDir(dgr.path+PATH_TESTS+PATH_TEMPLATES, dgr.target+PATH_TESTS+PATH_TEMPLATES)
	utils.CopyDir(dgr.path+PATH_TESTS+PATH_RUNLEVELS, dgr.target+PATH_TESTS+PATH_RUNLEVELS)

	os.MkdirAll(dgr.target+PATH_TESTS+PATH_FILES+PATH_TESTS, 0777)
	for _, f := range files {
		if !f.IsDir() {
			if err := utils.CopyFile(dgr.path+PATH_TESTS+"/"+f.Name(), dgr.target+PATH_TESTS+PATH_FILES+PATH_TESTS+"/"+f.Name()); err != nil {
				panic(err)
			}
		}
	}

	ExecScript := strings.Replace(TEST_INIT_SCRIPT, "%%COMMAND%%", "'"+strings.Join(dgr.manifest.Aci.App.Exec, "' '")+"'", 1)
	ExecScript = strings.Replace(ExecScript, "%%CWD%%", "'"+dgr.manifest.Aci.App.WorkingDirectory+"'", 2)

	ioutil.WriteFile(dgr.target+PATH_TESTS+PATH_FILES+"/init.sh", []byte(ExecScript), 0777)

	fullname := spec.NewACFullName(dgr.manifest.NameAndVersion.Name() + "_test:" + dgr.manifest.NameAndVersion.Version())

	resultMountName, _ := types.NewACName("result")
	testAci, err := NewAciWithManifest(dgr.target+PATH_TESTS, dgr.args, spec.AciManifest{
		Aci: spec.AciDefinition{
			App: spec.DgrApp{
				Exec:        []string{"/init.sh"},
				MountPoints: []types.MountPoint{{Path: PATH_RESULT, Name: *resultMountName}},
			},
			Dependencies: []spec.ACFullname{BATS_ACI, dgr.manifest.NameAndVersion},
		},
		NameAndVersion: *fullname,
	})
	testAci.FullyResolveDep = false                        // this is required to run local tests without discovery
	testAci.target = dgr.target + PATH_TESTS + PATH_TARGET // this is required when target is deported
	testAci.rootfs = testAci.target + PATH_ROOTFS
	if err != nil {
		panic("Cannot build test aci" + err.Error())
	}
	return testAci, nil
}
