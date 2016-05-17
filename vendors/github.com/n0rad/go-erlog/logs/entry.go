package logs

import (
	"github.com/n0rad/go-erlog/data"
)

type Entry struct {
	Logger  Log
	Level   Level
	Fields  data.Fields
	Message string
	Err     error
}

func WithError(err error) *Entry {
	return &Entry{
		Logger: GetDefaultLog(),
		Err:    err,
	}
}

func WithField(name string, value interface{}) *Entry {
	return &Entry{
		Logger: GetDefaultLog(),
		Fields: data.WithField(name, value),
	}
}

func WithFields(fields data.Fields) *Entry {
	return &Entry{
		Logger: GetDefaultLog(),
		Fields: fields,
	}
}

func WithEF(err error, fields data.Fields) *Entry {
	return &Entry{
		Logger: GetDefaultLog(),
		Err:    err,
		Fields: fields,
	}
}

func WithF(fields data.Fields) *Entry {
	return WithFields(fields)
}

func WithE(err error) *Entry {
	return WithError(err)
}


///////////////////////////////////

func (e *Entry) WithFields(data data.Fields) *Entry {
	e.Fields = data
	return e
}

func (e *Entry) WithField(name string, value interface{}) *Entry {
	if e.Fields == nil {
		e.Fields = data.WithField(name, value)
	} else {
		e.Fields = e.Fields.WithField(name, value)
	}
	return e
}

func (e *Entry) WithLog(logger Log) *Entry {
	e.Logger = logger
	return e
}

func (e *Entry) Trace(msg string) {
	e.Level = TRACE
	e.Message = msg
	e.Logger.LogEntry(e)
}

func (e *Entry) Debug(msg string) {
	e.Level = DEBUG
	e.Message = msg
	e.Logger.LogEntry(e)
}

func (e *Entry) Info(msg string) {
	e.Level = INFO
	e.Message = msg
	e.Logger.LogEntry(e)
}

func (e *Entry) Warn(msg string) {
	e.Level = WARN
	e.Message = msg
	e.Logger.LogEntry(e)
}

func (e *Entry) Error(msg string) {
	e.Level = ERROR
	e.Message = msg
	e.Logger.LogEntry(e)
}

func (e *Entry) Panic(msg string) {
	e.Level = PANIC
	e.Message = msg
	e.Logger.LogEntry(e)
}

func (e *Entry) Fatal(msg string) {
	e.Level = FATAL
	e.Message = msg
	e.Logger.LogEntry(e)
}
