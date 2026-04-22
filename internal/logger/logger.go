package logger

import (
    "io"
    "log"
    "net"
    "os"
    "strings"
    "syscall"
)

type FilteredLogger struct {
    inner *log.Logger
    debug bool
    level int
}

const (
    LogLevelError = iota
    LogLevelWarn
    LogLevelInfo
    LogLevelDebug
)

func NewFilteredLogger(debug bool, logLevel string) *FilteredLogger {
    level := LogLevelInfo

    switch strings.ToLower(logLevel) {
    case "error":
        level = LogLevelError
    case "warn", "warning":
        level = LogLevelWarn
    case "info":
        level = LogLevelInfo
    case "debug":
        level = LogLevelDebug
        debug = true
    }

    return &FilteredLogger{
        inner: log.New(os.Stdout, "[vless] ", log.LstdFlags),
        debug: debug,
        level: level,
    }
}

func (l *FilteredLogger) Errorf(format string, args ...interface{}) {
    if l.level < LogLevelError {
        return
    }
    l.inner.Printf("[E]: "+format, args...)
}

func (l *FilteredLogger) Warnf(format string, args ...interface{}) {
    if l.level >= LogLevelWarn {
        l.inner.Printf("[W]: "+format, args...)
    }
}

func (l *FilteredLogger) Infof(format string, args ...interface{}) {
    if l.level >= LogLevelInfo {
        l.inner.Printf("[I]: "+format, args...)
    }
}

func (l *FilteredLogger) Info(msg string) {
    if l.level >= LogLevelInfo {
        l.inner.Printf("[I]: %s", msg)
    }
}

func (l *FilteredLogger) Debugf(format string, args ...interface{}) {
    if l.level >= LogLevelDebug {
        l.inner.Printf("[D]: "+format, args...)
    }
}

func isNoisyError(msg string) bool {
    noisyPatterns := []string{
        io.EOF.Error(),
        "broken pipe",
        "connection reset by peer",
        "use of closed network connection",
        syscall.EPIPE.Error(),
        net.ErrClosed.Error(),
        "i/o timeout",
        "context canceled",
    }

    for _, pattern := range noisyPatterns {
        if strings.Contains(msg, pattern) {
            return true
        }
    }
    return false
}
