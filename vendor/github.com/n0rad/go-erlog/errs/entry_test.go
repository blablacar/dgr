package errs

import (
	"fmt"
	o "github.com/onsi/gomega"
	"testing"
	"github.com/n0rad/go-erlog/data"
)

func TestConstruction(t *testing.T) {
	o.RegisterTestingT(t)

	path := "/bin/toto42"
	err := fmt.Errorf("%s", "erf")

	err1 := With("salut").WithErr(err).WithField("path", path)
	err2 := WithE(err, "salut").WithField("path", path)

	o.Expect(Is(err1, err2)).To(o.BeTrue())
}

func TestNulError(t *testing.T) {
	o.RegisterTestingT(t)

	var err error = nil
	e := WithEF(err, data.WithField("yop", "la"), "Url check failed")

	e.Error()
}