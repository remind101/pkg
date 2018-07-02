package nr

import (
	"fmt"
	"strings"

	"context"

	"github.com/remind101/newrelic"
	"github.com/remind101/pkg/reporter"
	"github.com/remind101/pkg/reporter/util"
)

// Ensure that Reporter implements the reporter.Reporter interface.
var _ reporter.Reporter = &Reporter{}

type Reporter struct{}

func NewReporter() *Reporter {
	return &Reporter{}
}

func (r *Reporter) ReportWithLevel(ctx context.Context, level string, err error) error {
	if tx, ok := newrelic.FromContext(ctx); ok {
		var (
			exceptionType   string
			errorMessage    string
			stackTrace      []string
			stackFrameDelim string
		)

		errorMessage = err.Error()
		stackFrameDelim = "\n"
		stackTrace = make([]string, 0)

		exceptionType = util.ClassName(err)
		if e, ok := err.(util.Causer); ok {
			exceptionType = util.ClassName(e.Cause())
		}

		if e, ok := err.(util.StackTracer); ok {
			for _, frame := range e.StackTrace() {
				stackTrace = append(stackTrace, fmt.Sprintf("%s:%d %n", frame, frame, frame))
			}
		}

		return tx.ReportError(exceptionType, errorMessage, strings.Join(stackTrace, stackFrameDelim), stackFrameDelim)
	}
	return nil
}

func (r *Reporter) Wait() {
	// Doesn't look like we can wait for New Relic to flush.
}
