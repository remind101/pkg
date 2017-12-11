package service_client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/metrics"

	"golang.org/x/net/context"
)

var AggressiveTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   1 * time.Second,
		KeepAlive: 90 * time.Second,
	}).DialContext,
	TLSHandshakeTimeout: 3 * time.Second,
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 8,
	IdleConnTimeout:     90 * time.Second,
}

type clientMetrics struct {
	Transport http.RoundTripper
	prefix    string
}

func (t *clientMetrics) RoundTrip(req *http.Request) (*http.Response, error) {
	count := func(key string, delta int64, tags map[string]string) {
		metrics.Count("hermes."+t.prefix+"."+key, delta, tags, 1.0)
	}

	trace := &httptrace.ClientTrace{
		// GotConn is called after a successful connection is
		// obtained. There is no hook for failure to obtain a
		// connection; instead, use the error from
		// Transport.RoundTrip.
		GotConn: func(info httptrace.GotConnInfo) {
			count("GotConn", 1, map[string]string{
				"reused":   fmt.Sprintf("%t", info.Reused),
				"was_idle": fmt.Sprintf("%t", info.WasIdle),
			})
		},
		// PutIdleConn is called when the connection is returned to
		// the idle pool. If err is nil, the connection was
		// successfully returned to the idle pool. If err is non-nil,
		// it describes why not. PutIdleConn is not called if
		// connection reuse is disabled via Transport.DisableKeepAlives.
		// PutIdleConn is called before the caller's Response.Body.Close
		// call returns.
		// For HTTP/2, this hook is not currently used.
		PutIdleConn: func(err error) {
			count("PutIdleConn", 1, map[string]string{
				"error": fmt.Sprintf("%t", err != nil),
			})
		},
		// ConnectStart is called when a new connection's Dial begins.
		// If net.Dialer.DualStack (IPv6 "Happy Eyeballs") support is
		// enabled, this may be called multiple times.
		ConnectStart: func(network, addr string) {
			count("ConnectStart", 1, nil)
		},
		// ConnectDone is called when a new connection's Dial
		// completes. The provided err indicates whether the
		// connection completedly successfully.
		// If net.Dialer.DualStack ("Happy Eyeballs") support is
		// enabled, this may be called multiple times.
		ConnectDone: func(network, addr string, err error) {
			count("ConnectDone", 1, map[string]string{
				"error": fmt.Sprintf("%t", err != nil),
			})
		},
		// DNSStart is called when a DNS lookup begins.
		DNSStart: func(info httptrace.DNSStartInfo) {
			count("DNSStart", 1, nil)
		},
		// DNSDone is called when a DNS lookup ends.
		DNSDone: func(info httptrace.DNSDoneInfo) {
			count("DNSDone", 1, map[string]string{
				"error":     fmt.Sprintf("%t", info.Err != nil),
				"coalesced": fmt.Sprintf("%t", info.Coalesced),
			})
		},
	}
	ctx := httptrace.WithClientTrace(req.Context(), trace)
	resp, err := t.Transport.RoundTrip(req.WithContext(ctx))
	if err != nil {
		return resp, err
	}
	if resp.Close {
		count("ConnectionClosed", 1, nil)
	}
	return resp, err
}

type ServiceClient interface {
	Do(ctx context.Context, method, path string, jsonData interface{}, targetObject interface{}) error
	DoWithBearerAuth(ctx context.Context, method, path, access_token string, jsonData, targetObject interface{}) error
}

type serviceClient struct {
	serviceURL *url.URL
	client     *httpx.Client
}

func NewServiceClient(serviceURL string) *serviceClient {
	u, err := url.Parse(serviceURL)
	if err != nil {
		panic(err)
	}
	httpClient := &http.Client{Transport: &clientMetrics{AggressiveTransport, u.Host}}
	client := httpx.NewServiceClient(serviceURL, httpClient)

	return &serviceClient{
		serviceURL: u,
		client:     client,
	}
}

func (c *serviceClient) Do(ctx context.Context, method, path string, jsonData interface{}, targetObject interface{}) error {
	return c.do(ctx, method, path, "", jsonData, targetObject)
}

func (c *serviceClient) DoWithBearerAuth(ctx context.Context, method, path, token string, jsonData interface{}, targetObject interface{}) error {
	return c.do(ctx, method, path, token, jsonData, targetObject)
}

func (c *serviceClient) do(ctx context.Context, method, path, token string, jsonData interface{}, targetObject interface{}) error {
	requestPath := httpx.ParseURL(c.serviceURL, path)
	req, err := httpx.NewJSONRequest(method, requestPath, jsonData)

	if err != nil {
		return err
	}

	if token == "" {
		c.setBasicAuth(req)
	} else {
		c.setBearerAuth(req, token)
	}

	resp, err := c.client.Do(ctx, req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return err
	}
	if err = c.checkResponse(resp); err != nil {
		return err
	}
	if targetObject == nil {
		_, err := io.Copy(ioutil.Discard, resp.Body)
		return err
	}

	return json.NewDecoder(resp.Body).Decode(targetObject)
}

func (c *serviceClient) setBasicAuth(req *http.Request) {
	if c.serviceURL.User == nil {
		return
	}

	password, _ := c.serviceURL.User.Password()
	req.SetBasicAuth(c.serviceURL.User.Username(), password)
}

func (c *serviceClient) setBearerAuth(req *http.Request, token string) {
	req.Header.Add("Authentication", fmt.Sprintf("Bearer %s", token))
}

func (c *serviceClient) checkResponse(resp *http.Response) error {
	if code := resp.StatusCode; 200 <= code && code <= 299 {
		return nil
	}
	return &httpx.HTTPError{Path: resp.Request.URL.String(), StatusCode: resp.StatusCode}
}
