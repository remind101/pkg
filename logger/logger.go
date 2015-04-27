// package logger is a package that provides a structured logger that's
// context.Context aware.
package logger

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/net/context"
)

// Logger represents a structured leveled logger.
type Logger interface {
	Debug(ctx context.Context, msg string, pairs ...interface{})
	Info(ctx context.Context, msg string, pairs ...interface{})
	Warn(ctx context.Context, msg string, pairs ...interface{})
	Error(ctx context.Context, msg string, pairs ...interface{})
	Crit(ctx context.Context, msg string, pairs ...interface{})
}

// logger is an implementation of the Logger interface backed by the stdlib's
// logging facility. This is a fairly naive implementation, and it's probably
// better to use something like https://github.com/inconshreveable/log15 which
// offers real structure logging.
type logger struct {
	*log.Logger
}

// New wraps the log.Logger to implement the Logger interface.
func New(l *log.Logger) Logger {
	return &logger{
		Logger: l,
	}
}

// Log logs the pairs in logfmt. It will treat consecutive arguments as a key
// value pair. Given the input:
func (l *logger) Log(msg string, pairs ...interface{}) {
	m := l.message(pairs...)
	l.Println(msg, m)
}

func (l *logger) Debug(ctx context.Context, msg string, pairs ...interface{}) { l.Log(msg, pairs...) }
func (l *logger) Info(ctx context.Context, msg string, pairs ...interface{})  { l.Log(msg, pairs...) }
func (l *logger) Warn(ctx context.Context, msg string, pairs ...interface{})  { l.Log(msg, pairs...) }
func (l *logger) Error(ctx context.Context, msg string, pairs ...interface{}) { l.Log(msg, pairs...) }
func (l *logger) Crit(ctx context.Context, msg string, pairs ...interface{})  { l.Log(msg, pairs...) }

func (l *logger) message(pairs ...interface{}) string {
	if len(pairs) == 1 {
		return fmt.Sprintf("%v", pairs[0])
	}

	var parts []string

	for i := 0; i < len(pairs); i += 2 {
		// This conditional means that the pairs are uneven and we've
		// reached the end of iteration. We treat the last value as a
		// simple string message. Given an input pair as:
		//
		//	["key", "value", "message"]
		//
		// The output will be:
		//
		//	key=value message
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
		l.Info(ctx, msg, pairs...)
	})
}

func Debug(ctx context.Context, msg string, pairs ...interface{}) {
	withLogger(ctx, func(l Logger) {
		l.Debug(ctx, msg, pairs...)
	})
}

func Warn(ctx context.Context, msg string, pairs ...interface{}) {
	withLogger(ctx, func(l Logger) {
		l.Warn(ctx, msg, pairs...)
	})
}

func Error(ctx context.Context, msg string, pairs ...interface{}) {
	withLogger(ctx, func(l Logger) {
		l.Error(ctx, msg, pairs...)
	})
}

func Crit(ctx context.Context, msg string, pairs ...interface{}) {
	withLogger(ctx, func(l Logger) {
		l.Crit(ctx, msg, pairs...)
	})
}

func withLogger(ctx context.Context, fn func(l Logger)) {
	if l, ok := FromContext(ctx); ok {
		fn(l)
	}
}

// StructuredLogger represents a structured logger that is not context.Context
// aware, like log15.
type StructuredLogger interface {
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
	Crit(string, ...interface{})
}

// wrappedLogger wraps a structured logger that doesn't take a context.Context
// as the first argument, like log15, as a Logger.
type wrappedLogger struct {
	logger StructuredLogger
}

// Wrap wraps a StructuredLogger as a Logger.
func Wrap(l StructuredLogger) Logger {
	return &wrappedLogger{l}
}

func (l *wrappedLogger) Info(ctx context.Context, msg string, pairs ...interface{}) {
	l.logger.Info(msg, pairs...)
}
func (l *wrappedLogger) Debug(ctx context.Context, msg string, pairs ...interface{}) {
	l.logger.Debug(msg, pairs...)
}
func (l *wrappedLogger) Warn(ctx context.Context, msg string, pairs ...interface{}) {
	l.logger.Warn(msg, pairs...)
}
func (l *wrappedLogger) Error(ctx context.Context, msg string, pairs ...interface{}) {
	l.logger.Error(msg, pairs...)
}
func (l *wrappedLogger) Crit(ctx context.Context, msg string, pairs ...interface{}) {
	l.logger.Crit(msg, pairs...)
}

type key int

const (
	loggerKey key = iota
)
