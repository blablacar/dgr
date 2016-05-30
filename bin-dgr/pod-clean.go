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
		aci, err := p.toPodAci(e)
		if err != nil {
			return
		}

		aci.Clean()
	}

}
