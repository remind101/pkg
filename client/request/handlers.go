package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"

	"github.com/99designs/httpsignatures-go"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/remind101/pkg/httpx"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

// Handlers represents lists of request handlers for each phase in the request
// lifecycle.
type Handlers struct {
	Build            HandlerList
	Sign             HandlerList
	Send             HandlerList
	ValidateResponse HandlerList
	Decode           HandlerList
	DecodeError      HandlerList
	Complete         HandlerList
}

// DefaultHandlers defines a basic request configuration that assumes JSON
// requests and responses.
func DefaultHandlers() Handlers {
	return Handlers{
		Build:            NewHandlerList(JSONBuilder),
		Sign:             NewHandlerList(),
		Send:             NewHandlerList(WithTracing(BaseSender)),
		ValidateResponse: NewHandlerList(),
		Decode:           NewHandlerList(JSONDecoder),
		DecodeError:      NewHandlerList(),
		Complete:         NewHandlerList(),
	}
}

// Copy returns a copy of a Handlers instance.
func (h Handlers) Copy() Handlers {
	return Handlers{
		Build:            h.Build.copy(),
		Sign:             h.Sign.copy(),
		Send:             h.Send.copy(),
		ValidateResponse: h.ValidateResponse.copy(),
		Decode:           h.Decode.copy(),
		DecodeError:      h.DecodeError.copy(),
		Complete:         h.Complete.copy(),
	}
}

// HandlerList manages a list of request handlers.
type HandlerList struct {
	list []Handler
}

// NewHandlerList constructs a new HandlerList with the given handlers.
func NewHandlerList(hh ...Handler) HandlerList {
	return HandlerList{
		list: append([]Handler{}, hh...),
	}
}

// Run calls each request handler in order.
func (hl *HandlerList) Run(r *Request) {
	for _, h := range hl.list {
		h.Fn(r)
	}
}

// Append adds a handler to the end of the list.
func (hl *HandlerList) Append(h Handler) {
	hl.list = append(hl.list, h)
}

// Prepend adds a handler to the front of the list.
func (hl *HandlerList) Prepend(h Handler) {
	hl.list = append([]Handler{h}, hl.list...)
}

// Clear truncates a handler list.
func (hl *HandlerList) Clear() {
	hl.list = []Handler{}
}

func (hl *HandlerList) copy() HandlerList {
	n := HandlerList{}
	if len(hl.list) == 0 {
		return n
	}

	n.list = append(make([]Handler, 0, len(hl.list)), hl.list...)
	return n
}

// Handler defines a request handler.
type Handler struct {
	Name string
	Fn   func(*Request)
}

// BaseSender sends a request using the http.Client.
var BaseSender = Handler{
	Name: "BaseSender",
	Fn: func(r *Request) {
		var err error
		r.HTTPResponse, err = r.HTTPClient.Do(r.HTTPRequest)
		if err != nil {
			handleSendError(r, err)
		}
	},
}

var reStatusCode = regexp.MustCompile(`^(\d{3})`)

func handleSendError(r *Request, err error) {
	// Prevent leaking if an HTTPResponse was returned.
	if r.HTTPResponse != nil {
		r.HTTPResponse.Body.Close()
	}

	// Capture the case where url.Error is returned for error processing
	// response. e.g. 301 without location header comes back as string
	// error and r.HTTPResponse is nil. Other URL redirect errors will
	// comeback in a similar method.
	if e, ok := err.(*url.Error); ok && e.Err != nil {
		if s := reStatusCode.FindStringSubmatch(e.Err.Error()); s != nil {
			code, _ := strconv.ParseInt(s[1], 10, 64)
			r.HTTPResponse = &http.Response{
				StatusCode: int(code),
				Status:     http.StatusText(int(code)),
				Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
			}
			return
		}
	}

	if r.HTTPResponse == nil {
		// Add a dummy request response object to ensure the HTTPResponse
		// value is consistent.
		r.HTTPResponse = &http.Response{
			StatusCode: int(0),
			Status:     http.StatusText(int(0)),
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
		}
	}

	// Catch all other request errors.
	r.Error = errors.Wrap(err, "send request failed")
}

