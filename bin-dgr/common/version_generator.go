package common

import (
	"fmt"
	"time"

	"github.com/n0rad/go-erlog/logs"
)

func GenerateVersion(aciHome string) string {
	version := generateDate()
	if hash, err := GitHash(aciHome); err == nil {
		version += "-v" + hash
	}
	return version
}

func generateDate() string {
	return fmt.Sprintf("%s", time.Now().Format("20060102.150405"))
}

func GitHash(path string) (string, error) {
	out, _, err := ExecCmdGetStdoutAndStderr("git", "-C", path, "rev-parse", "--short", "HEAD")
	if err != nil {
		logs.WithE(err).WithField("path", path).Debug("Failed to get git hash from path")
		return "", err
	}
	return out, nil
}
