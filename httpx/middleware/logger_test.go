// Thanks negroni!
package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogger(t *testing.T) {
	b := new(bytes.Buffer)

	h := LogTo(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
	}), stdLogger(b))

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	h.ServeHTTP(resp, req)

	if got, want := b.String(), "request_id= request.start method=GET path=/\nrequest_id= request.complete status=201\n"; got != want {
		t.Fatalf("%s; want %s", got, want)
	}
}
