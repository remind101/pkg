package middleware

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/remind101/pkg/httpx"
)

type timeoutTest struct {
	Handler  httpx.Handler
	Duration time.Duration
	Err      error
	Code     int
	Body     string
	Panic    error
}

func TestTimeoutHandler(t *testing.T) {
	tests := []timeoutTest{
		{ // Success
			Handler: httpx.HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
				rw.WriteHeader(http.StatusOK)
				fmt.Fprintln(rw, "Hello")
				return nil
			}),
			Duration: 50 * time.Millisecond,
			Err:      nil,
			Code:     http.StatusOK,
			Body:     "Hello\n",
		},
		{ // Non 2xx Status Code
			Handler: httpx.HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
				rw.WriteHeader(http.StatusNotFound)
				return nil
			}),
			Duration: 50 * time.Millisecond,
			Code:     http.StatusNotFound,
			Err:      nil,
		},
		{ // Error
			Handler: httpx.HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
				return errors.New("boom")
			}),
			Duration: 50 * time.Millisecond,
			Err:      errors.New("boom"),
		},
		{ // Panic
			Handler: httpx.HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
				panic(errors.New("boom"))
			}),
			Duration: 50 * time.Millisecond,
			Panic:    errors.New("boom"),
		},
		{ // Timeout
			Handler: httpx.HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
				time.Sleep(100 * time.Millisecond)
				rw.WriteHeader(http.StatusOK)
				return nil
			}),
			Duration: 50 * time.Millisecond,
			Err:      ErrHandlerTimeout,
		},
	}

	for _, tt := range tests {
		runTimeoutTest(tt, t)
	}
}

func compareError(t *testing.T, got, want interface{}) {
	t.Helper()
	if got == nil && want != nil || got != nil && want == nil {
		t.Errorf("got: %#v; expected %#v", got, want)
	}
	if got == nil && want == nil {
		return
	}

	if g, w := got.(error).Error(), want.(error).Error(); g != w {
		t.Errorf("got: %#v; expected %#v", g, w)
	}
}

func runTimeoutTest(tt timeoutTest, t *testing.T) {
	defer func() {
		v := recover()
		compareError(t, v, tt.Panic)
	}()
	th := TimeoutHandler(tt.Handler, tt.Duration)
	ctx := context.Background()
	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()
	err := th.ServeHTTPContext(ctx, resp, req)

	compareError(t, err, tt.Err)

	if tt.Code > 0 {
		if got, want := resp.Result().StatusCode, tt.Code; got != want {
			t.Errorf("got: %#v; expected %#v", got, want)
		}
	}

	if tt.Body != "" {
		b, _ := ioutil.ReadAll(resp.Result().Body)
		if got, want := string(b), tt.Body; got != want {
			t.Errorf("got: %#v; expected %#v", got, want)
		}
	}

}
