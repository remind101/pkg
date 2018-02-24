package client_test

import (
	"encoding/json"
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

func newTestRequest(method string, path string, params interface{}, data interface{}) *client.Request {
	httpReq, _ := http.NewRequest(method, path, nil)
	return client.NewRequest(httpReq, client.DefaultHandlers(), params, &data)
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

// Test Basic Auth
func TestBasicAuther(t *testing.T) {
	r := newTestRequest("GET", "/foo", nil, nil)
	r.Handlers.Build.Append(client.BasicAuther("user", "pass"))

	sendRequest(r, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			t.Error("expected basic auth info to be present")
		}
		if got, want := user, "user"; got != want {
			t.Errorf("got %s; expected %s", got, want)
		}
		if got, want := pass, "pass"; got != want {
			t.Errorf("got %s; expected %s", got, want)
		}
	}))
}

// Test Adding Headers
// Test Forwarding Headers
// Test Tracing
// Test Parsing the status code
// Test metrics?
// Test Bearer Auth

func sendRequest(r *client.Request, h http.Handler) {
	s := httptest.NewServer(h)
	defer s.Close()
	replaceURL(r, s.URL)
	r.Send()
}

func replaceURL(r *client.Request, rawURL string) {
	u, _ := url.Parse(rawURL)
	r.HTTPRequest.URL.Host = u.Host
	r.HTTPRequest.URL.Scheme = u.Scheme
}
