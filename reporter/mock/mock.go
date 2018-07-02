package mock

import "context"

type Reporter struct {
	Calls []Params
}

type Params struct {
	Ctx   context.Context
	Level string
	Err   error
}

func NewReporter() *Reporter {
	return &Reporter{
		Calls: make([]Params, 0),
	}
}

func (r *Reporter) ReportWithLevel(ctx context.Context, level string, err error) error {
	r.Calls = append(r.Calls, Params{ctx, level, err})
	return nil
}

func (r *Reporter) Flush() {
	// Nothing to do.
}
