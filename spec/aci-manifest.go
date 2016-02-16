package spec

import (
	"github.com/appc/spec/schema/types"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
)

type DgrBuild struct {
	Image types.ACIdentifier `json:"image"`
}

func (b *DgrBuild) NoBuildImage() bool {
	return b.Image == ""
}

type AciManifest struct {
	NameAndVersion ACFullname    `json:"name"`
	From           interface{}   `json:"from"`
	Build          DgrBuild      `json:"build"`
	Aci            AciDefinition `json:"aci"`
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
