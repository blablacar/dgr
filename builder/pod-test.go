package builder

import (
	log "github.com/Sirupsen/logrus"
)

func (p *Pod) Test() {
	log.Info("Testing POD", p.manifest.Name)

	checkVersion := make(chan bool, 1)

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e), &checkVersion)
		if err != nil {
			panic(err)
		}
		aci.PodName = &p.manifest.Name
		aci.Test()
	}

	for range p.manifest.Pod.Apps {
		<-checkVersion
	}

}
