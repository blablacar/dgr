package builder

import (
	log "github.com/Sirupsen/logrus"
	"os"
)

func (cnt *Aci) UpdateConf() error {
	log.Info("Updating Image : ", cnt.manifest.NameAndVersion)

	os.MkdirAll(cnt.rootfs, 0777)

	cnt.writeImgManifest()

	cnt.copyRunlevelsScripts()
	cnt.copyAttributes()
	cnt.copyConfd()
	cnt.copyFiles()

	cnt.tarAci(false)

	return nil
}
