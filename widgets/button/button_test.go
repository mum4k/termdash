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

package button

import (
	"errors"
	"image"
	"sync"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/canvas/testcanvas"
	"github.com/mum4k/termdash/internal/cell"
	"github.com/mum4k/termdash/internal/draw"
	"github.com/mum4k/termdash/internal/draw/testdraw"
	"github.com/mum4k/termdash/internal/keyboard"
	"github.com/mum4k/termdash/internal/mouse"
	"github.com/mum4k/termdash/internal/terminalapi"
	"github.com/mum4k/termdash/internal/widgetapi"
	"github.com/mum4k/termdash/internal/faketerm"
)

// callbackTracker tracks whether callback was called.
type callbackTracker struct {
	// wantErr when set to true, makes callback return an error.
	wantErr bool

	// called asserts whether the callback was called.
	called bool

	// count is the number of times the callback was called.
	count int

	// mu protects the tracker.
	mu sync.Mutex
}

// callback is the callback function.
func (ct *callbackTracker) callback() error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if ct.wantErr {
		return errors.New("ct.wantErr set to true")
	}

	ct.count++
	ct.called = true
	return nil
}

func TestButton(t *testing.T) {
	tests := []struct {
		desc     string
		text     string
		callback *callbackTracker
		opts     []Option
		events   []terminalapi.Event
		canvas   image.Rectangle

		// timeSince is used to replace time.Since for tests, leave nil to use
		// the original.
		timeSince func(time.Time) time.Duration

		want            func(size image.Point) *faketerm.Terminal
		wantCallback    *callbackTracker
		wantNewErr      bool
		wantDrawErr     bool
		wantCallbackErr bool
	}{
		{
			desc:       "New fails with nil callback",
			canvas:     image.Rect(0, 0, 1, 1),
			wantNewErr: true,
		},
		{
			desc:     "New fails with negative keyUpDelay",
			callback: &callbackTracker{},
			opts: []Option{
				KeyUpDelay(-1 * time.Second),
			},
			canvas:     image.Rect(0, 0, 1, 1),
			wantNewErr: true,
		},
		{
			desc:     "New fails with zero Height",
			callback: &callbackTracker{},
			opts: []Option{
				Height(0),
			},
			canvas:     image.Rect(0, 0, 1, 1),
			wantNewErr: true,
		},
		{
			desc:     "New fails with zero Width",
			callback: &callbackTracker{},
			opts: []Option{
				Width(0),
			},
			canvas:     image.Rect(0, 0, 1, 1),
			wantNewErr: true,
		},
		{
			desc:        "draw fails on canvas too small",
			callback:    &callbackTracker{},
			text:        "hello",
			canvas:      image.Rect(0, 0, 1, 1),
			wantDrawErr: true,
		},
		{
			desc:     "draws button in up state",
			callback: &callbackTracker{},
			text:     "hello",
			canvas:   image.Rect(0, 0, 8, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Shadow.
				testcanvas.MustSetAreaCells(cvs, image.Rect(1, 1, 8, 4), 's', cell.BgColor(cell.ColorNumber(240)))

				// Button.
				testcanvas.MustSetAreaCells(cvs, image.Rect(0, 0, 7, 3), 'x', cell.BgColor(cell.ColorNumber(117)))

				// Text.
				testdraw.MustText(cvs, "hello", image.Point{1, 1},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorBlack),
						cell.BgColor(cell.ColorNumber(117))),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{},
		},
		{
			desc:     "draws button in down state due to a mouse event",
			callback: &callbackTracker{},
			text:     "hello",
			canvas:   image.Rect(0, 0, 8, 4),
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Button.
				testcanvas.MustSetAreaCells(cvs, image.Rect(1, 1, 8, 4), 'x', cell.BgColor(cell.ColorNumber(117)))

				// Text.
				testdraw.MustText(cvs, "hello", image.Point{2, 2},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorBlack),
						cell.BgColor(cell.ColorNumber(117))),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{},
		},
		{
			desc:     "mouse triggered the callback",
			callback: &callbackTracker{},
			text:     "hello",
			canvas:   image.Rect(0, 0, 8, 4),
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Shadow.
				testcanvas.MustSetAreaCells(cvs, image.Rect(1, 1, 8, 4), 's', cell.BgColor(cell.ColorNumber(240)))

				// Button.
				testcanvas.MustSetAreaCells(cvs, image.Rect(0, 0, 7, 3), 'x', cell.BgColor(cell.ColorNumber(117)))

				// Text.
				testdraw.MustText(cvs, "hello", image.Point{1, 1},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorBlack),
						cell.BgColor(cell.ColorNumber(117))),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{
				called: true,
				count:  1,
			},
		},
		{
			desc:     "draws button in down state due to a keyboard event, callback triggered",
			callback: &callbackTracker{},
			text:     "hello",
			opts: []Option{
				Key(keyboard.KeyEnter),
			},
			canvas: image.Rect(0, 0, 8, 4),
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Button.
				testcanvas.MustSetAreaCells(cvs, image.Rect(1, 1, 8, 4), 'x', cell.BgColor(cell.ColorNumber(117)))

				// Text.
				testdraw.MustText(cvs, "hello", image.Point{2, 2},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorBlack),
						cell.BgColor(cell.ColorNumber(117))),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{
				called: true,
				count:  1,
			},
		},
		{
			desc:     "keyboard event ignored when no key specified",
			callback: &callbackTracker{},
			text:     "hello",
			canvas:   image.Rect(0, 0, 8, 4),
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Shadow.
				testcanvas.MustSetAreaCells(cvs, image.Rect(1, 1, 8, 4), 's', cell.BgColor(cell.ColorNumber(240)))

				// Button.
				testcanvas.MustSetAreaCells(cvs, image.Rect(0, 0, 7, 3), 'x', cell.BgColor(cell.ColorNumber(117)))

				// Text.
				testdraw.MustText(cvs, "hello", image.Point{1, 1},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorBlack),
						cell.BgColor(cell.ColorNumber(117))),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{},
		},
		{
			desc:     "keyboard event triggers the button, trigger time didn't expire so button is down",
			callback: &callbackTracker{},
			text:     "hello",
			opts: []Option{
				Key(keyboard.KeyEnter),
			},
			timeSince: func(time.Time) time.Duration {
				return 200 * time.Millisecond
			},
			canvas: image.Rect(0, 0, 8, 4),
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Button.
				testcanvas.MustSetAreaCells(cvs, image.Rect(1, 1, 8, 4), 'x', cell.BgColor(cell.ColorNumber(117)))

				// Text.
				testdraw.MustText(cvs, "hello", image.Point{2, 2},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorBlack),
						cell.BgColor(cell.ColorNumber(117))),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{
				called: true,
				count:  1,
			},
		},
		{
			desc:     "keyboard event triggers the button, custom trigger time expired so button is up",
			callback: &callbackTracker{},
			text:     "hello",
			opts: []Option{
				Key(keyboard.KeyEnter),
				KeyUpDelay(100 * time.Millisecond),
			},
			timeSince: func(time.Time) time.Duration {
				return 200 * time.Millisecond
			},
			canvas: image.Rect(0, 0, 8, 4),
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Shadow.
				testcanvas.MustSetAreaCells(cvs, image.Rect(1, 1, 8, 4), 's', cell.BgColor(cell.ColorNumber(240)))

				// Button.
				testcanvas.MustSetAreaCells(cvs, image.Rect(0, 0, 7, 3), 'x', cell.BgColor(cell.ColorNumber(117)))

				// Text.
				testdraw.MustText(cvs, "hello", image.Point{1, 1},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorBlack),
						cell.BgColor(cell.ColorNumber(117))),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{
				called: true,
				count:  1,
			},
		},
		{
			desc:     "keyboard event triggers the button multiple times",
			callback: &callbackTracker{},
			text:     "hello",
			opts: []Option{
				Key(keyboard.KeyEnter),
				KeyUpDelay(100 * time.Millisecond),
			},
			timeSince: func(time.Time) time.Duration {
				return 200 * time.Millisecond
			},
			canvas: image.Rect(0, 0, 8, 4),
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Shadow.
				testcanvas.MustSetAreaCells(cvs, image.Rect(1, 1, 8, 4), 's', cell.BgColor(cell.ColorNumber(240)))

				// Button.
				testcanvas.MustSetAreaCells(cvs, image.Rect(0, 0, 7, 3), 'x', cell.BgColor(cell.ColorNumber(117)))

				// Text.
				testdraw.MustText(cvs, "hello", image.Point{1, 1},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorBlack),
						cell.BgColor(cell.ColorNumber(117))),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{
				called: true,
				count:  3,
			},
		},
		{
			desc:     "mouse event triggers the button multiple times",
			callback: &callbackTracker{},
			text:     "hello",
			canvas:   image.Rect(0, 0, 8, 4),
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Shadow.
				testcanvas.MustSetAreaCells(cvs, image.Rect(1, 1, 8, 4), 's', cell.BgColor(cell.ColorNumber(240)))

				// Button.
				testcanvas.MustSetAreaCells(cvs, image.Rect(0, 0, 7, 3), 'x', cell.BgColor(cell.ColorNumber(117)))

				// Text.
				testdraw.MustText(cvs, "hello", image.Point{1, 1},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorBlack),
						cell.BgColor(cell.ColorNumber(117))),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{
				called: true,
				count:  2,
			},
		},
		{
			desc: "the callback returns an error after a mouse event",
			callback: &callbackTracker{
				wantErr: true,
			},
			text:   "hello",
			canvas: image.Rect(0, 0, 8, 4),
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
			},
			wantCallbackErr: true,
		},
		{
			desc: "the callback returns an error after a keyboard event",
			callback: &callbackTracker{
				wantErr: true,
			},
			text: "hello",
			opts: []Option{
				Key(keyboard.KeyEnter),
				KeyUpDelay(100 * time.Millisecond),
			},
			timeSince: func(time.Time) time.Duration {
				return 200 * time.Millisecond
			},
			canvas: image.Rect(0, 0, 8, 4),
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			wantCallbackErr: true,
		},
		{
			desc:     "draws button with custom height (infra gives smaller canvas)",
			callback: &callbackTracker{},
			text:     "hello",
			canvas:   image.Rect(0, 0, 8, 2),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Shadow.
				testcanvas.MustSetAreaCells(cvs, image.Rect(1, 1, 8, 2), 's', cell.BgColor(cell.ColorNumber(240)))

				// Button.
				testcanvas.MustSetAreaCells(cvs, image.Rect(0, 0, 7, 1), 'x', cell.BgColor(cell.ColorNumber(117)))

				// Text.
				testdraw.MustText(cvs, "hello", image.Point{1, 0},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorBlack),
						cell.BgColor(cell.ColorNumber(117))),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{},
		},
		{
			desc:     "button width adjusts to width (infra gives smaller canvas)",
			callback: &callbackTracker{},
			text:     "h",
			canvas:   image.Rect(0, 0, 4, 2),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Shadow.
				testcanvas.MustSetAreaCells(cvs, image.Rect(1, 1, 4, 2), 's', cell.BgColor(cell.ColorNumber(240)))

				// Button.
				testcanvas.MustSetAreaCells(cvs, image.Rect(0, 0, 3, 1), 'x', cell.BgColor(cell.ColorNumber(117)))

				// Text.
				testdraw.MustText(cvs, "h", image.Point{1, 0},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorBlack),
						cell.BgColor(cell.ColorNumber(117))),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{},
		},
		{
			desc:     "sets custom text color",
			callback: &callbackTracker{},
			text:     "hello",
			opts: []Option{
				TextColor(cell.ColorRed),
			},
			canvas: image.Rect(0, 0, 8, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Shadow.
				testcanvas.MustSetAreaCells(cvs, image.Rect(1, 1, 8, 4), 's', cell.BgColor(cell.ColorNumber(240)))

				// Button.
				testcanvas.MustSetAreaCells(cvs, image.Rect(0, 0, 7, 3), 'x', cell.BgColor(cell.ColorNumber(117)))

				// Text.
				testdraw.MustText(cvs, "hello", image.Point{1, 1},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorRed),
						cell.BgColor(cell.ColorNumber(117))),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{},
		},
		{
			desc:     "sets custom fill color",
			callback: &callbackTracker{},
			text:     "hello",
			opts: []Option{
				FillColor(cell.ColorRed),
			},
			canvas: image.Rect(0, 0, 8, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Shadow.
				testcanvas.MustSetAreaCells(cvs, image.Rect(1, 1, 8, 4), 's', cell.BgColor(cell.ColorNumber(240)))

				// Button.
				testcanvas.MustSetAreaCells(cvs, image.Rect(0, 0, 7, 3), 'x', cell.BgColor(cell.ColorRed))

				// Text.
				testdraw.MustText(cvs, "hello", image.Point{1, 1},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorBlack),
						cell.BgColor(cell.ColorRed)),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{},
		},
		{
			desc:     "sets custom shadow color",
			callback: &callbackTracker{},
			text:     "hello",
			opts: []Option{
				ShadowColor(cell.ColorRed),
			},
			canvas: image.Rect(0, 0, 8, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Shadow.
				testcanvas.MustSetAreaCells(cvs, image.Rect(1, 1, 8, 4), 's', cell.BgColor(cell.ColorRed))

				// Button.
				testcanvas.MustSetAreaCells(cvs, image.Rect(0, 0, 7, 3), 'x', cell.BgColor(cell.ColorNumber(117)))

				// Text.
				testdraw.MustText(cvs, "hello", image.Point{1, 1},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorBlack),
						cell.BgColor(cell.ColorNumber(117))),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{},
		},
	}

	buttonRune = 'x'
	shadowRune = 's'
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.timeSince != nil {
				timeSince = tc.timeSince
			} else {
				timeSince = time.Since
			}

			gotCallback := tc.callback
			var cFn CallbackFn
			if gotCallback == nil {
				cFn = nil
			} else {
				cFn = gotCallback.callback
			}
			b, err := New(tc.text, cFn, tc.opts...)
			if (err != nil) != tc.wantNewErr {
				t.Errorf("New => unexpected error: %v, wantNewErr: %v", err, tc.wantNewErr)
			}
			if err != nil {
				return
			}

			{
				// Draw once which initializes the mouse state machine with the current canvas area.
				c, err := canvas.New(tc.canvas)
				if err != nil {
					t.Fatalf("canvas.New => unexpected error: %v", err)
				}
				err = b.Draw(c)
				if (err != nil) != tc.wantDrawErr {
					t.Errorf("Draw => unexpected error: %v, wantDrawErr: %v", err, tc.wantDrawErr)
				}
				if err != nil {
					return
				}
			}

			for i, ev := range tc.events {
				switch e := ev.(type) {
				case *terminalapi.Mouse:
					err := b.Mouse(e)
					// Only the last event in test cases is the one that triggers the callback.
					if i == len(tc.events)-1 {
						if (err != nil) != tc.wantCallbackErr {
							t.Errorf("Mouse => unexpected error: %v, wantCallbackErr: %v", err, tc.wantCallbackErr)
						}
						if err != nil {
							return
						}
					} else {
						if err != nil {
							t.Fatalf("Mouse => unexpected error: %v", err)
						}
					}

				case *terminalapi.Keyboard:
					err := b.Keyboard(e)
					// Only the last event in test cases is the one that triggers the callback.
					if i == len(tc.events)-1 {
						if (err != nil) != tc.wantCallbackErr {
							t.Errorf("Keyboard => unexpected error: %v, wantCallbackErr: %v", err, tc.wantCallbackErr)
						}
						if err != nil {
							return
						}
					} else {
						if err != nil {
							t.Fatalf("Keyboard => unexpected error: %v", err)
						}
					}

				default:
					t.Fatalf("unsupported event type: %T", ev)
				}
			}

			c, err := canvas.New(tc.canvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			err = b.Draw(c)
			if (err != nil) != tc.wantDrawErr {
				t.Errorf("Draw => unexpected error: %v, wantDrawErr: %v", err, tc.wantDrawErr)
			}
			if err != nil {
				return
			}

			got, err := faketerm.New(c.Size())
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			if err := c.Apply(got); err != nil {
				t.Fatalf("Apply => unexpected error: %v", err)
			}

			var want *faketerm.Terminal
			if tc.want != nil {
				want = tc.want(c.Size())
			} else {
				want = faketerm.MustNew(c.Size())
			}

			if diff := faketerm.Diff(want, got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}

			if diff := pretty.Compare(tc.wantCallback, gotCallback); diff != "" {
				t.Errorf("CallbackFn => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		desc string
		text string
		opts []Option
		want widgetapi.Options
	}{
		{
			desc: "width is based on the text width by default",
			text: "hello world",
			want: widgetapi.Options{
				MinimumSize:  image.Point{14, 4},
				MaximumSize:  image.Point{14, 4},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeGlobal,
			},
		},
		{
			desc: "width supports full-width unicode characters",
			text: "■㈱の世界①",
			want: widgetapi.Options{
				MinimumSize:  image.Point{13, 4},
				MaximumSize:  image.Point{13, 4},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeGlobal,
			},
		},
		{
			desc: "width specified via WidthFor",
			text: "hello",
			opts: []Option{
				WidthFor("■㈱の世界①"),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{13, 4},
				MaximumSize:  image.Point{13, 4},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeGlobal,
			},
		},
		{
			desc: "custom height specified",
			text: "hello",
			opts: []Option{
				Height(10),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{8, 11},
				MaximumSize:  image.Point{8, 11},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeGlobal,
			},
		},
		{
			desc: "custom width specified",
			text: "hello",
			opts: []Option{
				Width(10),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{11, 4},
				MaximumSize:  image.Point{11, 4},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeGlobal,
			},
		},

		{
			desc: "doesn't want keyboard by default",
			text: "hello",
			want: widgetapi.Options{
				MinimumSize:  image.Point{8, 4},
				MaximumSize:  image.Point{8, 4},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeGlobal,
			},
		},
		{
			desc: "registers for focused keyboard events",
			text: "hello",
			opts: []Option{
				Key(keyboard.KeyEnter),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{8, 4},
				MaximumSize:  image.Point{8, 4},
				WantKeyboard: widgetapi.KeyScopeFocused,
				WantMouse:    widgetapi.MouseScopeGlobal,
			},
		},
		{
			desc: "registers for global keyboard events",
			text: "hello",
			opts: []Option{
				GlobalKey(keyboard.KeyEnter),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{8, 4},
				MaximumSize:  image.Point{8, 4},
				WantKeyboard: widgetapi.KeyScopeGlobal,
				WantMouse:    widgetapi.MouseScopeGlobal,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ct := &callbackTracker{}
			b, err := New(tc.text, ct.callback, tc.opts...)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}

			got := b.Options()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}

}
