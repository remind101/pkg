package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/httpx/middleware"
	"github.com/remind101/pkg/logger"
	"github.com/remind101/pkg/reporter"
	"golang.org/x/net/context"
)

func ok(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	io.WriteString(w, "Ok\n")
	logger.Log(ctx, "foo", "bar")
	return nil
}

func bad(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return &Error{ID: "bad_error", Err: errors.New("bad request")}
}

func boom(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	panic("boom")
}

func errorHandler(err error, w http.ResponseWriter, r *http.Request) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func main() {
	// An error reporter that will log errors to stdout.
	r := reporter.NewLogReporter()

	m := httpx.NewRouter()

	m.Handle("GET", "/ok", httpx.HandlerFunc(ok))
	m.Handle("GET", "/bad", httpx.HandlerFunc(bad))
	m.Handle("GET", "/boom", httpx.HandlerFunc(boom))

	var h httpx.Handler

	// Recover from panics, and report the recovered error to the reporter.
	h = middleware.Recover(m, r)

	// Handles any errors returned from handlers in a common format.
	h = middleware.HandleError(h, errorHandler)

	// Adds a logger to the context.Context that will log to stdout,
	// prefixed with the request id.
	h = middleware.NewLogger(h, os.Stdout)

	// Adds the request id to the context.
	h = middleware.ExtractRequestID(h)

	http.ListenAndServe(":8080", middleware.BackgroundContext(h))
}

type Error struct {
	ID  string
	Err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.ID, e.Err)
}