// JSONBuilder adds standard JSON headers to the request and encodes Params
// as JSON to the request body if the request is not a GET.
var JSONBuilder = Handler{
	Name: "JSONBuilder",
	Fn: func(r *Request) {
		r.HTTPRequest.Header.Set("Content-Type", "application/json")
		r.HTTPRequest.Header.Set("Accept", "application/json")

		if r.HTTPRequest.Method != "GET" && r.Params != nil {
			raw, err := json.Marshal(r.Params)
			if err != nil {
				r.Error = err
				return
			}
			r.HTTPRequest.ContentLength = int64(len(raw))
			r.HTTPRequest.Body = ioutil.NopCloser(bytes.NewReader(raw))
		}
	},
}

// JSONDecoder decodes a response as JSON.
var JSONDecoder = Handler{
	Name: "JSONDecoder",
	Fn: func(r *Request) {
		if r.HTTPResponse == nil {
			return
		}
		if r.HTTPResponse.Body != nil {
			defer func() {
				// Read the entire body, including any trailing garbage, so
				// this connection is in a good state for reuse.
				io.Copy(ioutil.Discard, r.HTTPResponse.Body)
				r.HTTPResponse.Body.Close()
			}()
		}
		if r.Data == nil {
			return
		}
		decoder := json.NewDecoder(r.HTTPResponse.Body)
		r.Error = decoder.Decode(r.Data)

		if decoder.More() && r.Error == nil {
			r.Error = fmt.Errorf("Response includes more than one JSON object")
		}
	},
}

// RequestSigner signs requests.
func RequestSigner(id, key string) Handler {
	return Handler{
		Name: "RequestSigner",
		Fn: func(r *Request) {
			r.Error = httpsignatures.DefaultSha256Signer.SignRequest(id, key, r.HTTPRequest)
		},
	}
}

// BasicAuther sets basic auth on a request.
func BasicAuther(username, password string) Handler {
	return Handler{
		Name: "BasicAuther",
		Fn: func(r *Request) {
			r.HTTPRequest.SetBasicAuth(username, password)
		},
	}
}

// HeadersFromContext adds headers with values from the request context.
// This is useful for forwarding headers to upstream services.
func HeadersFromContext(headers ...string) Handler {
	return Handler{
		Name: "HeadersFromContext",
		Fn: func(r *Request) {
			for _, header := range headers {
				r.HTTPRequest.Header.Add(header, httpx.Header(r.HTTPRequest.Context(), header))
			}
		},
	}
}

// RequestLogger dumps the entire request to stdout.
var RequestLogger = Handler{
	Name: "RequestLogger",
	Fn: func(r *Request) {
		b, err := httputil.DumpRequestOut(r.HTTPRequest, true)
		if err != nil {
			fmt.Printf("error dumping request: %s\n", err.Error())
			return
		}
		fmt.Println(string(b))
	},
}

// ResponseLogger dumps the entire response to stdout.
var ResponseLogger = Handler{
	Name: "ResponseLogger",
	Fn: func(r *Request) {
		b, err := httputil.DumpResponse(r.HTTPResponse, true)
		if err != nil {
			fmt.Printf("error dumping response: %s\n", err.Error())
			return
		}
		fmt.Println(string(b))
	},
}

// WithTracing returns a Send Handler that wraps another Send Handler in a trace
// span.
func WithTracing(h Handler) Handler {
	return Handler{
		Name: "TracedSender",
		Fn: func(r *Request) {
			span, ctx := opentracing.StartSpanFromContext(r.HTTPRequest.Context(), "client.request")
			opentracing.GlobalTracer().Inject(
				span.Context(),
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(r.HTTPRequest.Header),
			)
			defer span.Finish()
			r.HTTPRequest = r.HTTPRequest.WithContext(ctx)

			span.SetTag(ext.ResourceName, r.ClientInfo.ServiceName)
			span.SetTag("http.method", r.HTTPRequest.Method)
			span.SetTag("http.url", r.HTTPRequest.URL.Hostname()+r.HTTPRequest.URL.EscapedPath())
			span.SetTag("out.host", r.HTTPRequest.URL.Hostname())
			span.SetTag("out.port", r.HTTPRequest.URL.Port())

			h.Fn(r)

			if r.HTTPResponse != nil {
				span.SetTag("http.status_code", r.HTTPResponse.StatusCode)
			}

			if r.Error != nil {
				span.SetTag(ext.Error, r.Error)
			}
		},
	}
}
