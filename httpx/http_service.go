package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/remind101/pkg/retry"

	"golang.org/x/net/context"
)

type RoundTripper interface {
	RoundTrip(*http.Request, context.Context) (*http.Response, error)
}

type Client struct {
	innerClient		*http.Client
	retryTransport 	*RetryTransport
	baseURL 		*url.URL
}

type RetryTransport struct {
	Retrier     *retry.Retrier
	Transport	*http.Transport
	Client		*Client
}

func NewHTTPClient(serviceName string, baseURL *url.URL) *Client {
	retryTransport := NewRetryTransport(serviceName)

	client := &Client{
		innerClient: &http.Client{},
		baseURL: baseURL,
		retryTransport: retryTransport,
	}
	retryTransport.Client = client
	client.innerClient.Transport = retryTransport.Transport

	return client
}
func NewRetryTransport(serviceName string) *RetryTransport {
	retrier := retry.NewErrorTypeRetrier(serviceName,
		retry.DefaultBackOffOpts,
		(*net.OpError)(nil),
		(*RetryableHTTPError)(nil))

	// DefaultTransport but with different numerical settings
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 90 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 3 * time.Second,
	}

	return &RetryTransport{
		Retrier: retrier,
		Transport: transport,
	}
}

func (rt *RetryTransport) RoundTrip(req *http.Request, ctx context.Context) (*http.Response, error) {
	ret, err := rt.Retrier.Retry(func() (interface{}, error) {
		return rt.Client.call(req, ctx)
	})
	if err != nil {
		return nil, err
	}
	if resp, ok := ret.(*http.Response); ok {
		return resp, nil
	} else {
		panic("http.Response not of correct type")
	}
}

func (client *Client) Do(req *http.Request, ctx context.Context) (*http.Response, error) {
	return client.retryTransport.RoundTrip(req, ctx)
}

func (client *Client) call(req *http.Request, ctx context.Context) (*http.Response, error) {
	req.Header.Add("X-Request-Id", RequestID(ctx))

	resp, err := client.innerClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 500 {
		return nil, &RetryableHTTPError{Path: req.URL.String(), StatusCode: resp.StatusCode}
	} else if resp.StatusCode >= 300 {
		return nil, &HTTPError{Path: req.URL.String(), StatusCode: resp.StatusCode}
	}

	return resp, nil
}

func (client *Client) ParseURL(path string) string {
	urlString := client.UrlWithoutCreds(*client.baseURL)
	return strings.TrimRight(urlString, "/") + path
}

// Shows the URL without information about the current username and password
func (client *Client) UrlWithoutCreds(u url.URL) string {
	u.User = nil
	return u.String()
}

func (client *Client) BuildJSONRequest(method, path string, jsonData interface{}) (*http.Request, error) {
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

func (client *Client) marshalRequestBody(jsonData interface{}) (io.Reader, error) {
	if jsonData == nil {
		return nil, nil
	}
	body, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(body), nil
}

// HTTPError is for generic non-200 errors
type HTTPError struct {
	Path       string
	StatusCode int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http service returned a error code when " +
	"calling %s: %d", e.Path, e.StatusCode)
}

// RetryableHTTPError is used to represent error codes that can be allowed to retry
type RetryableHTTPError struct {
	Path       string
	StatusCode int
}

func (e *RetryableHTTPError) Error() string {
	return fmt.Sprintf("http service returned a >= 500 error code when " +
	"calling %s: %d. This request can be retried.", e.Path, e.StatusCode)
}
