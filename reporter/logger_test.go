package reporter

import (
	"bytes"
	"log"
	"testing"

	"context"

	"github.com/pkg/errors"
	"github.com/remind101/pkg/logger"
)

func TestLogReporter(t *testing.T) {
	tests := []struct {
		err error
		out string
	}{
		{errBoom, "request_id=1234 status=error  error=\"boom\" line=0 file=unknown\n"},
		{errors.WithStack(errBoom), "request_id=1234 status=error  error=\"boom\" line=20 file=logger_test.go\n"},
	}

	for i, tt := range tests {
		b := new(bytes.Buffer)
		l := logger.New(log.New(b, "request_id=1234 ", 0), logger.INFO)
		h := &LogReporter{}

		ctx := logger.WithLogger(context.Background(), l)
		if err := h.ReportWithLevel(ctx, "error", tt.err); err != nil {
			t.Fatal(err)
		}

		if got, want := b.String(), tt.out; got != want {
			t.Fatalf("#%d: Output => %s; want %s", i, got, want)
		}
	}
}
