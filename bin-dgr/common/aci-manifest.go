package common

import (
	"github.com/appc/spec/aci"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func ExtractManifestContentFromAci(aciPath string) ([]byte, error) {
	fields := data.WithField("file", aciPath)
	input, err := os.Open(aciPath)
	if err != nil {
		return nil, errs.WithEF(err, fields, "Cannot open file")
	}
	defer input.Close()

	tr, err := aci.NewCompressedTarReader(input)
	if err != nil {
		return nil, errs.WithEF(err, fields, "Cannot open file as tar")
	}

Tar:
	for {
		hdr, err := tr.Next()
		switch err {
		case io.EOF:
			break Tar
		case nil:
			if filepath.Clean(hdr.Name) == aci.ManifestFile {
				bytes, err := ioutil.ReadAll(tr)
				if err != nil {
					return nil, errs.WithEF(err, fields, "Cannot read manifest content in tar")
				}
				return bytes, nil
			}
		default:
			return nil, errs.WithEF(err, fields, "error reading tarball file")
		}
	}
	return nil, errs.WithEF(err, fields, "Cannot found manifest in file")
}

func ExtractManifestFromAci(aciPath string) (*schema.ImageManifest, error) {
	fields := data.WithField("file", aciPath)
	content, err := ExtractManifestContentFromAci(aciPath)
	if err != nil {
		return nil, errs.WithEF(err, fields, "Cannot extract aci manifest content from file")
	}
	im := &schema.ImageManifest{}

	err = im.UnmarshalJSON(content)
	if err != nil {
		return nil, errs.WithEF(err, fields.WithField("content", string(content)), "Cannot unmarshall json content")
	}
	return im, nil
}

func WriteAciManifest(m *AciManifest, targetFile string, projectName string) error {
	fields := data.WithField("name", m.NameAndVersion.String())
	name, err := types.NewACIdentifier(projectName)
	if err != nil {
		return errs.WithEF(err, fields, "aci name is not a valid identifier for rkt")
	}

	labels := types.Labels{}
	if m.NameAndVersion.Version() != "" {
		labels = append(labels, types.Label{Name: "version", Value: m.NameAndVersion.Version()})
	}
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

	//dgrBuilderIdentifier, _ := types.NewACIdentifier(ManifestDrgBuilder)
	dgrVersionIdentifier, _ := types.NewACIdentifier(ManifestDrgVersion)
	buildDateIdentifier, _ := types.NewACIdentifier("build-date")
	im.Annotations.Set(*dgrVersionIdentifier, DgrVersion)
	//im.Annotations.Set(*dgrBuilderIdentifier, m.Builder.Image.String())
	im.Annotations.Set(*buildDateIdentifier, time.Now().Format(time.RFC3339))
	im.Dependencies, err = ToAppcDependencies(m.Aci.Dependencies)
	if err != nil {
		return errs.WithEF(err, fields, "Failed to prepare dependencies for manifest")
	}
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
		return errs.WithEF(err, fields.WithField("object", im), "Failed to marshal manifest")
	}
	err = ioutil.WriteFile(targetFile, buff, 0644)
	if err != nil {
		return errs.WithEF(err, fields.WithField("file", targetFile), "Failed to write manifest file")
	}
	return nil
}

func ToAppcDependencies(dependencies []ACFullname) (types.Dependencies, error) {
	appcDependencies := types.Dependencies{}
	for _, dep := range dependencies {
		id, err := types.NewACIdentifier(dep.Name())
		if err != nil {
			return nil, errs.WithEF(err, data.WithField("name", dep.Name()), "invalid identifer name for rkt")
		}
		t := types.Dependency{ImageName: *id}
		if dep.Version() != "" {
			t.Labels = types.Labels{}
			t.Labels = append(t.Labels, types.Label{Name: "version", Value: dep.Version()})
		}

		appcDependencies = append(appcDependencies, t)
	}
	return appcDependencies, nil
}
