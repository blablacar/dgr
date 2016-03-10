package main

import (
	"github.com/appc/spec/schema"
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/ghodss/yaml"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

const PATH_GRAPH_DOT = "/graph.dot"
const PATH_INSTALLED = "/installed"
const PATH_IMAGE_ACI = "/image.aci"
const PATH_IMAGE_ACI_ZIP = "/image-zip.aci"
const PATH_TARGET = "/target"
const PATH_ACI_MANIFEST = "/aci-manifest.yml"
const PATH_MANIFEST_JSON = "/manifest.json"
const PATH_TMP = "/tmp"

const PATH_BUILDER = "/builder"
const PATH_BUILDER_UUID = "/builder.uuid"

const PATH_TEST_BUILDER = "/test-builder"
const PATH_TEST_BUILDER_UUID = "/test-builder.uuid"

const MANIFEST_DRG_BUILDER = "dgr-builder"
const MANIFEST_DRG_VERSION = "dgr-version"

const PREFIX_TEST_BUILDER = "test-builder/"
const PREFIX_BUILDER = "builder/"

type Aci struct {
	fields          data.Fields
	path            string
	target          string
	podName         *common.ACFullname
	manifest        *AciManifest
	args            BuildArgs
	FullyResolveDep bool
}

func NewAciWithManifest(path string, args BuildArgs, manifest *AciManifest) (*Aci, error) {
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
	if Home.Config.TargetWorkDir != "" {
		currentAbsDir, err := filepath.Abs(Home.Config.TargetWorkDir + "/" + manifest.NameAndVersion.ShortName())
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
		FullyResolveDep: true,
	}

	froms, err := manifest.GetFroms()
	if err != nil {
		logs.WithEF(err, aci.fields).Fatal("Invalid from data")
	}
	if len(froms) != 0 {
		if froms[0].String() == "" {
			logs.WithF(aci.fields).Warn("From is deprecated and empty, remove it")
		} else {
			logs.WithF(aci.fields).Warn("From is deprecated and processed as dependency. move from to dependencies")
			aci.manifest.Aci.Dependencies = append(froms, aci.manifest.Aci.Dependencies...)
		}
	}

	go aci.checkCompatibilityVersions()
	go aci.checkLatestVersions()
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
	return NewAciWithManifest(path, args, manifest)
}

//////////////////////////////////////////////////////////////////

func readAciManifest(manifestPath string) (*AciManifest, error) {
	manifest := AciManifest{Aci: AciDefinition{}}

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

func (aci *Aci) tarAci(path string, zip bool) error {
	target := PATH_IMAGE_ACI[1:]
	if zip {
		target = PATH_IMAGE_ACI_ZIP[1:]
	}
	dir, _ := os.Getwd()
	logs.WithField("path", path).Debug("chdir")
	os.Chdir(path)
	if err := common.Tar(zip, target, common.PATH_MANIFEST[1:], common.PATH_ROOTFS[1:]); err != nil {
		return errs.WithEF(err, aci.fields.WithField("path", path), "Failed to tar container")
	}
	logs.WithField("path", dir).Debug("chdir")
	os.Chdir(dir)
	return nil
}

func (aci *Aci) checkCompatibilityVersions() {
	for _, dep := range aci.manifest.Aci.Dependencies {
		depFields := aci.fields.WithField("dependency", dep.String())
		common.ExecCmdGetStdoutAndStderr("rkt", "--insecure-options=image", "fetch", dep.String())

		version, err := GetDependencyDgrVersion(dep)
		if err != nil {
			logs.WithEF(err, depFields).Error("Failed to check compatibility version of dependency")
		} else {
			if version < 55 {
				logs.WithF(aci.fields).
					WithField("dependency", dep).
					WithField("require", ">=55").
					Error("dependency was not build with a compatible version of dgr")
			}
		}

	}
}

func GetDependencyDgrVersion(acName common.ACFullname) (int, error) {
	depFields := data.WithField("dependency", acName.String())
	out, stderr, err := common.ExecCmdGetStdoutAndStderr("rkt", "image", "cat-manifest", acName.String())
	if err != nil {
		return 0, errs.WithEF(err, depFields.WithField("stderr", stderr), "Dependency not found")
	}

	im := schema.ImageManifest{}
	if err := im.UnmarshalJSON([]byte(out)); err != nil {
		return 0, errs.WithEF(err, depFields.WithField("content", out), "Cannot read manifest cat by rkt image")
	}

	version, ok := im.Annotations.Get(MANIFEST_DRG_VERSION)
	var val int
	if ok {
		val, err = strconv.Atoi(version)
		if err != nil {
			return 0, errs.WithEF(err, depFields.WithField("version", version), "Failed to parse "+MANIFEST_DRG_VERSION+" from manifest")
		}
	}
	return val, nil
}

func (aci *Aci) giveBackUserRightsToTarget() {
	giveBackUserRights(aci.target)
}

func (aci *Aci) checkLatestVersions() {
	for _, dep := range aci.manifest.Aci.Dependencies {
		if dep.Version() == "" {
			continue
		}
		version, _ := dep.LatestVersion()
		if version != "" && Version(dep.Version()).LessThan(Version(version)) {
			logs.WithField("newer", dep.Name()+":"+version).
				WithField("current", dep.String()).
				Warn("Newer 'dependency' version")
		}
	}
}
