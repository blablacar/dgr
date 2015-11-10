package log

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	. "github.com/onsi/gomega"
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	RegisterTestingT(t)

	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&BlaFormatter{})
	logrus.SetLevel(logrus.DebugLevel)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	logrus.WithField("yopla", "boom").Debug("salut")
	logrus.WithField("yopla", "boom").Info("salut")
	logrus.WithField("yopla", "boom").Warn("salut")
	logrus.WithField("yopla", "boom").Error("salut")
	logrus.WithField("yopla", "boom").Panic("salut")
}

func TestReduce(t *testing.T) {
	RegisterTestingT(t)

	format := BlaFormatter{}

	Expect(format.reduceFilePath("com/blablacar/cnt/test.go", 50)).To(Equal("com/blablacar/cnt/test.go"))
	Expect(format.reduceFilePath("com/blablacar/cnt/test.go", 25)).To(Equal("com/blablacar/cnt/test.go"))
	Expect(format.reduceFilePath("com/blablacar/cnt/test.go", 24)).To(Equal("c/blablacar/cnt/test.go"))
	Expect(format.reduceFilePath("com/blablacar/cnt/test.go", 23)).To(Equal("c/blablacar/cnt/test.go"))
	Expect(format.reduceFilePath("com/blablacar/cnt/test.go", 22)).To(Equal("c/b/cnt/test.go"))
	Expect(format.reduceFilePath("com/blablacar/cnt/test.go", 12)).To(Equal("c/b/c/test.go"))
	Expect(format.reduceFilePath("com/blablacar/cnt/test.go", 10)).To(Equal("c/b/c/test.go"))
}
