// Package eventbus provides a high-performance type-safe event bus
// for Go with generics support.
//
// The package implements the Publisher-Subscriber pattern with the following
// key features:
//   - Type safety through Go 1.18+ generics
//   - Non-blocking event delivery
//   - Thread-safe implementation
//   - Automatic unsubscription on context cancellation
//
// Example usage:
//
//	type MyEvent struct {
//	    Message string
//	}
//
//	bus := events.NewEventBus()
//	ctx := context.Background()
//
//	ch, unsubscribe := events.Subscribe[MyEvent](bus, ctx, 10)
//	defer unsubscribe()
//
//	go func() {
//	    for event := range ch {
//	        fmt.Println("Received:", event.Message)
//	    }
//	}()
//
//	bus.Publish(ctx, MyEvent{Message: "Hello, World!"})
package eventbus

import (
	"context"
	"reflect"
	"sync"
)

// EventBus represents the event bus interface for publishing events.
// Event subscription is done through the Subscribe[T] function.
type EventBus interface {
	// Publish publishes an event to all subscribers of the corresponding type.
	// Uses non-blocking delivery - if a subscriber's channel is full,
	// the event will be dropped, preventing publisher blocking.
	//
	// Parameters:
	//   ctx - context for operation cancellation
	//   ev - event to publish (any type)
	//
	// If ev is nil, the publication is ignored.
	Publish(ctx context.Context, ev any)

	// subscribe - internal method for subscribing to events of a specific type.
	// Used through the typed Subscribe[T] function.
	subscribe(ctx context.Context, eventType reflect.Type, buf int) (any, func())
}

// Subscribe subscribes to events of type T and returns a channel for receiving
// events and an unsubscribe function.
//
// Parameters:
//
//	bus - event bus to subscribe to
//	ctx - context for automatic unsubscription on cancellation
//	buf - channel buffer size for events
//
// Returns:
//
//	<-chan T - read-only channel for receiving events of type T
//	func() - unsubscribe function that closes the channel and removes the subscriber
//
// When the context is cancelled, unsubscription is called automatically.
// The unsubscribe function can be called multiple times safely.
//
// Example:
//
//	ch, unsubscribe := Subscribe[MyEvent](bus, ctx, 100)
//	defer unsubscribe()
//
//	for event := range ch {
//	    // handle event
//	}
func Subscribe[T any](bus EventBus, ctx context.Context, buf int) (<-chan T, func()) {
	var zero T
	eventType := reflect.TypeOf(zero)

	ch, unsubscribe := bus.subscribe(ctx, eventType, buf)
	return ch.(chan T), unsubscribe
}

// eventBusImpl represents the concrete implementation of the event bus.
// Uses a map to store subscribers by event types and a mutex
// to ensure thread safety.
type eventBusImpl struct {
	mu          sync.RWMutex                  // mutex to protect the subscribers map
	subscribers map[reflect.Type][]subscriber // map of subscribers by event types
}

// subscriber represents a single subscriber to events of a specific type.
type subscriber struct {
	ch     any    // channel for sending events (typed through reflection)
	cancel func() // unsubscribe function
}

// NewEventBus creates and returns a new event bus.
// The returned bus is ready to use and is thread-safe.
//
// Example:
//
//	bus := NewEventBus()
//	// bus is ready to use
func NewEventBus() EventBus {
	return &eventBusImpl{
		subscribers: make(map[reflect.Type][]subscriber),
	}
}

// Publish publishes an event to all subscribers of the corresponding type.
//
// The implementation uses the following performance optimizations:
//   - Copying the subscriber list to minimize lock time
//   - Non-blocking sending through reflect.Value.TrySend
//   - Panic handling when sending to closed channels
//   - Context cancellation check before sending
//
// If the event is nil, the method returns without action.
// If a subscriber's channel is full, the event is dropped without blocking the publisher.
func (eb *eventBusImpl) Publish(ctx context.Context, ev any) {
	if ev == nil {
		return
	}

	eventType := reflect.TypeOf(ev)

	eb.mu.RLock()
	subs, exists := eb.subscribers[eventType]
	if !exists {
		eb.mu.RUnlock()
		return
	}

	// Create a copy of subscribers for safe iteration without prolonged mutex locking.
	// This is a key pattern for high performance.
	subscribersCopy := make([]subscriber, len(subs))
	copy(subscribersCopy, subs)
	eb.mu.RUnlock()

	// Send event to all subscribers from the copy
	for _, sub := range subscribersCopy {
		// Check if context is cancelled before attempting to send
		select {
		case <-ctx.Done():
			return
		default:
			// Use anonymous function with recover for safe sending.
			// This prevents panic if a subscriber unsubscribed and closed
			// their channel right after we copied the list,
			// but before we tried to send the event to it.
			// This is a known and expected race condition in this architecture.
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Panic when sending to a closed channel ("send on closed channel")
						// is expected in this race condition. We catch it to
						// prevent crashing the entire application. Debug-level logging
						// can be added if needed for debugging.
					}
				}()

				// Non-blocking sending through reflection.
				// If a subscriber's channel is full, the event will be dropped,
				// preventing the entire system from blocking due to one slow subscriber.
				chValue := reflect.ValueOf(sub.ch)

				// Additional check: if the channel is closed, TrySend will return false
				// without panic, but only if the channel is not nil
				if chValue.IsValid() && !chValue.IsNil() {
					eventValue := reflect.ValueOf(ev)
					chValue.TrySend(eventValue)
				}
			}()
		}
	}
}

// subscribe subscribes to events of a specific type (internal method).
//
// Creates a typed channel through reflection, registers the subscriber,
// and sets up automatic unsubscription on context cancellation.
//
// The unsubscribe function uses sync.Once to prevent multiple
// calls and ensures proper channel closure and subscriber removal
// from the list.
//
// Parameters:
//
//	ctx - context for automatic unsubscription
//	eventType - event type to subscribe to (obtained through reflection)
//	buf - channel buffer size
//
// Returns:
//
//	any - channel interface (cast to type in Subscribe[T])
//	func() - unsubscribe function
func (eb *eventBusImpl) subscribe(ctx context.Context, eventType reflect.Type, buf int) (any, func()) {
	// Create channel for events
	chType := reflect.ChanOf(reflect.BothDir, eventType)
	ch := reflect.MakeChan(chType, buf)

	// Create unsubscribe function
	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			eb.mu.Lock()
			defer eb.mu.Unlock()

			subs, exists := eb.subscribers[eventType]
			if !exists {
				ch.Close()
				return
			}

			// Remove subscriber from the list
			for i, sub := range subs {
				if reflect.ValueOf(sub.ch).Pointer() == ch.Pointer() {
					eb.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
					break
				}
			}

			// If no more subscribers, remove type from map
			if len(eb.subscribers[eventType]) == 0 {
				delete(eb.subscribers, eventType)
			}

			ch.Close()
		})
	}

	// Add subscriber
	eb.mu.Lock()
	eb.subscribers[eventType] = append(eb.subscribers[eventType], subscriber{
		ch:     ch.Interface(),
		cancel: unsubscribe,
	})
	eb.mu.Unlock()

	// Start goroutine for unsubscription on context cancellation
	go func() {
		<-ctx.Done()
		unsubscribe()
	}()

	return ch.Interface(), unsubscribe
}
