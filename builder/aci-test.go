package builder
import (
	"github.com/blablacar/cnt/log"
	"os"
	"github.com/blablacar/cnt/bats"
	"github.com/blablacar/cnt/utils"
)

func (cnt *Img) Test() {
	log.Get().Info("Testing " + cnt.manifest.NameAndVersion)
	if _, err := os.Stat(cnt.target + "/image.aci"); os.IsNotExist(err) {
		if err := cnt.Build(); err != nil {
			log.Get().Panic("Cannot Install since build failed")
		}
	}

	// BATS
	os.MkdirAll(cnt.target + "/test", 0777)
	bats.WriteBats(cnt.target + "/test")
	utils.ExecCmd("rkt", "--insecure-skip-verify=true", "run", cnt.target + "/image.aci") // TODO missing command override that will arrive in next RKT version
}
