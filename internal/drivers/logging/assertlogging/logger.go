package assertlogging

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alexZaicev/go-ftp-client/internal/drivers/logging"
)

// assertableLog is a log statement to be matched against an ExpectedLog instance.
type assertableLog struct {
	msg    string
	fields logging.Fields
}

// logLevel describes the log levels a log statement can utilize.
type logLevel string

const (
	errorLogLevel logLevel = "error"
	warnLogLevel  logLevel = "warn"
	infoLogLevel  logLevel = "info"
	debugLogLevel logLevel = "debug"
)

// newAssertableLog creates a new assertableLog instance.
func newAssertableLog(msg string, fields logging.Fields) *assertableLog {
	return &assertableLog{
		msg:    msg,
		fields: fields,
	}
}

func (l *assertableLog) String() string {
	var fields []string
	for fieldName, fieldValue := range l.fields {
		fields = append(fields, fmt.Sprintf("%s:%+v", fieldName, fieldValue))
	}
	sort.Strings(fields) // make output deteministic
	return fmt.Sprintf("message=%q, fields=map[%s]", l.msg, strings.Join(fields, " "))
}

// addLog adds the given assertable log at the given level to the root logger instance.
func (l *Logger) addLog(log *assertableLog, level logLevel) {
	if l.parent != nil {
		l.parent.addLog(log, level)
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()
	switch level {
	case errorLogLevel:
		l.errorLogs = append(l.errorLogs, log)

	case warnLogLevel:
		l.warnLogs = append(l.warnLogs, log)

	case infoLogLevel:
		l.infoLogs = append(l.infoLogs, log)

	case debugLogLevel:
		l.debugLogs = append(l.debugLogs, log)
	}
}

// Non-allocating compile time check to ensure the logging.Logger interface is implemented
// correctly.
var _ logging.Logger = &Logger{}

// Error creates a log statement at ERROR level.
func (l *Logger) Error(msg string) {
	l.addLog(newAssertableLog(msg, l.fields), errorLogLevel)
}

// Warn creates a log statement at WARN level.
func (l *Logger) Warn(msg string) {
	l.addLog(newAssertableLog(msg, l.fields), warnLogLevel)
}

// Info creates a log statement at INFO level.
func (l *Logger) Info(msg string) {
	l.addLog(newAssertableLog(msg, l.fields), infoLogLevel)
}

// Debug creates a log statement at DEBUG level.
func (l *Logger) Debug(msg string) {
	l.addLog(newAssertableLog(msg, l.fields), debugLogLevel)
}

// WithField adds a single log field to a log statement.
func (l *Logger) WithField(key string, value interface{}) logging.Logger {
	return l.WithFields(logging.Fields{key: value})
}

// WithError adds a log field with a key of "error" to a log statement.
func (l *Logger) WithError(err error) logging.Logger {
	return l.WithField(logging.ErrKey, err)
}

// WithFields adds multiple log fields to a log statement.
func (l *Logger) WithFields(fields logging.Fields) logging.Logger {
	child := &Logger{parent: l}

	for key, value := range l.fields {
		if _, ok := fields[key]; !ok {
			fields[key] = value
		}
	}

	child.fields = fields
	return child
}

// Flush flushes any pending log statements. This is a no-op as logs are stored in memory.
func (l *Logger) Flush() error {
	return nil
}
