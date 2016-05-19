package main

import (
	"encoding/json"
	"errors"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"os"
	"sync"
)

const pathPodManifestJson = "/pod-manifest.json"

func (p *Pod) CleanAndBuild() error {
	p.Clean()
	logs.WithF(p.fields).Info("Building")

	os.RemoveAll(p.target)
	os.MkdirAll(p.target, 0777)

	p.preparePodVersion()
	apps, err := p.processAcis()
	if err != nil {
		return err
	}

	if err := p.writePodManifest(apps); err != nil {
		return err
	}
	return nil
}

func (p *Pod) preparePodVersion() {
	if p.manifest.Name.Version() == "" {
		p.manifest.Name = *common.NewACFullName(p.manifest.Name.Name() + ":" + common.GenerateVersion(p.path))
	}
}

func (p *Pod) processAcis() ([]schema.RuntimeApp, error) {
	apps := make([]schema.RuntimeApp, len(p.manifest.Pod.Apps))
	errors := make([]error, len(p.manifest.Pod.Apps))
	var wg sync.WaitGroup
	for i, e := range p.manifest.Pod.Apps {
		wg.Add(1)
		f := func(i int, e common.RuntimeApp) {
			defer wg.Done()

			app, err := p.processAci(e)
			if app != nil {
				apps[i] = *app
			}
			errors[i] = err
		}

		if p.args.SerialBuild {
			f(i, e)
		} else {
			go f(i, e)
		}
	}
	wg.Wait()

	failed := 0
	for _, err := range errors {
		if err != nil {
			logs.WithE(err).Error("Aci process failed")
			failed++
		}
	}
	if failed > 0 {
		return apps, errs.With("Aci process failed")
	}
	return apps, nil
}

func (p *Pod) processAci(e common.RuntimeApp) (*schema.RuntimeApp, error) {
	aci, err := p.buildAci(e)
	if err != nil {
		return nil, err
	}

	name, err := types.NewACName(e.Name)
	if err != nil {
		return nil, errs.WithEF(err, p.fields.WithField("name", e.Name), "Invalid name format")
	}

	sum, err := Sha512sum(aci.target + pathImageAci)
	if err != nil {
		return nil, errs.WithEF(err, p.fields.WithField("file", aci.target+pathImageAci), "Failed to calculate sha512 of aci")
	}

	tmp, _ := types.NewHash("sha512-" + sum)

	labels := types.Labels{}
	labels = append(labels, types.Label{Name: "version", Value: aci.manifest.NameAndVersion.Version()})
	identifier, _ := types.NewACIdentifier(aci.manifest.NameAndVersion.Name())
	ttmp := schema.RuntimeImage{Name: identifier, ID: *tmp, Labels: labels}

	e.App.Group = aci.manifest.Aci.App.Group
	e.App.User = aci.manifest.Aci.App.User
	if e.App.User == "" {
		e.App.User = "0"
	}
	if e.App.Group == "" {
		e.App.Group = "0"
	}

	return &schema.RuntimeApp{
		Name:  *name,
		Image: ttmp,
		App: &types.App{
			Exec:              e.App.Exec,
			User:              e.App.User,
			Group:             e.App.Group,
			WorkingDirectory:  e.App.WorkingDirectory,
			SupplementaryGIDs: e.App.SupplementaryGIDs,
			Environment:       e.App.Environment,
			MountPoints:       e.App.MountPoints,
			Ports:             e.App.Ports,
			Isolators:         e.App.Isolators,
		},
		Mounts:      e.Mounts,
		Annotations: e.Annotations}, nil
}

