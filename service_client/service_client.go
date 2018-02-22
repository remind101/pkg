package service_client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	httpsignatures "github.com/99designs/httpsignatures-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/remind101/pkg/httpx"
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

type ServiceClient interface {
	Do(ctx context.Context, method, path string, jsonData interface{}, targetObject interface{}) error
	DoWithBearerAuth(ctx context.Context, method, path, access_token string, jsonData, targetObject interface{}) error
}

type serviceClient struct {
	serviceURL       *url.URL
	client           *httpx.Client
	signer           *httpsignatures.Signer
	keyId            string
	key              string
	forwardedHeaders []string
	logScrubber      Scrubber
}

type ServiceClientOpts struct {
	IncludeForwardedHeaders []string
	SigningKeyId            string
	SigningKey              string
	Scrubber                Scrubber
}

func NewServiceClient(serviceURL string) *serviceClient {
	return NewServiceClientWithOpts(serviceURL, ServiceClientOpts{})
}

func NewServiceClientWithOpts(serviceURL string, opts ServiceClientOpts) *serviceClient {
	u, err := url.Parse(serviceURL)
	if err != nil {
		panic(err)
	}
	httpClient := &http.Client{Transport: AggressiveTransport}
	client := httpx.NewServiceClient(serviceURL, httpClient)
	signer := httpsignatures.DefaultSha256Signer
	if opts.Scrubber == nil {
		opts.Scrubber = &NoopScrubber{}
	}

	return &serviceClient{
		serviceURL:       u,
		client:           client,
		signer:           signer,
		keyId:            opts.SigningKeyId,
		key:              opts.SigningKey,
		forwardedHeaders: opts.IncludeForwardedHeaders,
		logScrubber:      opts.Scrubber,
	}
}

// DoRequest performs the request and optionally will decode a json response into the targetObject.
// TODO use request.Context() instead of passing in a context.
func (c *serviceClient) DoRequest(ctx context.Context, req *http.Request, targetObject interface{}) error {
	req = req.WithContext(ctx)

	// Sign the request.
	if c.key != "" {
		c.setHttpSignature(req)
	}

	// Add forwarded headers.
	for _, header := range c.forwardedHeaders {
		req.Header.Add(header, httpx.Header(req.Context(), header))
	}

	// Add span to trace
	traceCloser := c.trace(req.Context(), req)

	resp, err := c.client.Do(req.Context(), req)
	traceCloser()
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
		// It is important to always read the response body. Otherwise the TCP
		// connection cannot be re-used for keep-alives.
		_, err := io.Copy(ioutil.Discard, resp.Body)
		return err
	}

	return json.NewDecoder(resp.Body).Decode(targetObject)
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

	return c.DoRequest(ctx, req, targetObject)
}

func (c *serviceClient) setBasicAuth(req *http.Request) {
	if c.serviceURL.User == nil {
		return
	}

	password, _ := c.serviceURL.User.Password()
	req.SetBasicAuth(c.serviceURL.User.Username(), password)
}

func (c *serviceClient) trace(ctx context.Context, req *http.Request) func() {
	span, ctx := opentracing.StartSpanFromContext(ctx, "client.request")
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))
	span.SetTag("uri", c.logScrubber.Scrub(req.URL.String()))
	return func() {
		span.Finish()
	}
}

func (c *serviceClient) setHttpSignature(req *http.Request) error {
	err := c.signer.SignRequest(c.keyId, c.key, req)
	if err != nil {
		return err
	}
	return nil
}

func (c *serviceClient) setBearerAuth(req *http.Request, token string) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
}

func (c *serviceClient) checkResponse(resp *http.Response) error {
	if code := resp.StatusCode; 200 <= code && code <= 299 {
		return nil
	}
	return &httpx.HTTPError{Path: resp.Request.URL.String(), StatusCode: resp.StatusCode}
}
