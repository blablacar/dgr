package main

import (
	"github.com/appc/spec/schema"
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/jhoonb/archivex"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

const pathGraphPng = "/graph.png"
const pathGraphDot = "/graph.dot"
const pathImageAci = "/image.aci"
const pathImageGzAci = "/image.gz.aci"
const pathImageGzAciAsc = "/image.gz.aci.asc"
const pathTarget = "/target"
const pathManifestJson = "/manifest.json"

const pathStage1 = "/stage1"

const pathBuilder = "/builder"
const pathBuilderUuid = "/builder.uuid"
const pathTesterUuid = "/tester.uuid"

const prefixTest = "test/"
const prefixBuilderStage1 = "builder-stage1/"

type Aci struct {
	fields          data.Fields
	path            string
	target          string
	podName         *common.ACFullname
	manifestTmpl    string
	manifest        *common.AciManifest
	args            BuildArgs
	FullyResolveDep bool
}

func NewAciWithManifest(path string, args BuildArgs, manifestTmpl string) (*Aci, error) {
	manifest, err := common.ProcessManifestTemplate(manifestTmpl, nil, false)
	if err != nil {
		return nil, errs.WithEF(err, data.WithField("content", manifestTmpl), "Failed to process manifest")
	}
	if manifest.NameAndVersion == "" {
		logs.WithField("path", path).Fatal("name is mandatory in manifest")
	}

	fields := data.WithField("aci", manifest.NameAndVersion.String())
	logs.WithF(fields).WithFields(data.Fields{"args": args, "path": path, "manifest": manifest}).Debug("New aci")

	fullPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errs.WithEF(err, fields, "Cannot get fullpath of project")
	}

	target := fullPath + pathTarget
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
		manifestTmpl:    manifestTmpl,
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
	manifest, err := ioutil.ReadFile(path + common.PathAciManifest)
	if err != nil {
		return nil, errs.WithEF(err, data.WithField("path", path+common.PathAciManifest), "Cannot read manifest")
	}
	return NewAciWithManifest(path, args, string(manifest))
}

//////////////////////////////////////////////////////////////////

func (aci *Aci) tarAci(path string) error {
	tar := new(archivex.TarFile)
	if err := tar.Create(path + pathImageAci); err != nil {
		return errs.WithEF(err, aci.fields.WithField("path", path+pathImageAci), "Failed to create image tar")
	}
	if err := tar.AddFile(path + common.PathManifest); err != nil {
		return errs.WithEF(err, aci.fields.WithField("path", path+common.PathManifest), "Failed to add manifest to tar")
	}
	if err := tar.AddAll(path+common.PathRootfs, false); err != nil {
		return errs.WithEF(err, aci.fields.WithField("path", path+common.PathRootfs), "Failed to add rootfs to tar")
	}
	if err := tar.Close(); err != nil {
		return errs.WithEF(err, aci.fields.WithField("path", path), "Failed to tar aci")
	}
	os.Rename(path+pathImageAci+".tar", path+pathImageAci)
	return nil
}

func (aci *Aci) zipAci() error {
	if _, err := os.Stat(aci.target + pathImageGzAci); err == nil {
		return nil
	}
	if stdout, stderr, err := common.ExecCmdGetStdoutAndStderr("gzip", "-k", aci.target+pathImageAci); err != nil {
		return errs.WithEF(err, aci.fields.WithField("path", aci.target+pathImageAci).WithField("stdout", stdout).WithField("stderr", stderr), "Failed to zip aci")
	}
	if err := common.ExecCmd("mv", aci.target+pathImageAci+".gz", aci.target+pathImageGzAci); err != nil {
		return errs.WithEF(err, aci.fields.WithField("from", aci.target+pathImageAci+".gz").
			WithField("to", aci.target+pathImageGzAci), "Failed to rename zip aci")
	}
	return nil
}

func (aci *Aci) checkCompatibilityVersions() {
	for _, dep := range aci.manifest.Aci.Dependencies {
		depFields := aci.fields.WithField("dependency", dep.String())

		logs.WithF(aci.fields).WithField("dependency", dep.String()).Info("Fetching dependency")
		Home.Rkt.Fetch(dep.String())
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

func (aci *Aci) checkLatestVersions() {
	CheckLatestVersion(aci.manifest.Aci.Dependencies, "dependency")
	CheckLatestVersion(aci.manifest.Builder.Dependencies, "builder dependency")
	CheckLatestVersion(aci.manifest.Tester.Builder.Dependencies, "tester builder dependency")
	CheckLatestVersion(aci.manifest.Tester.Aci.Dependencies, "tester dependency")
}

func CheckLatestVersion(deps []common.ACFullname, warnText string) {
	for _, dep := range deps {
		if dep.Version() == "" {
			continue
		}
		version, _ := dep.LatestVersion()
		if version != "" && common.Version(dep.Version()).LessThan(common.Version(version)) {
			logs.WithField("newer", dep.Name()+":"+version).
				WithField("current", dep.String()).
				Warn("Newer " + warnText + " version")
		}
	}
}

func GetDependencyDgrVersion(acName common.ACFullname) (int, error) {
	depFields := data.WithField("dependency", acName.String())

	out, err := Home.Rkt.CatManifest(acName.String())
	if err != nil {
		return 0, errs.WithEF(err, depFields, "Dependency not found")
	}

	im := schema.ImageManifest{}
	if err := im.UnmarshalJSON([]byte(out)); err != nil {
		return 0, errs.WithEF(err, depFields.WithField("content", out), "Cannot read manifest cat by rkt image")
	}

	version, ok := im.Annotations.Get(common.ManifestDrgVersion)
	var val int
	if ok {
		val, err = strconv.Atoi(version)
		if err != nil {
			return 0, errs.WithEF(err, depFields.WithField("version", version), "Failed to parse "+common.ManifestDrgVersion+" from manifest")
		}
	}
	return val, nil
}

func (aci *Aci) giveBackUserRightsToTarget() {
	giveBackUserRights(aci.target)
}
