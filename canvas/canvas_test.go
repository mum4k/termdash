package canvas

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/faketerm"
)

func TestNew(t *testing.T) {
	tests := []struct {
		desc     string
		area     image.Rectangle
		wantSize image.Point
		wantErr  bool
	}{
		{
			desc:    "area min has negative X",
			area:    image.Rect(-1, 0, 0, 0),
			wantErr: true,
		},
		{
			desc:    "area min has negative Y",
			area:    image.Rect(0, -1, 0, 0),
			wantErr: true,
		},
		{
			desc:    "area max has negative X",
			area:    image.Rect(0, 0, -1, 0),
			wantErr: true,
		},
		{
			desc:    "area max has negative Y",
			area:    image.Rect(0, 0, 0, -1),
			wantErr: true,
		},
		{
			desc:    "zero area is invalid",
			area:    image.Rect(0, 0, 0, 0),
			wantErr: true,
		},
		{
			desc:     "smallest valid size",
			area:     image.Rect(0, 0, 1, 1),
			wantSize: image.Point{1, 1},
		},
		{
			desc:     "rectangular canvas 3 by 4",
			area:     image.Rect(0, 0, 3, 4),
			wantSize: image.Point{3, 4},
		},
		{
			desc:     "non-zero based area",
			area:     image.Rect(1, 1, 2, 3),
			wantSize: image.Point{1, 2},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := New(tc.area)
			if (err != nil) != tc.wantErr {
				t.Errorf("New => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			got := c.Size()
			if diff := pretty.Compare(tc.wantSize, got); diff != "" {
				t.Errorf("Size => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestSetCellAndApply(t *testing.T) {
	tests := []struct {
		desc           string
		termSize       image.Point
		canvasArea     image.Rectangle
		point          image.Point
		r              rune
		opts           []cell.Option
		want           cell.Buffer // Expected back buffer in the fake terminal.
		wantSetCellErr bool
		wantApplyErr   bool
	}{
		{
			desc:           "setting cell outside the designated area",
			termSize:       image.Point{2, 2},
			canvasArea:     image.Rect(0, 0, 1, 1),
			point:          image.Point{0, 2},
			wantSetCellErr: true,
		},
		{
			desc:       "sets a top-left corner cell",
			termSize:   image.Point{3, 3},
			canvasArea: image.Rect(1, 1, 3, 3),
			point:      image.Point{0, 0},
			r:          'X',
			want: cell.Buffer{
				{
					cell.New(0),
					cell.New(0),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New('X'),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New(0),
					cell.New(0),
				},
			},
		},
		{
			desc:       "sets a top-right corner cell",
			termSize:   image.Point{3, 3},
			canvasArea: image.Rect(1, 1, 3, 3),
			point:      image.Point{1, 0},
			r:          'X',
			want: cell.Buffer{
				{
					cell.New(0),
					cell.New(0),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New(0),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New('X'),
					cell.New(0),
				},
			},
		},
		{
			desc:       "sets a bottom-left corner cell",
			termSize:   image.Point{3, 3},
			canvasArea: image.Rect(1, 1, 3, 3),
			point:      image.Point{0, 1},
			r:          'X',
			want: cell.Buffer{
				{
					cell.New(0),
					cell.New(0),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New(0),
					cell.New('X'),
				},
				{
					cell.New(0),
					cell.New(0),
					cell.New(0),
				},
			},
		},
		{
			desc:       "sets a bottom-right corner cell",
			termSize:   image.Point{3, 3},
			canvasArea: image.Rect(1, 1, 3, 3),
			point:      image.Point{1, 1},
			r:          'Z',
			want: cell.Buffer{
				{
					cell.New(0),
					cell.New(0),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New(0),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New(0),
					cell.New('Z'),
				},
			},
		},
		{
			desc:       "sets cell options",
			termSize:   image.Point{3, 3},
			canvasArea: image.Rect(1, 1, 3, 3),
			point:      image.Point{1, 1},
			r:          'A',
			opts: []cell.Option{
				cell.BgColor(cell.ColorRed),
			},
			want: cell.Buffer{
				{
					cell.New(0),
					cell.New(0),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New(0),
					cell.New(0),
				},
				{
					cell.New(0),
					cell.New(0),
					cell.New('A', cell.BgColor(cell.ColorRed)),
				},
			},
		},
		{
			desc:       "canvas size equals terminal size",
			termSize:   image.Point{1, 1},
			canvasArea: image.Rect(0, 0, 1, 1),
			point:      image.Point{0, 0},
			r:          'A',
			want: cell.Buffer{
				{
					cell.New('A'),
				},
			},
		},
		{
			desc:         "terminal too small for the area",
			termSize:     image.Point{1, 1},
			canvasArea:   image.Rect(0, 0, 2, 2),
			point:        image.Point{0, 0},
			r:            'A',
			wantApplyErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := New(tc.canvasArea)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}

			err = c.SetCell(tc.point, tc.r, tc.opts...)
			if (err != nil) != tc.wantSetCellErr {
				t.Errorf("SetCell => unexpected error: %v, wantSetCellErr: %v", err, tc.wantSetCellErr)
			}
			if err != nil {
				return
			}

			ft, err := faketerm.New(tc.termSize)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}
			err = c.Apply(ft)
			if (err != nil) != tc.wantApplyErr {
				t.Errorf("Apply => unexpected error: %v, wantApplyErr: %v", err, tc.wantApplyErr)
			}
			if err != nil {
				return
			}

			got := ft.BackBuffer()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("faketerm.BackBuffer => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestClear(t *testing.T) {
	c, err := New(image.Rect(1, 1, 3, 3))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	if err := c.SetCell(image.Point{0, 0}, 'X'); err != nil {
		t.Fatalf("SetCell => unexpected error: %v", err)
	}

	ft, err := faketerm.New(image.Point{3, 3})
	if err != nil {
		t.Fatalf("faketerm.New => unexpected error: %v", err)
	}
	// Set one cell outside of the canvas on the terminal.
	if err := ft.SetCell(image.Point{0, 0}, 'A'); err != nil {
		t.Fatalf("faketerm.SetCell => unexpected error: %v", err)
	}

	if err := c.Apply(ft); err != nil {
		t.Fatalf("Apply => unexpected error: %v", err)
	}

	want := cell.Buffer{
		{
			cell.New('A'),
			cell.New(0),
			cell.New(0),
		},
		{
			cell.New(0),
			cell.New('X'),
			cell.New(0),
		},
		{
			cell.New(0),
			cell.New(0),
			cell.New(0),
		},
	}
	got := ft.BackBuffer()
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("faketerm.BackBuffer before Clear => unexpected diff (-want, +got):\n%s", diff)
	}

	// Call Clear(), Apply() and verify that only the area belonging to the
	// canvas was cleared.
	if err := c.Clear(); err != nil {
		t.Fatalf("Clear => unexpected error: %v", err)
	}
	if err := c.Apply(ft); err != nil {
		t.Fatalf("Apply => unexpected error: %v", err)
	}

	want = cell.Buffer{
		{
			cell.New('A'),
			cell.New(0),
			cell.New(0),
		},
		{
			cell.New(0),
			cell.New(0),
			cell.New(0),
		},
		{
			cell.New(0),
			cell.New(0),
			cell.New(0),
		},
	}

	got = ft.BackBuffer()
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("faketerm.BackBuffer after Clear => unexpected diff (-want, +got):\n%s", diff)
	}
}
