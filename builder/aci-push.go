package builder
import (
	"github.com/blablacar/cnt/config"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/utils"
	"os"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/aci"
	"io"
	"path/filepath"
	"io/ioutil"
)

func (cnt *Cnt) Push() {
	cnt.checkBuilt()
	if config.GetConfig().Push.Type == "" {
		log.Get().Panic("Can't push, push is not configured in cnt global configuration file")
	}

	im := extractManifestFromAci(cnt.target + "/image.aci")
	val, _ := im.Labels.Get("version")
	utils.ExecCmd("curl", "-i",
		"-F", "r=releases",
		"-F", "hasPom=false",
		"-F", "e=aci",
		"-F", "g=com.blablacar.aci.linux.amd64",
		"-F", "p=aci",
		"-F", "v=" + val,
		"-F", "a=" + ShortNameId(im.Name),
		"-F", "file=@" + cnt.target + "/image.aci",
		"-u", config.GetConfig().Push.Username + ":" + config.GetConfig().Push.Password,
		config.GetConfig().Push.Url + "/service/local/artifact/maven/content")
}

func extractManifestFromAci(aciPath string) schema.ImageManifest {
	input, err := os.Open(aciPath)
	if err != nil {
		log.Get().Panic("cat-manifest: Cannot open %s: %v", aciPath, err)
	}
	defer input.Close()

	tr, err := aci.NewCompressedTarReader(input)
	if err != nil {
		log.Get().Panic("cat-manifest: Cannot open tar %s: %v", aciPath, err)
	}


	im := schema.ImageManifest{}

	Tar:
	for {
		hdr, err := tr.Next()
		switch err {
		case io.EOF:
			break Tar
		case nil:
			if filepath.Clean(hdr.Name) == aci.ManifestFile {
				bytes, err := ioutil.ReadAll(tr)
				if err != nil {
					log.Get().Panic(err)
				}

				err = im.UnmarshalJSON(bytes)
				if err != nil {
					log.Get().Panic(err)
				}
				return im
			}
		default:
			log.Get().Panic("error reading tarball: %v", err)
		}
	}
	log.Get().Panic("Cannot found manifest if aci");
	return im
}