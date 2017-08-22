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

func TestLatestUrl(t *testing.T) {
	RegisterTestingT(t)

	version, _ := checkForAcserver("/aci.awired.net/aci-glibc/aci-gli.aci")
	Expect(version).To(Equal(""))

	version, _ = checkForAcserver("/aci.awired.net/aci-glibc/aci-glibc-2.21-1-linux-amd64.aci")
	Expect(version).To(Equal("2.21-1"))

	version, _ = checkForNexus("/nexus/service/local/repositories/releases/content/com/blablacar/aci/linux/amd64/aci-debian/8.8-6/aci-debian-8.8-6.aci")
	Expect(version).To(Equal("8.8-6"))
}
