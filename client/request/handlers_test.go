package request_test

import (
	"bytes"
	"io"
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

func TestInvalidJSON(t *testing.T) {
	invalidJSONs := [][]byte{
		[]byte("{\"value\": \"hello\"}{\"value\": \"world\"}"),
		[]byte("{\"value\": \"hello\"}42"),
		[]byte("42"),
	}

	for _, invalidJSON := range invalidJSONs {
		var data JSONdata
		r := newTestRequest("POST", "/foo", nil, &data)
		sendRequest(r, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write(invalidJSON)
		}))

		if r.Error == nil {
			t.Errorf("malformed JSON %s should produce an error.", invalidJSON)
		}
	}
}

func TestValidJSON(t *testing.T) {
	validJSONs := [][]byte{
		[]byte("{\"value\": \"hello\"}"),
		[]byte("{\"value\": \"hello\"}\r\n"),
		[]byte("{\"value\": \"hello\"}\n"),
		[]byte("     {\"value\": \"hello\"}     "),
	}

	for _, validJSON := range validJSONs {
		var data JSONdata
		r := newTestRequest("POST", "/foo", nil, &data)
		sendRequest(r, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write(validJSON)
		}))

		if r.Error != nil {
			t.Errorf("valid JSON %s produced error: %v", validJSON, r.Error)
		}
		if data.Value != "hello" {
			t.Errorf("%s should have produced value \"hello\", got \"%v\"", validJSON, data.Value)
		}
	}
}

// for any io.Reader, .Read() may read some bytes and return EOF, or it may
// read zero bytes and return EOF. zeroReadEOFReader wraps any io.Reader such
// that it always does the latter.
//
// This is important because for an http connection to be reused, not only must
// all the bytes be read, but the response body must be read to EOF. Thus this
// is useful to test for handlers which might might read all the bytes (because
// they have some other means of knowing the length of the content) but fail to
// make the final call that returns zero bytes and EOF.
type zeroReadEOFReader struct {
	source  io.Reader
	sentEOF bool
}

func (c *zeroReadEOFReader) Close() error {
	c.source = nil
	return nil
}

func (c *zeroReadEOFReader) Read(buf []byte) (int, error) {
	bytesRead, err := c.source.Read(buf)
	if bytesRead != 0 && err == io.EOF {
		err = nil
	}
	c.sentEOF = err == io.EOF
	return bytesRead, err
}

func TestJSONReadsToEOF(t *testing.T) {
	var data JSONdata
	r := newTestRequest("POST", "/foo", nil, &data)
	r.HTTPResponse = &http.Response{}
	body := zeroReadEOFReader{source: bytes.NewReader([]byte("{\"value\": \"hello\"}"))}
	r.HTTPResponse.Body = &body
	request.JSONDecoder.Fn(r)
	if r.Error != nil {
		t.Errorf("handler returned error: %v", r.Error)
	}
	if data.Value != "hello" {
		t.Errorf("Value should be \"hello\", got %v", data.Value)
	}
	if !body.sentEOF {
		t.Error("handler did not read EOF")
	}
}

func TestJSONDecodeNoData(t *testing.T) {
	r := newTestRequest("POST", "/foo", nil, nil)
	r.HTTPResponse = &http.Response{}
	body := zeroReadEOFReader{source: bytes.NewReader([]byte("{\"value\": \"hello\"}"))}
	r.HTTPResponse.Body = &body
	request.JSONDecoder.Fn(r)
	if r.Error != nil {
		t.Errorf("handler returned error: %v", r.Error)
	}
	if !body.sentEOF {
		t.Error("handler did not read EOF")
	}
}
