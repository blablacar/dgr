package main

import (
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/logs"
)

func (p *Pod) Push() {
	logs.WithF(p.fields).Info("Pushing")

	p.Build()

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e))
		if err != nil {
			logs.WithEF(err, p.fields.WithField("name", e.Name)).Fatal("Cannot prepare aci")
		}
		aci.podName = &p.manifest.Name
		aci.Push()
	}

	if err := common.ExecCmd("curl", "-i",
		"-F", "r=releases",
		"-F", "hasPom=false",
		"-F", "e=pod",
		"-F", "g=com.blablacar.aci.linux.amd64",
		"-F", "p=pod",
		"-F", "v="+p.manifest.Name.Version(),
		"-F", "a="+p.manifest.Name.ShortName(),
		"-F", "file=@"+p.target+"/pod-manifest.json",
		"-u", Home.Config.Push.Username+":"+Home.Config.Push.Password,
		Home.Config.Push.Url+"/service/local/artifact/maven/content"); err != nil {
		logs.WithEF(err, p.fields).Fatal("Cannot push pod")
	}

}
