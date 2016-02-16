package builder

import (
	"github.com/blablacar/dgr/dgr"
	"github.com/blablacar/dgr/spec"
	"github.com/ghodss/yaml"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"path/filepath"
)

const POD_MANIFEST = "/pod-manifest.yml"

type Pod struct {
	fields   data.Fields
	path     string
	args     BuildArgs
	target   string
	manifest spec.PodManifest
}

func NewPod(path string, args BuildArgs) (*Pod, error) {
	fullPath, err := filepath.Abs(path)
	if err != nil {
		logs.WithE(err).WithField("path", path).Fatal("Cannot get fullpath")
	}

	manifest, err := readPodManifest(fullPath + POD_MANIFEST)
	if err != nil {
		manifest2, err2 := readPodManifest(fullPath + "/cnt-pod-manifest.yml")
		if err2 != nil {
			return nil, errs.WithEF(err, data.WithField("path", fullPath+POD_MANIFEST).WithField("err2", err2), "Failed to read pod manifest")
		}
		logs.WithField("old", "cnt-pod-manifest.yml").WithField("new", "pod-manifest.yml").Warn("You are using the old aci configuration file")
		manifest = manifest2
	}
	fields := data.WithField("pod", manifest.Name.String())

	target := path + PATH_TARGET
	if dgr.Home.Config.TargetWorkDir != "" {
		currentAbsDir, err := filepath.Abs(dgr.Home.Config.TargetWorkDir + "/" + manifest.Name.ShortName())
		if err != nil {
			logs.WithEF(err, fields).Panic("invalid target path")
		}
		target = currentAbsDir
	}

	pod := &Pod{
		fields:   fields,
		path:     fullPath,
		args:     args,
		target:   target,
		manifest: *manifest,
	}

	return pod, nil
}

func readPodManifest(manifestPath string) (*spec.PodManifest, error) {
	source, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	manifest := &spec.PodManifest{}
	err = yaml.Unmarshal([]byte(source), manifest)
	if err != nil {
		return nil, err
	}

	for i, app := range manifest.Pod.Apps {
		if app.Name == "" {
			manifest.Pod.Apps[i].Name = app.Dependencies[0].ShortName()
		}
	}
	//TODO check that there is no app name conflict
	return manifest, nil
}

func (p *Pod) toAciManifest(e spec.RuntimeApp) spec.AciManifest {
	fullname := spec.NewACFullName(p.manifest.Name.Name() + "_" + e.Name + ":" + p.manifest.Name.Version())
	return spec.AciManifest{
		Aci: spec.AciDefinition{
			Annotations:   e.Annotations,
			App:           e.App,
			Dependencies:  e.Dependencies,
			PathWhitelist: nil, // TODO
		},
		NameAndVersion: *fullname,
	}
}
