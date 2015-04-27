package logger

import (
	"log"
	"os"

	"golang.org/x/net/context"
)

func ExampleLogger_Log() {
	l := New(log.New(os.Stdout, "", 0))

	// Consecutive arguments after the message are treated as key value pairs.
	l.Info(context.Background(), "message", "key", "value")

	// Output:
	// message key=value
}
