package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"

	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/dgr/common"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

const pathPodManifestJson = "/pod-manifest.json"

func (p *Pod) Build() error {
	defer p.giveBackUserRightsToTarget()
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

func (p *Pod) CleanAndBuild() error {
	p.Clean()
	return p.Build()
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

		if !p.args.ParallelBuild {
			f(i, e)
		} else {
			go f(i, e)
		}
	}
	wg.Wait()

	failed := 0
	for _, err := range errors {
		if err != nil {
			failed++
		}
	}
	if failed > 0 {
		return apps, errs.With("Acis process failed").WithErrs(errors...)
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

	isolators, err := common.ToAppcIsolators(e.App.Isolators)
	if err != nil {
		return nil, errs.WithEF(err, p.fields, "Failed to prepare isolators")
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
			Isolators:         isolators,
		},
		Mounts:      e.Mounts,
		Annotations: e.Annotations}, nil
}

func (p *Pod) fillRuntimeAppFromDependencies(e *common.RuntimeApp) error {
	fields := p.fields.WithField("aci", e.Name)

	dependency := e.InheritDependencyPolicy.GetInheritDependency(*e)
	if dependency == nil {
		return nil
	}

	Home.Rkt.Fetch(dependency.String())
	manifest, err := Home.Rkt.GetManifest(dependency.String())
	if err != nil {
		return errs.WithEF(err, fields.WithField("dependency", dependency.String()), "Failed to get dependency manifest")
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
		res, err := common.FromAppcIsolators(manifest.App.Isolators)
		if err != nil {
			return errs.WithEF(err, fields, "Failed to replicate isolators from aci to pod")
		}
		e.App.Isolators = res
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
	return nil
}

func (p *Pod) buildAci(e common.RuntimeApp) (*Aci, error) {
	if err := p.fillRuntimeAppFromDependencies(&e); err != nil {
		return nil, err
	}

	aci, err := p.toPodAci(e)
	if err != nil {
		return nil, err
	}

	aci.Clean()

	// TODO attributes should be builder dependent only
	if empty, err := common.IsDirEmpty(p.path + "/attributes"); !empty && err == nil {
		path := aci.target + pathBuilder + common.PathRootfs + "/dgr/pod/attributes"
		if err := os.MkdirAll(path, 0777); err != nil {
			return nil, errs.WithEF(err, aci.fields.WithField("path", path), "Failed to create pod attributes directory in builder")
		}
		if err := common.CopyDir(p.path+"/attributes", path); err != nil {
			return nil, errs.WithEF(err, aci.fields, "Failed to copy pod attributes to aci builder")
		}
	}

	if err := aci.Build(); err != nil {
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
