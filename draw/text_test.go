package draw

import (
	"image"
	"testing"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/faketerm"
)

func TestText(t *testing.T) {
	tests := []struct {
		desc    string
		canvas  image.Rectangle
		text    string
		tb      TextBounds
		opts    []cell.Option
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc:   "start falls outside of the canvas",
			canvas: image.Rect(0, 0, 2, 2),
			tb: TextBounds{
				Start: image.Point{2, 2},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "text falls outside of the canvas on OverrunModeStrict",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "ab",
			tb: TextBounds{
				Start:   image.Point{0, 0},
				Overrun: OverrunModeStrict,
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "requested MaxX is negative",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "",
			tb: TextBounds{
				Start:   image.Point{0, 0},
				MaxX:    -1,
				Overrun: OverrunModeStrict,
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "requested MaxX is greater than canvas width",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "",
			tb: TextBounds{
				Start:   image.Point{0, 0},
				Overrun: OverrunModeStrict,
				MaxX:    2,
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "text falls outside of requested MaxX",
			canvas: image.Rect(0, 0, 3, 2),
			text:   "ab",
			tb: TextBounds{
				Start:   image.Point{1, 1},
				Overrun: OverrunModeStrict,
				MaxX:    2,
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "text is empty, nothing to do",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "",
			tb: TextBounds{
				Start:   image.Point{0, 0},
				Overrun: OverrunModeStrict,
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:   "draws text",
			canvas: image.Rect(0, 0, 3, 2),
			text:   "ab",
			tb: TextBounds{
				Start:   image.Point{1, 1},
				Overrun: OverrunModeStrict,
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{1, 1}, 'a')
				testcanvas.MustSetCell(c, image.Point{2, 1}, 'b')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws text with cell options",
			canvas: image.Rect(0, 0, 3, 2),
			text:   "ab",
			tb: TextBounds{
				Start:   image.Point{1, 1},
				Overrun: OverrunModeStrict,
			},
			opts: []cell.Option{
				cell.FgColor(cell.ColorRed),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{1, 1}, 'a', cell.FgColor(cell.ColorRed))
				testcanvas.MustSetCell(c, image.Point{2, 1}, 'b', cell.FgColor(cell.ColorRed))
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := canvas.New(tc.canvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			err = Text(c, tc.text, tc.tb, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("Text => unexpected error: %v, wantErr: %v", err, tc.wantErr)
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

			if diff := faketerm.Diff(tc.want(c.Size()), got); diff != "" {
				t.Errorf("Text => %v", diff)
			}
		})
	}
}
