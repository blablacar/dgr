package utils

import (
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/spec"
	"io/ioutil"
)

//
//func ReadManifest(path string) *schema.ImageManifest {
//    im := new(schema.ImageManifest)
//    content, err := ioutil.ReadFile(path)
//    if  err != nil {
//        panic(err)
////        config.GetConfig().Log.Panic("Cannot read manifest file", err)
//    }
//    im.UnmarshalJSON(content)
//    return im
//}

func WriteImageManifest(m *spec.AciManifest, targetFile string, projectName string, dgrVersion string) {
	name, err := types.NewACIdentifier(m.NameAndVersion.Name())
	if err != nil {
		panic(err)
	}

	version := m.NameAndVersion.Version()
	if version == "" {
		version = GenerateVersion()
	}

	labels := types.Labels{}
	labels = append(labels, types.Label{Name: "version", Value: version})
	labels = append(labels, types.Label{Name: "os", Value: "linux"})
	labels = append(labels, types.Label{Name: "arch", Value: "amd64"})

	if m.Aci.App.User == "" {
		m.Aci.App.User = "0"
	}
	if m.Aci.App.Group == "" {
		m.Aci.App.Group = "0"
	}

	im := schema.BlankImageManifest()
	im.Annotations = m.Aci.Annotations

	dgrVersionIdentifier, _ := types.NewACIdentifier("dgr-version")
	im.Annotations.Set(*dgrVersionIdentifier, dgrVersion)
	im.Dependencies = toAppcDependencies(m.Aci.Dependencies)
	im.Name = *name
	im.Labels = labels

	if len(m.Aci.App.Exec) == 0 {
		m.Aci.App.Exec = []string{"/dgr/bin/busybox", "sh"}
	}

	im.App = &types.App{
		Exec:             m.Aci.App.Exec,
		EventHandlers:    []types.EventHandler{{Name: "pre-start", Exec: []string{"/dgr/bin/prestart"}}},
		User:             m.Aci.App.User,
		Group:            m.Aci.App.Group,
		WorkingDirectory: m.Aci.App.WorkingDirectory,
		Environment:      m.Aci.App.Environment,
		MountPoints:      m.Aci.App.MountPoints,
		Ports:            m.Aci.App.Ports,
		Isolators:        m.Aci.App.Isolators,
	}

	buff, err := im.MarshalJSON()
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(targetFile, buff, 0644)
	if err != nil {
		panic(err)
	}
}

func toAppcDependencies(dependencies []spec.ACFullname) types.Dependencies {
	appcDependencies := types.Dependencies{}
	for _, dep := range dependencies {
		id, err := types.NewACIdentifier(dep.Name())
		if err != nil {
			panic(err)
		}
		t := types.Dependency{ImageName: *id}
		if dep.Version() != "" {
			t.Labels = types.Labels{}
			t.Labels = append(t.Labels, types.Label{Name: "version", Value: dep.Version()})
		}

		appcDependencies = append(appcDependencies, t)
	}
	return appcDependencies
}
