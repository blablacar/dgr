package builder
import (
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/utils"
	"github.com/blablacar/cnt/config"
)


func (p *Pod) Push() {
	log.Get().Info("Push POD", p.manifest.NameAndVersion)

	for _, e := range p.manifest.Pod.Apps {
		aci, err := NewAciWithManifest(p.path + "/" + e.Name, p.args, p.toAciManifest(e))
		if (err != nil) {
			log.Get().Panic(err)
		}
		aci.Push()
	}

	utils.ExecCmd("curl", "-i",
		"-F", "r=releases",
		"-F", "hasPom=false",
		"-F", "e=pod",
		"-F", "g=com.blablacar.aci.linux.amd64",
		"-F", "p=pod",
		"-F", "v=" + p.manifest.NameAndVersion.Version(),
		"-F", "a=" + p.manifest.NameAndVersion.ShortName(),
		"-F", "file=@" + p.target + POD_TARGET_MANIFEST,
		"-u", config.GetConfig().Push.Username + ":" + config.GetConfig().Push.Password,
		config.GetConfig().Push.Url + "/service/local/artifact/maven/content")

}
