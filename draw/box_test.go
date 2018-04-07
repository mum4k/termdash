package draw

import (
	"image"
	"testing"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/faketerm"
)

func TestBox(t *testing.T) {
	tests := []struct {
		desc    string
		canvas  image.Rectangle
		box     image.Rectangle
		ls      LineStyle
		opts    []cell.Option
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc:    "box is larger than canvas",
			canvas:  image.Rect(0, 0, 1, 1),
			box:     image.Rect(0, 0, 2, 2),
			ls:      LineStyleLight,
			wantErr: true,
		},
		{
			desc:    "box is too small",
			canvas:  image.Rect(0, 0, 2, 2),
			box:     image.Rect(0, 0, 1, 1),
			ls:      LineStyleLight,
			wantErr: true,
		},
		{
			desc:    "unsupported line style",
			canvas:  image.Rect(0, 0, 4, 4),
			box:     image.Rect(0, 0, 2, 2),
			ls:      LineStyle(-1),
			wantErr: true,
		},
		{
			desc:   "draws box around the canvas",
			canvas: image.Rect(0, 0, 4, 4),
			box:    image.Rect(0, 0, 4, 4),
			ls:     LineStyleLight,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, lineStyleChars[LineStyleLight][topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{0, 1}, lineStyleChars[LineStyleLight][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, lineStyleChars[LineStyleLight][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 3}, lineStyleChars[LineStyleLight][bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{1, 0}, lineStyleChars[LineStyleLight][hLine])
				testcanvas.MustSetCell(c, image.Point{1, 3}, lineStyleChars[LineStyleLight][hLine])

				testcanvas.MustSetCell(c, image.Point{2, 0}, lineStyleChars[LineStyleLight][hLine])
				testcanvas.MustSetCell(c, image.Point{2, 3}, lineStyleChars[LineStyleLight][hLine])

				testcanvas.MustSetCell(c, image.Point{3, 0}, lineStyleChars[LineStyleLight][topRightCorner])
				testcanvas.MustSetCell(c, image.Point{3, 1}, lineStyleChars[LineStyleLight][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 2}, lineStyleChars[LineStyleLight][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 3}, lineStyleChars[LineStyleLight][bottomRightCorner])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws box in the canvas",
			canvas: image.Rect(0, 0, 4, 4),
			box:    image.Rect(1, 1, 3, 3),
			ls:     LineStyleLight,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{1, 1}, lineStyleChars[LineStyleLight][topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{1, 2}, lineStyleChars[LineStyleLight][bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{2, 1}, lineStyleChars[LineStyleLight][topRightCorner])
				testcanvas.MustSetCell(c, image.Point{2, 2}, lineStyleChars[LineStyleLight][bottomRightCorner])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws box with cell options",
			canvas: image.Rect(0, 0, 4, 4),
			box:    image.Rect(1, 1, 3, 3),
			ls:     LineStyleLight,
			opts: []cell.Option{
				cell.FgColor(cell.ColorRed),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{1, 1},
					lineStyleChars[LineStyleLight][topLeftCorner], cell.FgColor(cell.ColorRed))
				testcanvas.MustSetCell(c, image.Point{1, 2},
					lineStyleChars[LineStyleLight][bottomLeftCorner], cell.FgColor(cell.ColorRed))

				testcanvas.MustSetCell(c, image.Point{2, 1},
					lineStyleChars[LineStyleLight][topRightCorner], cell.FgColor(cell.ColorRed))
				testcanvas.MustSetCell(c, image.Point{2, 2},
					lineStyleChars[LineStyleLight][bottomRightCorner], cell.FgColor(cell.ColorRed))

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

			err = Box(c, tc.box, tc.ls, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("Box => unexpected error: %v, wantErr: %v", err, tc.wantErr)
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
				t.Errorf("Box => %v", diff)
			}
		})
	}
}
