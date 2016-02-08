package builder

import (
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/cnt/dist"
	"github.com/blablacar/cnt/spec"
	"github.com/blablacar/cnt/utils"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
	"strings"
)

const BATS_ACI = "aci.blablacar.com/aci-bats:2"
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

func (cnt *Aci) Test() {
	cnt.Install()
	logs.WithF(cnt.fields).Info("Testing")

	if _, err := os.Stat(cnt.path + PATH_TESTS); err != nil {
		if cnt.args.NoTestFail {
			panic("Test directory does not exists but tests are mandatory")
		}
		logs.WithF(cnt.fields).Warn("Tests directory does not exists")
		return
	}

	cnt.importAciBats()

	testAci, err := cnt.prepareTestAci()
	if err != nil {
		panic(err)
	}
	testAci.Build()

	os.MkdirAll(cnt.target+PATH_TESTS+PATH_TARGET+PATH_RESULT, 0777)

	if err := utils.ExecCmd("rkt",
		"--insecure-options=image",
		"run",
		"--net=host",
		"--mds-register=false",
		"--no-overlay=true",
		"--volume=result,kind=host,source="+cnt.target+PATH_TESTS+PATH_TARGET+PATH_RESULT,
		cnt.target+PATH_TESTS+PATH_TARGET+"/image.aci"); err != nil {
		// rkt+systemd cannot exit with fail status yet
		panic(err)
	}

	cnt.checkResult()
}

func (cnt *Aci) checkResult() {
	files, err := ioutil.ReadDir(cnt.target + PATH_TESTS + PATH_TARGET + PATH_RESULT)
	if err != nil {
		panic("Cannot read test result directory" + err.Error())
	}
	testFound := false
	for _, f := range files {
		fullPath := cnt.target + PATH_TESTS + PATH_TARGET + PATH_RESULT + "/" + f.Name()
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
			logs.WithF(cnt.fields).WithField("file", f.Name()).Error("Failed test")
			os.Exit(2)
		}
	}

	if cnt.args.NoTestFail && !testFound {
		panic("No tests found")
	}
}

func (cnt *Aci) importAciBats() {
	if err := utils.ExecCmd("bash", "-c", "rkt image list --fields name --no-legend | grep -q "+BATS_ACI); err != nil {
		content, _ := dist.Asset("dist/bindata/aci-bats.aci")
		if err := ioutil.WriteFile("/tmp/aci-bats.aci", content, 0644); err != nil {
			panic(err)
		}
		utils.ExecCmd("rkt", "--insecure-options=image", "fetch", "/tmp/aci-bats.aci")
		os.Remove("/tmp/aci-bats.aci")
	}
}

func (cnt *Aci) prepareTestAci() (*Aci, error) {
	files, err := ioutil.ReadDir(cnt.path + PATH_TESTS)
	if err != nil {
		return nil, err
	}

	utils.CopyDir(cnt.path+PATH_TESTS+PATH_FILES, cnt.target+PATH_TESTS+PATH_FILES)
	utils.CopyDir(cnt.path+PATH_TESTS+PATH_ATTRIBUTES, cnt.target+PATH_TESTS+PATH_ATTRIBUTES)
	utils.CopyDir(cnt.path+PATH_TESTS+PATH_TEMPLATES, cnt.target+PATH_TESTS+PATH_TEMPLATES)
	utils.CopyDir(cnt.path+PATH_TESTS+PATH_RUNLEVELS, cnt.target+PATH_TESTS+PATH_RUNLEVELS)

	os.MkdirAll(cnt.target+PATH_TESTS+PATH_FILES+PATH_TESTS, 0777)
	for _, f := range files {
		if !f.IsDir() {
			if err := utils.CopyFile(cnt.path+PATH_TESTS+"/"+f.Name(), cnt.target+PATH_TESTS+PATH_FILES+PATH_TESTS+"/"+f.Name()); err != nil {
				panic(err)
			}
		}
	}

	ExecScript := strings.Replace(TEST_INIT_SCRIPT, "%%COMMAND%%", "'"+strings.Join(cnt.manifest.Aci.App.Exec, "' '")+"'", 1)
	ExecScript = strings.Replace(ExecScript, "%%CWD%%", "'"+cnt.manifest.Aci.App.WorkingDirectory+"'", 2)

	ioutil.WriteFile(cnt.target+PATH_TESTS+PATH_FILES+"/init.sh", []byte(ExecScript), 0777)

	fullname := spec.NewACFullName(cnt.manifest.NameAndVersion.Name() + "_test:" + cnt.manifest.NameAndVersion.Version())

	resultMountName, _ := types.NewACName("result")
	testAci, err := NewAciWithManifest(cnt.target+PATH_TESTS, cnt.args, spec.AciManifest{
		Aci: spec.AciDefinition{
			App: spec.CntApp{
				Exec:        []string{"/init.sh"},
				MountPoints: []types.MountPoint{{Path: PATH_RESULT, Name: *resultMountName}},
			},
			Dependencies: []spec.ACFullname{BATS_ACI, cnt.manifest.NameAndVersion},
		},
		NameAndVersion: *fullname,
	}, nil)
	testAci.FullyResolveDep = false                        // this is required to run local tests without discovery
	testAci.target = cnt.target + PATH_TESTS + PATH_TARGET // this is required when target is deported
	testAci.rootfs = testAci.target + PATH_ROOTFS
	if err != nil {
		panic("Cannot build test aci" + err.Error())
	}
	return testAci, nil
}
