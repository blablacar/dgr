package builder

import (
	"github.com/blablacar/cnt/config"
	"github.com/blablacar/cnt/utils"
	"strings"
)

func (aci *Aci) Push() {
	if config.GetConfig().Push.Type == "" {
		panic("Can't push, push is not configured in cnt global configuration file")
	}

	aci.CheckBuilt()
	if aci.args.Test {
		aci.args.Test = false
		aci.Test()
	}

	aci.tarAci(true)

	im := utils.ExtractManifestFromAci(aci.target + PATH_IMAGE_ACI_ZIP)
	val, _ := im.Labels.Get("version")
	if err := utils.ExecCmd("curl", "-f", "-i",
		"-F", "r=releases",
		"-F", "hasPom=false",
		"-F", "e=aci",
		"-F", "g=com.blablacar.aci.linux.amd64",
		"-F", "p=aci",
		"-F", "v="+val,
		"-F", "a="+strings.Split(string(im.Name), "/")[1],
		"-F", "file=@"+aci.target+PATH_IMAGE_ACI_ZIP,
		"-u", config.GetConfig().Push.Username+":"+config.GetConfig().Push.Password,
		config.GetConfig().Push.Url+"/service/local/artifact/maven/content"); err != nil {
		panic("Cannot push aci" + err.Error())
	}
}
