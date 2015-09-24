package builder

import (
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/utils"
	"io/ioutil"
	"os"
)

func (aci *Img) Install() string {
	aci.CheckBuilt()
	//TODO remove old installed for rkt ??
	os.Remove(aci.target + PATH_INSTALLED)
	hash, err := utils.ExecCmdGetOutput("rkt", "--insecure-skip-verify=true", "fetch", aci.target+PATH_IMAGE_ACI)
	if err != nil {
		log.Get().Panic("Cannot install", err)
	}
	ioutil.WriteFile(aci.target+PATH_INSTALLED, []byte(hash), 0644)
	return hash
}
