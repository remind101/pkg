package svc_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/httpx/errors"
	"github.com/remind101/pkg/httpx/middleware"
	"github.com/remind101/pkg/reporter/mock"
	"github.com/remind101/pkg/svc"
)

type handlerTest struct {
	HandlerOpts svc.HandlerOpts
	HandlerFunc func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error
	StatusCode  int
	Body        string
	ErrFrame    string
}

func TestStandardHandler(t *testing.T) {
	tests := []handlerTest{
		{ // Test panic
			HandlerOpts: svc.HandlerOpts{
				ErrorHandler: middleware.JSONReportingErrorHandler,
			},
			HandlerFunc: func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
				panic("I panicked")
			},
			StatusCode: 500,
			Body:       `{"error":"I panicked"}` + "\n",
			ErrFrame:   "svc_test.go:33",
		},
		{ // Test panic with timeout handler
			HandlerOpts: svc.HandlerOpts{
				ErrorHandler:   middleware.JSONReportingErrorHandler,
				HandlerTimeout: 10 * time.Millisecond,
			},
			HandlerFunc: func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
				panic("I panicked")
			},
			StatusCode: 500,
			Body:       `{"error":"I panicked"}` + "\n",
			ErrFrame:   "svc_test.go:45",
		},
		{ // Test timeout with timeout handler
			HandlerOpts: svc.HandlerOpts{
				ErrorHandler:   middleware.JSONReportingErrorHandler,
				HandlerTimeout: 10 * time.Millisecond,
			},
			HandlerFunc: func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
				time.Sleep(20 * time.Millisecond)
				return nil
			},
			StatusCode: 503,
			Body:       `{"error":"http: handler timeout"}` + "\n",
			ErrFrame:   "timeout.go:94",
		},
	}

	for _, tt := range tests {
		runHandlerTest(tt, t)
	}
}

func runHandlerTest(tt handlerTest, t *testing.T) {
	t.Helper()
	rep := mock.NewReporter()
	r := httpx.NewRouter()
	r.Handle("/", httpx.HandlerFunc(tt.HandlerFunc))
	tt.HandlerOpts.Router = r
	tt.HandlerOpts.Reporter = rep
	h := svc.NewStandardHandler(tt.HandlerOpts)

	s := httptest.NewServer(h)
	defer s.Close()

	req, _ := http.NewRequest("GET", s.URL, nil)
	req.Header.Add("X-Request-ID", "abc")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := resp.StatusCode, tt.StatusCode; got != want {
		t.Errorf("got %d; expected %d", got, want)
	}

	if tt.ErrFrame != "" {
		if got, want := len(rep.Calls), 1; got != want {
			t.Error("expected error to be reported")
		}

		if _, ok := rep.Calls[0].Err.(*errors.Error); !ok {
			t.Fatal("expected error to be an *errors.Error")
		}

		e := rep.Calls[0].Err.(*errors.Error)
		frame := e.StackTrace()[0]

		if got, want := fmt.Sprintf("%v", frame), tt.ErrFrame; got != want {
			t.Errorf("got %s; expected %s", got, want)
		}
	} else {
		if got, want := len(rep.Calls), 0; got != want {
			t.Error("expected no error to be reported")
		}
	}
}
