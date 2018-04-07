package fakewidget

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/draw/testdraw"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widget"
)

// kEvents are keyboard events to send to the widget.
type kEvents struct {
	k       *terminalapi.Keyboard
	wantErr bool
}

// mEvents are mouse events to send to the widget.
type mEvents struct {
	m       *terminalapi.Mouse
	wantErr bool
}

func TestMirror(t *testing.T) {
	tests := []struct {
		desc    string
		kEvents []kEvents // Keyboard events to send before calling Draw().
		mEvents []mEvents // Mouse events to send before calling Draw().
		cvs     *canvas.Canvas
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
				testdraw.MustBox(cvs, cvs.Area(), draw.LineStyleLight)
				tb := draw.TextBounds{
					Start: image.Point{1, 1},
				}
				testdraw.MustText(cvs, "(7,3)", tb)
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
				testdraw.MustBox(cvs, cvs.Area(), draw.LineStyleLight)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "draws the last keyboard event",
			kEvents: []kEvents{
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
				testdraw.MustBox(cvs, cvs.Area(), draw.LineStyleLight)
				tb := draw.TextBounds{
					Start: image.Point{1, 1},
				}
				testdraw.MustText(cvs, "(8,4)", tb)
				tb = draw.TextBounds{
					Start: image.Point{1, 2},
				}
				testdraw.MustText(cvs, "KeyEnd", tb)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "skips the keyboard event if there isn't a line for it",
			kEvents: []kEvents{
				{
					k: &terminalapi.Keyboard{Key: keyboard.KeyEnd},
				},
			},
			cvs: testcanvas.MustNew(image.Rect(0, 0, 8, 3)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(cvs, cvs.Area(), draw.LineStyleLight)
				tb := draw.TextBounds{
					Start: image.Point{1, 1},
				}
				testdraw.MustText(cvs, "(8,3)", tb)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "draws the last mouse event",
			mEvents: []mEvents{
				{
					m: &terminalapi.Mouse{Button: mouse.ButtonLeft},
				},
				{
					m: &terminalapi.Mouse{Button: mouse.ButtonRight},
				},
			},
			cvs: testcanvas.MustNew(image.Rect(0, 0, 13, 5)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(cvs, cvs.Area(), draw.LineStyleLight)
				tb := draw.TextBounds{
					Start: image.Point{1, 1},
				}
				testdraw.MustText(cvs, "(13,5)", tb)
				tb = draw.TextBounds{
					Start: image.Point{1, 3},
				}
				testdraw.MustText(cvs, "ButtonRight", tb)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "skips the mouse event if there isn't a line for it",
			mEvents: []mEvents{
				{
					m: &terminalapi.Mouse{Button: mouse.ButtonRight},
				},
			},
			cvs: testcanvas.MustNew(image.Rect(0, 0, 13, 4)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(cvs, cvs.Area(), draw.LineStyleLight)
				tb := draw.TextBounds{
					Start: image.Point{1, 1},
				}
				testdraw.MustText(cvs, "(13,4)", tb)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "draws both keyboard and mouse events",
			kEvents: []kEvents{
				{
					k: &terminalapi.Keyboard{Key: keyboard.KeyEnter},
				},
			},
			mEvents: []mEvents{
				{
					m: &terminalapi.Mouse{Button: mouse.ButtonLeft},
				},
			},
			cvs: testcanvas.MustNew(image.Rect(0, 0, 12, 5)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(cvs, cvs.Area(), draw.LineStyleLight)
				tb := draw.TextBounds{
					Start: image.Point{1, 1},
				}
				testdraw.MustText(cvs, "(12,5)", tb)
				tb = draw.TextBounds{
					Start: image.Point{1, 2},
				}
				testdraw.MustText(cvs, "KeyEnter", tb)
				tb = draw.TextBounds{
					Start: image.Point{1, 3},
				}
				testdraw.MustText(cvs, "ButtonLeft", tb)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "KeyEsc and ButtonMiddle reset the last event and return error",
			kEvents: []kEvents{
				{
					k: &terminalapi.Keyboard{Key: keyboard.KeyEnter},
				},
				{
					k:       &terminalapi.Keyboard{Key: keyboard.KeyEsc},
					wantErr: true,
				},
			},
			mEvents: []mEvents{
				{
					m: &terminalapi.Mouse{Button: mouse.ButtonLeft},
				},
				{
					m:       &terminalapi.Mouse{Button: mouse.ButtonMiddle},
					wantErr: true,
				},
			},
			cvs: testcanvas.MustNew(image.Rect(0, 0, 12, 5)),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(cvs, cvs.Area(), draw.LineStyleLight)
				tb := draw.TextBounds{
					Start: image.Point{1, 1},
				}
				testdraw.MustText(cvs, "(12,5)", tb)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			w := New(widget.Options{})

			for _, kEv := range tc.kEvents {
				err := w.Keyboard(kEv.k)
				if (err != nil) != kEv.wantErr {
					t.Errorf("Keyboard => got error:%v, wantErr: %v", err, kEv.wantErr)
				}
			}

			for _, mEv := range tc.mEvents {
				err := w.Mouse(mEv.m)
				if (err != nil) != mEv.wantErr {
					t.Errorf("Mouse => got error:%v, wantErr: %v", err, mEv.wantErr)
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
	want := widget.Options{
		Ratio:        image.Point{1, 2},
		WantKeyboard: true,
	}

	w := New(want)
	got := w.Options()
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
	}
}
