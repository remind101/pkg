package request_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	httpsignatures "github.com/99designs/httpsignatures-go"
	"github.com/remind101/pkg/client/request"
	"github.com/remind101/pkg/httpx"
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
	return request.New(httpReq, request.DefaultHandlers(), params, &data)
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
	r.Handlers.Build.Append(request.BasicAuther("user", "pass"))

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
func TestAddingHeaders(t *testing.T) {
	r := newTestRequest("GET", "/foo", nil, nil)
	r.Handlers.Build.Append(request.Handler{
		Name: "Extra Headers",
		Fn: func(r *request.Request) {
			r.HTTPRequest.Header.Add("X-Ray", "123456789")
		},
	})
	sendRequest(r, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("X-Ray"), "123456789"; got != want {
			t.Errorf("got %s; expected %s", got, want)
		}
	}))
}

// Test Forwarding Headers
func TestForwardingHeaders(t *testing.T) {
	r := newTestRequest("GET", "/", nil, nil)
	// Add a header to request context
	ctx := httpx.WithHeader(r.HTTPRequest.Context(), "X-Request-Id", "123456789")
	r.HTTPRequest = r.HTTPRequest.WithContext(ctx)

	// Add handler to pull value from context and add as header
	r.Handlers.Build.Append(request.HeadersFromContext("X-Request-Id"))

	sendRequest(r, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("X-Request-Id"), "123456789"; got != want {
			t.Errorf("got %s; expected %s", got, want)
		}
	}))
}

// Test Request Signing
func TestRequestSinging(t *testing.T) {
	r := newTestRequest("GET", "/", nil, nil)
	r.Handlers.Sign.Append(request.RequestSigner("id", "key"))
	sendRequest(r, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		sig, err := httpsignatures.FromRequest(r)
		if err != nil {
			t.Error(err)
		}
		if !sig.IsValid("key", r) {
			t.Error("Expected signature to be valid")
		}
	}))
}

func TestDebugLogging(t *testing.T) {
	r := newTestRequest("GET", "/", nil, nil)
	r.Handlers.Send.Prepend(request.RequestLogger)
	r.Handlers.Send.Append(request.ResponseLogger)
	sendRequest(r, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
	}))
}

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
