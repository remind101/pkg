package reporter

import (
	"fmt"

	"github.com/remind101/pkg/logger"
	"golang.org/x/net/context"
)

// LogReporter is a Handler that logs the error to a log.Logger.
type LogReporter struct {
	// If true, the full stack trace will be printed to Stdout.
	PrintStack bool
}

func NewLogReporter() *LogReporter {
	return &LogReporter{}
}

// Report logs the error to the Logger.
func (h *LogReporter) Report(ctx context.Context, err error) error {
	logger.Error(ctx, "", "error", fmt.Sprintf(`"%v"`, err))

	if err, ok := err.(*Error); ok && h.PrintStack {
		for _, l := range err.Backtrace {
			fmt.Println("%s:%d", l.File, l.Line)
		}
	}

	return nil
}
