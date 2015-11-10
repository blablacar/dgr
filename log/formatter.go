package log

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/mgutz/ansi"
	"runtime"
	"sort"
	"strings"
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
var lvlColorPanic = ansi.ColorCode(":red+h")

type BlaFormatter struct {
}

func init() {
	_, file, _, _ := runtime.Caller(0)
	paths := strings.Split(file, "/")
	for i := 0; i < len(paths); i++ {
		if paths[i] == "github.com" {
			pathSkip = i + 2
			break
		}
	}
}

func (f *BlaFormatter) Format(entry *log.Entry) ([]byte, error) {
	file, line := f.findFileAndLine()
	f.prefixFieldClashes(entry.Data)
	keys := f.prepareKeys(entry)
	time := time.Now().Format("15:04:05")
	level := f.level(entry.Level)

	//	isColored := isTerminal && (runtime.GOOS != "windows")

	paths := strings.SplitN(file, "/", pathSkip + 1)

	b := &bytes.Buffer{}
	fmt.Fprintf(b, "%s %s%-5s%s %s%30s:%-3d%s %s%-44s%s",
		f.timeColor(entry.Level)(time),
		f.levelColor(entry.Level),
		level,
		reset,
		f.fileColor(entry.Level),
		f.reduceFilePath(paths[pathSkip], 30),
		line,
		reset,
		f.textColor(entry.Level),
		entry.Message,
		reset)
	for _, k := range keys {
		v := entry.Data[k]
		fmt.Fprintf(b, " %s%s%s=%+v", lvlColorInfo, k, reset, v)
	}
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *BlaFormatter) reduceFilePath(path string, max int) string {
	if len(path) <= max {
		return path
	}

	split := strings.Split(path, "/")
	splitlen := len(split)
	reducedSize := len(path)
	var buffer bytes.Buffer;
	for i, e := range split {
		if (reducedSize > max && i + 1 < splitlen) {
			buffer.WriteByte(e[0])
			reducedSize -= len(e) - 1
		} else {
			buffer.WriteString(e)
		}
		if (i + 1 < splitlen) {
			buffer.WriteByte('/')
		}
	}
	return buffer.String()
}

func (f *BlaFormatter) findFileAndLine() (string, int) {
	var file string
	var line int
	for i := 5;; i++ {
		_, file, line, _ = runtime.Caller(i)
		if !strings.Contains(file, "Sirupsen/logrus") {
			break
		}
	}
	return file, line
}

func (f *BlaFormatter) prepareKeys(entry *log.Entry) []string {
	var keys []string = make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (f *BlaFormatter) level(level log.Level) string {
	levelText := strings.ToUpper(level.String())
	switch level {
	case log.InfoLevel:
	case log.WarnLevel:
		levelText = levelText[0:4]
	default:
		levelText = levelText[0:5]
	}
	return levelText
}

func (f *BlaFormatter) fileColor(level log.Level) string {
	switch level {
	case log.DebugLevel, log.InfoLevel:
		return fileColorFail
	default:
		return fileColorNormal
	}
}

func (f *BlaFormatter) textColor(level log.Level) string {
	switch level {
	case log.WarnLevel:
		return lvlColorWarn
	case log.ErrorLevel, log.FatalLevel, log.PanicLevel:
		return lvlColorError
	default:
		return ""
	}
}

func (f *BlaFormatter) timeColor(level log.Level) func(string) string {
	switch level {
	case log.DebugLevel, log.InfoLevel:
		return timeColorFail
	default:
		return timeColorNormal
	}
}

func (f *BlaFormatter) levelColor(level log.Level) string {
	switch level {
	case log.DebugLevel:
		return lvlColorDebug
	case log.WarnLevel:
		return lvlColorWarn
	case log.ErrorLevel:
		return lvlColorError
	case log.FatalLevel, log.PanicLevel:
		return lvlColorPanic
	default:
		return lvlColorInfo
	}
}

func needsQuoting(text string) bool {
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '-' || ch == '.') {
			return false
		}
	}
	return true
}

func (f *BlaFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {

	b.WriteString(key)
	b.WriteByte('=')

	switch value := value.(type) {
	case string:
		if needsQuoting(value) {
			b.WriteString(value)
		} else {
			fmt.Fprintf(b, "%q", value)
		}
	case error:
		errmsg := value.Error()
		if needsQuoting(errmsg) {
			b.WriteString(errmsg)
		} else {
			fmt.Fprintf(b, "%q", value)
		}
	default:
		fmt.Fprint(b, value)
	}

	b.WriteByte(' ')
}

func (f *BlaFormatter) prefixFieldClashes(data log.Fields) {
	_, ok := data["time"]
	if ok {
		data["fields.time"] = data["time"]
	}

	_, ok = data["msg"]
	if ok {
		data["fields.msg"] = data["msg"]
	}

	_, ok = data["level"]
	if ok {
		data["fields.level"] = data["level"]
	}
}
