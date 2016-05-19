package main

import "github.com/n0rad/go-erlog/logs"

func (p *Pod) Test() error {
	logs.WithF(p.fields).Info("Testing")

	for _, e := range p.manifest.Pod.Apps {
		tmpl, err := p.toAciManifestTemplate(e)
		if err != nil {
			return err
		}

		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, tmpl, p.checkWg)
		if err != nil {
			logs.WithEF(err, p.fields).WithField("name", e.Name).Fatal("Cannot prepare aci")
		}
		aci.podName = &p.manifest.Name
		if err := aci.Test(); err != nil {
			return err
		}
	}
	return nil
}
