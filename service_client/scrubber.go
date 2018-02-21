package service_client

import ()

// A Scrubber will process a string and remove PII making it safe for
// logging and metrics.
type Scrubber interface {
	Scrub(string) string
}

type NoopScrubber struct{}

func (s *NoopScrubber) Scrub(str string) string {
	return str
}
