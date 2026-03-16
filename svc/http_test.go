package svc_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/reporter"
	"github.com/remind101/pkg/svc"
)

func TestHTTPStack(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/boom", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusCreated)
		fmt.Fprintln(rw, r.Header.Get("X-Request-ID"))
	}).Methods("GET")

	h := svc.NewHTTPStack(r, svc.HTTPHandlerOpts{
		Reporter:     reporter.NewLogReporter(),
		ErrorHandler: httpx.Error,
	})

	req, _ := http.NewRequest("GET", "/boom", nil)
	req.Header.Add("X-Request-ID", "abc")

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if got, want := rw.Result().StatusCode, http.StatusCreated; got != want {
		t.Errorf("got %v; expected %v", got, want)
	}

	body, _ := ioutil.ReadAll(rw.Body)
	if got, want := string(body), "abc\n"; got != want {
		t.Errorf("got %v; expected %v", got, want)
	}
}
