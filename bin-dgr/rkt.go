package main

import (
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/blablacar/dgr/dist"
	"github.com/n0rad/go-erlog/logs"
)

var aciBuilder = common.NewACFullName("blablacar.github.io/dgr/aci-builder")
var aciTester = common.NewACFullName("blablacar.github.io/dgr/aci-tester")
var importMutex = sync.Mutex{}
var builderImported = false
var testerImported = false

func ImportInternalBuilderIfNeeded(manifest *common.AciManifest) {
	if manifest.Builder.Image.String() == "" {
		manifest.Builder.Image = *aciBuilder

		importMutex.Lock()
		defer importMutex.Unlock()
		if builderImported {
			return
		}

		importInternalAci("aci-builder.aci")
		builderImported = true
	}
}

func ImportInternalTesterIfNeeded(manifest *common.AciManifest) {
	ImportInternalBuilderIfNeeded(manifest)
	if manifest.Tester.Builder.Image.String() == "" {
		manifest.Tester.Builder.Image = *aciTester

		importMutex.Lock()
		defer importMutex.Unlock()
		if testerImported {
			return
		}

		importInternalAci("aci-tester.aci")
		testerImported = true
	}
}

func importInternalAci(filename string) {
	filepath := "dist/bindata/" + filename
	content, err := dist.Asset(filepath)
	if err != nil {
		logs.WithE(err).WithField("aci", filepath).Fatal("Cannot found internal aci")
	}
	tmpFile := "/tmp/" + RandStringBytesMaskImpr(20) + ".aci"
	if err := ioutil.WriteFile(tmpFile, content, 0644); err != nil {
		logs.WithE(err).WithField("aci", filepath).Fatal("Failed to write tmp aci to /tmp/tmp.aci")
	}
	defer os.Remove(tmpFile)
	if _, err := Home.Rkt.FetchInsecure(tmpFile); err != nil {
		logs.WithE(err).Fatal("Failed to import internal image to rkt")
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringBytesMaskImpr(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
