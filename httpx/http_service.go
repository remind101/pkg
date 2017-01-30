package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/remind101/pkg/retry"
)

type RoundTripper interface {
	RoundTrip(context.Context, *http.Request) (*http.Response, error)
}

// Client is an extension of http.Client with context.Context support.
type Client struct {
	Transport RoundTripper
}

var DefaultHTTPTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout:   15 * time.Second,
		KeepAlive: 90 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 3 * time.Second,
}

var DefaultHTTPClient = &http.Client{
	Transport: DefaultHTTPTransport,
}

func NewDefaultServiceClient(serviceName string) *Client {
	return NewServiceClient(serviceName, nil)
}

// NewClient returns a new Client instance that will use the given http.Client
// to perform round trips
func NewClient(c *http.Client) *Client {
	return &Client{Transport: &Transport{Client: c}}
}

// NewServiceClient returns an httpx.Client that has the following behavior:
//
//      1. Request ids will be added to outgoing requests within the
//         X-Request-Id header.
//      2. Any 500 errors will be retried.
//
// The optional *http.Client parameter can be used to override the default client.
func NewServiceClient(serviceName string, c *http.Client) *Client {
	if c == nil {
		c = DefaultHTTPClient
	}

	retrier := retry.NewErrorTypeRetrier(serviceName,
		retry.DefaultBackOffOpts,
		(*net.OpError)(nil),
		(*RetryableHTTPError)(nil))

	return &Client{
		Transport: &RequestIDTransport{
			Transport: NewRetryTransport(retrier, &Transport{Client: c}),
		},
	}
}

// Do performs the request and returns the response.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.Transport.RoundTrip(ctx, req)
}

// Transport is an implementation of the RoundTripper interface that uses an
// http.Client from the standard lib.
type Transport struct {
	*http.Client
}

// TODO: add support for context.Context cancellations
func (t *Transport) RoundTrip(ctx context.Context, req *http.Request) (*http.Response, error) {
	return t.Client.Do(req)
}

// RetryTransport is an implementation of the RoundTripper interface that
// retries requests.
type RetryTransport struct {
	*retry.Retrier
	MethodsToRetry map[string]bool
	Transport      RoundTripper
}

// NewRetryTransport returns a RetryTransport that will retry idempotent HTTP
// requests (GET/HEAD) using the given retrier.
func NewRetryTransport(retrier *retry.Retrier, transport RoundTripper) *RetryTransport {
	return &RetryTransport{
		Retrier: retrier,
		MethodsToRetry: map[string]bool{
			"":     true, // http.Client treats empty methods the same as "GET"
			"GET":  true,
			"HEAD": true,
		},
		Transport: transport,
	}
}

func (t *RetryTransport) RoundTrip(ctx context.Context, req *http.Request) (*http.Response, error) {
	if !t.MethodsToRetry[req.Method] {
		return t.Transport.RoundTrip(ctx, req)
	}

	resp, err := t.Retrier.Retry(func() (interface{}, error) {
		resp, err := t.Transport.RoundTrip(ctx, req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode >= 500 {
			return resp, &RetryableHTTPError{Path: req.URL.String(), StatusCode: resp.StatusCode}
		}

		return resp, nil
	})

	if resp == nil {
		return nil, err
	} else if response, ok := resp.(*http.Response); ok {
		return response, nil
	} else {
		panic("Response is non-nil and not of an expected type")
	}
}

// RequestIDTransport is an http.RoundTripper implementation that adds the
// embedded request id to outgoing http requests.
type RequestIDTransport struct {
	Transport RoundTripper
}

func (t *RequestIDTransport) RoundTrip(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.Header.Add("X-Request-Id", RequestID(ctx))
	return t.Transport.RoundTrip(ctx, req)
}

// NewJSONRequest generates a new http.Request with the body set to the json
// encoding of v.
func NewJSONRequest(method, path string, v interface{}) (*http.Request, error) {
	var r io.Reader
	if v != nil {
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, path, r)
	if err != nil {
		return nil, err
	}
	if v != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func ParseURL(baseURL *url.URL, path string) string {
	URLString := URLWithoutCreds(*baseURL)
	return strings.TrimRight(URLString, "/") + path
}

// Shows the URL without information about the current username and password
func URLWithoutCreds(u url.URL) string {
	u.User = nil
	return u.String()
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
