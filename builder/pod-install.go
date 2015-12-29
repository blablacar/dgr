package builder

func (p *Pod) Install() {
	p.log.Info("Installing")

	p.Build()

	checkVersion := make(chan bool, 1)

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e), &checkVersion)
		if err != nil {
			p.log.WithError(err).WithField("name", e.Name).Fatal("Cannot prepare aci")
		}
		aci.podName = &p.manifest.Name
		aci.Install()
	}

	for range p.manifest.Pod.Apps {
		<-checkVersion
	}

}
