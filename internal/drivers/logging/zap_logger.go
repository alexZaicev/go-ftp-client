package logging

import (
	"errors"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log output keys that are only used in the Zap implementation. These are based on the standardized
// keys listed in https://pages.github.hpe.com/cloud/storage-design/docs/logging.html#log-fields.
const (
	callerKey     = "caller"
	levelKey      = "level"
	messageKey    = "message"
	nameKey       = "name"
	stacktraceKey = "stacktrace"
	timestampKey  = "timestamp"
)

// ZapJSONLogger is an implementation of the logging repository that
// uses zap's sugared logger.
type ZapJSONLogger struct {
	logger *zap.Logger
}

// NewZapJSONLogger creates a zap based logger that implements to Logger repository defined
// in this package. The logger should be flushed before the application exits.
func NewZapJSONLogger(logLevel string, hooks ...func(zapcore.Entry) error) (*ZapJSONLogger, error) {
	zapConfig, err := stdJSONLoggerConfig(logLevel)
	if err != nil {
		return nil, err
	}

	return zapLoggerFromConfig(zapConfig, hooks...)
}

// Error logs an error level message. Logs at this level implicitly add a stacktrace field.
func (z *ZapJSONLogger) Error(msg string) {
	z.logger.Error(msg)
}

// Warn logs an warning level message.
func (z *ZapJSONLogger) Warn(msg string) {
	z.logger.Warn(msg)
}

// Info logs an info level message.
func (z *ZapJSONLogger) Info(msg string) {
	z.logger.Info(msg)
}

// Debug logs a debug level message
func (z *ZapJSONLogger) Debug(msg string) {
	z.logger.Debug(msg)
}

// WithFields returns a new logger with the specified key-value pairs attached for
// subsequent logging operations.
// This function returns a repositories logger interface rather than the explicit
// ZapJSONLogger to allow it to satisfy the Logger interface
func (z *ZapJSONLogger) WithFields(fields Fields) Logger {
	fieldList := make([]zap.Field, len(fields))
	i := 0
	for key, value := range fields {
		fieldList[i] = zap.Any(key, value)
		i++
	}

	return &ZapJSONLogger{
		z.logger.With(fieldList...),
	}
}

// WithField returns a new logger with the specified key-value pair attached for
// subsequent logging operations.
// This function returns a repositories logger interface rather than the explicit
// ZapJSONLogger to allow it to satisfy the Logger interface
func (z *ZapJSONLogger) WithField(key string, value interface{}) Logger {
	return &ZapJSONLogger{
		z.logger.With(zap.Any(key, value)),
	}
}

// WithError provides a wrapper around WithField to add an error field to the logger,
// ensuring consistency of error message keys. It will also unwrap the error, unlike a
// normal WithField call.
func (z *ZapJSONLogger) WithError(err error) Logger {
	unwrapper := unwrapInfoExtractor(1000) // nolint:gomnd // arbitrary exit condition to avoid infinite loop
	msg := unwrapper(err)
	return z.WithField(ErrKey, msg)
}

// Flush syncs that zap logger.
func (z *ZapJSONLogger) Flush() error {
	return z.logger.Sync()
}

func zapLoggerFromConfig(config zap.Config, hooks ...func(zapcore.Entry) error) (*ZapJSONLogger, error) {
	logHooks := zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.RegisterHooks(core, hooks...)
	})

	logger, err := config.Build(
		// ignore the logger itself when providing caller
		zap.AddCallerSkip(1),
		logHooks,
	)
	if err != nil {
		return nil, err
	}

	return &ZapJSONLogger{
		logger: logger,
	}, nil
}

func stdJSONLoggerConfig(logLevel string) (zap.Config, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		return zap.Config{}, err
	}

	return zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        timestampKey,
			LevelKey:       levelKey,
			MessageKey:     messageKey,
			NameKey:        nameKey,
			StacktraceKey:  stacktraceKey,
			LineEnding:     "\n",
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.FullCallerEncoder,
		},
	}, nil
}

// UnwrapInfoExtractor creates an ErrInfoExtractor function that unwraps an error
// to the specified depth, combining all the messages together into one string.
func unwrapInfoExtractor(maxDepth int) ErrInfoExtractor {
	return func(err error) string {
		if err == nil {
			return ""
		}

		builder := strings.Builder{}
		builder.WriteString(err.Error())

		for i := 1; i < maxDepth; i++ {
			err = errors.Unwrap(err)
			if err == nil {
				return builder.String()
			}

			builder.WriteString(": " + err.Error())
		}

		return builder.String()
	}
}
