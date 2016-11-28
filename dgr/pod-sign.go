package main

import (
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

func (p *Pod) Sign() error {
	logs.WithF(p.fields).Info("Signing")

	for _, e := range p.manifest.Pod.Apps {
		aci, err := p.toPodAci(e)
		if err != nil {
			return err
		}

		if err := aci.Sign(); err != nil {
			return errs.WithEF(err, aci.fields, "sign of pod's aci failed")
		}
	}
	return nil
}
