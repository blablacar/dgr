package main

import (
	"github.com/appc/spec/schema"
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/jhoonb/archivex"
	gzip "github.com/klauspost/pgzip"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

const pathGraphPng = "/graph.png"
const pathGraphDot = "/graph.dot"
const pathImageAci = "/image.aci"
const pathImageAciAsc = "/image.aci.asc"
const pathImageGzAci = "/image.gz.aci"
const pathImageGzAciAsc = "/image.gz.aci.asc"
const pathTarget = "/target"
const pathManifestJson = "/manifest.json"

const pathStage1 = "/stage1"

const pathBuilder = "/builder"
const pathBuilderUuid = "/builder.uuid"
const pathTesterUuid = "/tester.uuid"
const pathVersion = "/version"

const suffixAsc = ".asc"
const prefixTest = "test/"
const prefixBuilderStage1 = "builder-stage1/"

type Aci struct {
	checkWg         *sync.WaitGroup
	fields          data.Fields
	path            string
	target          string
	podName         *common.ACFullname
	manifestTmpl    string
	manifest        *common.AciManifest
	args            BuildArgs
	FullyResolveDep bool
}

func NewAciWithManifest(path string, args BuildArgs, manifestTmpl string, checkWg *sync.WaitGroup) (*Aci, error) {
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
		checkWg:         checkWg,
	}

	return aci, nil
}

func NewAci(path string, args BuildArgs, checkWg *sync.WaitGroup) (*Aci, error) {
	manifest, err := ioutil.ReadFile(path + common.PathAciManifest)
	if err != nil {
		return nil, errs.WithEF(err, data.WithField("path", path+common.PathAciManifest), "Cannot read manifest")
	}
	return NewAciWithManifest(path, args, string(manifest), checkWg)
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
	if err := tar.AddAll(path+common.PathRootfs, true); err != nil {
		return errs.WithEF(err, aci.fields.WithField("path", path+common.PathRootfs), "Failed to add rootfs to tar")
	}
	if err := tar.Close(); err != nil {
		return errs.WithEF(err, aci.fields.WithField("path", path), "Failed to tar aci")
	}
	os.Rename(path+pathImageAci+".tar", path+pathImageAci)
	return nil
}

func (aci *Aci) zipAci() error {
	source := aci.target + pathImageAci
	target := aci.target + pathImageGzAci
	if _, err := os.Stat(target); err == nil {
		return nil
	}

	logs.WithF(aci.fields).Info("Gzipping aci")
	reader, err := os.Open(source)
	if err != nil {
		return errs.WithEF(err, aci.fields.WithField("path", source), "Failed to open unziped aci")
	}
	filename := filepath.Base(source)
	writer, err := os.Create(target)
	if err != nil {
		return errs.WithEF(err, aci.fields.WithField("path", target), "Failed to create file descriptor for image zip")
	}
	defer writer.Close()
	archiver := gzip.NewWriter(writer)
	archiver.SetConcurrency(100000, 10)
	archiver.Name = filename
	defer archiver.Close()
	_, err = io.Copy(archiver, reader)
	if err != nil {
		return errs.WithEF(err, aci.fields.WithField("path", target), "Failed to zip aci")
	}
	return nil
}

func (aci *Aci) checkCompatibilityVersions() {
	defer aci.checkWg.Done()
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
	defer aci.checkWg.Done()
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
