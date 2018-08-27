package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/mocktracer"
)

func TestSQSCarrier(t *testing.T) {
	tracer := mocktracer.New()
	opentracing.SetGlobalTracer(tracer)

	span := opentracing.StartSpan("test span")
	m := &sqs.Message{}
	err := opentracing.GlobalTracer().Inject(
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
	if newSpan.(*mocktracer.MockSpan).ParentID != span.(*mocktracer.MockSpan).SpanContext.SpanID {
		t.Error("ParentID didn't match original spanID")
		return
	}
}

func TestSNSCarrier(t *testing.T) {
	tracer := mocktracer.New()
	opentracing.SetGlobalTracer(tracer)

	span := opentracing.StartSpan("test span")
	m := &sns.PublishInput{}
	err := opentracing.GlobalTracer().Inject(
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
	if newSpan.(*mocktracer.MockSpan).ParentID != span.(*mocktracer.MockSpan).SpanContext.SpanID {
		t.Error("ParentID didn't match original spanID")
		return
	}
}
