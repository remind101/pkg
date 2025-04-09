// Package stream provides types that make it easier to perform streaming io.
package stream

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Heartbeat sends the null character periodically, to keep the connection alive.
// Deprecated: Use HeartbeatWithContext instead.
func Heartbeat(outStream io.Writer, interval time.Duration) chan struct{} {
	stop := make(chan struct{})
	t := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-t.C:
				fmt.Fprintf(outStream, "\x00")
				continue
			case <-stop:
				t.Stop()
				return
			}
		}
	}()

	return stop
}

// HeartbeatWithContext sends the null character periodically to keep the connection alive.
// The heartbeat will stop when the provided context is canceled or when the returned
// stop channel is closed.
func HeartbeatWithContext(ctx context.Context, outStream io.Writer, interval time.Duration) chan struct{} {
	stop := make(chan struct{})
	t := time.NewTicker(interval)

	go func() {
		defer t.Stop()
		for {
			select {
			case <-t.C:
				fmt.Fprintf(outStream, "\x00")
			case <-stop:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return stop
}
