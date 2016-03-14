package main

import (
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
	"strings"
)

const PATH_TESTS_TARGET = "/tests-target"
const PATH_TESTS_RESULT = "/tests-result"
const MOUNT_ACNAME = "test-result"
const STATUS_SUFFIX = "_status"
const FILE_END_OF_TESTS = "end-of-tests"

func (aci *Aci) Test() error {
	defer aci.giveBackUserRightsToTarget()
	if err := aci.Install(); err != nil {
		return err
	}

	logs.WithF(aci.fields).Info("Testing")

	ImportInternalBuilderIfNeeded(aci.manifest) // TODO remove this hack
	ImportInternalTesterIfNeeded(aci.manifest)  // TODO import builder dependency

	logs.WithF(aci.fields).Info("Building test aci")
	if err := aci.buildTestAci(); err != nil {
		return err
	}

	logs.WithF(aci.fields).Info("Running test aci")
	if err := aci.runTestAci(); err != nil {
		return err
	}

	logs.WithF(aci.fields).Info("Checking result")
	if err := aci.checkResult(); err != nil {
		return err
	}
	return nil
}

func (aci *Aci) checkResult() error {
	files, err := ioutil.ReadDir(aci.target + PATH_TESTS_RESULT)
	if err != nil {
		return errs.WithEF(err, aci.fields, "Cannot read test result directory")
	}
	testFound := false
	getToTheEnd := false
	for _, f := range files {
		if f.Name() == FILE_END_OF_TESTS {
			getToTheEnd = true
			continue
		}
		testFields := aci.fields.WithField("file", f.Name())
		fullPath := aci.target + PATH_TESTS_RESULT + "/" + f.Name()
		content, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return errs.WithEF(err, testFields, "Cannot read result file")
		}
		if !strings.HasSuffix(f.Name(), STATUS_SUFFIX) {
			if testFound == false && string(content) != "1..0\n" {
				testFound = true
			}
			continue
		}
		if string(content) != "0\n" {
			return errs.WithEF(err, testFields, "Failed test")
		}
	}
	if !getToTheEnd {
		return errs.WithF(aci.fields, "Something goes wrong while running tests")
	}

	if Args.NoTestFail && !testFound {
		return errs.WithEF(err, aci.fields, "No tests found")
	}
	return nil
}

func (aci *Aci) runTestAci() error {
	os.MkdirAll(aci.target+PATH_TESTS_RESULT, 0777)

	if err := Home.Rkt.Run([]string{"--set-env=" + common.ENV_LOG_LEVEL + "=" + logs.GetLevel().String(),
		"--net=host",
		"--mds-register=false",
		"--volume=" + MOUNT_ACNAME + ",kind=host,source=" + aci.target + PATH_TESTS_RESULT,
		aci.target + PATH_TESTS_TARGET + PATH_IMAGE_ACI,
		"--exec", "/test",
	}); err != nil {
		// rkt+systemd cannot exit with fail status yet, so will not happen
		return errs.WithEF(err, aci.fields, "run of test aci failed")
	}
	return nil
}

func (aci *Aci) buildTestAci() error {
	fullname := common.NewACFullName(PREFIX_TEST_BUILDER + aci.manifest.NameAndVersion.Name() + ":" + aci.manifest.NameAndVersion.Version())
	resultMountName, _ := types.NewACName(MOUNT_ACNAME)
	testAci, err := NewAciWithManifest(aci.path, aci.args, &AciManifest{
		Build: BuildDefinition{Image: aci.manifest.TestBuilder},
		Aci: AciDefinition{
			App: DgrApp{
				Exec:             aci.manifest.Aci.App.Exec,
				MountPoints:      []types.MountPoint{{Path: PATH_TESTS_RESULT, Name: *resultMountName}},
				WorkingDirectory: aci.manifest.Aci.App.WorkingDirectory,
			},
			Dependencies: []common.ACFullname{aci.manifest.NameAndVersion},
		},
		NameAndVersion: *fullname,
	})
	if err != nil {
		return errs.WithEF(err, aci.fields, "Failed to prepare test's build aci")
	}

	testAci.FullyResolveDep = false // this is required to run local tests without discovery
	testAci.target = aci.target + PATH_TESTS_TARGET

	if err := testAci.CleanAndBuild(); err != nil {
		return errs.WithEF(err, aci.fields, "Build of test aci failed")
	}
	return nil
}
