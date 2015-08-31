package spec
import (
	"github.com/appc/spec/schema/types"
)

type CntBuild struct {
	Image types.ACIdentifier                `json:"image"`
}

func (b *CntBuild) NoBuildImage() bool {
	return b.Image == ""
}

type AciManifest struct {
	NameAndVersion ACFullname                  `json:"name"`
	From           ACFullname                  `json:"from"`
	Build          CntBuild                    `json:"build"`
	Aci            AciDefinition               `json:"aci"`
}

type AciDefinition struct {
	App           *CntApp            `json:"app,omitempty"`
	Annotations   types.Annotations  `json:"annotations,omitempty"`
	Dependencies  []ACFullname       `json:"dependencies,omitempty"`
	PathWhitelist []string           `json:"pathWhitelist,omitempty"`
}

type CntApp struct {
	Exec             types.Exec           `json:"exec"`
	EventHandlers    []types.EventHandler `json:"eventHandlers,omitempty"`
	User             string         `json:"user"`
	Group            string         `json:"group"`
	WorkingDirectory string         `json:"workingDirectory,omitempty"`
	Environment      types.Environment    `json:"environment,omitempty"`
	MountPoints      []types.MountPoint   `json:"mountPoints,omitempty"`
	Ports            []types.Port         `json:"ports,omitempty"`
	Isolators        types.Isolators      `json:"isolators,omitempty"`
}
