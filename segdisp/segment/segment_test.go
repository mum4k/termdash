package segment

import (
	"fmt"
	"image"
	"testing"

	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/canvas/braille"
	"github.com/mum4k/termdash/canvas/braille/testbraille"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw/testdraw"
	"github.com/mum4k/termdash/terminal/faketerm"
)

func TestDraw(t *testing.T) {
	tests := []struct {
		desc       string
		opts       []Option
		cellCanvas image.Rectangle // Canvas in cells that will be converted to braille canvas for drawing.
		ar         image.Rectangle
		st         SegmentType
		want       func(size image.Point) *faketerm.Terminal
		wantErr    bool
	}{
		{
			desc:       "fails on area with negative Min.X",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(-1, 0, 1, 1),
			st:         SegmentTypeHorizontal,
			wantErr:    true,
		},
		{
			desc:       "fails on area with negative Min.Y",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, -1, 1, 1),
			st:         SegmentTypeHorizontal,
			wantErr:    true,
		},
		{
			desc:       "fails on area with negative Max.X",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rectangle{image.Point{0, 0}, image.Point{-1, 1}},
			st:         SegmentTypeHorizontal,
			wantErr:    true,
		},
		{
			desc:       "fails on area with negative Max.Y",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rectangle{image.Point{0, 0}, image.Point{1, -1}},
			st:         SegmentTypeHorizontal,
			wantErr:    true,
		},
		{
			desc:       "fails on area with zero Dx()",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 0, 1),
			st:         SegmentTypeHorizontal,
			wantErr:    true,
		},
		{
			desc:       "fails on area with zero Dy()",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 0),
			st:         SegmentTypeHorizontal,
			wantErr:    true,
		},
		{
			desc:       "fails on unsupported segment type",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 2),
			st:         SegmentType(0),
			wantErr:    true,
		},
		{
			desc:       "fails on area larger than the canvas",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 3, 1),
			st:         SegmentTypeHorizontal,
			wantErr:    true,
		},
		{
			desc: "sets cell options",
			opts: []Option{
				CellOpts(
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorGreen),
				),
			},
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 1),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0}, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorGreen))
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 1x1",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 1),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 1x2",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 2),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{0, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 1x3",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 3),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{0, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{0, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 1x4",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 4),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{0, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{0, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 3}, image.Point{0, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 1x5",
			cellCanvas: image.Rect(0, 0, 1, 2),
			ar:         image.Rect(0, 0, 1, 5),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{0, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{0, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 3}, image.Point{0, 3})
				testdraw.MustBrailleLine(bc, image.Point{0, 4}, image.Point{0, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 2x1",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 1),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{1, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 2x2",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 2),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{1, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 2x3",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 3),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{1, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{1, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 2x4",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 4),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{1, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{1, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 3}, image.Point{1, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 2x5",
			cellCanvas: image.Rect(0, 0, 1, 2),
			ar:         image.Rect(0, 0, 2, 5),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{1, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{1, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 3}, image.Point{1, 3})
				testdraw.MustBrailleLine(bc, image.Point{0, 4}, image.Point{1, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 3x1",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 1),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{2, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 3x2",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 2),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 3x3",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 3),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 2}, image.Point{1, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 3x4",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 4),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{2, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{1, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 3x5",
			cellCanvas: image.Rect(0, 0, 2, 2),
			ar:         image.Rect(0, 0, 3, 5),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{1, 1}, image.Point{1, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{2, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{1, 3})
				testdraw.MustBrailleLine(bc, image.Point{1, 4}, image.Point{1, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 4x1",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 1),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{3, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 4x2",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 2),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{3, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 4x3",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 3),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{3, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 2}, image.Point{2, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 4x4",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{3, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{3, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{2, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 4x5",
			cellCanvas: image.Rect(0, 0, 2, 2),
			ar:         image.Rect(0, 0, 4, 5),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{1, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{3, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{2, 3})
				testdraw.MustBrailleLine(bc, image.Point{1, 4}, image.Point{2, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 5x1",
			cellCanvas: image.Rect(0, 0, 3, 1),
			ar:         image.Rect(0, 0, 5, 1),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{4, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 5x2",
			cellCanvas: image.Rect(0, 0, 3, 1),
			ar:         image.Rect(0, 0, 5, 2),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 5x3",
			cellCanvas: image.Rect(0, 0, 3, 1),
			ar:         image.Rect(0, 0, 5, 3),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 2}, image.Point{3, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 5x4",
			cellCanvas: image.Rect(0, 0, 3, 1),
			ar:         image.Rect(0, 0, 5, 4),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{4, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{3, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 5x5",
			cellCanvas: image.Rect(0, 0, 3, 2),
			ar:         image.Rect(0, 0, 5, 5),
			st:         SegmentTypeHorizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{1, 1}, image.Point{3, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{4, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{3, 3})
				testdraw.MustBrailleLine(bc, image.Point{2, 4}, image.Point{2, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s st:%v", tc.desc, tc.st), func(t *testing.T) {
			bc, err := braille.New(tc.cellCanvas)
			if err != nil {
				t.Fatalf("braille.New => unexpected error: %v", err)
			}

			err = Draw(bc, tc.ar, tc.st, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("Draw => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			size := area.Size(tc.cellCanvas)
			want := faketerm.MustNew(size)
			if tc.want != nil {
				want = tc.want(size)
			}

			got, err := faketerm.New(size)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}
			if err := bc.Apply(got); err != nil {
				t.Fatalf("bc.Apply => unexpected error: %v", err)
			}
			if diff := faketerm.Diff(want, got); diff != "" {
				t.Fatalf("BrailleLine => %v", diff)
			}
		})
	}
}

func TestAdjustHoriz(t *testing.T) {
	tests := []struct {
		desc      string
		start     image.Point
		end       image.Point
		segWidth  int
		adjust    int
		wantStart image.Point
		wantEnd   image.Point
	}{
		{
			desc: "no change for zero adjustment",
		},
		{
			desc:      "safe adjustments, points don't cross",
			start:     image.Point{0, 0},
			end:       image.Point{5, 0},
			segWidth:  6,
			adjust:    1,
			wantStart: image.Point{1, 0},
			wantEnd:   image.Point{4, 0},
		},
		{
			desc:      "safe adjustments, points land on each other",
			start:     image.Point{0, 0},
			end:       image.Point{4, 0},
			segWidth:  5,
			adjust:    2,
			wantStart: image.Point{2, 0},
			wantEnd:   image.Point{2, 0},
		},

		{
			desc:      "points cross, width divides evenly",
			start:     image.Point{0, 0},
			end:       image.Point{5, 0},
			segWidth:  6,
			adjust:    3,
			wantStart: image.Point{2, 0},
			wantEnd:   image.Point{3, 0},
		},
		{
			desc:      "points cross, width divides oddly",
			start:     image.Point{0, 0},
			end:       image.Point{6, 0},
			segWidth:  7,
			adjust:    4,
			wantStart: image.Point{3, 0},
			wantEnd:   image.Point{3, 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotStart, gotEnd := adjustHoriz(tc.start, tc.end, tc.segWidth, tc.adjust)
			if !gotStart.Eq(tc.wantStart) || !gotEnd.Eq(tc.wantEnd) {
				t.Errorf("adjustHoriz(%v, %v, %v, %v) => %v, %v, want %v, %v", tc.start, tc.end, tc.segWidth, tc.adjust, gotStart, gotEnd, tc.wantStart, tc.wantEnd)
			}

		})
	}
}
