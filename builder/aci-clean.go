package builder
import (
	"github.com/blablacar/cnt/log"
	"os"
)

func (cnt *Img) Clean() {
	log.Get().Info("Cleaning " + cnt.manifest.Aci.Name)
	if err := os.RemoveAll(cnt.target + "/"); err != nil {
		log.Get().Panic("Cannot clean " + cnt.manifest.Aci.Name, err)
	}
}
