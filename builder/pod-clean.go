package builder
import (
	"github.com/blablacar/cnt/log"
	"os"
)

func (p *Pod) Clean() {
	log.Get().Info("Cleaning POD", p.manifest.NameAndVersion)

	if err := os.RemoveAll(p.target + "/"); err != nil {
		log.Get().Panic("Cannot clean", p.manifest.NameAndVersion, err)
	}

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path + "/" + e.Name, p.args, p.toAciManifest(e))
		if (err != nil) {
			log.Get().Panic(err)
		}
		aci.Clean()
	}
}
