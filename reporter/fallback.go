package reporter

import "context"

type FallbackReporter struct {
	// The first reporter to call.
	Reporter Reporter

	// This reporter will be used to report an error if the first Reporter
	// fails for some reason.
	Fallback Reporter
}

func (r *FallbackReporter) ReportWithLevel(ctx context.Context, level string, err error) error {
	if err2 := r.Reporter.ReportWithLevel(ctx, level, err); err2 != nil {
		r.Fallback.ReportWithLevel(ctx, level, err2)
		return err2
	}

	return nil
}

func (r *FallbackReporter) Wait() {
	r.Reporter.Flush()
	r.Fallback.Flush()
}
