package common

import (
	"github.com/appc/spec/aci"
	"github.com/appc/spec/schema"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func ExtractManifestContentFromAci(aciPath string) ([]byte, error) {
	fields := data.WithField("file", aciPath)
	input, err := os.Open(aciPath)
	if err != nil {
		return nil, errs.WithEF(err, fields, "Cannot open file")
	}
	defer input.Close()

	tr, err := aci.NewCompressedTarReader(input)
	if err != nil {
		return nil, errs.WithEF(err, fields, "Cannot open file as tar")
	}

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
					return nil, errs.WithEF(err, fields, "Cannot read manifest content in tar")
				}
				return bytes, nil
			}
		default:
			return nil, errs.WithEF(err, fields, "error reading tarball file")
		}
	}
	return nil, errs.WithEF(err, fields, "Cannot found manifest in file")
}

func ExtractManifestFromAci(aciPath string) (*schema.ImageManifest, error) {
	fields := data.WithField("file", aciPath)
	content, err := ExtractManifestContentFromAci(aciPath)
	if err != nil {
		return nil, errs.WithEF(err, fields, "Cannot extract aci manifest content from file")
	}
	im := &schema.ImageManifest{}

	err = im.UnmarshalJSON(content)
	if err != nil {
		return nil, errs.WithEF(err, fields.WithField("content", string(content)), "Cannot unmarshall json content")
	}
	return im, nil
}
