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

func (aci *Img) checkInstalled() string {
	if _, err := os.Stat(aci.target + PATH_INSTALLED); os.IsNotExist(err) {
		return aci.Install()
	}
	content, err := ioutil.ReadFile(aci.target + PATH_INSTALLED)
	if err != nil {
		log.Get().Panic("Cannot get content of "+PATH_INSTALLED, err)
	}
	return string(content)
}
