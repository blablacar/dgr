package main

import (
	"github.com/blablacar/dgr/utils"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const info_template = `package dgr

func init() {
	Version = "X.X.X"
	CommitHash = "HASH"
	BuildDate = "DATE"
}`

func main() {
	hash := utils.GitHash()

	version := os.Getenv("VERSION")
	if version == "" {
		panic("You must set dgr version into VERSION env to generate. ex: # VERSION=1.0 go generate")
	}
	buildDate := time.Now()

	res := strings.Replace(info_template, "X.X.X", string(version), 1)
	res = strings.Replace(res, "HASH", hash, 1)
	res = strings.Replace(res, "DATE", buildDate.Format(time.RFC3339), 1)

	ioutil.WriteFile("dgr/version.go", []byte(res), 0644)
}