func (p *Pod) fillRuntimeAppFromDependencies(e *common.RuntimeApp) error {
	fields := p.fields.WithField("aci", e.Name)
	if len(e.Dependencies) > 1 && len(e.App.Exec) == 0 {
		return errs.WithF(fields, "There is more than 1 dependency, manifest aci must be set explicitly")
	}

	if len(e.Dependencies) == 1 {
		Home.Rkt.Fetch(e.Dependencies[0].String())
		manifestStr, err := Home.Rkt.CatManifest(e.Dependencies[0].String())
		if err != nil {
			return errs.WithEF(err, fields.WithField("dependency", e.Dependencies[0].String()), "Failed to get dependency manifest")
		}
		manifest := schema.ImageManifest{}
		if err := json.Unmarshal([]byte(manifestStr), &manifest); err != nil {
			return errs.WithEF(err, fields.WithField("content", manifestStr), "Failed to unmarshal stage1 manifest received from rkt")
		}

		if len(e.App.Exec) == 0 {
			e.App.Exec = manifest.App.Exec
		}
		if e.App.User == "" {
			e.App.User = manifest.App.User
		}
		if e.App.Group == "" {
			e.App.Group = manifest.App.Group
		}
		if e.App.WorkingDirectory == "" {
			e.App.WorkingDirectory = manifest.App.WorkingDirectory
		}
		if len(e.App.SupplementaryGIDs) == 0 {
			e.App.SupplementaryGIDs = manifest.App.SupplementaryGIDs
		}
		if len(e.App.Isolators) == 0 {
			e.App.Isolators = manifest.App.Isolators
		}
		if len(e.App.Ports) == 0 {
			e.App.Ports = manifest.App.Ports
		}
		if len(e.App.MountPoints) == 0 {
			e.App.MountPoints = manifest.App.MountPoints
		}
		if len(e.App.Environment) == 0 {
			e.App.Environment = manifest.App.Environment
		}

		anns := e.Annotations
		e.Annotations = manifest.Annotations
		for _, ann := range anns {
			e.Annotations.Set(ann.Name, ann.Value)
		}
	}
	return nil
}

func (p *Pod) buildAci(e common.RuntimeApp) (*Aci, error) {
	if err := p.fillRuntimeAppFromDependencies(&e); err != nil {
		return nil, err
	}

	path := p.path + "/" + e.Name
	if dir, err := os.Stat(path); err != nil || !dir.IsDir() {
		path = p.target + "/" + e.Name
		if err := os.Mkdir(path, 0777); err != nil {
			return nil, errs.WithEF(err, p.fields.WithField("path", path), "Cannot created pod's aci directory")
		}
	}

	tmpl, err := p.toAciManifestTemplate(e)
	if err != nil {
		return nil, err
	}
	aci, err := NewAciWithManifest(path, p.args, tmpl, p.checkWg)
	if err != nil {
		return nil, errs.WithEF(err, p.fields.WithField("aci", path), "Failed to prepare aci")
	}
	aci.podName = &p.manifest.Name
	if err := aci.CleanAndBuild(); err != nil {
		return nil, errs.WithEF(err, p.fields.WithField("name", e.Name), "build of  pod's aci failed")
	}
	return aci, nil
}

func (p *Pod) writePodManifest(apps []schema.RuntimeApp) error {
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
	return WritePodManifest(&manifest, p.target+pathPodManifestJson)
}

func WritePodManifest(im *schema.PodManifest, targetFile string) error {
	buff, err := json.MarshalIndent(im, "", "  ")
	if err != nil {
		return errs.WithEF(err, data.WithField("object", im), "Failed to marshal manifest")
	}
	err = ioutil.WriteFile(targetFile, []byte(buff), 0644)
	if err != nil {
		return errs.WithEF(err, data.WithField("file", targetFile), "Failed to write pod manifest")
	}
	return nil
}

func (p *Pod) getVolumeMountValue(mountName types.ACName) (*types.Volume, error) {
	for _, volume := range p.manifest.Pod.Volumes {
		if volume.Name.Equals(mountName) {
			return &volume, nil
		}
	}
	return nil, errors.New("Volume mount point not set :" + mountName.String())
}
