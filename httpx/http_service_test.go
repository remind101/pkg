package httpx

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/remind101/pkg/retry"
)

// A fake http.RoundTripper that returns responses fed to it from a channel.
type MockTransport struct {
	passedRequest       *http.Request
	requestWasCancelled bool
	responses           chan *http.Response
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.passedRequest = req
	return (<-m.responses), nil
}

func (m *MockTransport) CancelRequest(req *http.Request) {
	m.requestWasCancelled = true
}

func TestRequestIDTransport(t *testing.T) {
	mockTransport := &MockTransport{responses: make(chan *http.Response)}
	mockClient := &http.Client{Transport: mockTransport}
	client := NewServiceClient("service_name", mockClient)

	go func() {
		mockTransport.responses <- &http.Response{StatusCode: 200}
	}()

	ctx := WithRequestID(context.Background(), "request_id")

	req, _ := http.NewRequest("GET", "/", nil)
	resp, err := client.Do(ctx, req)
	if resp.StatusCode != 200 {
		t.Fatal("Expected a 200 response to be the final return value")
	}
	if err != nil {
		t.Fatal("Expected RoundTrip to return without error for simple 200 response")
	}
	if mockTransport.passedRequest.Header.Get("X-Request-Id") != "request_id" {
		t.Fatalf("Expected the context request id to be present in the request")
	}
}

func TestRetryTransport(t *testing.T) {
	mockTransport := &MockTransport{responses: make(chan *http.Response)}
	mockClient := &http.Client{Transport: mockTransport}
	client := NewServiceClient("service_name", mockClient)

	// Generate responses for mockTransport to return in response to Do() calls
	go func() {
		// Test 1: immediate successful 200 return
		mockTransport.responses <- &http.Response{StatusCode: 200}

		// Test 2: retryable error that is eventually successful
		mockTransport.responses <- &http.Response{StatusCode: 500}
		mockTransport.responses <- &http.Response{StatusCode: 500}
		mockTransport.responses <- &http.Response{StatusCode: 200}

		// Test 3: non-retryable error
		mockTransport.responses <- &http.Response{StatusCode: 400}

		// Test 4: retryable error not retried because of non-retryable method
		mockTransport.responses <- &http.Response{StatusCode: 500}
	}()

	// Test 1: 200 returned immediately
	req, _ := http.NewRequest("GET", "/", nil)
	resp, err := client.Do(context.Background(), req)
	if err != nil {
		t.Fatal("Expected RoundTrip to return without error for simple 200 response")
	}
	if resp.StatusCode != 200 {
		t.Fatal("Expected a 200 response to be the final return value")
	}

	// Test 2: should retry after the two 500 calls and eventually return 200
	req, _ = http.NewRequest("GET", "/path", nil)
	resp, err = client.Do(context.Background(), req)
	if err != nil {
		t.Fatal("Expected RoundTrip to return without error")
	}
	if resp.StatusCode != 200 {
		t.Fatalf("Expected the 500 and 200 test sequence to return with the 200 response but got %d", resp.StatusCode)
	}

	// Test 3: should not retry after non-retryable error 400
	req, _ = http.NewRequest("GET", "/another/path", nil)
	resp, err = client.Do(context.Background(), req)
	if err != nil {
		t.Fatal("Expected RoundTrip to return without error")
	}
	if resp.StatusCode != 400 {
		t.Fatalf("Expected response code 400 but got %d", resp.StatusCode)
	}

	// Test 4: should not retry despite retryable error because the request method is not in MethodsToRetry
	req, _ = http.NewRequest("DELETE", "/path/to/delete", nil)
	resp, err = client.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected RoundTrip to return without error")
	}
	if resp.StatusCode != 500 {
		t.Fatalf("For DELETE request, code 500 should have been returned with no retry")
	}

}

func TestRetryableRequestNotRetried(t *testing.T) {
	mockTransport := &MockTransport{responses: make(chan *http.Response)}
	mockClient := &http.Client{Transport: mockTransport}

	// Retrier with an empty list of errors to retry
	retrier := retry.NewErrorTypeRetrier("service_name", retry.DefaultBackOffOpts)

	client := &Client{
		Transport: &RequestIDTransport{
			Transport: NewRetryTransport(retrier, &Transport{Client: mockClient}),
		},
	}

	go func() {
		// Test 1: expect 500 to be returned and no retry to be made
		mockTransport.responses <- &http.Response{StatusCode: 500}
		mockTransport.responses <- &http.Response{StatusCode: 200}
	}()

	// Test 1: request is retryable but will not be retried
	req, _ := http.NewRequest("GET", "/", nil)
	resp, err := client.Do(context.Background(), req)
	if err != nil {
		t.Fatal("Expected non-retried RoundTrip not to return a RetryableHTTPError")
	}
	if resp.StatusCode != 500 {
		t.Fatal("Expected a 500 response to be the final return value")
	}
}

func TestJSONRequests(t *testing.T) {
	// Empty request test
	req, err := NewJSONRequest("GET", "/", "")
	if err != nil {
		t.Fatal("Empty JSON request should have been built without error")
	}
	if req.Header.Get("content-type") != "application/json" {
		t.Fatal("JSON request should have the correct content-type header")
	}

	// Non-empty JSON request test
	jsonRequestBody := "{field:value}"
	req, err = NewJSONRequest("POST", "/path", jsonRequestBody)
	if err != nil {
		t.Fatal("Non-empty JSON request should have been built without error")
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		t.Fatal("Failed to read HTTP request body")
	}
	if !strings.Contains(string(body), jsonRequestBody) {
		t.Fatal("JSON data was not included in http request")
	}
}

func TestCredentialRemoval(t *testing.T) {
	baseURL, _ := url.Parse("http://user:pass@base_url.com")
	path := "/path/"

	expectedURL := "http://base_url.com/path/"
	actualURL := ParseURL(baseURL, path)
	if actualURL != expectedURL {
		t.Fatalf("Expected ParseURL to return %s but got %s", expectedURL, actualURL)
	}
}
