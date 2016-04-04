package common

import (
	"bufio"
	"bytes"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/bin-templater/template"
	"github.com/ghodss/yaml"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
)

type PodManifest struct {
	Name ACFullname     `json:"name"`
	Pod  *PodDefinition `json:"pod"`
}

type PodDefinition struct {
	Apps        []RuntimeApp        `json:"apps"`
	Volumes     []types.Volume      `json:"volumes"`
	Isolators   []types.Isolator    `json:"isolators"`
	Annotations types.Annotations   `json:"annotations"`
	Ports       []types.ExposedPort `json:"ports"`
}

type RuntimeApp struct {
	Dependencies []ACFullname      `json:"dependencies"`
	Name         string            `json:"name"`
	App          DgrApp            `json:"app"`
	Mounts       []schema.Mount    `json:"mounts"`
	Annotations  types.Annotations `json:"annotations"`
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BuildDefinition struct {
	Image        ACFullname   `json:"image"`
	Dependencies []ACFullname `json:"dependencies"`
}

type AciManifest struct {
	NameAndVersion ACFullname      `json:"name"`
	From           interface{}     `json:"from"`
	Builder        BuildDefinition `json:"builder"`
	Aci            AciDefinition   `json:"aci"`
	Tester         TestManifest    `json:"tester"`
}

type TestManifest struct {
	Builder BuildDefinition `json:"builder"`
	Aci     AciDefinition   `json:"aci"`
}

func (m *AciManifest) GetFroms() ([]ACFullname, error) {
	var froms []ACFullname
	switch v := m.From.(type) {
	case string:
		froms = []ACFullname{*NewACFullName(m.From.(string))}
	case []interface{}:
		for _, from := range m.From.([]interface{}) {
			froms = append(froms, *NewACFullName(from.(string)))
		}
	case nil:
		return froms, nil
	default:
		return nil, errs.WithF(data.WithField("type", v), "Invalid from type format")
	}
	return froms, nil
}

type AciDefinition struct {
	App           DgrApp            `json:"app,omitempty"`
	Annotations   types.Annotations `json:"annotations,omitempty"`
	Dependencies  []ACFullname      `json:"dependencies,omitempty"`
	PathWhitelist []string          `json:"pathWhitelist,omitempty"`
}

type DgrApp struct {
	Exec             types.Exec         `json:"exec"`
	User             string             `json:"user"`
	Group            string             `json:"group"`
	WorkingDirectory string             `json:"workingDirectory,omitempty"`
	Environment      types.Environment  `json:"environment,omitempty"`
	MountPoints      []types.MountPoint `json:"mountPoints,omitempty"`
	Ports            []types.Port       `json:"ports,omitempty"`
	Isolators        types.Isolators    `json:"isolators,omitempty"`
}

func ReadAciManifest(manifestPath string) (*AciManifest, error) {
	manifest := AciManifest{Aci: AciDefinition{}}
	fields := data.WithField("file", manifestPath)

	source, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	template, err := template.NewTemplating(nil, manifestPath, string(source))
	if err != nil {
		return nil, errs.WithEF(err, fields, "Failed to load templating of manifest")
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	if err := template.Execute(writer, nil); err != nil {
		return nil, errs.WithEF(err, fields, "Failed to template manifest")
	}
	if err := writer.Flush(); err != nil {
		return nil, errs.WithEF(err, fields, "Failed to flush buffer")
	}

	templated := b.Bytes()
	if logs.IsDebugEnabled() {
		logs.WithField("content", string(templated)).Debug("Templated manifest")
	}

	err = yaml.Unmarshal(templated, &manifest)
	if err != nil {
		return nil, errs.WithEF(err, fields, "Cannot unmarshall manifest")
	}

	return &manifest, nil
}
