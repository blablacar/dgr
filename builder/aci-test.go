package builder

import (
	"github.com/blablacar/cnt/bats"
	"github.com/blablacar/cnt/log"
	"os"
)

func (cnt *Img) Test() {
	log.Get().Info("Testing " + cnt.manifest.NameAndVersion)
	if _, err := os.Stat(cnt.target + "/image.aci"); os.IsNotExist(err) {
		if err := cnt.Build(); err != nil {
			log.Get().Panic("Cannot Install since build failed")
		}
	}

	// prepare runner in target
	// run contauner with mout mpoint
	// run real service in background
	// run tests
	//

	// BATS
	os.MkdirAll(cnt.target+"/test", 0777)
	bats.WriteBats(cnt.target + "/test")

	//	if err := utils.ExecCmd("systemd-nspawn", "--directory=" + cnt.rootfs, "--capability=all",
	//		"--bind=" + cnt.target + "/:/target", "--share-system", "target/build.sh"); err != nil {
	//		log.Get().Panic("Build step did not succeed", err)
	//
	//
	//		utils.ExecCmd("rkt", "--insecure-skip-verify=true", "run", cnt.target + "/image.aci") // TODO missing command override that will arrive in next RKT version
	//	}
}
