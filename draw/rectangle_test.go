package draw

import (
	"image"
	"testing"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/faketerm"
)

func TestRectangle(t *testing.T) {
	t.Skip()
	tests := []struct {
		desc    string
		canvas  image.Rectangle
		rect    image.Rectangle
		opts    []RectangleOption
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc:   "draws a 1x1 rectangle",
			canvas: image.Rect(0, 0, 2, 2),
			rect:   image.Rect(0, 0, 1, 1),
			opts: []RectangleOption{
				RectChar('x'),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "sets cell options",
			canvas: image.Rect(0, 0, 2, 2),
			rect:   image.Rect(0, 0, 1, 1),
			opts: []RectangleOption{
				RectChar('x'),
				RectCellOpts(
					cell.FgColor(cell.ColorBlue),
					cell.BgColor(cell.ColorRed),
				),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws a larger rectangle",
			canvas: image.Rect(0, 0, 10, 10),
			rect:   image.Rect(0, 0, 3, 2),
			opts: []RectangleOption{
				RectChar('o'),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "rectangle not in the corner of the canvas",
			canvas: image.Rect(0, 0, 10, 10),
			rect:   image.Rect(1, 1, 9, 3),
			opts: []RectangleOption{
				RectChar('o'),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
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

			err = Rectangle(c, tc.rect, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("Rectangle => unexpected error: %v, wantErr: %v", err, tc.wantErr)
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
				t.Errorf("Rectangle => %v", diff)
			}
		})
	}
}
