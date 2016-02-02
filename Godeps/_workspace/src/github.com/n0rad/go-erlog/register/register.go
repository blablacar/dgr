package register

import (
	"github.com/n0rad/go-erlog"
	"github.com/n0rad/go-erlog/logs"
)

func init() {
	logs.RegisterLoggerFactory(erlog.NewErlogFactory())
}
