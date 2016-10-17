package errs

import (
	"bytes"
	"fmt"
	"github.com/n0rad/go-erlog/data"
	"runtime"
)

var MaxStackDepth = 50

type EntryError struct {
	Fields  data.Fields
	Message string
	Errs    []error
	Stack   []uintptr
	frames  []StackFrame
}

func With(message string) *EntryError {
	return fill(&EntryError{
		Message: message,
	})
}

func WithF(fields data.Fields, msg string) *EntryError {
	return fill(&EntryError{
		Fields:  fields,
		Message: msg,
	})
}

func WithE(err error, msg string) *EntryError {
	return fill(&EntryError{
		Errs:    []error{err},
		Message: msg,
	})
}

func WithEF(err error, fields data.Fields, msg string) *EntryError {
	return fill(&EntryError{
		Errs:    []error{err},
		Fields:  fields,
		Message: msg,
	})
}

func fill(entry *EntryError) *EntryError {
	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(2, stack[:])
	entry.Stack = stack[:length]
	return entry
}

///////////////////////////////////////////////

func Is(e1 error, e2 error) bool {
	if e1 == e2 {
		return true
	}

	ee1, ok1 := e1.(*EntryError)
	ee2, ok2 := e2.(*EntryError)
	if ok1 && ok2 && ee1.Message == ee2.Message {
		return true
	}

	if e1.Error() == e2.Error() {
		return true
	}

	return false
}

//////////////////////////////////////////////

func (e *EntryError) WithFields(data data.Fields) *EntryError {
	e.Fields = data
	return e
}

func (e *EntryError) WithErr(err error) *EntryError {
	if err != nil {
		e.Errs = append(e.Errs, err)
	}
	return e
}

func (e *EntryError) WithErrs(errs ...error) *EntryError {
	for _, err := range errs {
		if err != nil {
			e.Errs = append(e.Errs, err)
		}
	}
	return e
}

func (e *EntryError) WithField(name string, value interface{}) *EntryError {
	if e.Fields == nil {
		e.Fields = data.WithField(name, value)
	} else {
		e.Fields = e.Fields.WithField(name, value)
	}
	return e
}

func (e *EntryError) WithMessage(msg string) *EntryError {
	e.Message = msg
	return e
}

func (e *EntryError) Error() string {
	var buffer bytes.Buffer
	buffer.WriteString(e.Message)
	if e.Fields != nil {
		for key := range e.Fields {
			buffer.WriteString(" ")
			buffer.WriteString(key)
			buffer.WriteString("=")
			buffer.WriteString(fmt.Sprintf("%+v", e.Fields[key]))
		}
	}
	buffer.WriteString("\n")
	if e.Errs != nil {
		buffer.WriteString("Caused by : ")
		for i, err := range e.Errs {
			if err == nil {
				continue
			}
			if i > 0 {
				buffer.WriteString("And\n")
			}
			buffer.WriteString(err.Error())
			buffer.WriteString("\n")
		}
	}
	return buffer.String()
}

func (e *EntryError) String() string {
	return e.Error()
}

//
//func (e *EntryError) Stack() []byte {
//	buf := bytes.Buffer{}
//
//	for _, frame := range e.StackFrames() {
//		buf.WriteString(frame.String())
//	}
//
//	return buf.Bytes()
//}
//
//func (e *EntryError) StackFrames() []StackFrame {
//	if e.frames == nil {
//		e.frames = make([]StackFrame, len(e.stack))
//		for i, pc := range e.stack {
//			e.frames[i] = NewStackFrame(pc)
//		}
//	}
//	return e.frames
//}
