package builder

import (
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/spec"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"path/filepath"
)

const POD_MANIFEST = "/cnt-pod-manifest.yml"

//const POD_TARGET_MANIFEST = "/pod-manifest.json"

type Pod struct {
	path     string
	args     BuildArgs
	target   string
	manifest spec.PodManifest
}

func OpenPod(path string, args BuildArgs) (*Pod, error) {
	pod := new(Pod)

	if fullPath, err := filepath.Abs(path); err != nil {
		log.Get().Panic("Cannot get fullpath of project", err)
	} else {
		pod.path = fullPath
	}
	pod.args = args
	pod.target = pod.path + PATH_TARGET
	pod.readManifest(pod.path + POD_MANIFEST)
	return pod, nil
}

func (p *Pod) readManifest(manifestPath string) {
	source, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		log.Get().Panic(err)
	}

	err = yaml.Unmarshal([]byte(source), &p.manifest)
	if err != nil {
		log.Get().Panic(err)
	}

	for i, app := range p.manifest.Pod.Apps {
		if app.Name == "" {
			p.manifest.Pod.Apps[i].Name = app.Dependencies[0].ShortName()
		}
	}

	//TODO check that there is no app name conflict

	log.Get().Trace("Pod manifest : ", p.manifest.Name, p.manifest)
}

func (p *Pod) toAciManifest(e spec.RuntimeApp) spec.AciManifest {
	fullname, _ := spec.NewACFullName(p.manifest.Name.Name() + "_" + e.Name + ":" + p.manifest.Name.Version())
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
