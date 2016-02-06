package httpx

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/context"
)

type MockTransport struct {
	responses chan *http.Response
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return (<-m.responses), nil
}

func TestClientRetries(t *testing.T) {
	url, _ := url.Parse("http://base_url.com")
	client := NewHTTPClient("service_name", url)
	mockTransport := &MockTransport{responses: make(chan *http.Response)}
	client.innerClient.Transport = mockTransport

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
	}()

	// Test 1: 200 returned immediately
	req, _ := http.NewRequest("GET", "/", nil)
	resp, err := client.Do(req, context.Background())
	if err != nil {
		t.Fatal("Expected RoundTrip to return without error for simple 200 response")
	}
	if resp.StatusCode != 200 {
		t.Fatal("Expected a 200 response to be the final return value")
	}

	// Test 2: should retry after the two 500 calls and eventually return 200
	req, _ = http.NewRequest("POST", "/path", nil)
	resp, err = client.Do(req, context.Background())
	if err != nil {
		t.Fatal("Expected RoundTrip to return without error")
	}
	if resp.StatusCode != 200 {
		t.Fatalf("Expected the 500 and 200 test sequence to return with the 200 response but got %d", resp.StatusCode)
	}

	// Test 3: should not retry after non-retryable error 400 and should return the error instead
	req, _ = http.NewRequest("DELETE", "/another/path", nil)
	resp, err = client.Do(req, context.Background())
	if resp != nil {
		t.Fatal("Status 400 should have resulted in nil response")
	}
	if err, ok := err.(*HTTPError); ok {
		if err.StatusCode != 400 {
			t.Fatalf("Expected error code 400 but got %d", err.StatusCode)
		}
	} else {
		t.Fatal("Response 400 returned with an error, but not of the expected type HTTPError")
	}
}

func TestJSONRequests(t *testing.T) {
	baseURL, _ := url.Parse("http://base_url.com")
	client := NewHTTPClient("service_name", baseURL)

	// Empty request test
	req, err := client.BuildJSONRequest("GET", "/", "")
	if err != nil {
		t.Fatal("Empty JSON request should have been built without error")
	}
	if req.Header.Get("content-type") != "application/json" {
		t.Fatal("JSON request should have the correct content-type header")
	}

	// Non-empty JSON request test
	jsonRequestBody := "{field:value}"
	req, err = client.BuildJSONRequest("POST", "/path", jsonRequestBody)
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
	client := NewHTTPClient("service_name", baseURL)
	path := "/path/"

	expectedUrl := "http://base_url.com/path/"
	actualUrl := client.ParseURL(path)
	if actualUrl != expectedUrl {
		t.Fatalf("Expected ParseURL to return %s but got %s", expectedUrl, actualUrl)
	}
}
