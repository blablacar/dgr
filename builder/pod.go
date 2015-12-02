package builder

import (
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/cnt/cnt"
	"github.com/blablacar/cnt/spec"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"path/filepath"
)

const POD_MANIFEST = "/cnt-pod-manifest.yml"

type Pod struct {
	log      log.Entry
	path     string
	args     BuildArgs
	target   string
	manifest spec.PodManifest
}

func NewPod(path string, args BuildArgs) (*Pod, error) {
	fullPath, err := filepath.Abs(path)
	if err != nil {
		panic("Cannot get fullpath of project" + err.Error())
	}

	manifest := readPodManifest(fullPath + POD_MANIFEST)
	podLog := log.WithField("pod", manifest.Name.String())

	target := path + PATH_TARGET
	if cnt.Home.Config.TargetWorkDir != "" {
		currentAbsDir, err := filepath.Abs(cnt.Home.Config.TargetWorkDir + "/" + manifest.Name.ShortName())
		if err != nil {
			podLog.WithError(err).Panic("invalid target path")
		}
		target = currentAbsDir
	}

	pod := &Pod{
		log:      *podLog,
		path:     fullPath,
		args:     args,
		target:   target,
		manifest: *manifest,
	}

	return pod, nil
}

func readPodManifest(manifestPath string) *spec.PodManifest {
	source, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		panic(err)
	}

	manifest := &spec.PodManifest{}
	err = yaml.Unmarshal([]byte(source), manifest)
	if err != nil {
		panic(err)
	}

	for i, app := range manifest.Pod.Apps {
		if app.Name == "" {
			manifest.Pod.Apps[i].Name = app.Dependencies[0].ShortName()
		}
	}
	//TODO check that there is no app name conflict
	return manifest
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
