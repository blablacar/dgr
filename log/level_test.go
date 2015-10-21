package log

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestLevelStringConversion(t *testing.T) {
	RegisterTestingT(t)

	Expect(LogLevel("error")).To(Equal(&ERROR))
	Expect(LogLevel("waRn")).To(Equal(&WARN))
	Expect(LogLevel("iNfo")).To(Equal(&INFO))
	Expect(LogLevel("debug")).To(Equal(&DEBUG))
	Expect(LogLevel("TRACE")).To(Equal(&TRACE))
}

func TestLevelEnableFor(t *testing.T) {
	RegisterTestingT(t)

	Expect(ERROR.IsEnableFor(WARN)).To(BeTrue())
	Expect(WARN.IsEnableFor(WARN)).To(BeTrue())
	Expect(INFO.IsEnableFor(WARN)).To(BeFalse())
	Expect(DEBUG.IsEnableFor(WARN)).To(BeFalse())
	Expect(TRACE.IsEnableFor(WARN)).To(BeFalse())
}
