package log
import (
	"fmt"
	"os"
)

type Log interface {

	Panic(args ...interface{})
	Error(args ...interface{})
	Warn(args ...interface{})
	Info(args ...interface{})
	Debug(args ...interface{})
	Trace(args ...interface{})

	Panicf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Tracef(format string, args ...interface{})

	//TODO isEnable
}

type dummyLog struct {}
func (l dummyLog) log(msg string) {println(msg)}

func (l dummyLog) Panic(args ...interface{}) {l.log(fmt.Sprint(args)); os.Exit(1)}
func (l dummyLog) Error(args ...interface{}) {l.log(fmt.Sprint(args))}
func (l dummyLog) Warn(args ...interface{}) {l.log(fmt.Sprint(args))}
func (l dummyLog) Info(args ...interface{}) {l.log(fmt.Sprint(args))}
func (l dummyLog) Debug(args ...interface{}) {l.log(fmt.Sprint(args))}
func (l dummyLog) Trace(args ...interface{}) {l.log(fmt.Sprint(args))}

func (l dummyLog) Panicf(f string, args ...interface{}) {l.log(fmt.Sprintf(f, args)); os.Exit(1)}
func (l dummyLog) Errorf(f string, args ...interface{}) {l.log(fmt.Sprintf(f, args))}
func (l dummyLog) Warnf(f string, args ...interface{}) {l.log(fmt.Sprintf(f, args))}
func (l dummyLog) Infof(f string, args ...interface{}) {l.log(fmt.Sprintf(f, args))}
func (l dummyLog) Debugf(f string, args ...interface{}) {l.log(fmt.Sprintf(f, args))}
func (l dummyLog) Tracef(f string, args ...interface{}) {l.log(fmt.Sprintf(f, args))}

var currentLogger Log = dummyLog{}

func Get() Log {
	return currentLogger
}

func Set(newLogger Log) {
	currentLogger = newLogger;
}
