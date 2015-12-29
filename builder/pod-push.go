package builder

import (
	"github.com/blablacar/cnt/cnt"
	"github.com/blablacar/cnt/utils"
)

func (p *Pod) Push() {
	p.log.Info("Pushing")

	p.Build()

	checkVersion := make(chan bool, 1)

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e), &checkVersion)
		if err != nil {
			p.log.WithError(err).WithField("name", e.Name).Fatal("Cannot prepare aci")
		}
		aci.podName = &p.manifest.Name
		aci.Push()
	}

	for range p.manifest.Pod.Apps {
		<-checkVersion
	}

	if err := utils.ExecCmd("curl", "-i",
		"-F", "r=releases",
		"-F", "hasPom=false",
		"-F", "e=pod",
		"-F", "g=com.blablacar.aci.linux.amd64",
		"-F", "p=pod",
		"-F", "v="+p.manifest.Name.Version(),
		"-F", "a="+p.manifest.Name.ShortName(),
		"-F", "file=@"+p.target+"/pod-manifest.json",
		"-u", cnt.Home.Config.Push.Username+":"+cnt.Home.Config.Push.Password,
		cnt.Home.Config.Push.Url+"/service/local/artifact/maven/content"); err != nil {
		p.log.WithError(err).Fatal("Cannot push pod")
	}

}
