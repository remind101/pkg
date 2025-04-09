// package logger is a package that provides a structured logger that's
// context.Context aware.
package logger

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level int

const (
	OFF Level = iota
	ERROR
	WARN
	INFO
	DEBUG
)

func ParseLevel(lvl string) Level {
	switch strings.ToLower(lvl) {
	case "off":
		return OFF
	case "error":
		return ERROR
	case "warn":
		return WARN
	case "info":
		return INFO
	case "debug":
		return DEBUG
	default:
		return DEBUG
	}
}

func FormatLevel(level Level) string {
	switch level {
	case OFF:
		return "off"
	case ERROR:
		return "error"
	case WARN:
		return "warn"
	case INFO:
		return "info"
	case DEBUG:
		return "debug"
	default:
		return "debug"
	}
}

// Logger represents a structured leveled logger.
type Logger interface {
	Debug(msg string, pairs ...interface{})
	Info(msg string, pairs ...interface{})
	Warn(msg string, pairs ...interface{})
	Error(msg string, pairs ...interface{})
	With(pairs ...interface{}) Logger
}

var DefaultLogLevel = INFO
var DefaultLogger Logger

func init() {
	// Initialize with zap logger
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapLogger, _ := config.Build()
	DefaultLogger = NewZapLogger(zapLogger.Sugar(), DefaultLogLevel)
}

// Modern implementation using zap
type zapLogger struct {
	Level
	*zap.SugaredLogger
}

// NewZapLogger creates a Logger implementation backed by zap
func NewZapLogger(l *zap.SugaredLogger, ll Level) Logger {
	return &zapLogger{
		Level:         ll,
		SugaredLogger: l,
	}
}

// With returns a new logger with the given key value pairs added to each log message.
func (l *zapLogger) With(pairs ...interface{}) Logger {
	return &zapLogger{
		Level:         l.Level,
		SugaredLogger: l.SugaredLogger.With(pairs...),
	}
}

func (l *zapLogger) Debug(msg string, pairs ...interface{}) {
	if DEBUG <= l.Level {
		l.SugaredLogger.Debugw(msg, pairs...)
	}
}

func (l *zapLogger) Info(msg string, pairs ...interface{}) {
	if INFO <= l.Level {
		l.SugaredLogger.Infow(msg, pairs...)
	}
}

func (l *zapLogger) Error(msg string, pairs ...interface{}) {
	if ERROR <= l.Level {
		l.SugaredLogger.Errorw(msg, pairs...)
	}
}

func (l *zapLogger) Warn(msg string, pairs ...interface{}) {
	if WARN <= l.Level {
		l.SugaredLogger.Warnw(msg, pairs...)
	}
}

// Legacy logger implementation for backward compatibility
type stdLogger struct {
	Level
	*log.Logger
	ctxPairs []interface{} // Contextual key value pairs that will be prepended to the log message.
}

// New wraps the log.Logger to implement the Logger interface.
func New(l *log.Logger, ll Level) Logger {
	return &stdLogger{
		Logger:   l,
		Level:    ll,
		ctxPairs: []interface{}{},
	}
}

// With returns a new logger with the given key value pairs added to each log message.
func (l *stdLogger) With(pairs ...interface{}) Logger {
	return &stdLogger{
		Logger:   l.Logger,
		Level:    l.Level,
		ctxPairs: append(l.ctxPairs, pairs...),
	}
}

func (l *stdLogger) Log(level Level, msg string, pairs ...interface{}) {
	if level <= l.Level {
		msg = "status=" + FormatLevel(level) + " " + msg
		m := l.message(pairs...)
		l.Println(msg, m)
	}
}

func (l *stdLogger) Debug(msg string, pairs ...interface{}) { l.Log(DEBUG, msg, pairs...) }
func (l *stdLogger) Info(msg string, pairs ...interface{})  { l.Log(INFO, msg, pairs...) }
func (l *stdLogger) Error(msg string, pairs ...interface{}) { l.Log(ERROR, msg, pairs...) }
func (l *stdLogger) Warn(msg string, pairs ...interface{})  { l.Log(WARN, msg, pairs...) }

func (l *stdLogger) message(pairs ...interface{}) string {
	pairs = append(l.ctxPairs, pairs...)

	if len(pairs) == 1 {
		return fmt.Sprintf("%v", pairs[0])
	}

	var parts []string

	for i := 0; i < len(pairs); i += 2 {
		// This conditional means that the pairs are uneven and we've
		// reached the end of iteration. We treat the last value as a
		// simple string message. Given an input pair as:
		//
		//      ["key", "value", "message"]
		//
		// The output will be:
		//
		//      key=value message
		if len(pairs) == i+1 {
			parts = append(parts, fmt.Sprintf("%v", pairs[i]))
		} else {
			parts = append(parts, fmt.Sprintf("%s=%v", pairs[i], pairs[i+1]))
		}
	}

	return strings.Join(parts, " ")
}

// WithLogger inserts a log.Logger into the provided context.
func WithLogger(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext returns a log.Logger from the context.
func FromContext(ctx context.Context) (Logger, bool) {
	l, ok := ctx.Value(loggerKey).(Logger)
	return l, ok
}

func Info(ctx context.Context, msg string, pairs ...interface{}) {
	withLogger(ctx, func(l Logger) {
		l.Info(msg, pairs...)
	})
}

func Debug(ctx context.Context, msg string, pairs ...interface{}) {
	withLogger(ctx, func(l Logger) {
		l.Debug(msg, pairs...)
	})
}

func Warn(ctx context.Context, msg string, pairs ...interface{}) {
	withLogger(ctx, func(l Logger) {
		l.Warn(msg, pairs...)
	})
}

func Error(ctx context.Context, msg string, pairs ...interface{}) {
	withLogger(ctx, func(l Logger) {
		l.Error(msg, pairs...)
	})
}

func withLogger(ctx context.Context, fn func(l Logger)) {
	if l, ok := FromContext(ctx); ok {
		fn(l)
	} else {
		fn(DefaultLogger)
	}
}

type key int

const (
	loggerKey key = iota
)
