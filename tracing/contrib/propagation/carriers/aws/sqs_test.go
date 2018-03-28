package aws

import (
	"testing"

	dd_opentracing "github.com/DataDog/dd-trace-go/opentracing"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func TestSQSCarrier(t *testing.T) {
	config := dd_opentracing.NewConfiguration()
	tracer, _, err := dd_opentracing.NewTracer(config)
	opentracing.SetGlobalTracer(tracer)

	span := opentracing.StartSpan("test span")
	m := &sqs.Message{}
	err = opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.TextMap,
		SQSCarrier(m),
	)
	if err != nil {
		t.Error(err)
		return
	}
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.TextMap,
		SQSCarrier(m),
	)
	if err != nil {
		t.Error(err)
		return
	}
	newSpan := opentracing.StartSpan("newSpan", ext.RPCServerOption(wireContext))
	if newSpan.(*dd_opentracing.Span).Span.ParentID != span.(*dd_opentracing.Span).Span.SpanID {
		t.Error("ParentID didn't match original spanID")
		return
	}
}

func TestSNSCarrier(t *testing.T) {
	config := dd_opentracing.NewConfiguration()
	tracer, _, err := dd_opentracing.NewTracer(config)
	opentracing.SetGlobalTracer(tracer)

	span := opentracing.StartSpan("test span")
	m := &sns.PublishInput{}
	err = opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.TextMap,
		SNSCarrier(m),
	)
	if err != nil {
		t.Error(err)
		return
	}
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.TextMap,
		SNSCarrier(m),
	)
	if err != nil {
		t.Error(err)
		return
	}
	newSpan := opentracing.StartSpan("newSpan", ext.RPCServerOption(wireContext))
	if newSpan.(*dd_opentracing.Span).Span.ParentID != span.(*dd_opentracing.Span).Span.SpanID {
		t.Error("ParentID didn't match original spanID")
		return
	}
}
