package log

import (
	"errors"
	"strings"
)

type Level struct {
	Name  string
	value int
}

var ERROR = Level{Name: "ERROR", value: 0}
var WARN = Level{Name: "WARN", value: 1}
var INFO = Level{Name: "INFO", value: 2}
var DEBUG = Level{Name: "DEBUG", value: 3}
var TRACE = Level{Name: "TRACE", value: 4}

var levels = []Level{
	ERROR,
	WARN,
	INFO,
	DEBUG,
	TRACE,
}

func (l *Level) IsEnableFor(level Level) bool {
	return level.value <= l.value
}

func LogLevel(level string) (*Level, error) {
	for _, current := range levels {
		if strings.EqualFold(current.Name, level) {
			return &current, nil
		}
	}
	return nil, errors.New("Unknown log level :" + level)
}
