package utils
import (
	"fmt"
	"time"
)

func GenerateVersion() string {
	return generateDate() + "-" + gitHash();
}

func generateDate() string {
	return fmt.Sprintf("%s", time.Now().Format("20060102.150405"))
}


func gitHash() string {
	out, _ := ExecCmdGetOutput("git", "rev-parse", "--short", "HEAD")
	return out;
}
