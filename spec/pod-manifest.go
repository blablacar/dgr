package spec
import (
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
)

type PodManifest struct {
	NameAndVersion ACFullname           `json:"name"`
	Pod            *PodDefinition          `json:"pod"`
}

type PodDefinition struct {
	Apps        []RuntimeApp        `json:"apps"`
	Volumes     []types.Volume      `json:"volumes"`
	Isolators   []types.Isolator    `json:"isolators"`
	Annotations types.Annotations   `json:"annotations"`
	Ports       []types.ExposedPort `json:"ports"`
}

type RuntimeApp struct {
	Image       ACFullname               `json:"image"`
	Name        string					 `json:"name"`
	App         *types.App               `json:"app"`
	Mounts      []schema.Mount           `json:"mounts"`
	Annotations types.Annotations        `json:"annotations"`
}
