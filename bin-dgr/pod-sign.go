package main

import (
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

func (p *Pod) Sign() error {
	logs.WithF(p.fields).Info("Signing")

	for _, e := range p.manifest.Pod.Apps {
		appFields := p.fields.WithField("name", e.Name)
		tmpl, err := p.toAciManifestTemplate(e)
		if err != nil {
			return err
		}

		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, tmpl, p.checkWg)
		if err != nil {
			logs.WithEF(err, appFields).Error("Cannot prepare aci to sign")
		}
		aci.podName = &p.manifest.Name
		if err := aci.Sign(); err != nil {
			return errs.WithEF(err, appFields, "sign of pod's aci failed")
		}
	}
	return nil
}
