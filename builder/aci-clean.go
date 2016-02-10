package builder

import (
	"github.com/n0rad/go-erlog/logs"
	"os"
)

func (aci *Aci) Clean() {
	logs.WithF(aci.fields).Debug("Cleaning")

	aci.checkCompatibilityVersions()
	aci.checkLatestVersions()

	if err := os.RemoveAll(aci.target + "/"); err != nil {
		panic("Cannot clean " + aci.manifest.NameAndVersion.String() + err.Error())
	}
}
