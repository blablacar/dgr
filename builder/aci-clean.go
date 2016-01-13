package builder

import (
	"github.com/n0rad/go-erlog/logs"
	"os"
)

func (aci *Aci) Clean() {
	logs.WithF(aci.fields).Debug("Cleaning")
	checkVersion := make(chan bool, 1)
	go aci.checkLatestVersions(&checkVersion)
	if err := os.RemoveAll(aci.target + "/"); err != nil {
		panic("Cannot clean " + aci.manifest.NameAndVersion.String() + err.Error())
	}
	<-checkVersion
}
