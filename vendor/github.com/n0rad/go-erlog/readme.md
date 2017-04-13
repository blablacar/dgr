# Go erlog

*this is experimental*

In a fusion of logrus and go-errors logic with improvement to provide:
- a logging api supporting fields and stacktrace
- a logging implementation supporting multiple appenders
- log level by appenders
- trace log level
- fields to errors
- stacktrace to errors
- errors to log transformation without losing fields or stackstrace

# install

```shell
go get github.com/n0rad/go-erlog
```

# usage

basic :
```go
package main

import (
	"github.com/n0rad/go-erlog/log" // the api
	_ "github.com/n0rad/go-erlog/register" // use erlog implementation, with default appender (colored to stderr)
)

func main() {
	log.SetLevel(log.TRACE) // default is INFO

	log.Trace("I'm trace")
	log.Debug("I'm debug")
	log.Info("I'm info")
	log.Warn("I'm warn")
	log.Error("I'm error")

	func() {
		defer func() { recover() }()
		func() { log.Panic("I'm panic") }()
	}()

	log.Fatal("I'm fatal")
}
```

will produce

![basic](https://raw.githubusercontent.com/n0rad/go-erlog/master/docs/basic.png)

advanced :

