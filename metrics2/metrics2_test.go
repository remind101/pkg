package metrics2

import (
	"reflect"
	"testing"
)

func must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func assertDeepEqual(t *testing.T, name string, got, want interface{}) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("[%s], got:\n\t%#v\nwant:\n\t%#v\n", name, got, want)
	}
}

func TestMetricTags(t *testing.T) {
	r := newFakeMetricsReporter()
	Reporter = r
	defer resetReporter()

	must(t, Count("testcount", 42, nil, 1.0))
	assertDeepEqual(t, "no app name and no tags", r.LastCountMetric, &intMetric{
		metric{"testcount", map[string]string{}, 1.0},
		42,
	})

	must(t, Count("testcount", 42, map[string]string{"foo": "bar"}, 1.0))
	assertDeepEqual(t, "no app name and some tags", r.LastCountMetric, &intMetric{
		metric{"testcount", map[string]string{"foo": "bar"}, 1.0},
		42,
	})

	SetAppName("testapp")
	defer resetDefaultTags()

	must(t, Count("testcount", 42, nil, 1.0))
	assertDeepEqual(t, "app name and no tags", r.LastCountMetric, &intMetric{
		metric{"testcount", map[string]string{"app": "testapp"}, 1.0},
		42,
	})

	must(t, Count("testcount", 42, map[string]string{"foo": "bar"}, 1.0))
	assertDeepEqual(t, "app name and some tags", r.LastCountMetric, &intMetric{
		metric{"testcount", map[string]string{"app": "testapp", "foo": "bar"}, 1.0},
		42,
	})
}

func TestTime(t *testing.T) {
	r := newFakeMetricsReporter()
	Reporter = r
	defer resetReporter()

	tm := Time("mytiming", map[string]string{"yolo": "bolo"}, 1.0)
	tm.SetTags(map[string]string{"foo": "bar"})
	tm.Done()

	last := r.LastTimeInMillisecondsMetric
	gotTags := last.Tags
	wantTags := map[string]string{"yolo": "bolo", "foo": "bar"}

	if !reflect.DeepEqual(gotTags, wantTags) {
		t.Errorf("got tags:\n\t%#v\nwanted these:\n\t%#v", gotTags, wantTags)
	}
}
