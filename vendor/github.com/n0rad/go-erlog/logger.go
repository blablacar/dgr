package erlog

import (
	"fmt"
	"github.com/n0rad/go-erlog/logs"
	"os"
	"strings"
)

type ErlogFactory struct {
	defaultLog *ErlogLogger
	logs       map[string]*ErlogLogger
}

func NewErlogFactory() *ErlogFactory {
	return &ErlogFactory{
		defaultLog: newLog(),
		logs:       make(map[string]*ErlogLogger, 10),
	}
}

func (l *ErlogFactory) GetLog(name string) logs.Log {
	if name == "" {
		return l.defaultLog
	}
	log := l.logs[name]
	if log == nil {
		log = newLog()
		l.logs[name] = log
	}
	return log
}

type ErlogLogger struct {
	Appenders []Appender
	Level     logs.Level
}

func newLog() *ErlogLogger {
	return &ErlogLogger{
		Appenders: []Appender{NewErlogWriterAppender(os.Stderr)},
		Level:     logs.INFO,
	}
}

func (l *ErlogLogger) log(event *LogEvent) {
	for _, appender := range l.Appenders {
		appender.Fire(event)
	}
}

func (l *ErlogLogger) Trace(message ...string) {
	if logs.TRACE.IsEnableFor(l.Level) {
		l.logS(logs.TRACE, message...)
	}
}

func (l *ErlogLogger) Debug(message ...string) {
	if logs.DEBUG.IsEnableFor(l.Level) {
		l.logS(logs.DEBUG, message...)
	}
}
func (l *ErlogLogger) Info(message ...string) {
	if logs.INFO.IsEnableFor(l.Level) {
		l.logS(logs.INFO, message...)
	}
}
func (l *ErlogLogger) Warn(message ...string) {
	if logs.WARN.IsEnableFor(l.Level) {
		l.logS(logs.WARN, message...)
	}
}
func (l *ErlogLogger) Error(message ...string) {
	if logs.ERROR.IsEnableFor(l.Level) {
		l.logS(logs.ERROR, message...)
	}
}
func (l *ErlogLogger) Panic(message ...string) {
	if logs.PANIC.IsEnableFor(l.Level) {
		l.logS(logs.PANIC, message...)
	}
	panic(strings.Join(message, " "))
}
func (l *ErlogLogger) Fatal(message ...string) {
	if logs.FATAL.IsEnableFor(l.Level) {
		l.logS(logs.FATAL, message...)
	}
	os.Exit(1)
}

func (l *ErlogLogger) Tracef(format string, message ...interface{}) {
	if logs.TRACE.IsEnableFor(l.Level) {
		l.logS(logs.TRACE, fmt.Sprintf(format, message))
	}
}

func (l *ErlogLogger) Debugf(format string, message ...interface{}) {
	if logs.DEBUG.IsEnableFor(l.Level) {
		l.logS(logs.DEBUG, fmt.Sprintf(format, message))
	}
}

func (l *ErlogLogger) Infof(format string, message ...interface{}) {
	if logs.INFO.IsEnableFor(l.Level) {
		l.logS(logs.INFO, fmt.Sprintf(format, message))
	}
}

func (l *ErlogLogger) Warnf(format string, message ...interface{}) {
	if logs.WARN.IsEnableFor(l.Level) {
		l.logS(logs.WARN, fmt.Sprintf(format, message))
	}
}

func (l *ErlogLogger) Errorf(format string, message ...interface{}) {
	if logs.ERROR.IsEnableFor(l.Level) {
		l.logS(logs.ERROR, fmt.Sprintf(format, message))
	}
}

func (l *ErlogLogger) Panicf(format string, message ...interface{}) {
	if logs.PANIC.IsEnableFor(l.Level) {
		l.logS(logs.PANIC, fmt.Sprintf(format, message))
	}
	panic(fmt.Sprintf(format, message))
}

func (l *ErlogLogger) Fatalf(format string, message ...interface{}) {
	if logs.FATAL.IsEnableFor(l.Level) {
		l.logS(logs.FATAL, fmt.Sprintf(format, message))
	}
	os.Exit(1)
}

func (l *ErlogLogger) logS(level logs.Level, msg ...string) {
	l.log(NewLogEvent(&logs.Entry{Level: level, Message: strings.Join(msg, " ")}))
}

func (l *ErlogLogger) LogEntry(entry *logs.Entry) {
	if entry.Level.IsEnableFor(l.Level) {
		l.log(NewLogEvent(entry))
	}
	if entry.Level == logs.PANIC {
		panic(entry.Message)
	} else if entry.Level == logs.FATAL {
		os.Exit(1)
	}
}

func (l *ErlogLogger) GetLevel() logs.Level      { return l.Level }
func (l *ErlogLogger) SetLevel(level logs.Level) { l.Level = level }

func (l *ErlogLogger) IsTraceEnabled() bool                 { return logs.TRACE.IsEnableFor(l.Level) }
func (l *ErlogLogger) IsDebugEnabled() bool                 { return logs.DEBUG.IsEnableFor(l.Level) }
func (l *ErlogLogger) IsInfoEnabled() bool                  { return logs.INFO.IsEnableFor(l.Level) }
func (l *ErlogLogger) IsWarnEnabled() bool                  { return logs.WARN.IsEnableFor(l.Level) }
func (l *ErlogLogger) IsErrorEnabled() bool                 { return logs.ERROR.IsEnableFor(l.Level) }
func (l *ErlogLogger) IsPanicEnabled() bool                 { return logs.PANIC.IsEnableFor(l.Level) }
func (l *ErlogLogger) IsFatalEnabled() bool                 { return logs.FATAL.IsEnableFor(l.Level) }
func (l *ErlogLogger) IsLevelEnabled(level logs.Level) bool { return level.IsEnableFor(l.Level) }
