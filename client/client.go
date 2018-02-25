package client

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/remind101/pkg/client/request"
)

// Client  - Holds request handlers, and a client and builds requests using them.
// client.NewRequest(operation, params, data) => creates new request, passes in handlers
type Client struct {
	Endpoint   string
	HTTPClient *http.Client
	Handlers   request.Handlers
}

// Timeout specifies a time limit for requests made by this Client.
func Timeout(t time.Duration) func(*Client) {
	return func(c *Client) {
		c.HTTPClient.Timeout = t
	}
}

// RoundTripper sets a custom transport on the underlying http Client.
func RoundTripper(r http.RoundTripper) func(*Client) {
	return func(c *Client) {
		c.HTTPClient.Transport = r
	}
}

// BasicAuth adds basic auth to every request made by this client.
func BasicAuth(username, password string) func(*Client) {
	return func(c *Client) {
		c.Handlers.Build.Append(request.BasicAuther(username, password))
	}
}

// RequestSigning adds a handler to sign requests.
func RequestSigning(id, key string) func(*Client) {
	return func(c *Client) {
		c.Handlers.Sign.Append(request.RequestSigner(id, key))
	}
}

// DebugLogging adds logging of the enitre request and response.
func DebugLogging(c *Client) {
	c.Handlers.Send.Prepend(request.RequestLogger)
	c.Handlers.Send.Append(request.ResponseLogger)
}

// New returns a new client.
func New(endpoint string, options ...func(*Client)) *Client {
	c := &Client{
		Endpoint: endpoint,
		Handlers: request.DefaultHandlers(),
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   1 * time.Second,
					KeepAlive: 90 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout: 3 * time.Second,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 8,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}

	// Apply options
	for _, option := range options {
		option(c)
	}

	return c
}

func (c *Client) NewRequest(ctx context.Context, method, path string, params interface{}, data interface{}) *request.Request {
	httpReq, _ := http.NewRequest(method, path, nil)
	httpReq = httpReq.WithContext(ctx)
	httpReq.URL, _ = url.Parse(c.Endpoint + path)

	r := request.New(httpReq, c.Handlers.Copy(), params, data)
	r.HTTPClient = c.HTTPClient
	return r
}
