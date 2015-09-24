package builder

import (
	"github.com/blablacar/cnt/config"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/utils"
)

func (p *Pod) Push() {
	log.Get().Info("Push POD", p.manifest.Name)

	p.Build()

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path+"/"+e.Name, p.args, p.toAciManifest(e))
		if err != nil {
			log.Get().Panic(err)
		}
		aci.PodName = &p.manifest.Name
		aci.Push()
	}

	if err := utils.ExecCmd("curl", "-i",
		"-F", "r=releases",
		"-F", "hasPom=false",
		"-F", "e=service",
		"-F", "g=com.blablacar.aci.linux.amd64",
		"-F", "p=service",
		"-F", "v="+p.manifest.Name.Version(),
		"-F", "a="+p.manifest.Name.ShortName(),
		"-F", "file=@"+p.target+"/"+p.manifest.Name.ShortName()+"@.service",
		"-u", config.GetConfig().Push.Username+":"+config.GetConfig().Push.Password,
		config.GetConfig().Push.Url+"/service/local/artifact/maven/content"); err != nil {
		log.Get().Panic("Cannot push pod", err)
	}

}
