package httpservice

import (
	"encoding/json"
	"fmt"
	"github.com/remind101/pkg/retry"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPService provides a way to make http calls with an aggressive transport
type HTTPService interface {
	Call(method string, path string, jsonData interface{}) ([]byte, error)
	CallWithRetries(method string, path string, jsonData interface{}) ([]byte, error)
}

// HTTPClient contains the retrier, baseURL and the actual httpClient.
type HTTPClient struct {
	client      *http.Client
	baseURL     *url.URL
	retrier     *retry.Retrier
	serviceName string
}

// NewHTTPClient returns a new HTTPClient for the given service name and the base url
// The baseURL will be used to prefix the paths provided in Call and CallWithRetries
func NewHTTPClient(serviceName string, baseURL *url.URL) *HTTPClient {
	aggressiveTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   1 * time.Second, // !!!
			KeepAlive: 90 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 3 * time.Second,
	}

	return &HTTPClient{
		baseURL: baseURL,
		client: &http.Client{
			Transport: aggressiveTransport,
		},
		retrier: retry.NewErrorTypeRetrier(serviceName,
			retry.DefaultBackOffOpts,
			(*net.OpError)(nil),
			(*RetryableHTTPError)(nil)),
	}
}

// CallWithRetries makes calls to the network with the default retry logic
func (client *HTTPClient) CallWithRetries(method string, path string, jsonData interface{}) ([]byte, error) {
	res, err := client.retrier.Retry(func() (interface{}, error) {
		return client.Call(method, path, jsonData)
	})
	if err != nil {
		return []byte{}, err
	}
	return res.([]byte), nil
}

// Call formats the network request and returns the raw data as the response
func (client *HTTPClient) Call(method string, path string, jsonData interface{}) ([]byte, error) {
	req, err := client.buildJSONRequest(method, client.parseURL(path), jsonData)
	if err != nil {
		return nil, err
	}

	resp, err := client.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, &RetryableHTTPError{Path: path, StatusCode: resp.StatusCode}
	} else if resp.StatusCode >= 300 {
		return nil, &HTTPError{Path: path, StatusCode: resp.StatusCode}
	}

	return ioutil.ReadAll(resp.Body)
}

func (client *HTTPClient) parseURL(path string) string {
	baseURL := client.urlWithoutCreds(*client.baseURL)
	return strings.TrimRight(baseURL, "/") + path
}

func (client *HTTPClient) urlWithoutCreds(u url.URL) string {
	u.User = nil
	return u.String()
}

func (client *HTTPClient) buildJSONRequest(method, path string, jsonData interface{}) (*http.Request, error) {

	requestBody, err := client.marshalRequestBody(jsonData)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(method, path, requestBody)
	if err != nil {
		return nil, err
	}

	if requestBody != nil {
		request.Header.Set("content-type", "application/json")
	}

	return request, nil
}

func (client *HTTPClient) marshalRequestBody(jsonData interface{}) (io.Reader, error) {
	if jsonData == nil {
		return nil, nil
	}
	body, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(string(body)), nil
}

// HTTPError is for generic non-200 errors
type HTTPError struct {
	Path       string
	StatusCode int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http service returned a error code when "+
		"calling %s: %d", e.Path, e.StatusCode)
}

// RetryableHTTPError is used to represent error codes that can be allowed to retry
type RetryableHTTPError struct {
	Path       string
	StatusCode int
}

func (e *RetryableHTTPError) Error() string {
	return fmt.Sprintf("http service returned a >= 500 error code when "+
		"calling %s: %d. This request can be retried.", e.Path, e.StatusCode)
}
