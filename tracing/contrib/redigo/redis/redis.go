// COPIED from https://github.com/DataDog/dd-trace-go/tree/master/contrib/garyburd/redigo
// and modified to use opentracing.
//
// TODO: Remove once dd-trace-go switches their contrib packages to use opentracing.
package redis

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/opentracing/opentracing-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

// DialURL connects to a Redis server at the given URL using the Redis
// URI scheme. URLs should follow the draft IANA specification for the
// scheme (https://www.iana.org/assignments/uri-schemes/prov/redis).
// The returned redis.Conn is traced.
func DialURL(rawurl string, options ...interface{}) (redis.Conn, error) {
	dialOpts, cfg := parseOptions(options...)
	u, err := url.Parse(rawurl)
	if err != nil {
		return Conn{}, err
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
		port = "6379"
	}
	if host == "" {
		host = "localhost"
	}
	network := "tcp"
	c, err := redis.DialURL(rawurl, dialOpts...)
	tc := Conn{c, &params{cfg, network, host, port}}
	return tc, err
}

// params contains fields and metadata useful for command tracing
type params struct {
	config  *dialConfig
	network string
	host    string
	port    string
}

// Conn is an implementation of the redis.Conn interface that supports tracing
type Conn struct {
	redis.Conn
	*params
}

// Do wraps redis.Conn.Do. It sends a command to the Redis server and returns the received reply.
// In the process it emits a span containing key information about the command sent.
// When passed a context.Context as the final argument, Do will ensure that any span created
// inherits from this context. The rest of the arguments are passed through to the Redis server unchanged.
func (tc Conn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	var (
		ctx context.Context
		ok  bool
	)
	if n := len(args); n > 0 {
		ctx, ok = args[n-1].(context.Context)
		if ok {
			args = args[:n-1]
		}
	}

	if ctx == nil {
		ctx = context.Background()
	}

	span := tc.newChildSpan(ctx)
	defer func() {
		if err != nil {
			span.SetTag(ext.Error, err)
		}
		span.Finish()
	}()

	span.SetTag("redis.args_length", strconv.Itoa(len(args)))

	if len(commandName) > 0 {
		span.SetTag(ext.ResourceName, commandName)
	} else {
		// When the command argument to the Do method is "", then the Do method will flush the output buffer
		// See https://pkg.go.dev/github.com/gomodule/redigo/redis#hdr-Pipelining
		span.SetTag(ext.ResourceName, "conn.flush")
	}
	var b bytes.Buffer
	b.WriteString(commandName)
	for _, arg := range args {
		b.WriteString(" ")
		switch arg := arg.(type) {
		case string:
			b.WriteString(arg)
		case int:
			b.WriteString(strconv.Itoa(arg))
		case int32:
			b.WriteString(strconv.FormatInt(int64(arg), 10))
		case int64:
			b.WriteString(strconv.FormatInt(arg, 10))
		case fmt.Stringer:
			b.WriteString(arg.String())
		}
	}
	span.SetTag("redis.command", b.String())
	return tc.Conn.Do(commandName, args...)
}

// newChildSpan creates a span inheriting from the given context. It adds to the span useful metadata about the traced Redis connection
func (tc Conn) newChildSpan(ctx context.Context) opentracing.Span {
	p := tc.params
	span, _ := opentracing.StartSpanFromContext(ctx, "redis.command")
	span.SetTag(ext.ServiceName, p.config.serviceName)
	span.SetTag(ext.SpanType, "cache")
	span.SetTag("out.network", p.network)
	span.SetTag("out.port", p.port)
	span.SetTag("out.host", p.host)

	return span
}

// parseOptions parses a set of arbitrary options (which can be of type redis.DialOption
// or the local DialOption) and returns the corresponding redis.DialOption set as well as
// a configured dialConfig.
func parseOptions(options ...interface{}) ([]redis.DialOption, *dialConfig) {
	dialOpts := []redis.DialOption{}
	cfg := new(dialConfig)
	defaults(cfg)
	for _, opt := range options {
		switch o := opt.(type) {
		case redis.DialOption:
			dialOpts = append(dialOpts, o)
		case DialOption:
			o(cfg)
		}
	}
	return dialOpts, cfg
}
