package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/httpx/middleware"
	"github.com/remind101/pkg/logger"
	"github.com/remind101/pkg/reporter"
	"github.com/remind101/pkg/retry"
	"golang.org/x/net/context"
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
	m.Handle("/auth", middleware.BasicAuth(httpx.HandlerFunc(ok), "user", "pass", "realm")).Methods("GET")

	var h httpx.Handler

	// Recover from panics, and report the recovered error to the reporter.
	h = middleware.Recover(m, r)

	// Handles any errors returned from handlers in a common format.
	h = middleware.HandleError(h, errorHandler)

	// Adds a logger to the context.Context that will log to stdout,
	// prefixed with the request id.
	h = middleware.LogTo(h, middleware.StdoutLogger)

	http.ListenAndServe(":8080", middleware.BackgroundContext(h))
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
