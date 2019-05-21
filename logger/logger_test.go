package logger

import (
	"bytes"
	"log"
	"testing"

	"context"
)

func TestLogger(t *testing.T) {
	msg := "message"

	tests := []struct {
		in  []interface{}
		out string
	}{
		{[]interface{}{"key", "value"}, "status=info message key=value\n"},
		{[]interface{}{"this is a message"}, "status=info message this is a message\n"},
		{[]interface{}{"key", "value", "message"}, "status=info message key=value message\n"},
		{[]interface{}{"count", 1}, "status=info message count=1\n"},
		{[]interface{}{"b", 1, "a", 1}, "status=info message b=1 a=1\n"},
		{[]interface{}{}, "status=info message \n"},
	}

	for _, tt := range tests {
		out := testInfo(msg, tt.in...)
		if got, want := out, tt.out; got != want {
			t.Fatalf("Log => %q; want %q", got, want)
		}
	}
}

func testInfo(msg string, pairs ...interface{}) string {
	b := new(bytes.Buffer)
	l := New(log.New(b, "", 0), INFO)
	l.Info(msg, pairs...)
	return b.String()
}

// check that when set to a high level (WARN), a lower log (INFO) doesnt print
func TestLogLevel(t *testing.T) {

	b := new(bytes.Buffer)
	l := New(log.New(b, "", 0), ERROR)
	Info(WithLogger(context.Background(), l), "test info")
	if got, want := b.String(), ""; got != want {
		t.Fatalf("ontext Logger => %q; want %q", got, want)
	}
}

func TestWith(t *testing.T) {
	b := new(bytes.Buffer)
	l := New(log.New(b, "", 0), INFO)
	lw := l.With("request_id", "abc")
	lw.Info("message", "count", 1)

	if got, want := b.String(), "status=info message request_id=abc count=1\n"; got != want {
		t.Fatalf("With Logger => %q; want %q", got, want)
	}
}

func TestWithContextLogger(t *testing.T) {
	b := new(bytes.Buffer)
	l := New(log.New(b, "", 0), INFO)
	Info(WithLogger(context.Background(), l), "test")
	if got, want := b.String(), "status=info test \n"; got != want {
		t.Fatalf("Without Context Logger => %q; want %q", got, want)
	}
}

func TestWithoutContextLogger(t *testing.T) {
	origFallBackLogger := DefaultLogger
	defer func() { DefaultLogger = origFallBackLogger }()
	b := new(bytes.Buffer)
	DefaultLogger = New(log.New(b, "", 0), INFO)
	Info(context.Background(), "test")
	if got, want := b.String(), "status=info test \n"; got != want {
		t.Fatalf("Without Context Logger => %q; want %q", got, want)
	}
}
