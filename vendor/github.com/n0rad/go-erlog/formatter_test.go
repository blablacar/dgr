package erlog

//import (
//	"fmt"
//	o "github.com/onsi/gomega"
//	"os"
//	"testing"
//)
//
//func TestLog(t *testing.T) {
//	o.RegisterTestingT(t)
//
//	SetOutput(os.Stdout)
//	logrus.SetFormatter(&DefaultFormatter{})
//	logrus.SetLevel(logrus.DebugLevel)
//
//	defer func() {
//		if r := recover(); r != nil {
//			fmt.Println("Recovered in f", r)
//		}
//	}()
//
//	logrus.WithField("yopla", "boom").Debug("salut")
//	logrus.WithField("yopla", "boom").Info("salut")
//	logrus.WithField("yopla", "boom").Warn("salut")
//	logrus.WithField("yopla", "boom").Error("salut")
//	logrus.WithField("yopla", "boom").Panic("salut")
//}
//
//func TestReduce(t *testing.T) {
//	o.RegisterTestingT(t)
//
//	format := DefaultFormatter{}
//
//	o.Expect(format.reduceFilePath("com/blablacar/cnt/test.go", 50)).To(o.Equal("com/blablacar/cnt/test.go"))
//	o.Expect(format.reduceFilePath("com/blablacar/cnt/test.go", 25)).To(o.Equal("com/blablacar/cnt/test.go"))
//	o.Expect(format.reduceFilePath("com/blablacar/cnt/test.go", 24)).To(o.Equal("c/blablacar/cnt/test.go"))
//	o.Expect(format.reduceFilePath("com/blablacar/cnt/test.go", 23)).To(o.Equal("c/blablacar/cnt/test.go"))
//	o.Expect(format.reduceFilePath("com/blablacar/cnt/test.go", 22)).To(o.Equal("c/b/cnt/test.go"))
//	o.Expect(format.reduceFilePath("com/blablacar/cnt/test.go", 12)).To(o.Equal("c/b/c/test.go"))
//	o.Expect(format.reduceFilePath("com/blablacar/cnt/test.go", 10)).To(o.Equal("c/b/c/test.go"))
//}
