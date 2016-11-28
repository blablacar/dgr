package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/ghodss/yaml"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

const pathPodManifestYml = "/pod-manifest.yml"

type Pod struct {
	checkWg  *sync.WaitGroup
	fields   data.Fields
	path     string
	args     BuildArgs
	target   string
	manifest common.PodManifest
}

func NewPod(path string, args BuildArgs, checkWg *sync.WaitGroup) (*Pod, error) {
	if (args.CatchOnError || args.CatchOnStep) && args.ParallelBuild {
		args.ParallelBuild = false
	}

	fullPath, err := filepath.Abs(path)
	if err != nil {
		logs.WithE(err).WithField("path", path).Fatal("Cannot get fullpath")
	}

	manifest, err := readPodManifest(fullPath + pathPodManifestYml)
	if err != nil {
		return nil, errs.WithEF(err, data.WithField("path", fullPath+pathPodManifestYml), "Failed to read pod manifest")
	}
	fields := data.WithField("pod", manifest.Name.String())

	target := path + pathTarget
	if Home.Config.TargetWorkDir != "" {
		currentAbsDir, err := filepath.Abs(Home.Config.TargetWorkDir + "/" + manifest.Name.ShortName())
		if err != nil {
			logs.WithEF(err, fields).Panic("invalid target path")
		}
		target = currentAbsDir
	}

	pod := &Pod{
		checkWg:  checkWg,
		fields:   fields,
		path:     fullPath,
		args:     args,
		target:   target,
		manifest: *manifest,
	}

	return pod, nil
}

func readPodManifest(manifestPath string) (*common.PodManifest, error) {
	source, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	manifest := &common.PodManifest{}
	err = yaml.Unmarshal([]byte(source), manifest)
	if err != nil {
		return nil, err
	}

	for i, app := range manifest.Pod.Apps {
		if app.Name == "" {
			manifest.Pod.Apps[i].Name = app.Dependencies[0].TinyName()
		}
	}
	//TODO check that there is no app name conflict
	return manifest, nil
}

func (p *Pod) findAciDirectory(e common.RuntimeApp) (string, error) {
	path := p.path + "/" + e.Name
	if dir, err := os.Stat(path); err != nil || !dir.IsDir() {
		path = p.target + "/" + e.Name
		if err := os.MkdirAll(path, 0777); err != nil {
			return "", errs.WithEF(err, p.fields.WithField("path", path), "Cannot created pod's aci directory")
		}
	}
	return path, nil
}

func (p *Pod) toPodAci(e common.RuntimeApp) (*Aci, error) {
	tmpl, err := p.toAciManifestTemplate(e)
	if err != nil {
		return nil, err
	}
	dir, err := p.findAciDirectory(e)
	if err != nil {
		return nil, err
	}

	aci, err := NewAciWithManifest(dir, p.args, tmpl, p.checkWg)
	if err != nil {
		return nil, errs.WithEF(err, p.fields.WithField("aci-dir", dir), "Failed to prepare aci")
	}
	aci.podName = &p.manifest.Name
	return aci, err
}

func (p *Pod) toAciManifestTemplate(e common.RuntimeApp) (string, error) {
	fullname := common.NewACFullName(p.manifest.Name.Name() + "_" + e.Name + ":" + p.manifest.Name.Version())
	manifest := &common.AciManifest{
		Aci: common.AciDefinition{
			Annotations:   e.Annotations,
			App:           e.App,
			Dependencies:  e.Dependencies,
			PathWhitelist: nil, // TODO
		},
		NameAndVersion: *fullname,
	}
	content, err := yaml.Marshal(manifest)
	if err != nil {
		return "", errs.WithEF(err, p.fields, "Failed to prepare manifest template")
	}
	return string(content), nil
}

func (p *Pod) giveBackUserRightsToTarget() {
	giveBackUserRights(p.target)
}
