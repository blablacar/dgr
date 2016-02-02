package logs

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type DummyLog struct {
	Out io.Writer
}

func (d *DummyLog) GetLog(name string) Log { return d }

func (d *DummyLog) Tracef(format string, msg ...interface{}) {
	d.log(TRACE, fmt.Sprintf(format, msg...))
}
func (d *DummyLog) Debugf(format string, msg ...interface{}) {
	d.log(DEBUG, fmt.Sprintf(format, msg...))
}
func (d *DummyLog) Infof(format string, msg ...interface{}) {
	d.log(INFO, fmt.Sprintf(format, msg...))
}
func (d *DummyLog) Warnf(format string, msg ...interface{}) {
	d.log(WARN, fmt.Sprintf(format, msg...))
}
func (d *DummyLog) Errorf(format string, msg ...interface{}) {
	d.log(ERROR, fmt.Sprintf(format, msg...))
}
func (d *DummyLog) Panicf(format string, msg ...interface{}) {
	d.log(PANIC, fmt.Sprintf(format, msg...))
}
func (d *DummyLog) Fatalf(format string, msg ...interface{}) {
	d.log(FATAL, fmt.Sprintf(format, msg...))
}

func (d *DummyLog) Trace(msg ...string) { d.log(TRACE, msg...) }
func (d *DummyLog) Debug(msg ...string) { d.log(DEBUG, msg...) }
func (d *DummyLog) Info(msg ...string)  { d.log(INFO, msg...) }
func (d *DummyLog) Warn(msg ...string)  { d.log(WARN, msg...) }
func (d *DummyLog) Error(msg ...string) { d.log(ERROR, msg...) }
func (d *DummyLog) Panic(msg ...string) { d.log(PANIC, msg...); panic(msg) }
func (d *DummyLog) Fatal(msg ...string) { d.log(FATAL, msg...); os.Exit(1) }
func (d *DummyLog) log(level Level, msg ...string) {
	d.LogEntry(&Entry{Level: level, Message: strings.Join(msg, " ")})
}
func (d *DummyLog) LogEntry(entry *Entry)         { fmt.Fprintf(d.Out, "%s: %s\n", entry.Level, entry.Message) }
func (d *DummyLog) GetLevel() Level               { return TRACE }
func (d *DummyLog) SetLevel(lvl Level)            { d.Error("Dummy log cannot set level") }
func (d *DummyLog) IsLevelEnabled(lvl Level) bool { return true }
func (d *DummyLog) IsTraceEnabled() bool          { return true }
func (d *DummyLog) IsDebugEnabled() bool          { return true }
func (d *DummyLog) IsInfoEnabled() bool           { return true }
func (d *DummyLog) IsWarnEnabled() bool           { return true }
func (d *DummyLog) IsErrorEnabled() bool          { return true }
func (d *DummyLog) IsPanicEnabled() bool          { return true }
func (d *DummyLog) IsFatalEnabled() bool          { return true }
