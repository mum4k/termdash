// Copyright 2018 Google Inc.
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

package termdash

import (
	"context"
	"fmt"
	"image"
	"sync"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/internal/canvas/testcanvas"
	"github.com/mum4k/termdash/internal/event"
	"github.com/mum4k/termdash/internal/event/eventqueue"
	"github.com/mum4k/termdash/internal/event/testevent"
	"github.com/mum4k/termdash/internal/faketerm"
	"github.com/mum4k/termdash/internal/fakewidget"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/barchart"
	"github.com/mum4k/termdash/widgets/gauge"
)

// Example shows how to setup and run termdash with periodic redraw.
func Example() {
	// Create the terminal.
	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	// Create some widgets.
	bc, err := barchart.New()
	if err != nil {
		panic(err)
	}
	g, err := gauge.New()
	if err != nil {
		panic(err)
	}

	// Create the container with two widgets.
	c, err := container.New(
		t,
		container.SplitVertical(
			container.Left(
				container.PlaceWidget(bc),
			),
			container.Right(
				container.PlaceWidget(g),
			),
			container.SplitPercent(30),
		),
	)
	if err != nil {
		panic(err)
	}

	// Termdash runs until the context expires.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Run(ctx, t, c); err != nil {
		panic(err)
	}
}

// Example shows how to setup and run termdash with manually triggered redraw.
func Example_triggered() {
	// Create the terminal.
	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	// Create a widget.
	bc, err := barchart.New()
	if err != nil {
		panic(err)
	}

	// Create the container with a widget.
	c, err := container.New(
		t,
		container.PlaceWidget(bc),
	)
	if err != nil {
		panic(err)
	}

	// Create the controller and disable periodic redraw.
	ctrl, err := NewController(t, c)
	if err != nil {
		panic(err)
	}
	// Close the controller and termdash once it isn't required anymore.
	defer ctrl.Close()

	// Redraw the terminal manually.
	if err := ctrl.Redraw(); err != nil {
		panic(err)
	}
}

// errorHandler just stores the last error received.
type errorHandler struct {
	err error
	mu  sync.Mutex
}

func (eh *errorHandler) get() error {
	eh.mu.Lock()
	defer eh.mu.Unlock()
	return eh.err
}

func (eh *errorHandler) handle(err error) {
	eh.mu.Lock()
	defer eh.mu.Unlock()
	eh.err = err
}

// keySubscriber just stores the last pressed key.
type keySubscriber struct {
	received terminalapi.Keyboard
	mu       sync.Mutex
}

func (ks *keySubscriber) get() terminalapi.Keyboard {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	return ks.received
}

func (ks *keySubscriber) receive(k *terminalapi.Keyboard) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	ks.received = *k
}

// mouseSubscriber just stores the last mouse event.
type mouseSubscriber struct {
	received terminalapi.Mouse
	mu       sync.Mutex
}

func (ms *mouseSubscriber) get() terminalapi.Mouse {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.received
}

func (ms *mouseSubscriber) receive(m *terminalapi.Mouse) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.received = *m
}

type eventHandlers struct {
	handler  errorHandler
	keySub   keySubscriber
	mouseSub mouseSubscriber
}

func TestRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc   string
		size   image.Point
		opts   func(*eventHandlers) []Option
		events []terminalapi.Event
		// The number of expected processed events, used for synchronization.
		// Equals len(events) * number of subscribers for the event type.
		wantProcessed int
		// function to execute after the test case, can do additional comparison.
		after   func(*eventHandlers) error
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc: "draws the dashboard until closed",
			size: image.Point{60, 10},
			opts: func(*eventHandlers) []Option {
				return []Option{
					RedrawInterval(1),
				}
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc: "fails when the widget doesn't draw due to size too small",
			size: image.Point{1, 1},
			opts: func(*eventHandlers) []Option {
				return []Option{
					RedrawInterval(1),
				}
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				return ft
			},
			wantErr: true,
		},
		{
			desc: "forwards mouse events to container",
			size: image.Point{60, 10},
			opts: func(*eventHandlers) []Option {
				return []Option{
					RedrawInterval(1),
				}
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			wantProcessed: 2,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{
						WantMouse: widgetapi.MouseScopeWidget,
					},
					&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				)
				return ft
			},
		},
		{
			desc: "forwards keyboard events to container",
			size: image.Point{60, 10},
			opts: func(*eventHandlers) []Option {
				return []Option{
					RedrawInterval(1),
				}
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			wantProcessed: 2,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{
						WantKeyboard: widgetapi.KeyScopeFocused,
						WantMouse:    widgetapi.MouseScopeWidget,
					},
					&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				)
				return ft
			},
		},
		{
			desc: "forwards input errors to the error handler",
			size: image.Point{60, 10},
			opts: func(eh *eventHandlers) []Option {
				return []Option{
					RedrawInterval(1),
					ErrorHandler(eh.handler.handle),
				}
			},
			events: []terminalapi.Event{
				terminalapi.NewError("input error"),
			},
			wantProcessed: 1,
			after: func(eh *eventHandlers) error {
				if want := "input error"; eh.handler.get().Error() != want {
					return fmt.Errorf("errorHandler got %v, want %v", eh.handler.get(), want)
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc: "forwards keyboard events to the subscriber",
			size: image.Point{60, 10},
			opts: func(eh *eventHandlers) []Option {
				return []Option{
					RedrawInterval(1),
					KeyboardSubscriber(eh.keySub.receive),
				}
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyF1},
			},
			wantProcessed: 3,
			after: func(eh *eventHandlers) error {
				want := terminalapi.Keyboard{Key: keyboard.KeyF1}
				if diff := pretty.Compare(want, eh.keySub.get()); diff != "" {
					return fmt.Errorf("keySubscriber got unexpected value, diff (-want, +got):\n%s", diff)
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{
						WantKeyboard: widgetapi.KeyScopeFocused,
					},
					&terminalapi.Keyboard{Key: keyboard.KeyF1},
				)
				return ft
			},
		},
		{
			desc: "forwards mouse events to the subscriber",
			size: image.Point{60, 10},
			opts: func(eh *eventHandlers) []Option {
				return []Option{
					RedrawInterval(1),
					MouseSubscriber(eh.mouseSub.receive),
				}
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonWheelUp},
			},
			wantProcessed: 3,
			after: func(eh *eventHandlers) error {
				want := terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonWheelUp}
				if diff := pretty.Compare(want, eh.mouseSub.get()); diff != "" {
					return fmt.Errorf("mouseSubscriber got unexpected value, diff (-want, +got):\n%s", diff)
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{
						WantMouse: widgetapi.MouseScopeWidget,
					},
					&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonWheelUp},
				)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			tc := tc
			t.Parallel()

			handlers := &eventHandlers{
				handler:  errorHandler{},
				keySub:   keySubscriber{},
				mouseSub: mouseSubscriber{},
			}

			eq := eventqueue.New()
			for _, ev := range tc.events {
				eq.Push(ev)
			}

			got, err := faketerm.New(tc.size, faketerm.WithEventQueue(eq))
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			cont, err := container.New(
				got,
				container.PlaceWidget(fakewidget.New(widgetapi.Options{
					WantKeyboard: widgetapi.KeyScopeFocused,
					WantMouse:    widgetapi.MouseScopeWidget,
				})),
			)
			if err != nil {
				t.Fatalf("container.New => unexpected error: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			eds := event.NewDistributionSystem()
			opts := tc.opts(handlers)
			opts = append(opts, withEDS(eds))
			err = Run(ctx, got, cont, opts...)
			cancel()
			if (err != nil) != tc.wantErr {
				t.Errorf("Run => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if err := testevent.WaitFor(5*time.Second, func() error {
				if got, want := eds.Processed(), tc.wantProcessed; got != want {
					return fmt.Errorf("the event distribution system processed %d events, want %d", got, want)
				}
				return nil
			}); err != nil {
				t.Fatalf("testevent.WaitFor => %v", err)
			}

			if tc.after != nil {
				if err := tc.after(handlers); err != nil {
					t.Errorf("after => unexpected error: %v", err)
				}
			}

			if diff := faketerm.Diff(tc.want(got.Size()), got); diff != "" {
				t.Errorf("Run => %v", diff)
			}
		})
	}
}

