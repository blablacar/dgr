package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/blablacar/dgr/dgr/common"
	"github.com/jhoonb/archivex"
	gzip "github.com/klauspost/pgzip"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
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

const pathBuilderUuid = "/builder.uuid"
const pathTesterUuid = "/tester.uuid"
const pathVersion = "/version"
const pathUpdated = "/updated"

const suffixAsc = ".asc"
const prefixTest = "test/"
const prefixBuilderStage1 = "builder-stage1/"

type Aci struct {
	fields          data.Fields
	path            string
	target          string
	podName         *common.ACFullname
	manifest        *common.AciManifest
	manifestTmpl    string
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
		manifest:        manifest,
		manifestTmpl:    manifestTmpl,
		target:          target,
		FullyResolveDep: true,
	}

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

func (aci *Aci) giveBackUserRightsToTarget() {
	giveBackUserRights(aci.target)
}
