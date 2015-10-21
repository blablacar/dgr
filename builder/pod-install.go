package builder

import (
	"github.com/blablacar/cnt/log"
)

func (p *Pod) Install() {
	log.Info("Installing POD", p.manifest.Name)

	p.Build()

	checkVersion := make(chan bool, 1)

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e), &checkVersion)
		if err != nil {
			panic(err)
		}
		aci.PodName = &p.manifest.Name
		aci.Install()
	}

	for range p.manifest.Pod.Apps {
		<-checkVersion
	}

}
