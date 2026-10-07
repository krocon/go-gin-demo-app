package tasks

import "sync"

// subscriberBuffer is how many events a slow subscriber may lag behind
// before further events are dropped for it.
const subscriberBuffer = 64

// EventBus is a tiny in-process pub/sub built on channels:
// routes publish events, every connected WebSocket / SSE client subscribes.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[chan Event]struct{}
}

func NewEventBus() *EventBus {
	return &EventBus{subscribers: make(map[chan Event]struct{})}
}

// Subscribe returns a channel that receives every future event and a function
// to unsubscribe. Always call unsubscribe (typically with defer) when done.
func (b *EventBus) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, subscriberBuffer)

	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			b.mu.Lock()
			delete(b.subscribers, ch)
			close(ch)
			b.mu.Unlock()
		})
	}
	return ch, unsubscribe
}

// Publish sends the event to all subscribers. It never blocks:
// if a subscriber's buffer is full, that subscriber misses the event.
func (b *EventBus) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- e:
		default: // slow consumer – drop instead of blocking everyone
		}
	}
}
