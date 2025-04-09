package metrics

import "time"

// timer is a utility for timing code blocks and reporting metrics
type timer struct {
	start time.Time
	end   time.Time
	name  string
	tags  map[string]string
	rate  float64
}

// Start begins the timer
func (t *timer) Start() {
	t.start = time.Now()
}

// Done stops the timer and reports the elapsed time
func (t *timer) Done() {
	t.end = time.Now()
	milliseconds := float64(t.end.Sub(t.start).Nanoseconds()) / float64(time.Millisecond)
	TimeInMilliseconds(t.name, milliseconds, t.tags, t.rate)
}

// SetTags sets or updates the tags for this timer
func (t *timer) SetTags(tags map[string]string) {
	if t.tags == nil {
		t.tags = make(map[string]string, len(tags))
	}
	for k, v := range tags {
		t.tags[k] = v
	}
}

// DoneWithTags stops the timer, sets the tags, and reports the elapsed time
func (t *timer) DoneWithTags(tags map[string]string) {
	t.SetTags(tags)
	t.Done()
}
