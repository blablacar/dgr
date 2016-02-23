package main

import (
	"fmt"
	"github.com/blablacar/dgr/bin-dgr/common"
	"time"
)

func GenerateVersion() string {
	return generateDate() + "-v" + GitHash()
}

func generateDate() string {
	return fmt.Sprintf("%s", time.Now().Format("20060102.150405"))
}

func GitHash() string {
	out, _ := common.ExecCmdGetOutput("git", "rev-parse", "--short", "HEAD")
	return out
}
