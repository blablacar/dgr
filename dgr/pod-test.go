package main

import "github.com/n0rad/go-erlog/logs"

func (p *Pod) Test() error {
	logs.WithF(p.fields).Info("Testing")

	for _, e := range p.manifest.Pod.Apps {
		aci, err := p.toPodAci(e)
		if err != nil {
			return err
		}

		if err := aci.Test(); err != nil {
			return err
		}
	}
	return nil
}
