package builder

import (
	"github.com/blablacar/cnt/log"
)

func (p *Pod) Install() {
	log.Get().Info("Installing POD", p.manifest.Name)

	p.Build()

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e))
		if err != nil {
			log.Get().Panic(err)
		}
		aci.Install()
	}

}
