package builder

import (
	"github.com/n0rad/go-erlog/logs"
	"os"
)

func (aci *Aci) Update() error {
	logs.WithF(aci.fields).Info("Updating")

	os.MkdirAll(aci.rootfs, 0777)

	aci.copyInternals()
	aci.writeAciManifest()

	aci.copyRunlevelsScripts()
	aci.copyAttributes()
	aci.copyConfd()
	aci.copyFiles()

	aci.tarAci(false)

	return nil
}
