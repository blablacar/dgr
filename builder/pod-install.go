package builder

import "github.com/n0rad/go-erlog/logs"

func (p *Pod) Install() {
	logs.WithF(p.fields).Info("Installing")

	p.Build()

	checkVersion := make(chan bool, 1)

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e), &checkVersion)
		if err != nil {
			logs.WithEF(err, p.fields.WithField("name", e.Name)).Fatal("Cannot prepare aci")
		}
		aci.podName = &p.manifest.Name
		aci.Install()
	}

	for range p.manifest.Pod.Apps {
		<-checkVersion
	}

}
