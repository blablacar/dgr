package main

import (
	"github.com/blablacar/dgr/bin-dgr/common"
	"io/ioutil"
	"os"
)

func (aci *Aci) Install() {
	aci.CheckBuilt()
	if aci.args.Test {
		aci.args.Test = false
		aci.Test()
	}
	os.Remove(aci.target + PATH_INSTALLED)
	hash, err := common.ExecCmdGetOutput("rkt", "--insecure-options=image", "fetch", aci.target+PATH_IMAGE_ACI)
	if err != nil {
		panic("Cannot install" + err.Error())
	}
	ioutil.WriteFile(aci.target+PATH_INSTALLED, []byte(hash), 0644)
}
