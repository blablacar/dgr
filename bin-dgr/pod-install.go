package main

import "github.com/n0rad/go-erlog/logs"

func (p *Pod) Install() error {
	logs.WithF(p.fields).Info("Installing")

	if err := p.Build(); err != nil {
		return err
	}

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e))
		if err != nil {
			logs.WithEF(err, p.fields.WithField("name", e.Name)).Fatal("Cannot prepare aci")
		}
		aci.podName = &p.manifest.Name
		if err := aci.Install(); err != nil {
			return err
		}
	}
	return nil
}
