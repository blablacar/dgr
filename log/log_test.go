package log

import (
	"bytes"
	. "github.com/onsi/gomega"
	"io"
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	RegisterTestingT(t)

	r, w, _ := os.Pipe()
	Logger = dummyLog{out: w}

	Error("this is a error")
	Warn("this is a warn")
	Info("this is a info")
	Debug("this is a debug")
	Trace("this is a trace")

	Expect(getPipeContent(r, w)).To(Equal(`ERROR: this is a error
WARN: this is a warn
INFO: this is a info
DEBUG: this is a debug
TRACE: this is a trace
`))
}

func getPipeContent(r *os.File, w *os.File) string {
	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()
	w.Close()
	return <-outC
}
