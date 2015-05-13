
# middleware
    import "github.com/remind101/pkg/httpx/middleware"





## Variables
``` go
var DefaultErrorHandler = func(err error, w http.ResponseWriter, r *http.Request) {
    http.Error(w, err.Error(), http.StatusInternalServerError)
}
```
DefaultErrorHandler is an error handler that will respond with the error
message and a 500 status.

``` go
var DefaultGenerator = context.Background
```
DefaultGenerator is the default context generator. Defaults to just use
context.Background().

``` go
var DefaultRequestIDExtractor = HeaderExtractor([]string{"X-Request-Id", "Request-Id"})
```
DefaultRequestIDExtractor is the default function to use to extract a request
id from an http.Request.

``` go
var StdoutLogger = stdLogger(os.Stdout)
```
StdoutLogger is a logger.Logger generator that generates a logger that writes
to stdout.


## func HeaderExtractor
``` go
func HeaderExtractor(headers []string) func(*http.Request) string
```
HeaderExtractor returns a function that can extract a request id from a list
of headers.


## func InsertLogger
``` go
func InsertLogger(h httpx.Handler, f func(context.Context, *http.Request) logger.Logger) httpx.Handler
```
InsertLogger returns an httpx.Handler middleware that will call f to generate
a logger, then insert it into the context.


## func LogTo
``` go
func LogTo(h httpx.Handler, f func(context.Context, *http.Request) logger.Logger) httpx.Handler
```
LogTo is an httpx middleware that wraps the handler to insert a logger and
log the request to it.



## type Background
``` go
type Background struct {
    // Generate will be called to generate a context.Context for the
    // request.
    Generate func() context.Context
    // contains filtered or unexported fields
}
```
Background is middleware that implements the http.Handler interface to inject
an initial context object. Use this as the entry point from an http.Handler
server.









### func BackgroundContext
``` go
func BackgroundContext(h httpx.Handler) *Background
```



### func (\*Background) ServeHTTP
``` go
func (h *Background) ServeHTTP(w http.ResponseWriter, r *http.Request)
```
ServeHTTP implements the http.Handler interface.



### func (\*Background) ServeHTTPContext
``` go
func (h *Background) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error
```


## type Error
``` go
type Error struct {
    // ErrorHandler is a function that will be called when a handler returns
    // an error.
    ErrorHandler func(error, http.ResponseWriter, *http.Request)
    // contains filtered or unexported fields
}
```
Error is an httpx.Handler that will handle errors with an ErrorHandler.









### func HandleError
``` go
func HandleError(h httpx.Handler, f func(error, http.ResponseWriter, *http.Request)) *Error
```
HandleError returns a new Error middleware that uses f as the ErrorHandler.


### func NewError
``` go
func NewError(h httpx.Handler) *Error
```



### func (\*Error) ServeHTTPContext
``` go
func (h *Error) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error
```
ServeHTTPContext implements the httpx.Handler interface.



## type Logger
``` go
type Logger struct {
    // contains filtered or unexported fields
}
```
Logger is middleware that logs the request details to the logger.Logger
embedded within the context.









### func Log
``` go
func Log(h httpx.Handler) *Logger
```



### func (\*Logger) ServeHTTPContext
``` go
func (h *Logger) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error
```


## type Recovery
``` go
type Recovery struct {
    // Reporter is a Reporter that will be inserted into the context. It
    // will also be used to report panics.
    reporter.Reporter
    // contains filtered or unexported fields
}
```
Recovery is a middleware that will recover from panics and return the error.









### func Recover
``` go
func Recover(h httpx.Handler, r reporter.Reporter) *Recovery
```



### func (\*Recovery) ServeHTTPContext
``` go
func (h *Recovery) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error)
```
ServeHTTPContext implements the httpx.Handler interface. It recovers from
panics and returns an error for upstream middleware to handle.



## type RequestID
``` go
type RequestID struct {
    // Extractor is a function that can extract a request id from an
    // http.Request. The zero value is a function that will pull a request
    // id from the `X-Request-ID` or `Request-ID` headers.
    Extractor func(*http.Request) string
    // contains filtered or unexported fields
}
```
RequestID is middleware that extracts a request id from the headers and
inserts it into the context.









### func ExtractRequestID
``` go
func ExtractRequestID(h httpx.Handler) *RequestID
```



### func (\*RequestID) ServeHTTPContext
``` go
func (h *RequestID) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error
```
ServeHTTPContext implements the httpx.Handler interface. It extracts a
request id from the headers and inserts it into the context.



## type ResponseWriter
``` go
type ResponseWriter interface {
    http.ResponseWriter
    http.Flusher
    // Status returns the status code of the response or 0 if the response has not been written.
    Status() int
}
```
ResponseWriter is a wrapper around http.ResponseWriter that provides extra information about
the response.









### func NewResponseWriter
``` go
func NewResponseWriter(rw http.ResponseWriter) ResponseWriter
```
NewResponseWriter creates a ResponseWriter that wraps an http.ResponseWriter










- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)