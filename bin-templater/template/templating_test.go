package template

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestIsValidType(t *testing.T) {
	RegisterTestingT(t)

	Expect(IsType("zerzrze", "string")).To(BeTrue())
	Expect(IsType(true, "string")).To(BeFalse())
	Expect(IsType(true, "bool")).To(BeTrue())
	Expect(IsType(make(map[string]int), "map[string]int")).To(BeTrue())
	Expect(IsType([]string{}, "[]string")).To(BeTrue())
	Expect(IsArray([]string{})).To(BeTrue())
	Expect(IsMap(make(map[string]int))).To(BeTrue())
	Expect(IsKind(make(map[string]int), "map")).To(BeTrue())
}
