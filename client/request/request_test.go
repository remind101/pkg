package request_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/remind101/pkg/client/metadata"
	"github.com/remind101/pkg/client/request"
)

type requestTest struct {
	Request     *request.Request
	HTTPHandler http.Handler
	Test        func(*request.Request)
}

type JSONparams struct {
	Param string `json:"param"`
}

type JSONdata struct {
	Value string `json:"value"`
}

func newTestRequest(method string, path string, params interface{}, data interface{}) *request.Request {
	httpReq, _ := http.NewRequest(method, path, nil)
	info := metadata.ClientInfo{ServiceName: "TestService", Endpoint: "http://localhost"}
	return request.New(httpReq, info, request.DefaultHandlers(), params, &data)
}

func TestRequestDefaults(t *testing.T) {
	var data JSONdata
	r := newTestRequest("POST", "/foo", JSONparams{Param: "hello"}, &data)

	sendRequest(r, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Verify JSON headers.
		if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
			t.Errorf("got %s; expected %s", got, want)
		}
		if got, want := r.Header.Get("Accept"), "application/json"; got != want {
			t.Errorf("got %s; expected %s", got, want)
		}

		// Verify JSON encoded request body.
		var params JSONparams
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			t.Error(err)
		}

		// Encode JSON response body.
		response := JSONdata{Value: params.Param}
		json.NewEncoder(rw).Encode(response)
	}))

	if r.Error != nil {
		t.Error(r.Error)
	}

	// Verify decoded JSON response.
	if got, want := data.Value, "hello"; got != want {
		t.Errorf("got %s; expected %s", got, want)
	}
}

// TODO
// Test Parsing the status code (create custom error type)
// Test Logging
// Test Metrics
// Test Failure Modes (failure to encode input, failure to send, failed response, failure to decode, multiple failures?)
// Test Context insertion
// Test Retry

func sendRequest(r *request.Request, h http.Handler) {
	s := httptest.NewServer(h)
	defer s.Close()
	replaceURL(r, s.URL)
	r.Send()
}

func replaceURL(r *request.Request, rawURL string) {
	u, _ := url.Parse(rawURL)
	r.HTTPRequest.URL.Host = u.Host
	r.HTTPRequest.URL.Scheme = u.Scheme
}
