package builder

import "github.com/n0rad/go-erlog/logs"

func (p *Pod) Test() {
	logs.WithF(p.fields).Info("Testing")

	checkVersion := make(chan bool, 1)
	checkCompat := make(chan bool, 1)

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e), &checkVersion, &checkCompat)
		if err != nil {
			logs.WithEF(err, p.fields).WithField("name", e.Name).Fatal("Cannot prepare aci")
		}
		aci.podName = &p.manifest.Name
		aci.Test()
	}

	for range p.manifest.Pod.Apps {
		<-checkVersion
	}
	for range p.manifest.Pod.Apps {
		<-checkCompat
	}

}
