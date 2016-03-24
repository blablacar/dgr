package main

import (
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
)

type DgrCommand interface {
	CleanAndBuild() error
	CleanAndTry() error
	Clean()
	Push() error
	Install() ([]string, error)
	Test() error
	Graph() error
	Init() error
}

type PodManifest struct {
	Name common.ACFullname `json:"name"`
	Pod  *PodDefinition    `json:"pod"`
}

type PodDefinition struct {
	Apps        []RuntimeApp        `json:"apps"`
	Volumes     []types.Volume      `json:"volumes"`
	Isolators   []types.Isolator    `json:"isolators"`
	Annotations types.Annotations   `json:"annotations"`
	Ports       []types.ExposedPort `json:"ports"`
}

type RuntimeApp struct {
	Dependencies []common.ACFullname `json:"dependencies"`
	Name         string              `json:"name"`
	App          DgrApp              `json:"app"`
	Mounts       []schema.Mount      `json:"mounts"`
	Annotations  types.Annotations   `json:"annotations"`
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BuildDefinition struct {
	Image        common.ACFullname   `json:"image"`
	Dependencies []common.ACFullname `json:"dependencies"`
}

type AciManifest struct {
	NameAndVersion common.ACFullname `json:"name"`
	From           interface{}       `json:"from"`
	Builder        BuildDefinition   `json:"builder"`
	Aci            AciDefinition     `json:"aci"`
	Tester         TestManifest      `json:"tester"`
}

type TestManifest struct {
	Builder BuildDefinition `json:"builder"`
	Aci     AciDefinition   `json:"aci"`
}

func (m *AciManifest) GetFroms() ([]common.ACFullname, error) {
	var froms []common.ACFullname
	switch v := m.From.(type) {
	case string:
		froms = []common.ACFullname{*common.NewACFullName(m.From.(string))}
	case []interface{}:
		for _, from := range m.From.([]interface{}) {
			froms = append(froms, *common.NewACFullName(from.(string)))
		}
	case nil:
		return froms, nil
	default:
		return nil, errs.WithF(data.WithField("type", v), "Invalid from type format")
	}
	return froms, nil
}

type AciDefinition struct {
	App           DgrApp              `json:"app,omitempty"`
	Annotations   types.Annotations   `json:"annotations,omitempty"`
	Dependencies  []common.ACFullname `json:"dependencies,omitempty"`
	PathWhitelist []string            `json:"pathWhitelist,omitempty"`
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