func TestController(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc   string
		size   image.Point
		opts   []Option
		events []terminalapi.Event
		// The number of expected processed events, used for synchronization.
		// Equals len(events) * number of subscribers for the event type.
		wantProcessed int
		apiEvents     func(*fakewidget.Mirror) // Calls to the API of the widget.
		controls      func(*Controller) error
		want          func(size image.Point) *faketerm.Terminal
		wantErr       bool
	}{
		{
			desc: "event triggers a redraw",
			size: image.Point{60, 10},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			wantProcessed: 2,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{
						WantKeyboard: widgetapi.KeyScopeFocused,
						WantMouse:    widgetapi.MouseScopeWidget,
					},
					&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				)
				return ft

			},
		},
		{
			desc: "controller triggers redraw",
			size: image.Point{60, 10},
			apiEvents: func(mi *fakewidget.Mirror) {
				mi.Text("hello")
			},
			controls: func(ctrl *Controller) error {
				return ctrl.Redraw()
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				mirror := fakewidget.New(widgetapi.Options{})
				mirror.Text("hello")
				fakewidget.MustDrawWithMirror(
					mirror,
					ft,
					testcanvas.MustNew(ft.Area()),
				)
				return ft
			},
		},
		{
			desc: "ignores periodic redraw via the controller",
			size: image.Point{60, 10},
			opts: []Option{
				RedrawInterval(1),
			},
			apiEvents: func(mi *fakewidget.Mirror) {
				mi.Text("hello")
			},
			controls: func(ctrl *Controller) error {
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc: "does not redraw unless triggered when periodic disabled",
			size: image.Point{60, 10},
			apiEvents: func(mi *fakewidget.Mirror) {
				mi.Text("hello")
			},
			controls: func(ctrl *Controller) error {
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc: "fails when redraw fails",
			size: image.Point{1, 1},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc: "resizes the terminal",
			size: image.Point{60, 10},
			events: []terminalapi.Event{
				&terminalapi.Resize{Size: image.Point{70, 10}},
			},
			wantProcessed: 1,
			controls: func(ctrl *Controller) error {
				return ctrl.Redraw()
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(image.Point{70, 10})

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{},
				)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			tc := tc
			t.Parallel()

			eq := eventqueue.New()
			for _, ev := range tc.events {
				eq.Push(ev)
			}

			got, err := faketerm.New(tc.size, faketerm.WithEventQueue(eq))
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			mi := fakewidget.New(widgetapi.Options{
				WantKeyboard: widgetapi.KeyScopeFocused,
				WantMouse:    widgetapi.MouseScopeWidget,
			})
			cont, err := container.New(
				got,
				container.PlaceWidget(mi),
			)
			if err != nil {
				t.Fatalf("container.New => unexpected error: %v", err)
			}

			eds := event.NewDistributionSystem()
			opts := tc.opts
			opts = append(opts, withEDS(eds))
			ctrl, err := NewController(got, cont, opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("NewController => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if tc.apiEvents != nil {
				tc.apiEvents(mi)
			}

			if err := testevent.WaitFor(5*time.Second, func() error {
				if got, want := eds.Processed(), tc.wantProcessed; got != want {
					return fmt.Errorf("the event distribution system processed %d events, want %d", got, want)
				}
				return nil
			}); err != nil {
				t.Fatalf("testevent.WaitFor => %v", err)
			}

			if tc.controls != nil {
				if err := tc.controls(ctrl); err != nil {
					t.Errorf("controls => unexpected error: %v", err)
				}
			}
			ctrl.Close()

			if diff := faketerm.Diff(tc.want(got.Size()), got); diff != "" {
				t.Errorf("Run => %v", diff)
			}
		})
	}
}
