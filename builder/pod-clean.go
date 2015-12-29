package builder

import (
	"os"
)

func (p *Pod) Clean() {
	p.log.Info("Cleaning")

	if err := os.RemoveAll(p.target + "/"); err != nil {
		panic("Cannot clean" + p.manifest.Name.String() + err.Error())
	}

	checkVersion := make(chan bool, 1)

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e), &checkVersion)
		if err != nil {
			p.log.WithError(err).WithField("name", e.Name).Fatal("Cannot prepare aci")
		}
		aci.podName = &p.manifest.Name
		aci.Clean()
	}

	for range p.manifest.Pod.Apps {
		<-checkVersion
	}

}
