package utils

import (
	"github.com/appc/spec/schema"
	"github.com/blablacar/cnt/log"
	"io/ioutil"
)

const POD_MANIFEST = `{
  "acVersion": "0.6.1",
  "acKind": "PodManifest"
}`

func BasicPodManifest() *schema.PodManifest {
	im := new(schema.PodManifest)
	im.UnmarshalJSON([]byte(POD_MANIFEST))
	return im
}

func WritePodManifest(im *schema.PodManifest, targetFile string) {
	buff, err := im.MarshalJSON()
	if err != nil {
		log.Get().Panic(err)
	}
	err = ioutil.WriteFile(targetFile, []byte(buff), 0644)
	if err != nil {
		log.Get().Panic(err)
	}
}
