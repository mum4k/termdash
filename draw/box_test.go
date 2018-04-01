package draw

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/canvas"
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
		want    cell.Buffer
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
			want: cell.Buffer{
				{
					cell.New(lineStyleChars[LineStyleLight][topLeftCorner]),
					cell.New(lineStyleChars[LineStyleLight][vLine]),
					cell.New(lineStyleChars[LineStyleLight][vLine]),
					cell.New(lineStyleChars[LineStyleLight][bottomLeftCorner]),
				},
				{
					cell.New(lineStyleChars[LineStyleLight][hLine]),
					cell.New(0),
					cell.New(0),
					cell.New(lineStyleChars[LineStyleLight][hLine]),
				},
				{
					cell.New(lineStyleChars[LineStyleLight][hLine]),
					cell.New(0),
					cell.New(0),
					cell.New(lineStyleChars[LineStyleLight][hLine]),
				},
				{
					cell.New(lineStyleChars[LineStyleLight][topRightCorner]),
					cell.New(lineStyleChars[LineStyleLight][vLine]),
					cell.New(lineStyleChars[LineStyleLight][vLine]),
					cell.New(lineStyleChars[LineStyleLight][bottomRightCorner]),
				},
			},
		},
		{
			desc:   "draws box in the canvas",
			canvas: image.Rect(0, 0, 4, 4),
			box:    image.Rect(1, 1, 3, 3),
			ls:     LineStyleLight,
			want: cell.Buffer{
				{
					cell.New(0),
					cell.New(0),
					cell.New(0),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New(lineStyleChars[LineStyleLight][topLeftCorner]),
					cell.New(lineStyleChars[LineStyleLight][bottomLeftCorner]),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New(lineStyleChars[LineStyleLight][topRightCorner]),
					cell.New(lineStyleChars[LineStyleLight][bottomRightCorner]),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New(0),
					cell.New(0),
					cell.New(0),
				},
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
			want: cell.Buffer{
				{
					cell.New(0),
					cell.New(0),
					cell.New(0),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New(
						lineStyleChars[LineStyleLight][topLeftCorner],
						cell.FgColor(cell.ColorRed),
					),
					cell.New(
						lineStyleChars[LineStyleLight][bottomLeftCorner],
						cell.FgColor(cell.ColorRed),
					),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New(
						lineStyleChars[LineStyleLight][topRightCorner],
						cell.FgColor(cell.ColorRed),
					),
					cell.New(
						lineStyleChars[LineStyleLight][bottomRightCorner],
						cell.FgColor(cell.ColorRed),
					),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New(0),
					cell.New(0),
					cell.New(0),
				},
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

			ft, err := faketerm.New(c.Size())
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			if err := c.Apply(ft); err != nil {
				t.Fatalf("Apply => unexpected error: %v", err)
			}

			got := ft.BackBuffer()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Logf("Box => got output:\n%s", ft)
				t.Errorf("Box => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
