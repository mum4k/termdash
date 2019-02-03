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

package sixteen

import (
	"image"
	"sort"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/braille/testbraille"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw/segdisp/segment"
	"github.com/mum4k/termdash/draw/segdisp/segment/testsegment"
	"github.com/mum4k/termdash/terminal/faketerm"
)

func TestDraw(t *testing.T) {
	tests := []struct {
		desc       string
		opts       []Option
		drawOpts   []Option
		cellCanvas image.Rectangle
		// If not nil, it is called before Draw is called and can set, clear or
		// toggle segments or characters.
		update        func(*Display) error
		want          func(size image.Point) *faketerm.Terminal
		wantErr       bool
		wantUpdateErr bool
	}{
		{
			desc:       "fails for area not wide enough",
			cellCanvas: image.Rect(0, 0, MinCols-1, MinRows),
			wantErr:    true,
		},
		{
			desc:       "fails for area not tall enough",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows-1),
			wantErr:    true,
		},
		{
			desc:       "fails to set invalid segment (too small)",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(Segment(-1))
			},
			wantUpdateErr: true,
		},
		{
			desc:       "fails to set invalid segment (too large)",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(Segment(segmentMax))
			},
			wantUpdateErr: true,
		},
		{
			desc:       "fails to clear invalid segment (too small)",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.ClearSegment(Segment(-1))
			},
			wantUpdateErr: true,
		},
		{
			desc:       "fails to clear invalid segment (too large)",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.ClearSegment(Segment(segmentMax))
			},
			wantUpdateErr: true,
		},
		{
			desc:       "fails to toggle invalid segment (too small)",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.ToggleSegment(Segment(-1))
			},
			wantUpdateErr: true,
		},
		{
			desc:       "fails to toggle invalid segment (too large)",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.ToggleSegment(Segment(segmentMax))
			},
			wantUpdateErr: true,
		},
		{
			desc:       "empty when no segments set",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
		},
		{
			desc:       "smallest valid display 6x5, A1",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(A1)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(1, 0, 4, 1), segment.Horizontal) // A1
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, A2",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(A2)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(5, 0, 8, 1), segment.Horizontal) // A2
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, F",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(F)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(0, 1, 1, 8), segment.Vertical) // F
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, J",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(J)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(4, 1, 5, 8), segment.Vertical) // J
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, B",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(B)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(8, 1, 9, 8), segment.Vertical) // B
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, G1",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(G1)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(1, 8, 4, 9), segment.Horizontal) // G1
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, G2",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(G2)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(5, 8, 8, 9), segment.Horizontal) // G2
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, E",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(E)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(0, 9, 1, 16), segment.Vertical) // E
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, M",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(M)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(4, 9, 5, 16), segment.Vertical) // M
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, C",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(C)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(8, 9, 9, 16), segment.Vertical) // C
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, D1",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(D1)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(1, 16, 4, 17), segment.Horizontal) // D1
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, D2",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(D2)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(5, 16, 8, 17), segment.Horizontal) // D2
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, H",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(H)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustDiagonal(bc, image.Rect(1, 1, 4, 8), 1, segment.LeftToRight) // H
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, K",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(K)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustDiagonal(bc, image.Rect(5, 1, 8, 8), 1, segment.RightToLeft) // K
				testbraille.MustApply(bc, ft)
				return ft
			},
		},

		{
			desc:       "smallest valid display 6x5, N",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(N)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustDiagonal(bc, image.Rect(1, 9, 4, 16), 1, segment.RightToLeft) // N
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, L",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				return d.SetSegment(L)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustDiagonal(bc, image.Rect(5, 9, 8, 16), 1, segment.LeftToRight) // L
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, all segments",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				for _, s := range AllSegments() {
					if err := d.SetSegment(s); err != nil {
						return err
					}
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(1, 0, 4, 1), segment.Horizontal) // A1
				testsegment.MustHV(bc, image.Rect(5, 0, 8, 1), segment.Horizontal) // A2

				testsegment.MustHV(bc, image.Rect(0, 1, 1, 8), segment.Vertical) // F
				testsegment.MustHV(bc, image.Rect(4, 1, 5, 8), segment.Vertical) // J
				testsegment.MustHV(bc, image.Rect(8, 1, 9, 8), segment.Vertical) // B

				testsegment.MustHV(bc, image.Rect(1, 8, 4, 9), segment.Horizontal) // G1
				testsegment.MustHV(bc, image.Rect(5, 8, 8, 9), segment.Horizontal) // G2

				testsegment.MustHV(bc, image.Rect(0, 9, 1, 16), segment.Vertical) // E
				testsegment.MustHV(bc, image.Rect(4, 9, 5, 16), segment.Vertical) // M
				testsegment.MustHV(bc, image.Rect(8, 9, 9, 16), segment.Vertical) // C

				testsegment.MustHV(bc, image.Rect(1, 16, 4, 17), segment.Horizontal) // D1
				testsegment.MustHV(bc, image.Rect(5, 16, 8, 17), segment.Horizontal) // D2

				testsegment.MustDiagonal(bc, image.Rect(1, 1, 4, 8), 1, segment.LeftToRight)  // H
				testsegment.MustDiagonal(bc, image.Rect(5, 1, 8, 8), 1, segment.RightToLeft)  // K
				testsegment.MustDiagonal(bc, image.Rect(1, 9, 4, 16), 1, segment.RightToLeft) // N
				testsegment.MustDiagonal(bc, image.Rect(5, 9, 8, 16), 1, segment.LeftToRight) // L
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc: "smallest valid display 6x5, all segments, sets cell options provided to new",
			opts: []Option{
				CellOpts(
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorGreen),
				),
			},
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				for _, s := range AllSegments() {
					if err := d.SetSegment(s); err != nil {
						return err
					}
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				cOpts := []cell.Option{
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorGreen),
				}
				testsegment.MustHV(bc, image.Rect(1, 0, 4, 1), segment.Horizontal, segment.CellOpts(cOpts...)) // A1
				testsegment.MustHV(bc, image.Rect(5, 0, 8, 1), segment.Horizontal, segment.CellOpts(cOpts...)) // A2

				testsegment.MustHV(bc, image.Rect(0, 1, 1, 8), segment.Vertical, segment.CellOpts(cOpts...)) // F
				testsegment.MustHV(bc, image.Rect(4, 1, 5, 8), segment.Vertical, segment.CellOpts(cOpts...)) // J
				testsegment.MustHV(bc, image.Rect(8, 1, 9, 8), segment.Vertical, segment.CellOpts(cOpts...)) // B

				testsegment.MustHV(bc, image.Rect(1, 8, 4, 9), segment.Horizontal, segment.CellOpts(cOpts...)) // G1
				testsegment.MustHV(bc, image.Rect(5, 8, 8, 9), segment.Horizontal, segment.CellOpts(cOpts...)) // G2

				testsegment.MustHV(bc, image.Rect(0, 9, 1, 16), segment.Vertical, segment.CellOpts(cOpts...)) // E
				testsegment.MustHV(bc, image.Rect(4, 9, 5, 16), segment.Vertical, segment.CellOpts(cOpts...)) // M
				testsegment.MustHV(bc, image.Rect(8, 9, 9, 16), segment.Vertical, segment.CellOpts(cOpts...)) // C

				testsegment.MustHV(bc, image.Rect(1, 16, 4, 17), segment.Horizontal, segment.CellOpts(cOpts...)) // D1
				testsegment.MustHV(bc, image.Rect(5, 16, 8, 17), segment.Horizontal, segment.CellOpts(cOpts...)) // D2

				testsegment.MustDiagonal(bc, image.Rect(1, 1, 4, 8), 1, segment.LeftToRight, segment.DiagonalCellOpts(cOpts...))  // H
				testsegment.MustDiagonal(bc, image.Rect(5, 1, 8, 8), 1, segment.RightToLeft, segment.DiagonalCellOpts(cOpts...))  // K
				testsegment.MustDiagonal(bc, image.Rect(1, 9, 4, 16), 1, segment.RightToLeft, segment.DiagonalCellOpts(cOpts...)) // N
				testsegment.MustDiagonal(bc, image.Rect(5, 9, 8, 16), 1, segment.LeftToRight, segment.DiagonalCellOpts(cOpts...)) // L
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc: "smallest valid display 6x5, all segments, sets cell options provided to draw",
			drawOpts: []Option{
				CellOpts(
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorGreen),
				),
			},
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				for _, s := range AllSegments() {
					if err := d.SetSegment(s); err != nil {
						return err
					}
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				cOpts := []cell.Option{
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorGreen),
				}
				testsegment.MustHV(bc, image.Rect(1, 0, 4, 1), segment.Horizontal, segment.CellOpts(cOpts...)) // A1
				testsegment.MustHV(bc, image.Rect(5, 0, 8, 1), segment.Horizontal, segment.CellOpts(cOpts...)) // A2

				testsegment.MustHV(bc, image.Rect(0, 1, 1, 8), segment.Vertical, segment.CellOpts(cOpts...)) // F
				testsegment.MustHV(bc, image.Rect(4, 1, 5, 8), segment.Vertical, segment.CellOpts(cOpts...)) // J
				testsegment.MustHV(bc, image.Rect(8, 1, 9, 8), segment.Vertical, segment.CellOpts(cOpts...)) // B

				testsegment.MustHV(bc, image.Rect(1, 8, 4, 9), segment.Horizontal, segment.CellOpts(cOpts...)) // G1
				testsegment.MustHV(bc, image.Rect(5, 8, 8, 9), segment.Horizontal, segment.CellOpts(cOpts...)) // G2

				testsegment.MustHV(bc, image.Rect(0, 9, 1, 16), segment.Vertical, segment.CellOpts(cOpts...)) // E
				testsegment.MustHV(bc, image.Rect(4, 9, 5, 16), segment.Vertical, segment.CellOpts(cOpts...)) // M
				testsegment.MustHV(bc, image.Rect(8, 9, 9, 16), segment.Vertical, segment.CellOpts(cOpts...)) // C

				testsegment.MustHV(bc, image.Rect(1, 16, 4, 17), segment.Horizontal, segment.CellOpts(cOpts...)) // D1
				testsegment.MustHV(bc, image.Rect(5, 16, 8, 17), segment.Horizontal, segment.CellOpts(cOpts...)) // D2

				testsegment.MustDiagonal(bc, image.Rect(1, 1, 4, 8), 1, segment.LeftToRight, segment.DiagonalCellOpts(cOpts...))  // H
				testsegment.MustDiagonal(bc, image.Rect(5, 1, 8, 8), 1, segment.RightToLeft, segment.DiagonalCellOpts(cOpts...))  // K
				testsegment.MustDiagonal(bc, image.Rect(1, 9, 4, 16), 1, segment.RightToLeft, segment.DiagonalCellOpts(cOpts...)) // N
				testsegment.MustDiagonal(bc, image.Rect(5, 9, 8, 16), 1, segment.LeftToRight, segment.DiagonalCellOpts(cOpts...)) // L
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "clears the display",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				for _, s := range AllSegments() {
					if err := d.SetSegment(s); err != nil {
						return err
					}
				}
				d.Clear()
				return nil
			},
		},
		{
			desc:       "clears the display and sets cell options",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				d.Clear(CellOpts(cell.FgColor(cell.ColorBlue)))
				return d.SetSegment(A1)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(1, 0, 4, 1), segment.Horizontal, segment.CellOpts(cell.FgColor(cell.ColorBlue))) // A1
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "clears some segments",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				for _, s := range AllSegments() {
					if err := d.SetSegment(s); err != nil {
						return err
					}
				}
				for _, s := range []Segment{A1, A2, G1, G2, D1, D2, L} {
					if err := d.ClearSegment(s); err != nil {
						return err
					}
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(0, 1, 1, 8), segment.Vertical) // F
				testsegment.MustHV(bc, image.Rect(4, 1, 5, 8), segment.Vertical) // J
				testsegment.MustHV(bc, image.Rect(8, 1, 9, 8), segment.Vertical) // B

				testsegment.MustHV(bc, image.Rect(0, 9, 1, 16), segment.Vertical) // E
				testsegment.MustHV(bc, image.Rect(4, 9, 5, 16), segment.Vertical) // M
				testsegment.MustHV(bc, image.Rect(8, 9, 9, 16), segment.Vertical) // C

				testsegment.MustDiagonal(bc, image.Rect(1, 1, 4, 8), 1, segment.LeftToRight)  // H
				testsegment.MustDiagonal(bc, image.Rect(5, 1, 8, 8), 1, segment.RightToLeft)  // K
				testsegment.MustDiagonal(bc, image.Rect(1, 9, 4, 16), 1, segment.RightToLeft) // N

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "toggles some segments off",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				for _, s := range AllSegments() {
					if err := d.SetSegment(s); err != nil {
						return err
					}
				}
				for _, s := range []Segment{A1, A2, G1, G2, D1, D2, L} {
					if err := d.ToggleSegment(s); err != nil {
						return err
					}
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(0, 1, 1, 8), segment.Vertical) // F
				testsegment.MustHV(bc, image.Rect(4, 1, 5, 8), segment.Vertical) // J
				testsegment.MustHV(bc, image.Rect(8, 1, 9, 8), segment.Vertical) // B

				testsegment.MustHV(bc, image.Rect(0, 9, 1, 16), segment.Vertical) // E
				testsegment.MustHV(bc, image.Rect(4, 9, 5, 16), segment.Vertical) // M
				testsegment.MustHV(bc, image.Rect(8, 9, 9, 16), segment.Vertical) // C

				testsegment.MustDiagonal(bc, image.Rect(1, 1, 4, 8), 1, segment.LeftToRight)  // H
				testsegment.MustDiagonal(bc, image.Rect(5, 1, 8, 8), 1, segment.RightToLeft)  // K
				testsegment.MustDiagonal(bc, image.Rect(1, 9, 4, 16), 1, segment.RightToLeft) // N

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "toggles some segments on",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				for _, s := range []Segment{A1, A2, G1, G2, D1, D2, L} {
					if err := d.ToggleSegment(s); err != nil {
						return err
					}
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(1, 0, 4, 1), segment.Horizontal) // A1
				testsegment.MustHV(bc, image.Rect(5, 0, 8, 1), segment.Horizontal) // A2

				testsegment.MustHV(bc, image.Rect(1, 8, 4, 9), segment.Horizontal) // G1
				testsegment.MustHV(bc, image.Rect(5, 8, 8, 9), segment.Horizontal) // G2

				testsegment.MustHV(bc, image.Rect(1, 16, 4, 17), segment.Horizontal) // D1
				testsegment.MustHV(bc, image.Rect(5, 16, 8, 17), segment.Horizontal) // D2

				testsegment.MustDiagonal(bc, image.Rect(5, 9, 8, 16), 1, segment.LeftToRight) // L

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "set is idempotent",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				if err := d.SetSegment(A1); err != nil {
					return err
				}
				return d.SetSegment(A1)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(1, 0, 4, 1), segment.Horizontal) // A1
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "clear is idempotent",
			cellCanvas: image.Rect(0, 0, MinCols, MinRows),
			update: func(d *Display) error {
				if err := d.SetSegment(A1); err != nil {
					return err
				}
				if err := d.ClearSegment(A1); err != nil {
					return err
				}
				return d.ClearSegment(A1)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			d := New(tc.opts...)
			if tc.update != nil {
				err := tc.update(d)
				if (err != nil) != tc.wantUpdateErr {
					t.Errorf("tc.update => unexpected error: %v, wantUpdateErr: %v", err, tc.wantUpdateErr)
				}
				if err != nil {
					return
				}
			}

			cvs, err := canvas.New(tc.cellCanvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			{
				err := d.Draw(cvs, tc.drawOpts...)
				if (err != nil) != tc.wantErr {
					t.Errorf("Draw => unexpected error: %v, wantErr: %v", err, tc.wantErr)
				}
				if err != nil {
					return
				}
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
			if err := cvs.Apply(got); err != nil {
				t.Fatalf("bc.Apply => unexpected error: %v", err)
			}
			if diff := faketerm.Diff(want, got); diff != "" {
				t.Fatalf("Draw => %v", diff)
			}

		})
	}
}

