package event

import (
	"context"
	"fmt"
	"image"
	"sync"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminalapi"
)

// receiver receives events from the distribution system.
type receiver struct {
	mu sync.Mutex

	// events are the received events.
	events []terminalapi.Event
}

// newReceiver returns a new event receiver.
func newReceiver() *receiver {
	return &receiver{}
}

// receive receives an event.
func (r *receiver) receive(ev terminalapi.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events = append(r.events, ev)
}

// getEvents returns the received events.
func (r *receiver) getEvents() map[terminalapi.Event]bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	res := map[terminalapi.Event]bool{}
	for _, ev := range r.events {
		res[ev] = true
	}
	return res
}

// waitFor waits until the receiver receives the specified number of events or
// the timeout.
// Returns the received events in an unspecified order.
func (r *receiver) waitFor(want int, timeout time.Duration) (map[terminalapi.Event]bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tick := time.NewTimer(5 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			ev := r.getEvents()

			switch got := len(ev); {
			case got > want:
				return nil, fmt.Errorf("got %d events %v, want %d", got, ev, want)

			case got == want:
				return ev, nil
			}

		case <-ctx.Done():
			ev := r.getEvents()
			return nil, fmt.Errorf("while waiting for events, got %d so far: %v, err: %v", len(ev), ev, ctx.Err())
		}
	}
}

// subscriberCase holds test case specifics for one subscriber.
type subscriberCase struct {
	// filter is the subscribers filter.
	filter []terminalapi.Event

	// rec receives the events.
	rec *receiver

	// want are the expected events that should be delivered to this subscriber.
	want map[terminalapi.Event]bool

	// wantErr asserts whether we want an error from waitFor.
	wantErr bool
}

func TestDistributionSystem(t *testing.T) {
	tests := []struct {
		desc string
		// events will be sent down the distribution system.
		events []terminalapi.Event

		// subCase are the event subscribers and their expectations.
		subCase []*subscriberCase
	}{
		{
			desc: "no events and no subscribers",
		},
		{
			desc: "events and no subscribers",
			events: []terminalapi.Event{
				&terminalapi.Mouse{},
				&terminalapi.Keyboard{},
			},
		},
		{
			desc: "single subscriber, wants all events and gets them",
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				&terminalapi.Mouse{Position: image.Point{1, 1}},
				&terminalapi.Resize{Size: image.Point{2, 2}},
				terminalapi.NewError("error"),
			},
			subCase: []*subscriberCase{
				{
					filter: nil,
					rec:    newReceiver(),
					want: map[terminalapi.Event]bool{
						&terminalapi.Keyboard{Key: keyboard.KeyEnter}:   true,
						&terminalapi.Mouse{Position: image.Point{1, 1}}: true,
						&terminalapi.Resize{Size: image.Point{2, 2}}:    true,
						terminalapi.NewError("error"):                   true,
					},
				},
			},
		},
		{
			desc: "single subscriber, filters events",
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				&terminalapi.Mouse{Position: image.Point{1, 1}},
				&terminalapi.Resize{Size: image.Point{2, 2}},
			},
			subCase: []*subscriberCase{
				{
					filter: []terminalapi.Event{
						&terminalapi.Keyboard{},
						&terminalapi.Mouse{},
					},
					rec: newReceiver(),
					want: map[terminalapi.Event]bool{
						&terminalapi.Keyboard{Key: keyboard.KeyEnter}:   true,
						&terminalapi.Mouse{Position: image.Point{1, 1}}: true,
					},
				},
			},
		},
		{
			desc: "single subscriber, errors are always received",
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				&terminalapi.Mouse{Position: image.Point{1, 1}},
				&terminalapi.Resize{Size: image.Point{2, 2}},
				terminalapi.NewError("error"),
			},
			subCase: []*subscriberCase{
				{
					filter: []terminalapi.Event{
						&terminalapi.Keyboard{},
						&terminalapi.Mouse{},
					},
					rec: newReceiver(),
					want: map[terminalapi.Event]bool{
						&terminalapi.Keyboard{Key: keyboard.KeyEnter}:   true,
						&terminalapi.Mouse{Position: image.Point{1, 1}}: true,
						terminalapi.NewError("error"):                   true,
					},
				},
			},
		},
		{
			desc: "multiple subscribers and events",
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				&terminalapi.Keyboard{Key: keyboard.KeyEsc},
				&terminalapi.Mouse{Position: image.Point{0, 0}},
				&terminalapi.Mouse{Position: image.Point{1, 1}},
				&terminalapi.Resize{Size: image.Point{1, 1}},
				&terminalapi.Resize{Size: image.Point{2, 2}},
				terminalapi.NewError("error1"),
				terminalapi.NewError("error2"),
			},
			subCase: []*subscriberCase{
				{
					filter: []terminalapi.Event{
						&terminalapi.Keyboard{},
					},
					rec: newReceiver(),
					want: map[terminalapi.Event]bool{
						&terminalapi.Keyboard{Key: keyboard.KeyEnter}: true,
						&terminalapi.Keyboard{Key: keyboard.KeyEsc}:   true,
						terminalapi.NewError("error1"):                true,
						terminalapi.NewError("error2"):                true,
					},
				},
				{
					filter: []terminalapi.Event{
						&terminalapi.Mouse{},
						&terminalapi.Resize{},
					},
					rec: newReceiver(),
					want: map[terminalapi.Event]bool{
						&terminalapi.Mouse{Position: image.Point{0, 0}}: true,
						&terminalapi.Mouse{Position: image.Point{1, 1}}: true,
						&terminalapi.Resize{Size: image.Point{1, 1}}:    true,
						&terminalapi.Resize{Size: image.Point{2, 2}}:    true,
						terminalapi.NewError("error1"):                  true,
						terminalapi.NewError("error2"):                  true,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			eds := NewDistributionSystem()
			for _, sc := range tc.subCase {
				stop := eds.Subscribe(sc.filter, sc.rec.receive)
				defer stop()
			}

			for _, ev := range tc.events {
				eds.Event(ev)
			}

			for i, sc := range tc.subCase {
				got, err := sc.rec.waitFor(len(sc.want), 5*time.Second)
				if err != nil {
					t.Fatalf("sc.rec.waitFor[%d] => unexpected error: %v", i, err)
				}

				if diff := pretty.Compare(sc.want, got); diff != "" {
					t.Errorf("sc.rec.waitFor[%d] => unexpected diff (-want, +got):\n%s", i, diff)
				}
			}
		})
	}
}
