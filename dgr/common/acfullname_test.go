package common

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestAcFullname(t *testing.T) {
	RegisterTestingT(t)

	Expect(NewACFullName("blablacar.github.io/dgr/pod-cassandra:42").Name()).To(Equal("blablacar.github.io/dgr/pod-cassandra"))
	Expect(NewACFullName("blablacar.github.io/dgr/pod-cassandra:42").DomainName()).To(Equal("blablacar.github.io"))
	Expect(NewACFullName("blablacar.github.io/dgr/pod-cassandra:42").ShortName()).To(Equal("dgr/pod-cassandra"))
	Expect(NewACFullName("blablacar.github.io/dgr/pod-cassandra:42").Version()).To(Equal("42"))
}
