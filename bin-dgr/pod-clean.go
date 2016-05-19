package main

import (
	"github.com/n0rad/go-erlog/logs"
	"os"
)

func (p *Pod) Clean() {
	logs.WithF(p.fields).Info("Cleaning")

	if err := os.RemoveAll(p.target + "/"); err != nil {
		logs.WithEF(err, p.fields.WithField("dir", p.target)).Warn("Cannot clean directory")
	}

	for _, e := range p.manifest.Pod.Apps {
		tmpl, err := p.toAciManifestTemplate(e)
		if err != nil {
			logs.WithEF(err, p.fields).WithField("name", e.Name).Error("Failed to prepare manifest")
		}
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, tmpl, p.checkWg)
		if err != nil {
			logs.WithEF(err, p.fields).WithField("name", e.Name).Error("Cannot prepare aci to clean")
		}
		aci.podName = &p.manifest.Name
		aci.Clean()
	}

}
