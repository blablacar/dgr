package main

import (
	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
	"github.com/n0rad/go-erlog/data"
	"os"
)

func main() {
	logger := logs.GetLog("newlog") // another logger

	//	logger.(*erlog.ErlogLogger).Appenders[0].(*erlog.ErlogWriterAppender).Out = os.Stdout

	path := "/toto/config"
	if err := os.Mkdir(path, 0777); err != nil {
		logs.WithEF(err, data.WithField("dir", path)).Info("Failed to create config directory")

		logger.LogEntry(&logs.Entry{
			Fields:  data.WithField("dir", path),
			Level:   logs.INFO,
			Err:     err,
			Message: "Salut !1",
		})
	}

	logger.Info("Salut !2")
	logger.Info("Salut !3")

}
