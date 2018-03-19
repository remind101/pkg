package svc_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/logger"
	"github.com/remind101/pkg/reporter"
	"github.com/remind101/pkg/svc"
)

func TestStandardHandler(t *testing.T) {
	buf := logToBuffer()
	rep := reporter.NewLogReporter()
	r := httpx.NewRouter()
	r.Handle("/panic", httpx.HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
		var m []string
		fmt.Println(m[1])
		return nil
	}))
	h := svc.NewStandardHandler(svc.HandlerOpts{
		Router:   r,
		Reporter: rep,
	})
	s := httptest.NewServer(h)
	defer s.Close()

	req, _ := http.NewRequest("GET", s.URL+"/panic", nil)
	req.Header.Add("X-Request-ID", "abc")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}

	if got, want := resp.StatusCode, 500; got != want {
		t.Errorf("got %d; expected %d", got, want)
	}

	if got, want := buf.String(), " request_id=abc error=\"runtime error: index out of range\" line=24 file=svc_test.go\n"; got != want {
		t.Errorf("got %s; expected %s", got, want)
	}
}

func logToBuffer() *bytes.Buffer {
	lvl := logger.ERROR
	var buf bytes.Buffer
	l := logger.New(log.New(&buf, "", 0), lvl)
	logger.DefaultLogger = l
	return &buf
}
