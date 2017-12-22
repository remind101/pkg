package pubsub_test

import (
	"testing"

	. "github.com/remind101/pkg/pubsub"
	"golang.org/x/net/context"
)

func TestReceiveEventOnce(t *testing.T) {
	pub := NewPublisher()
	ch := pub.Subscribe()
	done := make(chan bool)
	received := []string{}

	go func() {
		for event := range ch {
			received = append(received, event.Payload.(string))
		}
		done <- true
	}()

	pub.EmitEvent(context.Background(), "test_event")
	pub.Close()

	// wait until goroutine is finished
	<-done

	if len(received) != 1 || received[0] != "test_event" {
		t.Fatalf("Expected to have received \"test_event\" exactly once.")
	}
}

func TestReceiveEventTwice(t *testing.T) {
	pub := NewPublisher()

	ch1 := pub.Subscribe()
	ch2 := pub.Subscribe()
	done := make(chan bool)

	received := []string{}

	go func() {
		for i := 0; i < 2; i++ {
			select {
			case event1 := <-ch1:
				received = append(received, event1.Payload.(string))
			case event2 := <-ch2:
				received = append(received, event2.Payload.(string))
			}
		}
		done <- true
	}()

	pub.EmitEvent(context.Background(), "test_event")
	pub.Close()

	// wait until goroutine is finished
	<-done

	if len(received) != 2 || received[0] != "test_event" || received[1] != "test_event" {
		t.Fatalf("Expected to have received \"test_event\" twice.")
	}
}
