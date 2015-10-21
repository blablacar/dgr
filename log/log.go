package log

import "os"

type Log interface {
	//Fatal() os.exit() should never be triggered by a logger
	//Panic()  panic() should never be triggered by a logger

	Error(args ...interface{})
	Warn(args ...interface{})
	Info(args ...interface{})
	Debug(args ...interface{})
	Trace(args ...interface{})

	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Tracef(format string, args ...interface{})
}

var Logger Log = &dummyLog{out: os.Stdout}

func Error(args ...interface{}) { Logger.Error(args...) }
func Warn(args ...interface{})  { Logger.Warn(args...) }
func Info(args ...interface{})  { Logger.Info(args...) }
func Debug(args ...interface{}) { Logger.Debug(args...) }
func Trace(args ...interface{}) { Logger.Trace(args...) }

func Errorf(format string, args ...interface{}) { Logger.Errorf(format, args...) }
func Warnf(format string, args ...interface{})  { Logger.Warnf(format, args...) }
func Infof(format string, args ...interface{})  { Logger.Infof(format, args...) }
func Debugf(format string, args ...interface{}) { Logger.Debugf(format, args...) }
func Tracef(format string, args ...interface{}) { Logger.Tracef(format, args...) }
