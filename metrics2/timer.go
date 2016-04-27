package metrics2

import "time"

type timer struct {
	start time.Time
	end   time.Time
	name  string
	tags  map[string]string
	rate  float64
}

func (t *timer) Start() {
	t.start = time.Now()
}

func (t *timer) Done() {
	t.end = time.Now()
	milliseconds := float64(t.end.Sub(t.start).Nanoseconds()) / float64(time.Millisecond)
	TimeInMilliseconds(t.name, milliseconds, t.tags, t.rate)
}

func (t *timer) SetTags(tags map[string]string) {
	if t.tags == nil {
		t.tags = make(map[string]string, 2)
	}
	for k, v := range tags {
		t.tags[k] = v
	}
}
