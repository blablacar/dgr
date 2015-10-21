package builder

import (
	"github.com/blablacar/cnt/utils"
	"io/ioutil"
	"os"
)

func (aci *Img) Install() string {
	aci.CheckBuilt()
	if aci.args.Test {
		aci.args.Test = false
		aci.Test()
	}
	os.Remove(aci.target + PATH_INSTALLED)
	hash, err := utils.ExecCmdGetOutput("rkt", "--insecure-skip-verify=true", "fetch", aci.target+PATH_IMAGE_ACI)
	if err != nil {
		panic("Cannot install" + err.Error())
	}
	ioutil.WriteFile(aci.target+PATH_INSTALLED, []byte(hash), 0644)
	return hash
}
