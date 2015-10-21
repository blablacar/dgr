package log

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestEnableFor(t *testing.T) {
	RegisterTestingT(t)

	Expect(LogLevel("error")).To(Equal(&ERROR))
	Expect(LogLevel("waRn")).To(Equal(&WARN))
	Expect(LogLevel("iNfo")).To(Equal(&INFO))
	Expect(LogLevel("debug")).To(Equal(&DEBUG))
	Expect(LogLevel("TRACE")).To(Equal(&TRACE))
}
