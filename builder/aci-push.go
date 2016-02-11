package builder

import (
	"github.com/blablacar/cnt/cnt"
	"github.com/blablacar/cnt/utils"
	"github.com/n0rad/go-erlog/logs"
	"strings"
)

func (aci *Aci) Push() {
	if cnt.Home.Config.Push.Type == "" {
		panic("Can't push, push is not configured in cnt global configuration file")
	}

	aci.CheckBuilt()
	if aci.args.Test {
		aci.args.Test = false
		aci.Test()
	}

	logs.WithF(aci.fields).Info("Gzipping aci before upload")
	aci.tarAci(true)

	logs.WithF(aci.fields).Info("Uploading aci")
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
		"-u", cnt.Home.Config.Push.Username+":"+cnt.Home.Config.Push.Password,
		cnt.Home.Config.Push.Url+"/service/local/artifact/maven/content"); err != nil {
		logs.WithEF(err, aci.fields).Error("Failed to push aci")
	}
}
