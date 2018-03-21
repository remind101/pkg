package hb2

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"context"

	"github.com/remind101/pkg/errctx"
	"github.com/remind101/pkg/reporter"
)

func TestHb2IsReporter(t *testing.T) {
	var _ reporter.Reporter = NewReporter(Config{})
}

func TestHb2ReportsErrorContext(t *testing.T) {
	h := newFakeHoneybadgerHandler(t)
	ts := httptest.NewServer(h)
	defer ts.Close()
	r := NewReporter(Config{Endpoint: ts.URL})

	boom := errors.New("The Error")
	tests := []struct {
		name    string
		err     error
		context map[string]interface{}
		request *http.Request
		path    string
		want    map[string]interface{}
	}{
		{
			name: "error with context",
			err:  boom,
			context: map[string]interface{}{
				"lol": "wut",
			},
			path: "request.context",
			want: map[string]interface{}{
				"lol": "wut",
			},
		},
		{
			name: "error during request",
			err:  boom,
			context: map[string]interface{}{
				"request_id": "1234",
			},
			request: func() *http.Request {
				form := url.Values{}
				form.Add("param1", "param1value")
				req, _ := http.NewRequest("GET", "/api/foo", nil)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Forwarded-For", "127.0.0.1")
				req.Header.Set("Authorization", "Basic shouldnotseeit")
				req.Form = form
				return req
			}(),
			path: "request",
			want: map[string]interface{}{
				"cgi_data": map[string]interface{}{
					"HTTP_CONTENT_TYPE":    "application/json",
					"HTTP_X_FORWARDED_FOR": "127.0.0.1",
					"REQUEST_METHOD":       "GET",
				},
				"context": map[string]interface{}{
					"request_id": "1234",
				},
				"params": map[string]interface{}{
					"param1": []interface{}{
						"param1value",
					},
				},
				"url": "/api/foo",
			},
		},
	}

	for _, tt := range tests {
		ctx := reporter.WithReporter(context.Background(), r)
		for k, v := range tt.context {
			ctx = errctx.WithInfo(ctx, k, v)
		}
		ctx = errctx.WithRequest(ctx, tt.request)
		reporter.Report(ctx, tt.err)

		select {
		case v := <-h.LastRequestBodyChan:
			got, want := atPath(v, tt.path), tt.want
			if !reflect.DeepEqual(got, want) {
				jsonGot, _ := json.MarshalIndent(got, "", "  ")
				jsonWant, _ := json.MarshalIndent(want, "", "  ")
				t.Errorf("[%s]\n got %s\nwant %s", tt.name, jsonGot, jsonWant)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("did not report to HB in 3 seconds")
		}
	}

}

func newFakeHoneybadgerHandler(t *testing.T) *fakeHoneybadgerHandler {
	return &fakeHoneybadgerHandler{make(chan map[string]interface{}, 1), t}
}

type fakeHoneybadgerHandler struct {
	LastRequestBodyChan chan map[string]interface{}
	T                   *testing.T
}

func (h *fakeHoneybadgerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	v := map[string]interface{}{}
	err := decoder.Decode(&v)
	if err != nil {
		h.T.Fatalf("not a valid json in request to Honeybadger")
	}
	h.LastRequestBodyChan <- v
	w.WriteHeader(http.StatusCreated)
}

func atPath(v map[string]interface{}, path string) map[string]interface{} {
	for _, p := range strings.Split(path, ".") {
		if p != "" {
			v = v[p].(map[string]interface{})
		}
	}
	return v
}
