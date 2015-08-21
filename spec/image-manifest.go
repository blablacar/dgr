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
	App           *types.App         `json:"app,omitempty"`
	Annotations   types.Annotations  `json:"annotations,omitempty"`
	Dependencies  types.Dependencies `json:"dependencies,omitempty"`
	PathWhitelist []string           `json:"pathWhitelist,omitempty"`
}