package runner
import (
	"github.com/blablacar/cnt/log"
	"strings"
	"bytes"
	"io"
	"os/exec"
	"github.com/blablacar/cnt/utils"
)

type Runner interface {

	prepare(target string) error

	run(target string) error

	release(target string) error
}
