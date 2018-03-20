package reporter

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"context"

	"github.com/remind101/pkg/errctx"
)

var errBoom = errors.New("boom")

func TestReport(t *testing.T) {
	r := ReporterFunc(func(ctx context.Context, level string, err error) error {
		e := err.(*errctx.Error)

		if e.Request.Header.Get("Content-Type") != "application/json" {
			t.Fatal("request information not set")
		}

		stack := e.StackTrace()
		var method string
		if stack != nil && len(stack) > 0 {
			method = fmt.Sprintf("%n", stack[0])
		}

		if got, want := method, "TestReport"; got != want {
			t.Fatalf("expected the first stacktrace method to be %v, got %v", want, got)
		}

		return nil
	})
	ctx := WithReporter(context.Background(), r)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Content-Type", "application/json")
	ctx = AddRequest(ctx, req)

	if err := ReportWithLevel(ctx, "error", errBoom); err != nil {
		t.Fatal(err)
	}
}

func TestReportWithLevel(t *testing.T) {
	var calledWithLevel string

	r := ReporterFunc(func(ctx context.Context, level string, err error) error {
		calledWithLevel = level
		return nil
	})
	ctx := WithReporter(context.Background(), r)
	err := ReportWithLevel(ctx, "warning", errBoom)

	if err != nil {
		t.Fatalf("unexpected error happened %v", err)
	}

	if got, want := calledWithLevel, "warning"; got != want {
		t.Fatalf("expected reporter to have been called with level %v, got %v", want, got)
	}
}

func TestReportWithNoReporterInContext(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Expected panic due to context without reporter, got no panic")
		}
	}()
	ctx := context.Background() // no reporter
	Report(ctx, errBoom)
}

func TestReportWithSensitiveData(t *testing.T) {
	r := ReporterFunc(func(ctx context.Context, level string, err error) error {
		e := err.(*errctx.Error)

		if e.Request.URL.Scheme != "http" {
			t.Fatalf("expected request.URL.Scheme to be \"http\", got: %v", e.Request.URL.Scheme)
		}

		if e.Request.URL.User != nil {
			t.Fatal("expected request.User to have been removed by the reporter")
		}

		if e.Request.URL.Host != "remind.com:80" {
			t.Fatalf("expected request.URL.Host to be \"remind.com:80\", got: %v", e.Request.URL.Host)
		}

		if e.Request.URL.Path != "/docs" {
			t.Fatalf("expected request.URL.Host to be \"/docs\", got: %v", e.Request.URL.Path)
		}

		if e.Request.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected request.Header[\"Content-type\"] to be \"application/json\", got: %v", e.Request.Header.Get("Content-Type"))
		}

		if e.Request.Header.Get("Authorization") != "" {
			t.Fatal("expected request.headers.Authorization to have been removed by the reporter")
		}

		if e.Request.Header.Get("Cookie") != "" {
			t.Fatal("expected request.headers.Cookie to have been removed by the reporter")
		}

		if len(e.Request.Cookies()) != 0 {
			t.Fatal("expected request.Cookies to have been removed by the reporter")
		}

		return nil
	})

	ctx := WithReporter(context.Background(), r)
	req, _ := http.NewRequest("GET", "http://user:pass@remind.com:80/docs", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "this-is-a-secret")
	req.Header.Set("Cookie", "r101_auth_token=this-is-sensitive")
	ctx = AddRequest(ctx, req)

	if err := ReportWithLevel(ctx, "error", errBoom); err != nil {
		t.Fatal(err)
	}
}

func TestReportWithFormData(t *testing.T) {
	r := ReporterFunc(func(ctx context.Context, level string, err error) error {
		e := err.(*errctx.Error)

		if e.Request.Form.Get("key") != "foo" {
			t.Fatalf("expected request.Form[\"key\"] to be \"foo\", got: %v", e.Request.Form.Get("key"))
		}

		if e.Request.Form.Get("username") != "admin" {
			t.Fatalf("expected request.Form[\"username\"] to be \"admin\", got: %v", e.Request.Form.Get("username"))
		}

		if e.Request.Form.Get("password") != "" {
			t.Fatal("expected request.Form[\"password\"] to have been removed by the reporter")
		}

		return nil
	})
	ctx := WithReporter(context.Background(), r)

	req, _ := http.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Form = url.Values{}
	req.Form.Add("key", "foo")
	req.Form.Add("username", "admin")
	req.Form.Add("password", "this-is-a-secret")
	ctx = AddRequest(ctx, req)

	if err := ReportWithLevel(ctx, "error", errBoom); err != nil {
		t.Fatal(err)
	}
}

func TestMonitor(t *testing.T) {
	ensureRepanicked := func() {
		if v := recover(); v == nil {
			t.Errorf("Must have panicked after reporting!")
		}
	}
	var reportedError error
	r := ReporterFunc(func(ctx context.Context, level string, err error) error {
		reportedError = err
		return nil
	})
	ctx := WithReporter(context.Background(), r)

	done := make(chan interface{})
	go func() {
		defer close(done)
		defer ensureRepanicked()
		defer Monitor(ctx)
		panic("oh noes!")
	}()
	<-done
	if reportedError == nil || reportedError.Error() != "panic: oh noes!" {
		t.Errorf("expected panic 'oh noes!' to be reported, got %#v", reportedError)
	}
}
