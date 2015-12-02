package utils

import (
	"github.com/appc/spec/aci"
	"github.com/appc/spec/schema"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func ExtractManifestFromAci(aciPath string) schema.ImageManifest {
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
