package logs

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"sync"
)

type LogFactory interface {
	GetLog(name string) Log
}

type Log interface {
	Trace(msg ...string)
	Debug(msg ...string)
	Info(msg ...string)
	Warn(msg ...string)
	Error(msg ...string)
	Panic(msg ...string)
	Fatal(msg ...string)

	Tracef(format string, msg ...interface{})
	Debugf(format string, msg ...interface{})
	Infof(format string, msg ...interface{})
	Warnf(format string, msg ...interface{})
	Errorf(format string, msg ...interface{})
	Panicf(format string, msg ...interface{})
	Fatalf(format string, msg ...interface{})

	LogEntry(entry *Entry)

	GetLevel() Level
	SetLevel(lvl Level)

	IsLevelEnabled(lvl Level) bool
	IsTraceEnabled() bool
	IsDebugEnabled() bool
	IsInfoEnabled() bool
	IsWarnEnabled() bool
	IsErrorEnabled() bool
	IsPanicEnabled() bool
	IsFatalEnabled() bool
}

var factory LogFactory = &DummyLog{Out: os.Stderr}
var mu sync.Mutex

func RegisterLoggerFactory(f LogFactory) {
	mu.Lock()
	if f == factory {
		return
	}

	if _, ok := factory.(*DummyLog); !ok {
		_, file, line, _ := runtime.Caller(1)
		fmt.Fprintf(os.Stderr, "Re-Registering the logger factory : %s:%d. There is already one registered : %s\n", file, line, reflect.TypeOf(factory))
	}
	factory = f
	mu.Unlock()
}
