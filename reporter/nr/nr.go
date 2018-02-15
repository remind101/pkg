package nr

import (
	"fmt"
	"strings"

	"github.com/remind101/newrelic"
	"github.com/remind101/pkg/reporter"
	"github.com/remind101/pkg/reporter/util"
	"context"
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

		if e, ok := err.(*reporter.Error); ok {
			exceptionType = util.ClassName(e.Err)

			for _, frame := range e.StackTrace() {
				stackTrace = append(stackTrace, fmt.Sprintf("%s:%d %n", frame, frame, frame))
			}

		}

		return tx.ReportError(exceptionType, errorMessage, strings.Join(stackTrace, stackFrameDelim), stackFrameDelim)
	}
	return nil
}
