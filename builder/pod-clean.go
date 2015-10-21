package builder

import (
	"github.com/blablacar/cnt/log"
	"os"
)

func (p *Pod) Clean() {
	log.Info("Cleaning POD", p.manifest.Name)

	if err := os.RemoveAll(p.target + "/"); err != nil {
		panic("Cannot clean" + p.manifest.Name.String() + err.Error())
	}

	checkVersion := make(chan bool, 1)

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e), &checkVersion)
		if err != nil {
			panic(err)
		}
		aci.PodName = &p.manifest.Name
		aci.Clean()
	}

	for range p.manifest.Pod.Apps {
		<-checkVersion
	}

}
