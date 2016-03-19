package erlog

import (
	"github.com/n0rad/go-erlog/logs"
)

type Appender interface {
	Fire(event *LogEvent)
	GetLevel() logs.Level
	SetLevel(level logs.Level)
}
