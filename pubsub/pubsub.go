package pubsub

import (
	"sync"

	"golang.org/x/net/context"
)

// Publisher defines the sender contract of a basic publish/subscribe design pattern.
type Publisher struct {
	sync.Mutex
	subChannels []chan *Event
}

type Event struct {
	Context context.Context
	Payload interface{}
}

func NewPublisher() *Publisher {
	return &Publisher{
		subChannels: []chan *Event{},
	}
}

// A single Publisher instance is able to take several subscribers.
// It returns a receive-only channel.
func (pub *Publisher) Subscribe() <-chan *Event {
	ch := make(chan *Event)
	pub.Lock()
	pub.subChannels = append(pub.subChannels, ch)
	pub.Unlock()
	return ch
}

// It closes all active channels.
func (pub *Publisher) Close() {
	pub.Lock()
	defer pub.Unlock()
	for _, ch := range pub.subChannels {
		close(ch)
	}
	pub.subChannels = []chan *Event{}
}

func (pub *Publisher) EmitEvent(ctx context.Context, payload interface{}) {
	pub.Lock()
	defer pub.Unlock()
	for _, ch := range pub.subChannels {
		ch <- &Event{Context: ctx, Payload: payload}
	}
}
