package builder

import (
	"github.com/blablacar/cnt/log"
	"os"
)

func (cnt *Img) UpdateConf() error {
	log.Info("Updating Image : ", cnt.manifest.NameAndVersion)

	os.MkdirAll(cnt.rootfs, 0777)

	cnt.writeImgManifest()
	cnt.writeCntManifest() // TODO move that, here because we update the version number to generated version

	cnt.copyRunlevelsScripts()
	cnt.copyAttributes()
	cnt.copyConfd()
	cnt.copyFiles()

	cnt.tarAci(false)

	return nil
}
