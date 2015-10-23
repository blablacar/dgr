package builder

import (
	log "github.com/Sirupsen/logrus"
	"os"
)

func (cnt *Img) Clean() {
	log.Info("Cleaning " + cnt.manifest.NameAndVersion)
	checkVersion := make(chan bool, 1)
	go cnt.checkLatestVersions(&checkVersion)
	if err := os.RemoveAll(cnt.target + "/"); err != nil {
		panic("Cannot clean " + cnt.manifest.NameAndVersion.String() + err.Error())
	}
	<-checkVersion
}
