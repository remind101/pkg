package stream

import (
	"context"
	"os"
	"testing"
	"time"
)

func ExampleHeartbeat() {
	w := os.Stdout
	defer close(Heartbeat(w, time.Second)) // close to cleanup resources
}

func ExampleHeartbeatWithContext() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	w := os.Stdout
	defer close(HeartbeatWithContext(ctx, w, time.Second)) // close to cleanup resources
}

func TestHeartbeatWithContext(t *testing.T) {
	// Create a context that will be canceled after a short time
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start the heartbeat
	stop := HeartbeatWithContext(ctx, &mockWriter{}, 10*time.Millisecond)

	// Wait for the context to be canceled
	<-ctx.Done()

	// Give a little time for the goroutine to exit
	time.Sleep(20 * time.Millisecond)

	// Closing the stop channel should be safe (not block) if the goroutine has exited
	close(stop)
}

// mockWriter is a simple io.Writer implementation for testing
type mockWriter struct{}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
