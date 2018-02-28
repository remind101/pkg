package request_test

import (
	"net/http"
	"testing"

	httpsignatures "github.com/99designs/httpsignatures-go"
	"github.com/remind101/pkg/client/request"
	"github.com/remind101/pkg/httpx"
)

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
