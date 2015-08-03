package main
import (
	"os"
	"runtime"
	"github.com/blablacar/cnt/commands"
)

//go:generate go run compile/info_generate.go

func main() {
	if os.Getuid() != 0 {
		println("Cnt needs to be run as root")
		os.Exit(1)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	commands.Execute();
}
