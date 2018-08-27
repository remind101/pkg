package redis

import (
	"context"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"github.com/go-redis/redis"
	"github.com/opentracing/opentracing-go"
)

func WrapClient(ctx context.Context, c *redis.Client, serviceName string) *redis.Client {
	if ctx == nil {
		return c
	}
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan == nil {
		return c
	}

	copy := c.WithContext(c.Context())
	copy.WrapProcess(func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			span, _ := opentracing.StartSpanFromContext(ctx, "redis.command")
			span.SetTag(ext.ServiceName, serviceName)
			span.SetTag(ext.SpanType, "cache")
			span.SetTag(ext.ResourceName, cmd.Name())
			span.SetTag("redis.command", cmd.String())
			defer span.Finish()

			err := oldProcess(cmd)
			if err != nil && err != redis.Nil {
				span.SetTag(ext.Error, err)
			}
			return err
		}
	})
	return copy
}

func NewClient(connstr string) *redis.Client {
	opt, err := redis.ParseURL(connstr)
	if err != nil {
		panic(err)
	}
	return redis.NewClient(opt)
}
