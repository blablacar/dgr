package main

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/dgr/common"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"gopkg.in/yaml.v2"
)

const pathTestsTarget = "/tests-target"
const pathTestsResult = "/tests-result"
const mountAcname = "test-result"
const statusSuffix = "_status"
const fileEndOfTests = "end-of-tests"

func (aci *Aci) Test() error {
	defer aci.giveBackUserRightsToTarget()
	hashAcis, err := aci.Install()
	if err != nil {
		return err
	}

	logs.WithF(aci.fields).Info("Testing")

	ImportInternalTesterIfNeeded(aci.manifest)

	logs.WithF(aci.fields).Info("Building test aci")
	hashTestAci, err := aci.buildTestAci()
	if err != nil {
		return err
	}

	logs.WithF(aci.fields).Info("Running test aci")
	if err := aci.runTestAci(hashTestAci, hashAcis); err != nil {
		return err
	}

	logs.WithF(aci.fields).Info("Checking result")
	if err := aci.checkResult(); err != nil {
		return err
	}
	return nil
}

func (aci *Aci) checkResult() error {
	files, err := ioutil.ReadDir(aci.target + pathTestsResult)
	if err != nil {
		return errs.WithEF(err, aci.fields, "Cannot read test result directory")
	}
	testFound := false
	getToTheEnd := false
	for _, f := range files {
		if f.Name() == fileEndOfTests {
			getToTheEnd = true
			continue
		}
		testFields := aci.fields.WithField("file", f.Name())
		fullPath := aci.target + pathTestsResult + "/" + f.Name()
		content, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return errs.WithEF(err, testFields, "Cannot read result file")
		}
		if !strings.HasSuffix(f.Name(), statusSuffix) {
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

func (aci *Aci) runTestAci(testerHash string, hashAcis []string) error {
	os.MkdirAll(aci.target+pathTestsResult, 0777)

	defer aci.cleanupTest(testerHash, hashAcis)
	if err := Home.Rkt.Run([]string{"--set-env=" + common.EnvLogLevel + "=" + logs.GetLevel().String(),
		"--net=host",
		"--mds-register=false",
		"--uuid-file-save=" + aci.target + pathTesterUuid,
		"--volume=" + mountAcname + ",kind=host,source=" + aci.target + pathTestsResult,
		testerHash,
		"--exec", "/test",
	}); err != nil {
		// rkt+systemd cannot exit with fail status yet, so will not happen
		return errs.WithEF(err, aci.fields, "Run of test aci failed")
	}
	return nil
}

func (aci *Aci) cleanupTest(testerHash string, hashAcis []string) {
	if !Args.KeepBuilder {
		if _, _, err := Home.Rkt.RmFromFile(aci.target + pathTesterUuid); err != nil {
			logs.WithEF(err, aci.fields).Warn("Failed to remove test container")
		}
	}

	for _, hash := range hashAcis {
		if err := Home.Rkt.ImageRm(hash); err != nil {
			logs.WithEF(err, aci.fields.WithField("hash", hash)).Warn("Failed to remove container image")
		}
	}

	if err := Home.Rkt.ImageRm(testerHash); err != nil {
		logs.WithEF(err, aci.fields.WithField("hash", testerHash)).Warn("Failed to remove test container image")
	}
}

func (aci *Aci) buildTestAci() (string, error) {
	manifest, err := common.ExtractManifestFromAci(aci.target + common.PathImageAci)
	if err != nil {
		return "", errs.WithEF(err, aci.fields.WithField("file", aci.target+common.PathImageAci), "Failed to extract manifest from aci")
	}

	name := prefixTest + manifest.Name.String()
	if version, ok := manifest.Labels.Get("version"); ok {
		name += ":" + version
	}

	fullname := common.NewACFullName(name)
	resultMountName, _ := types.NewACName(mountAcname)

	aciManifest := &common.AciManifest{
		Builder: aci.manifest.Tester.Builder,
		Aci: common.AciDefinition{
			App: common.DgrApp{
				Exec:              aci.manifest.Aci.App.Exec,
				MountPoints:       []types.MountPoint{{Path: pathTestsResult, Name: *resultMountName}},
				WorkingDirectory:  aci.manifest.Aci.App.WorkingDirectory,
				User:              aci.manifest.Aci.App.User,
				Group:             aci.manifest.Aci.App.Group,
				SupplementaryGIDs: aci.manifest.Aci.App.SupplementaryGIDs,
				Environment:       aci.manifest.Aci.App.Environment,
				Ports:             aci.manifest.Aci.App.Ports,
				Isolators:         aci.manifest.Aci.App.Isolators,
			},
			Dependencies:  append(aci.manifest.Tester.Aci.Dependencies, *common.NewACFullName(name[len(prefixTest):])),
			Annotations:   aci.manifest.Aci.Annotations,
			PathWhitelist: aci.manifest.Aci.PathWhitelist,
		},
		NameAndVersion: *fullname,
	}

	content, err := yaml.Marshal(aciManifest)
	if err != nil {
		return "", errs.WithEF(err, aci.fields, "Failed to marshall manifest for test aci")
	}

	testAci, err := NewAciWithManifest(aci.path, aci.args, string(content), aci.checkWg)
	if err != nil {
		return "", errs.WithEF(err, aci.fields, "Failed to prepare test's build aci")
	}

	testAci.FullyResolveDep = false // this is required to run local tests without discovery
	testAci.target = aci.target + pathTestsTarget

	if err := testAci.CleanAndBuild(); err != nil {
		return "", errs.WithEF(err, aci.fields, "Build of test aci failed")
	}
	hash, err := Home.Rkt.Fetch(aci.target + pathTestsTarget + pathImageAci)
	if err != nil {
		return "", errs.WithEF(err, aci.fields, "fetch of test aci failed")
	}
	return hash, nil
}
