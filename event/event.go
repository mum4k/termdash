// Copyright 2019 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package event provides a non-blocking event distribution and subscription
// system.
package event

import (
	"context"
	"reflect"
	"sync"

	"github.com/mum4k/termdash/event/eventqueue"
	"github.com/mum4k/termdash/terminalapi"
)

// Callback is a function provided by an event subscriber.
// It gets called with each event that passed the subscription filter.
// Implementations must be thread-safe, events come from a separate goroutine.
// Implementation should be light-weight, otherwise a slow-processing
// subscriber can build a long tail of events.
type Callback func(terminalapi.Event)

// subscriber represents a single subscriber.
type subscriber struct {
	// cb is the callback the subscriber receives events on.
	cb Callback

	// filter filters events towards the subscriber.
	// An empty filter receives all events.
	filter map[reflect.Type]bool

	// queue is a queue of events towards the subscriber.
	queue *eventqueue.Unbound

	// cancel when called terminates the goroutine that forwards events towards
	// this subscriber.
	cancel context.CancelFunc
}

// newSubscriber creates a new event subscriber.
func newSubscriber(filter []terminalapi.Event, cb Callback) *subscriber {
	f := map[reflect.Type]bool{}
	for _, ev := range filter {
		f[reflect.TypeOf(ev)] = true
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &subscriber{
		cb:     cb,
		filter: f,
		queue:  eventqueue.New(),
		cancel: cancel,
	}

	// Terminates when stop() is called.
	go s.run(ctx)
	return s
}

// run periodically forwards events towards the subscriber.
// Terminates when the context expires.
func (s *subscriber) run(ctx context.Context) {
	for {
		ev := s.queue.Pull(ctx)
		if ev != nil {
			s.cb(ev)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

// event forwards an event to the subscriber.
func (s *subscriber) event(ev terminalapi.Event) {
	if len(s.filter) == 0 {
		s.queue.Push(ev)
	}

	t := reflect.TypeOf(ev)
	if s.filter[t] {
		s.queue.Push(ev)
	}
}

// stop stops the event subscriber.
func (s *subscriber) stop() {
	s.cancel()
	s.queue.Close()
}

// DistributionSystem distributes events to subscribers.
//
// Subscribers can request filtering of events they get based on event type or
// subscribe to all events.
//
// The distribution system maintains a queue towards each subscriber, making
// sure that a single slow subscriber only slows itself down, rather than the
// entire application.
//
// This object is thread-safe.
type DistributionSystem struct {
	// subscribers subscribe to events.
	// maps subscriber id to subscriber.
	subscribers map[int]*subscriber

	// nextID is id for the next subscriber.
	nextID int

	// mu protects the distribution system.
	mu sync.RWMutex
}

// NewDistributionSystem creates a new event distribution system.
func NewDistributionSystem() *DistributionSystem {
	return &DistributionSystem{
		subscribers: map[int]*subscriber{},
	}
}

// Event should be called with events coming from the terminal.
// The distribution system will distribute these to all the subscribers.
func (eds *DistributionSystem) Event(ev terminalapi.Event) {
	eds.mu.RLock()
	defer eds.mu.RUnlock()

	for _, sub := range eds.subscribers {
		sub.event(ev)
	}
}

// StopFunc when called unsubscribes the subscriber from all events and
// releases resources tied to the subscriber.
type StopFunc func()

// Subscribe subscribes to events according to the filter.
// An empty filter indicates that the subscriber wishes to receive events of
// all kinds. If the filter is non-empty, only events of the provided type will
// be sent to the subscriber.
// Returns a function that allows the subscriber to unsubscribe.
func (eds *DistributionSystem) Subscribe(filter []terminalapi.Event, cb Callback) StopFunc {
	eds.mu.Lock()
	defer eds.mu.Unlock()

	id := eds.nextID
	eds.nextID++
	sub := newSubscriber(filter, cb)
	eds.subscribers[id] = sub

	return func() {
		eds.mu.Lock()
		defer eds.mu.Unlock()

		sub.stop()
		delete(eds.subscribers, id)
	}
}
