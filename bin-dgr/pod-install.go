package main

import "github.com/n0rad/go-erlog/logs"

func (p *Pod) Install() ([]string, error) {
	logs.WithF(p.fields).Info("Installing")

	hashs := []string{}

	if err := p.CleanAndBuild(); err != nil {
		return hashs, err
	}

	for _, e := range p.manifest.Pod.Apps {
		tmpl, err := p.toAciManifestTemplate(e)
		if err != nil {
			return nil, err
		}

		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, tmpl)
		if err != nil {
			logs.WithEF(err, p.fields.WithField("name", e.Name)).Fatal("Cannot prepare aci")
		}
		aci.podName = &p.manifest.Name
		hash, err := aci.Install()
		if err != nil {
			return hashs, err
		}
		hashs = append(hashs, hash...)
	}
	return hashs, nil
}
