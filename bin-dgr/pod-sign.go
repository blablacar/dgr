package main

import (
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

func (p *Pod) Sign() error {
	logs.WithF(p.fields).Info("Signing")

	for _, e := range p.manifest.Pod.Apps {
		appFields := p.fields.WithField("name", e.Name)
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e))
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
