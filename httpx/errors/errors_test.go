package errors_test

import (
	"context"
	gerrors "errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/remind101/pkg/httpx/errors"
)

var errBoom = gerrors.New("boom")

func TestNew(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Content-Type", "application/json")
	ctx := errors.WithRequest(context.Background(), req)
	ctx = errors.WithInfo(ctx, "foo", "bar")
	e := errors.New(ctx, errBoom, 0)
	r := e.Request()

	if r.Header.Get("Content-Type") != "application/json" {
		t.Fatal("request information not set")
	}

	if v := e.ContextData()["foo"]; !reflect.DeepEqual(v, "bar") {
		t.Fatal("expected contextual information to be set")
	}

	stack := e.StackTrace()
	var method string
	if stack != nil && len(stack) > 0 {
		method = fmt.Sprintf("%n", stack[0])
	}

	if got, want := method, "TestNew"; got != want {
		t.Fatalf("expected the first stacktrace method to be %v, got %v", want, got)
	}
}

func TestWithSensitiveData(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://user:pass@remind.com:80/docs", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "this-is-a-secret")
	req.Header.Set("Cookie", "r101_auth_token=this-is-sensitive")
	ctx := errors.WithRequest(context.Background(), req)
	e := errors.New(ctx, errBoom, 0)
	r := e.Request()

	if r.URL.Scheme != "http" {
		t.Fatalf("expected request.URL.Scheme to be \"http\", got: %v", r.URL.Scheme)
	}

	if r.URL.User != nil {
		t.Fatal("expected request.User to have been removed by the reporter")
	}

	if r.URL.Host != "remind.com:80" {
		t.Fatalf("expected request.URL.Host to be \"remind.com:80\", got: %v", r.URL.Host)
	}

	if r.URL.Path != "/docs" {
		t.Fatalf("expected request.URL.Host to be \"/docs\", got: %v", r.URL.Path)
	}

	if r.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("expected request.Header[\"Content-type\"] to be \"application/json\", got: %v", r.Header.Get("Content-Type"))
	}

	if r.Header.Get("Authorization") != "" {
		t.Fatal("expected request.headers.Authorization to have been removed by the reporter")
	}

	if r.Header.Get("Cookie") != "" {
		t.Fatal("expected request.headers.Cookie to have been removed by the reporter")
	}

	if len(r.Cookies()) != 0 {
		t.Fatal("expected request.Cookies to have been removed by the reporter")
	}
}

func TestWithFormData(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Form = url.Values{}
	req.Form.Add("key", "foo")
	req.Form.Add("username", "admin")
	req.Form.Add("password", "this-is-a-secret")
	ctx := errors.WithRequest(context.Background(), req)
	e := errors.New(ctx, errBoom, 0)
	r := e.Request()

	if r.Form.Get("key") != "foo" {
		t.Fatalf("expected request.Form[\"key\"] to be \"foo\", got: %v", r.Form.Get("key"))
	}

	if r.Form.Get("username") != "admin" {
		t.Fatalf("expected request.Form[\"username\"] to be \"admin\", got: %v", r.Form.Get("username"))
	}

	if r.Form.Get("password") != "" {
		t.Fatal("expected request.Form[\"password\"] to have been removed by the reporter")
	}
}

type panicTest struct {
	Fn     func()
	TestFn func(error)
}

func TestPanics(t *testing.T) {
	tests := []panicTest{
		{
			Fn: func() {},
			TestFn: func(err error) {
				if err != nil {
					t.Error("expected err to be nil")
				}
			},
		},
		{
			Fn: func() {
				panic("boom!")
			},
			TestFn: func(err error) {
				if err == nil {
					t.Error("expected err to not be nil")
				}
				e := err.(*errors.Error)
				if got, want := fmt.Sprintf("%v", e.StackTrace()[0]), "errors_test.go:127"; got != want {
					t.Errorf("got: %v; expected: %v", got, want)
				}
			},
		},
		{
			Fn: func() {
				panic(fmt.Errorf("boom!"))
			},
			TestFn: func(err error) {
				if err == nil {
					t.Error("expected err to not be nil")
				}
				e := err.(*errors.Error)
				if got, want := fmt.Sprintf("%v", e.StackTrace()[0]), "errors_test.go:141"; got != want {
					t.Errorf("got: %v; expected: %v", got, want)
				}
			},
		},
		{
			Fn: func() {
				ctx := context.Background()
				ctx = errors.WithInfo(ctx, "request_id", "1234")
				panic(errors.New(ctx, gerrors.New("boom"), 0))
			},
			TestFn: func(err error) {
				if err == nil {
					t.Error("expected err to not be nil")
				}
				e := err.(*errors.Error)
				if got, want := fmt.Sprintf("%v", e.StackTrace()[0]), "errors_test.go:157"; got != want {
					t.Errorf("got: %v; expected: %v", got, want)
				}

				if got, want := e.ContextData()["request_id"], "1234"; got != want {
					t.Errorf("got: %v; expected: %v", got, want)
				}
			},
		},
	}
	for _, tt := range tests {
		runPanicTest(tt)
	}
}

func runPanicTest(pt panicTest) {
	defer func() {
		err := errors.Recover(context.Background(), recover())
		pt.TestFn(err)
	}()

	pt.Fn()
}
