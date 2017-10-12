package logger

import (
	"bytes"
	"log"
	"testing"

	"golang.org/x/net/context"
)

func TestLogger(t *testing.T) {
	msg := "message"

	tests := []struct {
		in  []interface{}
		out string
	}{
		{[]interface{}{"key", "value"}, "message key=value\n"},
		{[]interface{}{"this is a message"}, "message this is a message\n"},
		{[]interface{}{"key", "value", "message"}, "message key=value message\n"},
		{[]interface{}{"count", 1}, "message count=1\n"},
		{[]interface{}{"b", 1, "a", 1}, "message b=1 a=1\n"},
		{[]interface{}{}, "message \n"},
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

func TestWithContextLogger(t *testing.T) {
	b := new(bytes.Buffer)
	l := New(log.New(b, "", 0), INFO)
	Info(WithLogger(context.Background(), l), "test")
	if got, want := b.String(), "test \n"; got != want {
		t.Fatalf("Without Context Logger => %q; want %q", got, want)
	}
}

func TestWithoutContextLogger(t *testing.T) {
	origFallBackLogger := DefaultLogger
	defer func() { DefaultLogger = origFallBackLogger }()
	b := new(bytes.Buffer)
	DefaultLogger = New(log.New(b, "", 0), INFO)
	Info(context.Background(), "test")
	if got, want := b.String(), "test \n"; got != want {
		t.Fatalf("Without Context Logger => %q; want %q", got, want)
	}
}
