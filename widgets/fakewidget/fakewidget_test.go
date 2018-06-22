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

package fakewidget

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/draw/testdraw"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// keyEvents are keyboard events to send to the widget.
type keyEvents struct {
	k       *terminalapi.Keyboard
	wantErr bool
}

// mouseEvents are mouse events to send to the widget.
type mouseEvents struct {
	m       *terminalapi.Mouse
	wantErr bool
}

func TestMirror(t *testing.T) {
	tests := []struct {
		desc        string
		keyEvents   []keyEvents   // Keyboard events to send before calling Draw().
		mouseEvents []mouseEvents // Mouse events to send before calling Draw().
		apiEvents   func(*Mirror) // External events via the widget's API.
		cvs         *canvas.Canvas
		want        func(size image.Point) *faketerm.Terminal
		wantErr     bool
	}{
		{
			desc: "canvas too small to draw a box",
			cvs:  testcanvas.MustNew(image.Rect(0, 0, 1, 1)),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc: "the canvas size text doesn't fit onto the line",
			cvs:  testcanvas.MustNew(image.Rect(0, 0, 3, 3)),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc: "draws the box and canvas size",
			cvs:  testcanvas.MustNew(image.Rect(0, 0, 7, 3)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, cvs.Area())
				testdraw.MustText(cvs, "(7,3)", image.Point{1, 1})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "draws the box, canvas size and custom text",
			apiEvents: func(mi *Mirror) {
				mi.Text("hi")
			},
			cvs: testcanvas.MustNew(image.Rect(0, 0, 9, 3)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, cvs.Area())
				testdraw.MustText(cvs, "(9,3)hi", image.Point{1, 1})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "skips canvas size if there isn't a line for it",
			cvs:  testcanvas.MustNew(image.Rect(0, 0, 3, 2)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, cvs.Area())
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "draws the last keyboard event",
			keyEvents: []keyEvents{
				{
					k: &terminalapi.Keyboard{Key: keyboard.KeyEnter},
				},
				{
					k: &terminalapi.Keyboard{Key: keyboard.KeyEnd},
				},
			},
			cvs: testcanvas.MustNew(image.Rect(0, 0, 8, 4)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, cvs.Area())
				testdraw.MustText(cvs, "(8,4)", image.Point{1, 1})
				testdraw.MustText(cvs, "KeyEnd", image.Point{1, 2})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "skips the keyboard event if there isn't a line for it",
			keyEvents: []keyEvents{
				{
					k: &terminalapi.Keyboard{Key: keyboard.KeyEnd},
				},
			},
			cvs: testcanvas.MustNew(image.Rect(0, 0, 8, 3)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, cvs.Area())
				testdraw.MustText(cvs, "(8,3)", image.Point{1, 1})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "draws the last mouse event",
			mouseEvents: []mouseEvents{
				{
					m: &terminalapi.Mouse{Button: mouse.ButtonLeft},
				},
				{
					m: &terminalapi.Mouse{
						Position: image.Point{1, 2},
						Button:   mouse.ButtonMiddle},
				},
			},
			cvs: testcanvas.MustNew(image.Rect(0, 0, 19, 5)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, cvs.Area())
				testdraw.MustText(cvs, "(19,5)", image.Point{1, 1})
				testdraw.MustText(cvs, "(1,2)ButtonMiddle", image.Point{1, 3})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "skips the mouse event if there isn't a line for it",
			mouseEvents: []mouseEvents{
				{
					m: &terminalapi.Mouse{Button: mouse.ButtonLeft},
				},
			},
			cvs: testcanvas.MustNew(image.Rect(0, 0, 13, 4)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, cvs.Area())
				testdraw.MustText(cvs, "(13,4)", image.Point{1, 1})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "draws both keyboard and mouse events",
			keyEvents: []keyEvents{
				{
					k: &terminalapi.Keyboard{Key: keyboard.KeyEnter},
				},
			},
			mouseEvents: []mouseEvents{
				{
					m: &terminalapi.Mouse{Button: mouse.ButtonLeft},
				},
			},
			cvs: testcanvas.MustNew(image.Rect(0, 0, 17, 5)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, cvs.Area())
				testdraw.MustText(cvs, "(17,5)", image.Point{1, 1})
				testdraw.MustText(cvs, "KeyEnter", image.Point{1, 2})
				testdraw.MustText(cvs, "(0,0)ButtonLeft", image.Point{1, 3})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "KeyEsc and ButtonRight reset the last event and return error",
			keyEvents: []keyEvents{
				{
					k: &terminalapi.Keyboard{Key: keyboard.KeyEnter},
				},
				{
					k:       &terminalapi.Keyboard{Key: keyboard.KeyEsc},
					wantErr: true,
				},
			},
			mouseEvents: []mouseEvents{
				{
					m: &terminalapi.Mouse{Button: mouse.ButtonLeft},
				},
				{
					m:       &terminalapi.Mouse{Button: mouse.ButtonRight},
					wantErr: true,
				},
			},
			cvs: testcanvas.MustNew(image.Rect(0, 0, 12, 5)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, cvs.Area())
				testdraw.MustText(cvs, "(12,5)", image.Point{1, 1})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			w := New(widgetapi.Options{})

			if tc.apiEvents != nil {
				tc.apiEvents(w)
			}

			for _, keyEv := range tc.keyEvents {
				err := w.Keyboard(keyEv.k)
				if (err != nil) != keyEv.wantErr {
					t.Errorf("Keyboard => got error:%v, wantErr: %v", err, keyEv.wantErr)
				}
			}

			for _, mouseEv := range tc.mouseEvents {
				err := w.Mouse(mouseEv.m)
				if (err != nil) != mouseEv.wantErr {
					t.Errorf("Mouse => got error:%v, wantErr: %v", err, mouseEv.wantErr)
				}
			}

			err := w.Draw(tc.cvs)
			if (err != nil) != tc.wantErr {
				t.Errorf("Draw => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			got := faketerm.MustNew(tc.cvs.Size())
			testcanvas.MustApply(tc.cvs, got)
			if diff := faketerm.Diff(tc.want(tc.cvs.Size()), got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	want := widgetapi.Options{
		Ratio:        image.Point{1, 2},
		WantKeyboard: true,
	}

	w := New(want)
	got := w.Options()
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
	}
}

func TestDraw(t *testing.T) {
	tests := []struct {
		desc    string
		opts    widgetapi.Options
		cvs     *canvas.Canvas
		events  []terminalapi.Event
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc: "canvas too small to draw a box",
			cvs:  testcanvas.MustNew(image.Rect(0, 0, 1, 1)),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc: "draws the box and canvas size",
			cvs:  testcanvas.MustNew(image.Rect(0, 0, 9, 3)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, cvs.Area())
				testdraw.MustText(cvs, "(9,3)", image.Point{1, 1})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "draws both keyboard and mouse events",
			opts: widgetapi.Options{
				WantKeyboard: true,
				WantMouse:    true,
			},
			cvs: testcanvas.MustNew(image.Rect(0, 0, 17, 5)),
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				&terminalapi.Mouse{Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, cvs.Area())
				testdraw.MustText(cvs, "(17,5)", image.Point{1, 1})
				testdraw.MustText(cvs, "KeyEnter", image.Point{1, 2})
				testdraw.MustText(cvs, "(0,0)ButtonLeft", image.Point{1, 3})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := faketerm.MustNew(tc.cvs.Size())
			err := Draw(got, tc.cvs, tc.opts, tc.events...)
			if (err != nil) != tc.wantErr {
				t.Errorf("Draw => got error:%v, wantErr: %v", err, tc.wantErr)
			}

			if diff := faketerm.Diff(tc.want(tc.cvs.Size()), got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}
		})
	}
}
