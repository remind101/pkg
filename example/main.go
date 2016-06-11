package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/httpx/middleware"
	"github.com/remind101/pkg/logger"
	"github.com/remind101/pkg/reporter"
	"github.com/remind101/pkg/retry"
	"golang.org/x/net/context"
)

func ok(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ip, err := ip(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info(ctx, "ip address", "ip", ip)

	fmt.Fprintf(w, "%s\n", ip)
}

func main() {
	// An error reporter that will log errors to stdout.
	r := reporter.NewLogReporter()

	m := mux.NewRouter()

	m.HandleFunc("/ok", ok).Methods("GET")
	m.Handle("/auth", middleware.BasicAuth(http.HandlerFunc(ok), "user", "pass", "realm")).Methods("GET")

	var h http.Handler

	// Recover from panics, and report the recovered error to the reporter.
	h = middleware.Recover(m, r)

	// Adds a logger to the context.Context that will log to stdout,
	// prefixed with the request id.
	h = middleware.LogTo(h, middleware.StdoutLogger)

	// Adds the request id to the context.
	h = middleware.ExtractRequestID(h)

	http.ListenAndServe(":8080", h)
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
