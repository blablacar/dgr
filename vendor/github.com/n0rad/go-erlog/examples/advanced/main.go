package main

import (
	"github.com/n0rad/go-erlog/log"
	_ "github.com/n0rad/go-erlog/register"
	"github.com/n0rad/go-erlog/with"
	"os"
)

func main() {
	logger := log.GetLog("newlog") // another logger

	//	logger.(*erlog.ErlogLogger).Appenders[0].(*erlog.ErlogWriterAppender).Out = os.Stdout

	path := "/toto/config"
	if err := os.Mkdir(path, 0777); err != nil {
		log.WithEF(err, with.WithField("dir", path)).Info("Failed to create config directory")

		logger.LogEntry(&log.Entry{
			Fields:  with.WithField("dir", path),
			Level:   log.INFO,
			Err:     err,
			Message: "Salut !1",
		})
	}

	logger.Info("Salut !2")
	logger.Info("Salut !3")

}
