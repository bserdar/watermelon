package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bserdar/watermelon/server"
)

// Logging is the factory for host-specific loggers. Configure this
// with the logdir, so the loggers for each host can create separate
// log files under that dir.
type Logging struct {
	Logdir string
}

// Logger is the host-specific logger
type Logger struct {
	sync.Mutex

	Host      *server.Host
	LogStdout bool
	OutFile   string
}

// Formatter is the log formatter
var Formatter = func(now time.Time, h *server.Host, msg string) string {
	return fmt.Sprintf("%s [%s] %s\n", now.Format("20060102T15:04:05.0000"), h.ID, msg)
}

// GetLogDir returns the new logdir using the logbase, pkgname, and current time
func GetLogDir(logbase, pkgName string) string {
	return filepath.Join(logbase, fmt.Sprintf("%s-%s", pkgName, time.Now().Format("200601021504050000")))
}

// New creates a new logger for the host, with the given logging
// directory. This will create a new file for each host under that
// directory. If stdout is true, the logs are also printed to stdout
func (l Logging) New(host *server.Host, stdout bool) server.Logger {
	lg := Logger{Host: host, OutFile: filepath.Join(l.Logdir, string(host.ID)), LogStdout: stdout}
	return &lg
}

// Print a log message
func (l *Logger) Print(v ...interface{}) {
	l.queue(fmt.Sprint(v...))
}

// Printf a log message
func (l *Logger) Printf(format string, v ...interface{}) {
	l.queue(fmt.Sprintf(format, v...))
}

func (l *Logger) queue(msg string) {
	formatted := Formatter(time.Now(), l.Host, msg)
	if len(formatted) > 0 {
		if l.LogStdout {
			log.Infof("[%s] %s", l.Host.ID, formatted)
		}
		l.Lock()
		defer l.Unlock()

		f, err := os.OpenFile(l.OutFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			wd, _ := os.Getwd()
			panic(fmt.Sprintf("Cannot open log file: %v, cwd: %s", err, wd))
		}
		defer f.Close()
		f.WriteString(formatted)
	}
}
