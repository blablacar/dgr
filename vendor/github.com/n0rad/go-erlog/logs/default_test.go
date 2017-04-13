package logs

import (
	"bytes"
	o "github.com/onsi/gomega"
	"os"
	"testing"
)

var buffer bytes.Buffer

func reset() {
	buffer = bytes.Buffer{}
	GetDefaultLog().(*DummyLog).Out = &buffer
}

func TestMain(m *testing.M) {
	reset()
	os.Exit(m.Run())
}

func TestLog(t *testing.T) {
	o.RegisterTestingT(t)
	reset()

	Trace("trace")
	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error")

	o.Expect(buffer.String()).To(o.Equal(`TRACE: trace
DEBUG: debug
INFO: info
WARN: warn
ERROR: error
`))
}
