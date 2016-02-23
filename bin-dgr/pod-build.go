package main

import (
	"errors"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
)

const PATH_POD_MANIFEST = "/pod-manifest.json"

func (p *Pod) Build() error {
	logs.WithF(p.fields).Info("Building")

	os.RemoveAll(p.target)
	os.MkdirAll(p.target, 0777)

	p.preparePodVersion()
	apps := p.processAci()

	p.writePodManifest(apps)
	return nil
}

func (p *Pod) preparePodVersion() {
	if p.manifest.Name.Version() == "" {
		p.manifest.Name = *common.NewACFullName(p.manifest.Name.Name() + ":" + GenerateVersion())
	}
}

func (p *Pod) processAci() []schema.RuntimeApp {
	apps := []schema.RuntimeApp{}
	for _, e := range p.manifest.Pod.Apps {

		aci := p.buildAci(e)

		name, _ := types.NewACName(e.Name)

		sum, err := Sha512sum(aci.target + "/image.aci")
		if err != nil {
			panic(err)
		}

		tmp, _ := types.NewHash("sha512-" + sum)

		labels := types.Labels{}
		labels = append(labels, types.Label{Name: "version", Value: aci.manifest.NameAndVersion.Version()})
		identifier, _ := types.NewACIdentifier(aci.manifest.NameAndVersion.Name())
		ttmp := schema.RuntimeImage{Name: identifier, ID: *tmp, Labels: labels}

		if e.App.User == "" {
			e.App.User = "0"
		}
		if e.App.Group == "" {
			e.App.Group = "0"
		}

		apps = append(apps, schema.RuntimeApp{
			Name:  *name,
			Image: ttmp,
			App: &types.App{
				Exec:             e.App.Exec,
				User:             e.App.User,
				Group:            e.App.Group,
				WorkingDirectory: e.App.WorkingDirectory,
				Environment:      e.App.Environment,
				MountPoints:      e.App.MountPoints,
				Ports:            e.App.Ports,
				Isolators:        e.App.Isolators,
			},
			Mounts:      e.Mounts,
			Annotations: e.Annotations})
	}

	return apps
}

func (p *Pod) buildAci(e RuntimeApp) *Aci {
	path := p.path + "/" + e.Name
	if dir, err := os.Stat(path); err != nil || !dir.IsDir() {
		if err := os.Mkdir(path, 0777); err != nil {
			logs.WithEF(err, p.fields).WithField("path", path).Fatal("Cannot created pod's aci directory")
		}
	}
	aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e))
	if err != nil {
		logs.WithEF(err, p.fields).WithField("aci", path).Fatal("Failed to prepare aci")
	}
	aci.podName = &p.manifest.Name
	aci.Build()
	return aci
}

func (p *Pod) writePodManifest(apps []schema.RuntimeApp) {
	m := p.manifest.Pod
	ver, _ := types.NewSemVer("0.6.1")
	manifest := schema.PodManifest{
		ACKind:      "PodManifest",
		ACVersion:   *ver,
		Apps:        apps,
		Volumes:     m.Volumes,
		Isolators:   m.Isolators,
		Annotations: m.Annotations,
		Ports:       m.Ports}
	WritePodManifest(&manifest, p.target+PATH_POD_MANIFEST)
}

func WritePodManifest(im *schema.PodManifest, targetFile string) {
	buff, err := im.MarshalJSON()
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(targetFile, []byte(buff), 0644)
	if err != nil {
		panic(err)
	}
}

func (p *Pod) getVolumeMountValue(mountName types.ACName) (*types.Volume, error) {
	for _, volume := range p.manifest.Pod.Volumes {
		if volume.Name.Equals(mountName) {
			return &volume, nil
		}
	}
	return nil, errors.New("Volume mount point not set :" + mountName.String())
}
