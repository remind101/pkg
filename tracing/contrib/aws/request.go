package aws

import (
	"context"
	"fmt"

	dd_ext "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"github.com/aws/aws-sdk-go/aws/request"
	opentracing "github.com/opentracing/opentracing-go"
)

// Option is a hook for adding span tags.
type Option func(opentracing.Span, *request.Request, error)

// Tracer holds tagging options for request tracing.
type Tracer struct {
	Opts []Option
}

// New returns a new AWS request tracer with default tagging.
func New(opts ...Option) *Tracer {
	opts = append([]Option{defaultTags}, opts...)
	return &Tracer{opts}
}

// Send wraps the aws request with an opentracing span.
//
// req, output := dynamoClient.PutItemRequest(input)
// err := t.Send(ctx, req)
func (t *Tracer) Send(ctx context.Context, r *request.Request, opts ...Option) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "client.request")
	defer span.Finish()

	r.HTTPRequest = r.HTTPRequest.WithContext(ctx)
	err := r.Send()

	opts = append(t.Opts, opts...)
	for _, fn := range opts {
		fn(span, r, err)
	}

	return err
}

func defaultTags(span opentracing.Span, r *request.Request, err error) {
	span.SetTag("resource.name", r.Operation.Name)
	span.SetTag("http.method", r.Operation.HTTPMethod)
	span.SetTag("http.url", r.ClientInfo.Endpoint+r.Operation.HTTPPath)
	span.SetTag("out.host", r.ClientInfo.Endpoint)
	span.SetTag("aws.operation", r.Operation.Name)
	span.SetTag("aws.retry_count", r.RetryCount)

	if r.HTTPResponse != nil {
		span.SetTag("http.status_code", r.HTTPResponse.StatusCode)
	}

	if err != nil {
		span.SetTag(dd_ext.Error, err)
		if _, ok := err.(fmt.Formatter); ok {
			span.SetTag(dd_ext.ErrorStack, fmt.Sprintf("%+v", err))
		}
	}
}