func TestRequired(t *testing.T) {
	tests := []struct {
		desc     string
		cellArea image.Rectangle
		want     image.Rectangle
		wantErr  bool
	}{
		{
			desc:     "fails when area isn't wide enough",
			cellArea: image.Rect(0, 0, MinCols-1, MinRows),
			wantErr:  true,
		},
		{
			desc:     "fails when area isn't tall enough",
			cellArea: image.Rect(0, 0, MinCols, MinRows-1),
			wantErr:  true,
		},
		{
			desc:     "returns same area when no adjustment needed",
			cellArea: image.Rect(0, 0, MinCols, MinRows),
			want:     image.Rect(0, 0, MinCols, MinRows),
		},
		{
			desc:     "adjusts width to aspect ratio",
			cellArea: image.Rect(0, 0, MinCols+100, MinRows),
			want:     image.Rect(0, 0, MinCols, MinRows),
		},
		{
			desc:     "adjusts height to aspect ratio",
			cellArea: image.Rect(0, 0, MinCols, MinRows+100),
			want:     image.Rect(0, 0, MinCols, MinRows),
		},
		{
			desc:     "adjusts larger area to aspect ratio",
			cellArea: image.Rect(0, 0, MinCols*2, MinRows*4),
			want:     image.Rect(0, 0, 12, 10),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := Required(tc.cellArea)
			if (err != nil) != tc.wantErr {
				t.Errorf("Required => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Required => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestAllSegments(t *testing.T) {
	want := []Segment{A1, A2, B, C, D1, D2, E, F, G1, G2, H, J, K, L, M, N}
	got := AllSegments()
	sort.Slice(got, func(i, j int) bool {
		return int(got[i]) < int(got[j])
	})
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("AllSegments => unexpected diff (-want, +got):\n%s", diff)
	}
}

func TestSupportsChars(t *testing.T) {
	tests := []struct {
		desc       string
		str        string
		wantRes    bool
		wantUnsupp []rune
	}{
		{
			desc:    "supports all chars in an empty string",
			wantRes: true,
		},
		{
			desc:    "supports all chars in the string",
			str:     " wW ",
			wantRes: true,
		},
		{
			desc:       "supports some chars in the string",
			str:        " w:!W :",
			wantRes:    false,
			wantUnsupp: []rune{':', '!'},
		},
		{
			desc:       "supports no chars in the string",
			str:        ":!()",
			wantRes:    false,
			wantUnsupp: []rune{':', '!', '(', ')'},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotRes, gotUnsupp := SupportsChars(tc.str)
			if gotRes != tc.wantRes {
				t.Errorf("SupportsChars(%q) => %v, %v, want %v, %v", tc.str, gotRes, gotUnsupp, tc.wantRes, tc.wantUnsupp)
			}

			sort.Slice(gotUnsupp, func(i, j int) bool {
				return gotUnsupp[i] < gotUnsupp[j]
			})
			sort.Slice(tc.wantUnsupp, func(i, j int) bool {
				return tc.wantUnsupp[i] < tc.wantUnsupp[j]
			})
			if diff := pretty.Compare(tc.wantUnsupp, gotUnsupp); diff != "" {
				t.Errorf("SupportsChars => unexpected unsupported characters returned, diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		desc string
		str  string
		want string
	}{
		{
			desc: "no alternation to empty string",
		},
		{
			desc: "all characters are supported",
			str:  " wW",
			want: " wW",
		},
		{
			desc: "some characters are supported",
			str:  " :w!W:",
			want: "  w W ",
		},
		{
			desc: "no characters are supported",
			str:  ":!()",
			want: "    ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := Sanitize(tc.str)
			if got != tc.want {
				t.Errorf("Sanitize => %q, want %q", got, tc.want)
			}
		})
	}
}
