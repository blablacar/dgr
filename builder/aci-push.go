package builder

import (
	"github.com/appc/spec/aci"
	"github.com/appc/spec/schema"
	"github.com/blablacar/cnt/config"
	"github.com/blablacar/cnt/utils"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (aci *Img) Push() {
	if config.GetConfig().Push.Type == "" {
		panic("Can't push, push is not configured in cnt global configuration file")
	}

	aci.CheckBuilt()
	if aci.args.Test {
		aci.args.Test = false
		aci.Test()
	}

	aci.tarAci(true)

	im := extractManifestFromAci(aci.target + PATH_IMAGE_ACI_ZIP)
	val, _ := im.Labels.Get("version")
	if err := utils.ExecCmd("curl", "-f", "-i",
		"-F", "r=releases",
		"-F", "hasPom=false",
		"-F", "e=aci",
		"-F", "g=com.blablacar.aci.linux.amd64",
		"-F", "p=aci",
		"-F", "v="+val,
		"-F", "a="+ShortNameId(im.Name),
		"-F", "file=@"+aci.target+PATH_IMAGE_ACI_ZIP,
		"-u", config.GetConfig().Push.Username+":"+config.GetConfig().Push.Password,
		config.GetConfig().Push.Url+"/service/local/artifact/maven/content"); err != nil {
		panic("Cannot push aci" + err.Error())
	}
}

func extractManifestFromAci(aciPath string) schema.ImageManifest {
	input, err := os.Open(aciPath)
	if err != nil {
		panic("cat-manifest: Cannot open %s: %v" + aciPath + err.Error())
	}
	defer input.Close()

	tr, err := aci.NewCompressedTarReader(input)
	if err != nil {
		panic("cat-manifest: Cannot open tar %s: %v" + aciPath + err.Error())
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
					panic(err)
				}

				err = im.UnmarshalJSON(bytes)
				if err != nil {
					panic(err)
				}
				return im
			}
		default:
			panic("error reading tarball: %v" + err.Error())
		}
	}
	panic("Cannot found manifest if aci")
	return im
}
