package builder

import (
	"github.com/n0rad/go-erlog/logs"
	"os"
)

func (cnt *Aci) UpdateConf() error {
	logs.WithF(cnt.fields).Info("Updating")

	os.MkdirAll(cnt.rootfs, 0777)

	cnt.writeImgManifest()

	cnt.copyRunlevelsScripts()
	cnt.copyAttributes()
	cnt.copyConfd()
	cnt.copyFiles()

	cnt.tarAci(false)

	return nil
}
