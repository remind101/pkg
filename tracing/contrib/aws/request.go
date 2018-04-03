package aws

import (
	"context"

	dd_opentracing "github.com/DataDog/dd-trace-go/opentracing"
	"github.com/aws/aws-sdk-go/aws/request"
	opentracing "github.com/opentracing/opentracing-go"
)

// Send wraps the aws request with an opentracing span.
//
// req, output := dynamoClient.PutItemRequest(input)
// err := aws.Send(ctx, req, func(s opentracing.Span){
//     s.SetTag("span.type", "db")
//     s.SetTag("service.name", "dynamodb")
// })
func Send(ctx context.Context, r *request.Request, fn func(opentracing.Span)) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "client.request")
	defer span.Finish()
	r.HTTPRequest = r.HTTPRequest.WithContext(ctx)

	span.SetTag(dd_opentracing.ResourceName, r.Operation.Name)
	span.SetTag("http.method", r.Operation.HTTPMethod)
	span.SetTag("http.url", r.ClientInfo.Endpoint+r.Operation.HTTPPath)
	span.SetTag("out.host", r.ClientInfo.Endpoint)
	span.SetTag("aws.operation", r.Operation.Name)

	// Set additional tags with the given function, for example:
	//
	// span.SetTag(dd_opentracing.SpanType, "db")
	// span.SetTag(dd_opentracing.ServiceName, "dynamodb.myapp")
	fn(span)

	err := r.Send()

	span.SetTag("aws.retry_count", r.RetryCount)

	if r.HTTPResponse != nil {
		span.SetTag("http.status_code", r.HTTPResponse.StatusCode)
	}

	if err != nil {
		span.SetTag(dd_opentracing.Error, err)
	}

	return err
}
