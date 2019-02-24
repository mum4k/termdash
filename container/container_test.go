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

package container

import (
	"fmt"
	"image"
	"sync"
	"testing"
	"time"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/internal/canvas/testcanvas"
	"github.com/mum4k/termdash/internal/event"
	"github.com/mum4k/termdash/internal/event/testevent"
	"github.com/mum4k/termdash/internal/faketerm"
	"github.com/mum4k/termdash/internal/keyboard"
	"github.com/mum4k/termdash/internal/mouse"
	"github.com/mum4k/termdash/internal/testdraw"
	"github.com/mum4k/termdash/internal/widgetapi"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/barchart"
	"github.com/mum4k/termdash/widgets/fakewidget"
)

// Example demonstrates how to use the Container API.
func Example() {
	bc, err := barchart.New()
	if err != nil {
		panic(err)
	}
	if _, err := New(
		/* terminal = */ nil,
		SplitVertical(
			Left(
				SplitHorizontal(
					Top(
						Border(draw.LineStyleLight),
					),
					Bottom(
						SplitHorizontal(
							Top(
								Border(draw.LineStyleLight),
							),
							Bottom(
								Border(draw.LineStyleLight),
							),
						),
					),
					SplitPercent(30),
				),
			),
			Right(
				Border(draw.LineStyleLight),
				PlaceWidget(bc),
			),
		),
	); err != nil {
		panic(err)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		desc             string
		termSize         image.Point
		container        func(ft *faketerm.Terminal) (*Container, error)
		wantContainerErr bool
		want             func(size image.Point) *faketerm.Terminal
	}{
		{
			desc:     "empty container",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft)
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "container with a border",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "horizontal split, children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(draw.LineStyleLight),
						),
						Bottom(
							Border(draw.LineStyleLight),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 10, 5))
				testdraw.MustBorder(cvs, image.Rect(0, 5, 10, 10))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "fails on horizontal split too small",
			termSize: image.Point{10, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(draw.LineStyleLight),
						),
						Bottom(
							Border(draw.LineStyleLight),
						),
						SplitPercent(0),
					),
				)
			},
			wantContainerErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "fails on horizontal split too large",
			termSize: image.Point{10, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(draw.LineStyleLight),
						),
						Bottom(
							Border(draw.LineStyleLight),
						),
						SplitPercent(100),
					),
				)
			},
			wantContainerErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "horizontal unequal split",
			termSize: image.Point{10, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(draw.LineStyleLight),
						),
						Bottom(
							Border(draw.LineStyleLight),
						),
						SplitPercent(20),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 10, 4))
				testdraw.MustBorder(cvs, image.Rect(0, 4, 10, 20))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "horizontal split, parent and children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					SplitHorizontal(
						Top(
							Border(draw.LineStyleLight),
						),
						Bottom(
							Border(draw.LineStyleLight),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testdraw.MustBorder(cvs, image.Rect(1, 1, 9, 5))
				testdraw.MustBorder(cvs, image.Rect(1, 5, 9, 9))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical split, children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 5, 10))
				testdraw.MustBorder(cvs, image.Rect(5, 0, 10, 10))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "fails on vertical split too small",
			termSize: image.Point{20, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
						SplitPercent(0),
					),
				)
			},
			wantContainerErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "fails on vertical split too large",
			termSize: image.Point{20, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
						SplitPercent(100),
					),
				)
			},
			wantContainerErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "vertical unequal split",
			termSize: image.Point{20, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
						SplitPercent(20),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 4, 10))
				testdraw.MustBorder(cvs, image.Rect(4, 0, 20, 10))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical split, parent and children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testdraw.MustBorder(cvs, image.Rect(1, 1, 5, 9))
				testdraw.MustBorder(cvs, image.Rect(5, 1, 9, 9))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "multi level split",
			termSize: image.Point{10, 16},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							SplitHorizontal(
								Top(
									Border(draw.LineStyleLight),
								),
								Bottom(
									SplitHorizontal(
										Top(
											Border(draw.LineStyleLight),
										),
										Bottom(
											Border(draw.LineStyleLight),
										),
									),
								),
							),
						),
						Right(
							Border(draw.LineStyleLight),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 5, 8))
				testdraw.MustBorder(cvs, image.Rect(0, 8, 5, 12))
				testdraw.MustBorder(cvs, image.Rect(0, 12, 5, 16))
				testdraw.MustBorder(cvs, image.Rect(5, 0, 10, 16))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "inherits border and focused color",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					BorderColor(cell.ColorRed),
					FocusedColor(cell.ColorBlue),
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.BorderCellOpts(cell.FgColor(cell.ColorBlue)),
				)
				testdraw.MustBorder(
					cvs,
					image.Rect(1, 1, 5, 9),
					draw.BorderCellOpts(cell.FgColor(cell.ColorRed)),
				)
				testdraw.MustBorder(
					cvs,
					image.Rect(5, 1, 9, 9),
					draw.BorderCellOpts(cell.FgColor(cell.ColorRed)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "splitting a container removes the widget",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(fakewidget.New(widgetapi.Options{})),
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					ft.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testdraw.MustBorder(cvs, image.Rect(1, 1, 5, 9))
				testdraw.MustBorder(cvs, image.Rect(5, 1, 9, 9))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "placing a widget removes container split",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
					),
					PlaceWidget(fakewidget.New(widgetapi.Options{})),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 10, 10))
				testdraw.MustText(cvs, "(10,10)", image.Point{1, 1})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := faketerm.New(tc.termSize)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			cont, err := tc.container(got)
			if (err != nil) != tc.wantContainerErr {
				t.Errorf("tc.container => unexpected error:%v, wantErr:%v", err, tc.wantContainerErr)
			}
			if err != nil {
				return
			}
			if err := cont.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			if diff := faketerm.Diff(tc.want(tc.termSize), got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}
		})
	}

}

