package logs

import (
	"fmt"
	"strings"
)

type Level uint8

const (
	// the program cannot continue. will log and exit(1)
	FATAL Level = iota
	// the routine cannot continue. will log and call go panic function
	PANIC
	// program continue but caller have received error or data lost or similar
	ERROR
	// recovered problem or non critical
	WARN
	// general info
	INFO
	// tell what is going on step by step
	DEBUG
	// log data content
	TRACE
)

func (level Level) String() string {
	switch level {
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	case PANIC:
		return "PANIC"
	}
	return "UNKNOWN"
}

func ParseLevel(lvl string) (Level, error) {
	lvl = strings.ToUpper(lvl)
	switch lvl {
	case "PANIC":
		return PANIC, nil
	case "FATAL":
		return FATAL, nil
	case "ERROR":
		return ERROR, nil
	case "WARN", "WARNING":
		return WARN, nil
	case "INFO":
		return INFO, nil
	case "DEBUG":
		return DEBUG, nil
	case "TRACE":
		return TRACE, nil
	}
	var l Level
	return l, fmt.Errorf("Not a valid level : %s", lvl) // not using errs to prevent cycle dep
}

func (l Level) IsEnableFor(level Level) bool {
	return level >= l
}
