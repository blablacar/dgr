package logs

import (
	o "github.com/onsi/gomega"
	"testing"
)

func TestLevelStringConversion(t *testing.T) {
	o.RegisterTestingT(t)

	o.Expect(ParseLevel("error")).To(o.Equal(ERROR))
	o.Expect(ParseLevel("waRn")).To(o.Equal(WARN))
	o.Expect(ParseLevel("iNfo")).To(o.Equal(INFO))
	o.Expect(ParseLevel("debug")).To(o.Equal(DEBUG))
	o.Expect(ParseLevel("TRACE")).To(o.Equal(TRACE))
}

func TestLevelEnableFor(t *testing.T) {
	o.RegisterTestingT(t)

	o.Expect(PANIC.IsEnableFor(WARN)).To(o.BeTrue())
	o.Expect(FATAL.IsEnableFor(WARN)).To(o.BeTrue())
	o.Expect(ERROR.IsEnableFor(WARN)).To(o.BeTrue())
	o.Expect(WARN.IsEnableFor(WARN)).To(o.BeTrue())
	o.Expect(INFO.IsEnableFor(WARN)).To(o.BeFalse())
	o.Expect(DEBUG.IsEnableFor(WARN)).To(o.BeFalse())
	o.Expect(TRACE.IsEnableFor(WARN)).To(o.BeFalse())
}

func TestUnknownLevel(t *testing.T) {
	o.RegisterTestingT(t)

	_, err := ParseLevel("does not exists")
	o.Expect(err).NotTo(o.BeNil())
}
