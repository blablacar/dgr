package main
import (
	"github.com/blablacar/cnt/utils"
	"time"
	"io/ioutil"
	"strings"
	"os"
)


const (info_template=`package cnt

func init() {
	Version = "X.X.X"
	CommitHash = "HASH"
	BuildDate = "DATE"
}`)

func main() {
	hash := utils.GitHash()

	version := os.Getenv("VERSION")
	if version == "" {
		panic("You must set cnt version into VERSION env to generate: # VERSION=1.0 go generate")
	}
	buildDate := time.Now()

	res := strings.Replace(info_template, "X.X.X", string(version), 1)
	res = strings.Replace(res, "HASH", hash, 1)
	res = strings.Replace(res, "DATE", buildDate.Format(time.RFC3339), 1)

	ioutil.WriteFile("cnt/version.go", []byte(res), 0644)
}
