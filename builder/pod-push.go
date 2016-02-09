package builder

import (
	"github.com/blablacar/cnt/cnt"
	"github.com/blablacar/cnt/utils"
	"github.com/n0rad/go-erlog/logs"
)

func (p *Pod) Push() {
	logs.WithF(p.fields).Info("Pushing")

	p.Build()

	checkVersion := make(chan bool, 1)
	checkCompat := make(chan bool, 1)

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e), &checkVersion, &checkCompat)
		if err != nil {
			logs.WithEF(err, p.fields.WithField("name", e.Name)).Fatal("Cannot prepare aci")
		}
		aci.podName = &p.manifest.Name
		aci.Push()
	}

	for range p.manifest.Pod.Apps {
		<-checkVersion
	}
	for range p.manifest.Pod.Apps {
		<-checkCompat
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
		logs.WithEF(err, p.fields).Fatal("Cannot push pod")
	}

}
