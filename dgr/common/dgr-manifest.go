package common

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/bin-templater/template"
	"github.com/ghodss/yaml"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

type PodManifest struct {
	Name ACFullname     `json:"name,omitempty" yaml:"name,omitempty"`
	Pod  *PodDefinition `json:"pod,omitempty" yaml:"pod,omitempty"`
}

type PodDefinition struct {
	Apps        []RuntimeApp        `json:"apps,omitempty" yaml:"apps,omitempty"`
	Volumes     []types.Volume      `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Isolators   []types.Isolator    `json:"isolators,omitempty" yaml:"isolators,omitempty"`
	Annotations types.Annotations   `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Ports       []types.ExposedPort `json:"ports,omitempty" yaml:"ports,omitempty"`
}

type InheritDependencyPolicy string

const (
	FIRST InheritDependencyPolicy = "first" // default
	LAST                          = "last"  // useful when you want to overlay with a dep as last but inherit from it
	NONE                          = "none"
)

func (i InheritDependencyPolicy) GetInheritDependency(r RuntimeApp) *ACFullname {
	if len(r.Dependencies) > 0 {
		if i == LAST {
			return &r.Dependencies[len(r.Dependencies)-1]
		} else if i == FIRST {
			return &r.Dependencies[0]
		}
	}
	return nil
}

type RuntimeApp struct {
	InheritDependencyPolicy InheritDependencyPolicy `json:"inheritDependencyPolicy,omitempty" yaml:"inheritDependencyPolicy,omitempty"`
	Dependencies            []ACFullname            `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Name                    string                  `json:"name,omitempty" yaml:"name,omitempty"`
	App                     DgrApp                  `json:"app,omitempty" yaml:"app,omitempty"`
	Mounts                  []schema.Mount          `json:"mounts,omitempty" yaml:"mounts,omitempty"`
	Annotations             types.Annotations       `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type MountInfo struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type BuilderDefinition struct {
	Image        ACFullname        `json:"image,omitempty" yaml:"image,omitempty"`
	Environment  types.Environment `json:"environment,omitempty" yaml:"environment,omitempty"`
	Dependencies []ACFullname      `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	MountPoints  []MountInfo       `json:"mountPoints,omitempty" yaml:"mountPoints,omitempty"`
}

type BuildDefinition struct {
	MountPoints []MountInfo `json:"mountPoints,omitempty" yaml:"mountPoints,omitempty"`
}

type AciManifest struct {
	NameAndVersion ACFullname        `json:"name,omitempty" yaml:"name,omitempty"`
	Builder        BuilderDefinition `json:"builder,omitempty" yaml:"builder,omitempty"`
	Build          BuildDefinition   `json:"build,omitempty" yaml:"build,omitempty"`
	Aci            AciDefinition     `json:"aci,omitempty" yaml:"aci,omitempty"`
	Tester         TestManifest      `json:"tester,omitempty" yaml:"tester,omitempty"`
}

type TestManifest struct {
	Builder BuilderDefinition `json:"builder,omitempty" yaml:"builder,omitempty"`
	Aci     AciDefinition     `json:"aci,omitempty" yaml:"aci,omitempty"`
}

type AciDefinition struct {
	App           DgrApp            `json:"app,omitempty" yaml:"app,omitempty"`
	Annotations   types.Annotations `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Dependencies  []ACFullname      `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	PathWhitelist []string          `json:"pathWhitelist,omitempty" yaml:"pathWhitelist,omitempty"`
}

type DgrApp struct {
	Exec              types.Exec         `json:"exec,omitempty" yaml:"exec,omitempty"`
	User              string             `json:"user,omitempty" yaml:"user,omitempty"`
	Group             string             `json:"group,omitempty" yaml:"group,omitempty"`
	SupplementaryGIDs []int              `json:"supplementaryGIDs,omitempty" yaml:"supplementaryGIDs,omitempty"`
	WorkingDirectory  string             `json:"workingDirectory,omitempty" yaml:"workingDirectory,omitempty"`
	Environment       types.Environment  `json:"environment,omitempty" yaml:"environment,omitempty"`
	MountPoints       []types.MountPoint `json:"mountPoints,omitempty" yaml:"mountPoints,omitempty"`
	Ports             []types.Port       `json:"ports,omitempty" yaml:"ports,omitempty"`
	Isolators         []Isolator         `json:"isolators,omitempty" yaml:"isolators,omitempty"`
}

type LinuxCapabilitiesSetValue struct {
	Set []types.LinuxCapability `json:"set"`
}

type Isolator struct {
	Name  string
	Value interface{}
}

func ProcessManifestTemplate(manifestContent string, data2 interface{}, checkNoValue bool) (*AciManifest, error) {
	manifest := AciManifest{Aci: AciDefinition{}}
	fields := data.WithField("source", manifestContent)

	template, err := template.NewTemplating(nil, "", manifestContent)
	if err != nil {
		return nil, errs.WithEF(err, fields, "Failed to load templating of manifest")
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	if err := template.Execute(writer, data2); err != nil {
		return nil, errs.WithEF(err, fields, "Failed to template manifest")
	}
	if err := writer.Flush(); err != nil {
		return nil, errs.WithEF(err, fields, "Failed to flush buffer")
	}

	templated := b.Bytes()
	if logs.IsDebugEnabled() {
		logs.WithField("content", string(templated)).Debug("Templated manifest")
	}

	if checkNoValue {
		scanner := bufio.NewScanner(bytes.NewReader(templated))
		scanner.Split(bufio.ScanLines)
		for i := 1; scanner.Scan(); i++ {
			text := scanner.Text()
			if bytes.Contains([]byte(text), []byte("<no value>")) {
				return nil, errs.WithF(fields.WithField("line", i).WithField("text", text), "Templating result of manifest have <no value>")
			}
		}
	}

	err = yaml.Unmarshal(templated, &manifest)
	if err != nil {
		return nil, errs.WithEF(err, fields, "Cannot unmarshall manifest")
	}

	if strings.Contains(manifest.NameAndVersion.Version(), "<no value>") {
		manifest.NameAndVersion = ACFullname(manifest.NameAndVersion.Name())
	}

	return &manifest, nil
}