// eventGroup is a group of events to be delivered with synchronization.
// I.e. the test execution waits until the specified number is processed before
// proceeding with test execution.
type eventGroup struct {
	events        []terminalapi.Event
	wantProcessed int
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

func TestKeyboard(t *testing.T) {
	tests := []struct {
		desc        string
		termSize    image.Point
		container   func(ft *faketerm.Terminal) (*Container, error)
		eventGroups []*eventGroup
		want        func(size image.Point) *faketerm.Terminal
		wantErr     bool
	}{
		{
			desc:     "event not forwarded if container has no widget",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft)
			},
			eventGroups: []*eventGroup{
				{
					events: []terminalapi.Event{
						&terminalapi.Keyboard{Key: keyboard.KeyEnter},
					},
					wantProcessed: 0,
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "event forwarded to focused container only",
			termSize: image.Point{40, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused})),
						),
						Right(
							SplitHorizontal(
								Top(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused})),
								),
								Bottom(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused})),
								),
							),
						),
					),
				)
			},
			eventGroups: []*eventGroup{
				// Move focus to the target container.
				{
					events: []terminalapi.Event{
						&terminalapi.Mouse{Position: image.Point{39, 19}, Button: mouse.ButtonLeft},
						&terminalapi.Mouse{Position: image.Point{39, 19}, Button: mouse.ButtonRelease},
					},
					wantProcessed: 2,
				},
				// Send the keyboard event.
				{
					events: []terminalapi.Event{
						&terminalapi.Keyboard{Key: keyboard.KeyEnter},
					},
					wantProcessed: 5,
				},
			},

			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				// Widgets that aren't focused don't get the key.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 0, 20, 20)),
					widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(20, 0, 40, 10)),
					widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused},
				)

				// The focused widget receives the key.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(20, 10, 40, 20)),
					widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused},
					&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				)
				return ft
			},
		},
		{
			desc:     "event forwarded to all widgets that requested global key scope",
			termSize: image.Point{40, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeGlobal})),
						),
						Right(
							SplitHorizontal(
								Top(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused})),
								),
								Bottom(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused})),
								),
							),
						),
					),
				)
			},
			eventGroups: []*eventGroup{
				// Move focus to the target container.
				{
					events: []terminalapi.Event{
						&terminalapi.Mouse{Position: image.Point{39, 19}, Button: mouse.ButtonLeft},
						&terminalapi.Mouse{Position: image.Point{39, 19}, Button: mouse.ButtonRelease},
					},
					wantProcessed: 2,
				},
				// Send the keyboard event.
				{
					events: []terminalapi.Event{
						&terminalapi.Keyboard{Key: keyboard.KeyEnter},
					},
					wantProcessed: 5,
				},
			},

			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				// Widget that isn't focused, but registered for global
				// keyboard events.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 0, 20, 20)),
					widgetapi.Options{WantKeyboard: widgetapi.KeyScopeGlobal},
					&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				)

				// Widget that isn't focused and only wants focused events.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(20, 0, 40, 10)),
					widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused},
				)

				// The focused widget receives the key.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(20, 10, 40, 20)),
					widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused},
					&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				)
				return ft
			},
		},
		{
			desc:     "event not forwarded if the widget didn't request it",
			termSize: image.Point{40, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeNone})),
				)
			},
			eventGroups: []*eventGroup{
				{
					events: []terminalapi.Event{
						&terminalapi.Keyboard{Key: keyboard.KeyEnter},
					},
					wantProcessed: 0,
				},
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
			desc:     "widget returns an error when processing the event",
			termSize: image.Point{40, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused})),
				)
			},
			eventGroups: []*eventGroup{
				{
					events: []terminalapi.Event{
						&terminalapi.Keyboard{Key: keyboard.KeyEsc},
					},
					wantProcessed: 2,
				},
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
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := faketerm.New(tc.termSize)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			c, err := tc.container(got)
			if err != nil {
				t.Fatalf("tc.container => unexpected error: %v", err)
			}

			eds := event.NewDistributionSystem()
			eh := &errorHandler{}
			// Subscribe to receive errors.
			eds.Subscribe([]terminalapi.Event{terminalapi.NewError("")}, func(ev terminalapi.Event) {
				eh.handle(ev.(*terminalapi.Error).Error())
			})

			c.Subscribe(eds)
			for _, eg := range tc.eventGroups {
				for _, ev := range eg.events {
					eds.Event(ev)
				}
				if err := testevent.WaitFor(5*time.Second, func() error {
					if got, want := eds.Processed(), eg.wantProcessed; got != want {
						return fmt.Errorf("the event distribution system processed %d events, want %d", got, want)
					}
					return nil
				}); err != nil {
					t.Fatalf("testevent.WaitFor => %v", err)
				}
			}

			if err := c.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			if diff := faketerm.Diff(tc.want(tc.termSize), got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}

			if err := eh.get(); (err != nil) != tc.wantErr {
				t.Errorf("errorHandler => unexpected error %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}

func TestMouse(t *testing.T) {
	tests := []struct {
		desc          string
		termSize      image.Point
		container     func(ft *faketerm.Terminal) (*Container, error)
		events        []terminalapi.Event
		want          func(size image.Point) *faketerm.Terminal
		wantProcessed int
		wantErr       bool
	}{
		{
			desc:     "mouse click outside of the terminal is ignored",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{-1, -1}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{10, 10}, Button: mouse.ButtonRelease},
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
			wantProcessed: 4,
		},
		{
			desc:     "event not forwarded if container has no widget",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantProcessed: 2,
		},
		{
			desc:     "event forwarded to container at that point",
			termSize: image.Point{50, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
						),
						Right(
							SplitHorizontal(
								Top(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
								),
								Bottom(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
								),
							),
						),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{49, 9}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{49, 9}, Button: mouse.ButtonRelease},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				// Widgets that aren't focused don't get the mouse clicks.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 0, 25, 20)),
					widgetapi.Options{},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(25, 10, 50, 20)),
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Keyboard{},
				)

				// The focused widget receives the key.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(25, 0, 50, 10)),
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{24, 9}, Button: mouse.ButtonLeft},
					&terminalapi.Mouse{Position: image.Point{24, 9}, Button: mouse.ButtonRelease},
				)
				return ft
			},
			wantProcessed: 8,
		},
		{
			desc:     "event not forwarded if the widget didn't request it",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeNone})),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
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
			wantProcessed: 1,
		},
		{
			desc:     "MouseScopeWidget, event not forwarded if it falls on the container's border",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(
						fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget}),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					ft.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(cvs, ft)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(1, 1, 19, 19)),
					widgetapi.Options{},
				)
				return ft
			},
			wantProcessed: 2,
		},
		{
			desc:     "MouseScopeContainer, event forwarded if it falls on the container's border",
			termSize: image.Point{21, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(
						fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeContainer}),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					ft.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(cvs, ft)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(1, 1, 20, 19)),
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{-1, -1}, Button: mouse.ButtonLeft},
				)
				return ft
			},
			wantProcessed: 2,
		},
		{
			desc:     "MouseScopeGlobal, event forwarded if it falls on the container's border",
			termSize: image.Point{21, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(
						fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeGlobal}),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					ft.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(cvs, ft)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(1, 1, 20, 19)),
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{-1, -1}, Button: mouse.ButtonLeft},
				)
				return ft
			},
			wantProcessed: 2,
		},
		{
			desc:     "MouseScopeWidget event not forwarded if it falls outside of widget's canvas",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: widgetapi.MouseScopeWidget,
							Ratio:     image.Point{2, 1},
						}),
					),
					AlignVertical(align.VerticalMiddle),
					AlignHorizontal(align.HorizontalCenter),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 5, 20, 15)),
					widgetapi.Options{},
				)
				return ft
			},
			wantProcessed: 2,
		},
		{
			desc:     "MouseScopeContainer event forwarded if it falls outside of widget's canvas",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: widgetapi.MouseScopeContainer,
							Ratio:     image.Point{2, 1},
						}),
					),
					AlignVertical(align.VerticalMiddle),
					AlignHorizontal(align.HorizontalCenter),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 5, 20, 15)),
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{-1, -1}, Button: mouse.ButtonLeft},
				)
				return ft
			},
			wantProcessed: 2,
		},
		{
			desc:     "MouseScopeGlobal event forwarded if it falls outside of widget's canvas",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: widgetapi.MouseScopeGlobal,
							Ratio:     image.Point{2, 1},
						}),
					),
					AlignVertical(align.VerticalMiddle),
					AlignHorizontal(align.HorizontalCenter),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 5, 20, 15)),
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{-1, -1}, Button: mouse.ButtonLeft},
				)
				return ft
			},
			wantProcessed: 2,
		},
		{
			desc:     "MouseScopeWidget event not forwarded if it falls to another container",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(),
						Bottom(
							PlaceWidget(
								fakewidget.New(widgetapi.Options{
									WantMouse: widgetapi.MouseScopeWidget,
									Ratio:     image.Point{2, 1},
								}),
							),
							AlignVertical(align.VerticalMiddle),
							AlignHorizontal(align.HorizontalCenter),
						),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 10, 20, 20)),
					widgetapi.Options{},
				)
				return ft
			},
			wantProcessed: 2,
		},
		{
			desc:     "MouseScopeContainer event not forwarded if it falls to another container",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(),
						Bottom(
							PlaceWidget(
								fakewidget.New(widgetapi.Options{
									WantMouse: widgetapi.MouseScopeContainer,
									Ratio:     image.Point{2, 1},
								}),
							),
							AlignVertical(align.VerticalMiddle),
							AlignHorizontal(align.HorizontalCenter),
						),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 10, 20, 20)),
					widgetapi.Options{},
				)
				return ft
			},
			wantProcessed: 2,
		},
		{
			desc:     "MouseScopeGlobal event forwarded if it falls to another container",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(),
						Bottom(
							PlaceWidget(
								fakewidget.New(widgetapi.Options{
									WantMouse: widgetapi.MouseScopeGlobal,
									Ratio:     image.Point{2, 1},
								}),
							),
							AlignVertical(align.VerticalMiddle),
							AlignHorizontal(align.HorizontalCenter),
						),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 10, 20, 20)),
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{-1, -1}, Button: mouse.ButtonLeft},
				)
				return ft
			},
			wantProcessed: 2,
		},
		{
			desc:     "mouse position adjusted relative to widget's canvas, vertical offset",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: widgetapi.MouseScopeWidget,
							Ratio:     image.Point{2, 1},
						}),
					),
					AlignVertical(align.VerticalMiddle),
					AlignHorizontal(align.HorizontalCenter),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 5}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 5, 20, 15)),
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				)
				return ft
			},
			wantProcessed: 2,
		},
		{
			desc:     "mouse poisition adjusted relative to widget's canvas, horizontal offset",
			termSize: image.Point{30, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: widgetapi.MouseScopeWidget,
							Ratio:     image.Point{9, 10},
						}),
					),
					AlignVertical(align.VerticalMiddle),
					AlignHorizontal(align.HorizontalCenter),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{6, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(6, 0, 24, 20)),
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				)
				return ft
			},
			wantProcessed: 2,
		},
		{
			desc:     "widget returns an error when processing the event",
			termSize: image.Point{40, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRight},
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
			wantProcessed: 3,
			wantErr:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := faketerm.New(tc.termSize)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			c, err := tc.container(got)
			if err != nil {
				t.Fatalf("tc.container => unexpected error: %v", err)
			}

			eds := event.NewDistributionSystem()
			eh := &errorHandler{}
			// Subscribe to receive errors.
			eds.Subscribe([]terminalapi.Event{terminalapi.NewError("")}, func(ev terminalapi.Event) {
				eh.handle(ev.(*terminalapi.Error).Error())
			})
			c.Subscribe(eds)
			for _, ev := range tc.events {
				eds.Event(ev)
			}
			if err := testevent.WaitFor(5*time.Second, func() error {
				if got, want := eds.Processed(), tc.wantProcessed; got != want {
					return fmt.Errorf("the event distribution system processed %d events, want %d", got, want)
				}
				return nil
			}); err != nil {
				t.Fatalf("testevent.WaitFor => %v", err)
			}

			if err := c.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			if diff := faketerm.Diff(tc.want(tc.termSize), got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}

			if err := eh.get(); (err != nil) != tc.wantErr {
				t.Errorf("errorHandler => unexpected error %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}
