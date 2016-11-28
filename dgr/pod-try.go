package main

import "github.com/n0rad/go-erlog/logs"

func (p *Pod) CleanAndTry() error {
	logs.WithF(p.fields).Info("Try")

	for _, e := range p.manifest.Pod.Apps {
		aci, err := p.toPodAci(e)
		if err != nil {
			return err
		}

		if err := aci.CleanAndTry(); err != nil {
			return err
		}
	}
	return nil
}
