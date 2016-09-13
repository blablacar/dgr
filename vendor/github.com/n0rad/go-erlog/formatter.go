package erlog

import (
	"bytes"
	"fmt"
	"github.com/mgutz/ansi"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

var pathSkip int = 0

var reset = ansi.ColorCode("reset")

var fileColorNormal = ansi.ColorCode("cyan+b")
var fileColorFail = ansi.ColorCode("cyan")

var timeColorNormal = ansi.ColorFunc("blue+b")
var timeColorFail = ansi.ColorFunc("blue")

var lvlColorError = ansi.ColorCode("red+b")
var lvlColorWarn = ansi.ColorCode("yellow+b")
var lvlColorInfo = ansi.ColorCode("green")
var lvlColorDebug = ansi.ColorCode("magenta")
var lvlColorTrace = ansi.ColorCode("blue")
var lvlColorPanic = ansi.ColorCode(":red+h")

type ErlogWriterAppender struct {
	Out      io.Writer
	Level    logs.Level
	mu       sync.Mutex
	useColor bool
}

func init() {
	_, file, _, _ := runtime.Caller(0)
	paths := strings.Split(file, "/")
	for i := 0; i < len(paths); i++ {
		if paths[i] == "github.com" {
			if paths[i-1] == "vendor" {
				pathSkip = i - 2
				break
			}
			pathSkip = i + 2
			break
		}
	}

}

func NewErlogWriterAppender(writer io.Writer) (f *ErlogWriterAppender) {
	return &ErlogWriterAppender{
		Out:      writer,
		useColor: terminal.IsTerminal(int(os.Stdout.Fd())),
	}
}

func (f *ErlogWriterAppender) GetLevel() logs.Level {
	return f.Level
}

func (f *ErlogWriterAppender) SetLevel(level logs.Level) {
	f.Level = level
}

func (f *ErlogWriterAppender) Fire(event *LogEvent) {
	keys := f.prepareKeys(event.Fields)
	var now string
	if f.useColor {
		now = time.Now().Format("15:04:05") // TODO prepare format ?
	} else {
		now = time.Now().Format("2006-01-02 15:04:05") // TODO prepare format ?
	}
	level := f.textLevel(event.Level)

	paths := strings.SplitN(event.File, "/", pathSkip+1)

	packagePath := event.File
	if len(paths) > pathSkip {
		packagePath = paths[pathSkip]
	}

	b := &bytes.Buffer{}
	if f.useColor {
		fmt.Fprintf(b, "%s %s%-5s%s %s%30s:%-3d%s %s%-44s%s",
			f.timeColor(event.Level)(now),
			f.levelColor(event.Level),
			level,
			reset,
			f.fileColor(event.Level),
			f.reduceFilePath(packagePath, 30),
			event.Line,
			reset,
			f.textColor(event.Level),
			event.Message,
			reset)
		for _, k := range keys {
			v := event.Entry.Fields[k]
			fmt.Fprintf(b, " %s%s%s=%+v", lvlColorInfo, k, reset, v)
		}
	} else {
		fmt.Fprintf(b, "%s %-5s %30s:%-3d %-44s",
			now,
			level,
			f.reduceFilePath(packagePath, 30),
			event.Line,
			event.Message,
		)
		for _, k := range keys {
			v := event.Entry.Fields[k]
			fmt.Fprintf(b, " %s=%+v", k, v)
		}
	}
	b.WriteByte('\n')

	f.logError(b, event, event.Err)

	f.mu.Lock()
	defer f.mu.Unlock()
	f.Out.Write(b.Bytes())
}

func (f *ErlogWriterAppender) logError(b *bytes.Buffer, event *LogEvent, errors error) {
	if errors == nil {
		return
	}

	for err := errors; err != nil; {
		if e, ok := err.(*errs.EntryError); ok {
			if e == nil {
				err = nil
				continue
			}
			path, line := findFileAndName(e.Stack)
			paths := strings.SplitN(path, "/", pathSkip+1)

			packagePath := event.File
			if len(paths) > pathSkip {
				packagePath = paths[pathSkip]
			}

			if f.useColor {
				fmt.Fprintf(b, "               %s%30s:%-3d%s %s%-44s%s",
					f.fileColor(event.Level),
					f.reduceFilePath(packagePath, 30),
					line,
					reset,
					f.textColor(event.Level),
					e.Message,
					reset)

				keys := f.prepareKeys(e.Fields)
				for _, k := range keys {
					v := e.Fields[k]
					fmt.Fprintf(b, " %s%s%s=%+v", lvlColorInfo, k, reset, v)
				}
			} else {
				fmt.Fprintf(b, "                          %30s:%-3d %-44s",
					f.reduceFilePath(packagePath, 30),
					line,
					e.Message,
				)

				keys := f.prepareKeys(e.Fields)
				for _, k := range keys {
					v := e.Fields[k]
					fmt.Fprintf(b, " %s=%+v", k, v)
				}
			}

			b.WriteByte('\n')

			if len(e.Errs) > 1 {
				for i, ee := range e.Errs {
					if i == 0 {
						b.WriteString("Caused By\n")
					} else {
						b.WriteString("And\n")
					}
					f.logError(b, event, ee)
					b.WriteString("\n")
				}
				err = nil
			} else if len(e.Errs) == 1 {
				err = e.Errs[0]
			} else {
				err = nil
			}
		} else {
			if f.useColor {
				fmt.Fprintf(b, "                                                  %s%s%s\n",
					f.textColor(event.Level),
					err.Error(),
					reset)
				err = nil
			} else {
				fmt.Fprintf(b, "                                                             %s\n",
					err.Error(),
				)
				err = nil
			}
		}
	}
}

func findFileAndName(ptrs []uintptr) (string, int) {
	var frame errs.StackFrame
	for i := 1; i < len(ptrs); i++ {
		frame = errs.NewStackFrame(ptrs[i])
		if !strings.Contains(frame.Package, "n0rad/go-erlog") {
			break
		}
		if strings.Contains(frame.Package, "n0rad/go-erlog/examples") {
			// TODO what to do with that ?
			break
		}
	}
	return frame.File, frame.LineNumber
}

func (f *ErlogWriterAppender) reduceFilePath(path string, max int) string {
	if len(path) <= max {
		return path
	}

	split := strings.Split(path, "/")
	splitlen := len(split)
	reducedSize := len(path)
	var buffer bytes.Buffer
	for i, e := range split {
		if len(e) > 0 && reducedSize > max && i+1 < splitlen {
			buffer.WriteByte(e[0])
			reducedSize -= len(e) - 1
		} else {
			buffer.WriteString(e)
		}
		if i+1 < splitlen {
			buffer.WriteByte('/')
		}
	}
	return buffer.String()
}

func (f *ErlogWriterAppender) prepareKeys(fields data.Fields) []string {
	var keys []string = make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (f *ErlogWriterAppender) textLevel(level logs.Level) string {
	levelText := strings.ToUpper(level.String())
	switch level {
	case logs.INFO:
	case logs.WARN:
		levelText = levelText[0:4]
	default:
		levelText = levelText[0:5]
	}
	return levelText
}

func (f *ErlogWriterAppender) fileColor(level logs.Level) string {
	switch level {
	case logs.DEBUG, logs.INFO, logs.TRACE:
		return fileColorFail
	default:
		return fileColorNormal
	}
}

func (f *ErlogWriterAppender) textColor(level logs.Level) string {
	switch level {
	case logs.WARN:
		return lvlColorWarn
	case logs.ERROR, logs.FATAL, logs.PANIC:
		return lvlColorError
	default:
		return ""
	}
}

func (f *ErlogWriterAppender) timeColor(level logs.Level) func(string) string {
	switch level {
	case logs.DEBUG, logs.INFO, logs.TRACE:
		return timeColorFail
	default:
		return timeColorNormal
	}
}

func (f *ErlogWriterAppender) levelColor(level logs.Level) string {
	switch level {
	case logs.TRACE:
		return lvlColorTrace
	case logs.DEBUG:
		return lvlColorDebug
	case logs.WARN:
		return lvlColorWarn
	case logs.ERROR:
		return lvlColorError
	case logs.FATAL, logs.PANIC:
		return lvlColorPanic
	default:
		return lvlColorInfo
	}
}
