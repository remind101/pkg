package reporter

import (
	"fmt"

	"context"

	"github.com/pkg/errors"
	"github.com/remind101/pkg/logger"
)

// LogReporter is a Handler that logs the error to a log.Logger.
type LogReporter struct{}

func NewLogReporter() *LogReporter {
	return &LogReporter{}
}

// Report logs the error to the Logger.
func (h *LogReporter) ReportWithLevel(ctx context.Context, level string, err error) error {
	var file, line string
	var stack errors.StackTrace

	if err_with_stack, ok := err.(stackTracer); ok {
		stack = err_with_stack.StackTrace()
	}
	if stack != nil && len(stack) > 0 {
		file = fmt.Sprintf("%s", stack[0])
		line = fmt.Sprintf("%d", stack[0])
	} else {
		file = "unknown"
		line = "0"
	}

	if level == "debug" {
		logger.Debug(ctx, "", "debug", fmt.Sprintf(`"%v"`, err), "line", line, "file", file)
	} else if level == "info" {
		logger.Info(ctx, "", "info", fmt.Sprintf(`"%v"`, err), "line", line, "file", file)
	} else {
		logger.Error(ctx, "", level, fmt.Sprintf(`"%v"`, err), "line", line, "file", file)
	}
	return nil
}
