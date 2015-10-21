package log

import (
	"fmt"
	"os"
)

type dummyLog struct {
	out *os.File
}

func (l dummyLog) log(lvl Level, msg string) {
	l.out.WriteString(lvl.Name)
	l.out.WriteString(": ")
	l.out.WriteString(msg)
	l.out.WriteString("\n")
}

func (l dummyLog) Error(args ...interface{}) { l.log(ERROR, fmt.Sprint(args...)) }
func (l dummyLog) Warn(args ...interface{})  { l.log(WARN, fmt.Sprint(args...)) }
func (l dummyLog) Info(args ...interface{})  { l.log(INFO, fmt.Sprint(args...)) }
func (l dummyLog) Debug(args ...interface{}) { l.log(DEBUG, fmt.Sprint(args...)) }
func (l dummyLog) Trace(args ...interface{}) { l.log(TRACE, fmt.Sprint(args...)) }

func (l dummyLog) Errorf(format string, args ...interface{}) {
	l.log(ERROR, fmt.Sprintf(format, args...))
}
func (l dummyLog) Warnf(format string, args ...interface{}) { l.log(WARN, fmt.Sprintf(format, args...)) }
func (l dummyLog) Infof(format string, args ...interface{}) { l.log(INFO, fmt.Sprintf(format, args...)) }
func (l dummyLog) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, fmt.Sprintf(format, args...))
}
func (l dummyLog) Tracef(format string, args ...interface{}) {
	l.log(TRACE, fmt.Sprintf(format, args...))
}
