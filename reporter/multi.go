package reporter

import "context"

// MultiReporter is an implementation of the Reporter interface that reports the
// error to multiple Reporters. If any of the individual error reporters returns
// an error, a MutliError will be returned.
type MultiReporter []Reporter

func (r MultiReporter) ReportWithLevel(ctx context.Context, level string, err error) error {
	var errors []error

	for _, reporter := range r {
		if err2 := reporter.ReportWithLevel(ctx, level, err); err2 != nil {
			errors = append(errors, err2)
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return &MultiError{Errors: errors}
}

func (r MultiReporter) Flush() {
	for _, reporter := range r {
		if f, ok := reporter.(flusher); ok {
			f.Flush()
		}
	}
}
