package builder

import (
	"os"
)

func (cnt *Aci) Clean() {
	cnt.log.Debug("Cleaning")
	checkVersion := make(chan bool, 1)
	go cnt.checkLatestVersions(&checkVersion)
	if err := os.RemoveAll(cnt.target + "/"); err != nil {
		panic("Cannot clean " + cnt.manifest.NameAndVersion.String() + err.Error())
	}
	<-checkVersion
}
