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
	"github.com/mum4k/termdash/canvas/testcanvas"
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
		{
			desc:       "segment width of two",
			cellCanvas: image.Rect(0, 0, MinCols*2, MinRows*2),
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

				testsegment.MustHV(bc, image.Rect(2, 0, 10, 2), segment.Horizontal)  // A1
				testsegment.MustHV(bc, image.Rect(12, 0, 20, 2), segment.Horizontal) // A2

				testsegment.MustHV(bc, image.Rect(0, 2, 2, 18), segment.Vertical)                             // F
				testsegment.MustHV(bc, image.Rect(10, 2, 12, 18), segment.Vertical, segment.SkipSlopesLTE(2)) // J
				testsegment.MustHV(bc, image.Rect(20, 2, 22, 18), segment.Vertical, segment.ReverseSlopes())  // B

				testsegment.MustHV(bc, image.Rect(2, 18, 10, 20), segment.Horizontal, segment.SkipSlopesLTE(2))  // G1
				testsegment.MustHV(bc, image.Rect(12, 18, 20, 20), segment.Horizontal, segment.SkipSlopesLTE(2)) // G2

				testsegment.MustHV(bc, image.Rect(0, 20, 2, 36), segment.Vertical)                             // E
				testsegment.MustHV(bc, image.Rect(10, 20, 12, 36), segment.Vertical, segment.SkipSlopesLTE(2)) // M
				testsegment.MustHV(bc, image.Rect(20, 20, 22, 36), segment.Vertical, segment.ReverseSlopes())  // C

				testsegment.MustHV(bc, image.Rect(2, 36, 10, 38), segment.Horizontal, segment.ReverseSlopes())  // D1
				testsegment.MustHV(bc, image.Rect(12, 36, 20, 38), segment.Horizontal, segment.ReverseSlopes()) // D2

				testsegment.MustDiagonal(bc, image.Rect(3, 3, 9, 17), 2, segment.LeftToRight)    // H
				testsegment.MustDiagonal(bc, image.Rect(13, 3, 19, 17), 2, segment.RightToLeft)  // K
				testsegment.MustDiagonal(bc, image.Rect(3, 21, 9, 35), 2, segment.RightToLeft)   // N
				testsegment.MustDiagonal(bc, image.Rect(13, 21, 19, 35), 2, segment.LeftToRight) // L
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "segment width of three",
			cellCanvas: image.Rect(0, 0, MinCols*3, MinRows*3),
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

				testsegment.MustHV(bc, image.Rect(2, 0, 16, 3), segment.Horizontal)  // A1
				testsegment.MustHV(bc, image.Rect(17, 0, 31, 3), segment.Horizontal) // A2

				testsegment.MustHV(bc, image.Rect(0, 2, 3, 28), segment.Vertical)   // F
				testsegment.MustHV(bc, image.Rect(15, 2, 18, 28), segment.Vertical) // J
				testsegment.MustHV(bc, image.Rect(30, 2, 33, 28), segment.Vertical) // B

				testsegment.MustHV(bc, image.Rect(2, 27, 16, 30), segment.Horizontal)  // G1
				testsegment.MustHV(bc, image.Rect(17, 27, 31, 30), segment.Horizontal) // G2

				testsegment.MustHV(bc, image.Rect(0, 29, 3, 55), segment.Vertical)   // E
				testsegment.MustHV(bc, image.Rect(15, 29, 18, 55), segment.Vertical) // M
				testsegment.MustHV(bc, image.Rect(30, 29, 33, 55), segment.Vertical) // C

				testsegment.MustHV(bc, image.Rect(2, 54, 16, 57), segment.Horizontal)  // D1
				testsegment.MustHV(bc, image.Rect(17, 54, 31, 57), segment.Horizontal) // D2

				testsegment.MustDiagonal(bc, image.Rect(3, 3, 15, 27), 3, segment.LeftToRight)   // H
				testsegment.MustDiagonal(bc, image.Rect(18, 3, 30, 27), 3, segment.RightToLeft)  // K
				testsegment.MustDiagonal(bc, image.Rect(3, 30, 15, 54), 3, segment.RightToLeft)  // N
				testsegment.MustDiagonal(bc, image.Rect(18, 30, 30, 54), 3, segment.LeftToRight) // L
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "segment with even width is changed to odd",
			cellCanvas: image.Rect(0, 0, MinCols*4, MinRows*4),
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

				testsegment.MustHV(bc, image.Rect(4, 0, 21, 5), segment.Horizontal)  // A1
				testsegment.MustHV(bc, image.Rect(24, 0, 41, 5), segment.Horizontal) // A2

				testsegment.MustHV(bc, image.Rect(0, 4, 5, 37), segment.Vertical)   // F
				testsegment.MustHV(bc, image.Rect(20, 4, 25, 37), segment.Vertical) // J
				testsegment.MustHV(bc, image.Rect(40, 4, 45, 37), segment.Vertical) // B

				testsegment.MustHV(bc, image.Rect(4, 36, 21, 41), segment.Horizontal)  // G1
				testsegment.MustHV(bc, image.Rect(24, 36, 41, 41), segment.Horizontal) // G2

				testsegment.MustHV(bc, image.Rect(0, 40, 5, 73), segment.Vertical)   // E
				testsegment.MustHV(bc, image.Rect(20, 40, 25, 73), segment.Vertical) // M
				testsegment.MustHV(bc, image.Rect(40, 40, 45, 73), segment.Vertical) // C

				testsegment.MustHV(bc, image.Rect(4, 72, 21, 77), segment.Horizontal)  // D1
				testsegment.MustHV(bc, image.Rect(24, 72, 41, 77), segment.Horizontal) // D2

				testsegment.MustDiagonal(bc, image.Rect(6, 6, 19, 35), 5, segment.LeftToRight)   // H
				testsegment.MustDiagonal(bc, image.Rect(26, 6, 39, 35), 5, segment.RightToLeft)  // K
				testsegment.MustDiagonal(bc, image.Rect(6, 42, 19, 71), 5, segment.RightToLeft)  // N
				testsegment.MustDiagonal(bc, image.Rect(26, 42, 39, 71), 5, segment.LeftToRight) // L
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "segment with odd width and e√en peak to peak distance is changed to odd",
			cellCanvas: image.Rect(0, 0, MinCols*7, MinRows*7),
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

				testsegment.MustHV(bc, image.Rect(7, 0, 39, 9), segment.Horizontal)  // A1
				testsegment.MustHV(bc, image.Rect(44, 0, 76, 9), segment.Horizontal) // A2

				testsegment.MustHV(bc, image.Rect(0, 7, 9, 67), segment.Vertical)   // F
				testsegment.MustHV(bc, image.Rect(37, 7, 46, 67), segment.Vertical) // J
				testsegment.MustHV(bc, image.Rect(74, 7, 83, 67), segment.Vertical) // B

				testsegment.MustHV(bc, image.Rect(7, 65, 39, 74), segment.Horizontal)  // G1
				testsegment.MustHV(bc, image.Rect(44, 65, 76, 74), segment.Horizontal) // G2

				testsegment.MustHV(bc, image.Rect(0, 72, 9, 132), segment.Vertical)   // E
				testsegment.MustHV(bc, image.Rect(37, 72, 46, 132), segment.Vertical) // M
				testsegment.MustHV(bc, image.Rect(74, 72, 83, 132), segment.Vertical) // C

				testsegment.MustHV(bc, image.Rect(7, 130, 39, 139), segment.Horizontal)  // D1
				testsegment.MustHV(bc, image.Rect(44, 130, 76, 139), segment.Horizontal) // D2

				testsegment.MustDiagonal(bc, image.Rect(10, 10, 36, 64), 9, segment.LeftToRight)  // H
				testsegment.MustDiagonal(bc, image.Rect(47, 10, 73, 64), 9, segment.RightToLeft)  // K
				testsegment.MustDiagonal(bc, image.Rect(10, 75, 36, 129), 9, segment.RightToLeft) // N
				testsegment.MustDiagonal(bc, image.Rect(47, 75, 73, 129), 9, segment.LeftToRight) // L
				testbraille.MustApply(bc, ft)
				return ft
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

// mustDrawSegments returns a fake terminal of the specified size with the
// segments drawn on it or panics.
func mustDrawSegments(size image.Point, seg ...Segment) *faketerm.Terminal {
	ft := faketerm.MustNew(size)
	cvs := testcanvas.MustNew(ft.Area())

	d := New()
	for _, s := range seg {
		if err := d.SetSegment(s); err != nil {
			panic(err)
		}
	}

	if err := d.Draw(cvs); err != nil {
		panic(err)
	}

	testcanvas.MustApply(cvs, ft)
	return ft
}

func TestSetCharacter(t *testing.T) {
	tests := []struct {
		desc string
		char rune
		// If not nil, it is called before Draw is called and can set, clear or
		// toggle segments or characters.
		update  func(*Display) error
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc:    "fails on unsupported character",
			char:    '.',
			wantErr: true,
		},
		{
			desc: "displays ' '",
			char: ' ',
		},
		{
			desc: "doesn't clear the display",
			update: func(d *Display) error {
				return d.SetSegment(A2)
			},
			char: 'W',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, F, E, N, L, C, B, A2)
			},
		},
		{
			desc: "displays '!'",
			char: '!',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, B, C)
			},
		},
		{
			desc: "displays '\"'",
			char: '"',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, J, B)
			},
		},
		{
			desc: "displays '#'",
			char: '#',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, J, B, G1, G2, M, C, D1, D2)
			},
		},
		{
			desc: "displays '$'",
			char: '$',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, J, G1, G2, M, C, D1, D2)
			},
		},
		{
			desc: "displays '%'",
			char: '%',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, F, J, K, G1, G2, N, M, C, D2)
			},
		},
		{
			desc: "displays '&'",
			char: '&',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, H, J, G1, E, L, D1, D2)
			},
		},
		{
			desc: "displays '",
			char: '\'',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, J)
			},
		},
		{
			desc: "displays '('",
			char: '(',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, K, L)
			},
		},
		{
			desc: "displays ')'",
			char: ')',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, H, N)
			},
		},
		{
			desc: "displays '*'",
			char: '*',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, H, J, K, G1, G2, N, M, L)
			},
		},
		{
			desc: "displays '+'",
			char: '+',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, J, G1, G2, M)
			},
		},
		{
			desc: "displays ','",
			char: ',',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, N)
			},
		},
		{
			desc: "displays '-'",
			char: '-',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, G1, G2)
			},
		},
		{
			desc: "displays '/'",
			char: '/',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, N, K)
			},
		},
		{
			desc: "displays '0'",
			char: '0',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, K, B, E, N, C, D1, D2)
			},
		},
		{
			desc: "displays '1'",
			char: '1',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, K, B, C)
			},
		},
		{
			desc: "displays '2'",
			char: '2',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, B, G1, G2, E, D1, D2)
			},
		},
		{
			desc: "displays '3'",
			char: '3',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, B, G2, C, D1, D2)
			},
		},
		{
			desc: "displays '4'",
			char: '4',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, F, B, G1, G2, C)
			},
		},
		{
			desc: "displays '5'",
			char: '5',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, G1, L, D1, D2)
			},
		},
		{
			desc: "displays '6'",
			char: '6',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, G1, G2, E, C, D1, D2)
			},
		},
		{
			desc: "displays '7'",
			char: '7',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, B, C)
			},
		},
		{
			desc: "displays '8'",
			char: '8',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, B, G1, G2, E, C, D1, D2)
			},
		},
		{
			desc: "displays '9'",
			char: '9',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, B, G1, G2, C, D1, D2)
			},
		},
		{
			desc: "displays ':'",
			char: ':',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, J, M)
			},
		},
		{
			desc: "displays ';'",
			char: ';',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, J, N)
			},
		},
		{
			desc: "displays '<'",
			char: '<',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, K, G1, L)
			},
		},
		{
			desc: "displays '='",
			char: '=',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, G1, G2, D1, D2)
			},
		},
		{
			desc: "displays '>'",
			char: '>',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, H, G2, N)
			},
		},
		{
			desc: "displays '?'",
			char: '?',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, B, G2, M)
			},
		},
		{
			desc: "displays '@'",
			char: '@',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, J, B, G2, E, D1, D2)
			},
		},
		{
			desc: "displays 'A'",
			char: 'A',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, B, G1, G2, E, C)
			},
		},
		{
			desc: "displays 'B'",
			char: 'B',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, J, B, G2, M, C, D1, D2)
			},
		},
		{
			desc: "displays 'C'",
			char: 'C',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, E, D1, D2)
			},
		},
		{
			desc: "displays 'D'",
			char: 'D',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, J, B, M, C, D1, D2)
			},
		},
		{
			desc: "displays 'E'",
			char: 'E',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, G1, E, D1, D2)
			},
		},
		{
			desc: "displays 'F'",
			char: 'F',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, G1, E)
			},
		},
		{
			desc: "displays 'G'",
			char: 'G',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, G2, E, C, D1, D2)
			},
		},
		{
			desc: "displays 'H'",
			char: 'H',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, F, B, G1, G2, E, C)
			},
		},
		{
			desc: "displays 'I'",
			char: 'I',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, J, M, D1, D2)
			},
		},
		{
			desc: "displays 'J'",
			char: 'J',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, B, E, C, D1, D2)
			},
		},
		{
			desc: "displays 'K'",
			char: 'K',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, F, K, G1, E, L)
			},
		},
		{
			desc: "displays 'L'",
			char: 'L',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, F, E, D1, D2)
			},
		},
		{
			desc: "displays 'M'",
			char: 'M',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, F, H, K, B, E, C)
			},
		},
		{
			desc: "displays 'N'",
			char: 'N',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, F, H, B, E, L, C)
			},
		},
		{
			desc: "displays 'O'",
			char: 'O',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, B, E, C, D1, D2)
			},
		},
		{
			desc: "displays 'P'",
			char: 'P',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, B, G1, G2, E)
			},
		},
		{
			desc: "displays 'Q'",
			char: 'Q',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, B, E, L, C, D1, D2)
			},
		},
		{
			desc: "displays 'R'",
			char: 'R',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, B, G1, G2, E, L)
			},
		},
		{
			desc: "displays 'S'",
			char: 'S',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, F, G1, G2, C, D1, D2)
			},
		},
		{
			desc: "displays 'T'",
			char: 'T',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, J, M)
			},
		},
		{
			desc: "displays 'U'",
			char: 'U',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, F, B, E, C, D1, D2)
			},
		},
		{
			desc: "displays 'V'",
			char: 'V',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, F, K, E, N)
			},
		},
		{
			desc: "displays 'W'",
			char: 'W',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, F, E, N, L, C, B)
			},
		},
		{
			desc: "displays 'X'",
			char: 'X',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, H, K, N, L)
			},
		},
		{
			desc: "displays 'Y'",
			char: 'Y',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, F, B, G1, G2, C, D1, D2)
			},
		},
		{
			desc: "displays 'Z'",
			char: 'Z',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, A2, K, N, D1, D2)
			},
		},
		{
			desc: "displays '['",
			char: '[',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A2, J, M, D2)
			},
		},
		{
			desc: "displays '\\'",
			char: '\\',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, H, L)
			},
		},
		{
			desc: "displays ']'",
			char: ']',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, A1, J, M, D1)
			},
		},
		{
			desc: "displays '^'",
			char: '^',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, N, L)
			},
		},
		{
			desc: "displays '_'",
			char: '_',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, D1, D2)
			},
		},
		{
			desc: "displays '`'",
			char: '`',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, H)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			d := New()
			if tc.update != nil {
				err := tc.update(d)
				if err != nil {
					t.Fatalf("tc.update => unexpected error: %v", err)
				}
			}

			{
				err := d.SetCharacter(tc.char)
				if (err != nil) != tc.wantErr {
					t.Errorf("SetCharacter => unexpected error: %v, wantErr: %v", err, tc.wantErr)
				}
				if err != nil {
					return
				}
			}

			ar := image.Rect(0, 0, MinCols, MinRows)
			cvs, err := canvas.New(ar)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			if err := d.Draw(cvs); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			size := area.Size(ar)
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
				t.Fatalf("SetCharacter => %v", diff)
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
			str:        " w.W :",
			wantRes:    false,
			wantUnsupp: []rune{'.'},
		},
		{
			desc:       "supports no chars in the string",
			str:        ".",
			wantRes:    false,
			wantUnsupp: []rune{'.'},
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
			str:  " w.W:",
			want: " w W:",
		},
		{
			desc: "no characters are supported",
			str:  ".",
			want: " ",
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
