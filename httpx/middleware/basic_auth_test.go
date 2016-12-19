package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/city/123.123.123.123", nil)
	req.SetBasicAuth("user", "pass")

	var h http.Handler

	const test_str = "sfaftfofsfhfi"

	h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(test_str))
	})

	h = BasicAuth(h, "user", "pass", "realm")
	h.ServeHTTP(resp, req)

	if got, want := resp.Body.String(), test_str; got != want {
		t.Fatalf("Body => %s; want %s", got, want)
	}
}

func TestBasicAuthUnauthorized(t *testing.T) {
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/city/123.123.123.123", nil)

	var h http.Handler

	const test_str = "sfaftfofsfhfi"

	h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(test_str))
	})

	h = BasicAuth(h, "user", "pass", "realm")
	h.ServeHTTP(resp, req)

	if resp.Code != 401 {
		t.Fatalf("Expected code 401")
	}
}
