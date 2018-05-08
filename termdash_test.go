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
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/eventqueue"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/fakewidget"
)

// Example shows how to setup and run termdash.
func Example() {
	// Create the terminal.
	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	wOpts := widgetapi.Options{
		MinimumSize:  fakewidget.MinimumSize,
		WantKeyboard: true,
		WantMouse:    true,
	}

	// Create the container with two fake widgets.
	c := container.New(
		t,
		container.SplitVertical(
			container.Left(
				container.PlaceWidget(fakewidget.New(wOpts)),
			),
			container.Right(
				container.PlaceWidget(fakewidget.New(wOpts)),
			),
		),
	)

	// Termdash runs until the context expires.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Run(ctx, t, c); err != nil {
		panic(err)
	}
}

// errorHandler just stores the last error received.
type errorHandler struct {
	err error
}

func (eh *errorHandler) handle(err error) {
	eh.err = err
}

// keySubscriber just stores the last pressed key.
type keySubscriber struct {
	received terminalapi.Keyboard
}

func (ks *keySubscriber) receive(k *terminalapi.Keyboard) {
	ks.received = *k
}

// mouseSubscriber just stores the last mouse event.
type mouseSubscriber struct {
	received terminalapi.Mouse
}

func (ms *mouseSubscriber) receive(m *terminalapi.Mouse) {
	ms.received = *m
}

func TestRun(t *testing.T) {
	var (
		handler  errorHandler
		keySub   keySubscriber
		mouseSub mouseSubscriber
	)

	tests := []struct {
		desc   string
		size   image.Point
		opts   []Option
		events []terminalapi.Event
		// function to execute after the test case, can do additional comparison.
		after   func() error
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc: "draws the dashboard until closed",
			size: image.Point{60, 10},
			opts: []Option{
				RedrawInterval(1),
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
			desc: "fails when the widget doesn't draw",
			size: image.Point{1, 1},
			opts: []Option{
				RedrawInterval(1),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				return ft
			},
			wantErr: true,
		},
		{
			desc: "resizes the terminal",
			size: image.Point{60, 10},
			opts: []Option{
				RedrawInterval(1),
			},
			events: []terminalapi.Event{
				&terminalapi.Resize{Size: image.Point{70, 10}},
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
		{
			desc: "forwards mouse events to container",
			size: image.Point{60, 10},
			opts: []Option{
				RedrawInterval(1),
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{
						WantMouse: true,
					},
					&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				)
				return ft
			},
		},
		{
			desc: "forwards keyboard events to container",
			size: image.Point{60, 10},
			opts: []Option{
				RedrawInterval(1),
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{
						WantKeyboard: true,
						WantMouse:    true,
					},
					&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				)
				return ft
			},
		},
		{
			desc: "forwards input errors to the error handler",
			size: image.Point{60, 10},
			opts: []Option{
				RedrawInterval(1),
				ErrorHandler(handler.handle),
			},
			events: []terminalapi.Event{
				terminalapi.NewError("input error"),
			},
			after: func() error {
				if want := "input error"; handler.err.Error() != want {
					return fmt.Errorf("errorHandler got %v, want %v", handler.err, want)
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
			opts: []Option{
				RedrawInterval(1),
				KeyboardSubscriber(keySub.receive),
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyF1},
			},
			after: func() error {
				want := terminalapi.Keyboard{Key: keyboard.KeyF1}
				if diff := pretty.Compare(want, keySub.received); diff != "" {
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
						WantKeyboard: true,
					},
					&terminalapi.Keyboard{Key: keyboard.KeyF1},
				)
				return ft
			},
		},
		{
			desc: "forwards mouse events to the subscriber",
			size: image.Point{60, 10},
			opts: []Option{
				RedrawInterval(1),
				MouseSubscriber(mouseSub.receive),
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonWheelUp},
			},
			after: func() error {
				want := terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonWheelUp}
				if diff := pretty.Compare(want, mouseSub.received); diff != "" {
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
						WantMouse: true,
					},
					&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonWheelUp},
				)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			handler = errorHandler{}
			keySub = keySubscriber{}
			mouseSub = mouseSubscriber{}

			eq := eventqueue.New()
			for _, ev := range tc.events {
				eq.Push(ev)
			}

			got, err := faketerm.New(tc.size, faketerm.WithEventQueue(eq))
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			cont := container.New(
				got,
				container.PlaceWidget(fakewidget.New(widgetapi.Options{
					WantKeyboard: true,
					WantMouse:    true,
				})),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
			err = Run(ctx, got, cont, tc.opts...)
			cancel()
			if (err != nil) != tc.wantErr {
				t.Errorf("Run => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if tc.after != nil {
				if err := tc.after(); err != nil {
					t.Errorf("after => unexpected error: %v", err)
				}
			}
			if diff := faketerm.Diff(tc.want(got.Size()), got); diff != "" {
				t.Errorf("Run => %v", diff)
			}
		})
	}
}
