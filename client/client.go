package client

import (
	"net"
	"net/http"
	"time"
)

type Client struct {
	client *http.Client
}

// Timeout specifies a time limit for requests made by this Client.
func Timeout(t time.Duration) func(*Client) {
	return func(c *Client) {
		c.client.Timeout = t
	}
}

// RoundTripper sets a custom transport on the underlying http Client.
func RoundTripper(r http.RoundTripper) func(*Client) {
	return func(c *Client) {
		c.client.Transport = r
	}
}

// New returns a new client.
func New(options ...func(*Client)) *Client {
	c := &Client{
		client: &http.Client{
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

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Apply request options
	// for _, option := range c.requestOptions {
	// 	option(req)
	// }

	return c.client.Do(req)
}
