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
