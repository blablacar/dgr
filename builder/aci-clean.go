package builder

import (
	"github.com/blablacar/cnt/log"
	"os"
)

func (cnt *Img) Clean() {
	log.Get().Info("Cleaning " + cnt.manifest.NameAndVersion)
	checkVersion := make(chan bool, 1)
	go cnt.checkLatestVersions(&checkVersion)
	if err := os.RemoveAll(cnt.target + "/"); err != nil {
		log.Get().Panic("Cannot clean "+cnt.manifest.NameAndVersion, err)
	}
	<-checkVersion
}
