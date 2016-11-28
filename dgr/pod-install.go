package main

import "github.com/n0rad/go-erlog/logs"

func (p *Pod) Install() ([]string, error) {
	logs.WithF(p.fields).Info("Installing")

	hashs := []string{}

	if err := p.CleanAndBuild(); err != nil {
		return hashs, err
	}

	for _, e := range p.manifest.Pod.Apps {
		aci, err := p.toPodAci(e)
		if err != nil {
			return nil, err
		}

		hash, err := aci.Install()
		if err != nil {
			return hashs, err
		}
		hashs = append(hashs, hash...)
	}
	return hashs, nil
}
