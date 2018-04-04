package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"context"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/httpx/middleware"
	"github.com/remind101/pkg/logger"
	"github.com/remind101/pkg/reporter"
	"github.com/remind101/pkg/retry"
	"github.com/remind101/pkg/svc"
)

func ok(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ip, err := ip(ctx)
	if err != nil {
		return err
	}

	logger.Info(ctx, "ip address", "ip", ip)

	_, err = fmt.Fprintf(w, "%s\n", ip)
	return err
}

func bad(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return &Error{ID: "bad_error", Err: errors.New("bad request")}
}

func boom(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	panic("boom")
}

func new(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	panic(errors.New("new error"))
}

type CustomError string

func (e CustomError) Error() string {
	return string(e)
}

func custom(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	panic(errors.WithStack(CustomError("custom error")))
}

func inner() error {
	return errors.New("inner")
}

func wrap(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	err := inner()
	panic(errors.Wrap(err, "this is a wrap"))
}

func errorHandler(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func main() {
	// An error reporter that will log errors to stdout.
	r := reporter.NewLogReporter()

	m := httpx.NewRouter()

	m.HandleFunc("/ok", ok).Methods("GET")
	m.HandleFunc("/bad", bad).Methods("GET")
	m.HandleFunc("/boom", boom).Methods("GET")
	m.HandleFunc("/new", new).Methods("GET")
	m.HandleFunc("/custom", custom).Methods("GET")
	m.HandleFunc("/wrap", wrap).Methods("GET")
	m.Handle("/auth", middleware.BasicAuth(httpx.HandlerFunc(ok), "user", "pass", "realm")).Methods("GET")

	h := svc.NewStandardHandler(svc.HandlerOpts{
		Router:         m,
		Reporter:       r,
		ErrorHandler:   httpx.Error,
		HandlerTimeout: 5 * time.Second,
	})

	svc.RunServer(h, "8080", 5*time.Second)
}

type Error struct {
	ID  string
	Err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.ID, e.Err)
}

// ip returns your ip.
func ip(ctx context.Context) (string, error) {
	req, err := http.NewRequest("GET", "http://api.ipify.org?format=text", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Request-ID", httpx.RequestID(ctx))

	retrier := retry.NewRetrier("ip", retry.DefaultBackOffOpts, retry.RetryOnAnyError)
	val, err := retrier.Retry(func() (interface{}, error) { return http.DefaultClient.Do(req) })
	if err != nil {
		return "", err
	}
	resp := val.(*http.Response)
	defer resp.Body.Close()

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}
