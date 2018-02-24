package client_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/remind101/pkg/client"
)

type requestTest struct {
	Request     *client.Request
	HTTPHandler http.Handler
	Test        func(*client.Request)
}

type JSONparams struct {
	Param string `json:"param"`
}

type JSONdata struct {
	Value string `json:"value"`
}

func TestRequestDefaults(t *testing.T) {
	httpReq, _ := http.NewRequest("POST", "/foo", nil)
	var data JSONdata
	r := client.NewRequest(httpReq, client.DefaultHandlers(), JSONparams{Param: "hello"}, &data)

	runRequestTest(requestTest{
		Request: r,
		HTTPHandler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
				t.Errorf("got %s; expected %s", got, want)
			}
			if got, want := r.Header.Get("Accept"), "application/json"; got != want {
				t.Errorf("got %s; expected %s", got, want)
			}

			var params JSONparams
			err := json.NewDecoder(r.Body).Decode(&params)
			if err != nil {
				t.Error(err)
			}
			fmt.Fprintf(rw, `{"value":"%s"}`, params.Param)
		}),
		Test: func(r *client.Request) {
			if r.Error != nil {
				t.Error(r.Error)
			}
			if got, want := data.Value, "hello"; got != want {
				t.Errorf("got %s; expected %s", got, want)
			}
		},
	})
}

func runRequestTest(test requestTest) {
	s := httptest.NewServer(test.HTTPHandler)
	defer s.Close()
	replaceURL(test.Request, s.URL)
	test.Request.Send()
	test.Test(test.Request)
}

func replaceURL(r *client.Request, rawURL string) {
	u, _ := url.Parse(rawURL)
	r.HTTPRequest.URL.Host = u.Host
	r.HTTPRequest.URL.Scheme = u.Scheme
}
