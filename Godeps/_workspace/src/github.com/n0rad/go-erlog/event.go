package erlog

import (
	"github.com/n0rad/go-erlog/logs"
	"runtime"
	"strings"
	"time"
)

type LogEvent struct {
	logs.Entry
	Depth int
	Time  time.Time
	File  string
	Line  int
}

func NewLogEvent(entry *logs.Entry) *LogEvent {
	var file string
	var line int
	var ok bool
	for i := 2; ; i++ {
		_, file, line, ok = runtime.Caller(i)
		if !ok {
			file = "???"
			line = 0
		}
		if !strings.Contains(file, "n0rad/go-erlog") {
			break
		}
		if strings.Contains(file, "n0rad/go-erlog/examples") { // TODO what to do with that ?
			break
		}
	}

	return &LogEvent{
		Entry: *entry,
		File:  file,
		Line:  line,
	}
}
