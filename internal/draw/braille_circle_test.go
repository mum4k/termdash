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

package draw

import (
	"image"
	"testing"

	"termdash/cell"
	"termdash/internal/area"
	"termdash/internal/canvas/braille"
	"termdash/internal/canvas/braille/testbraille"
	"termdash/internal/faketerm"
)

// mustBrailleLine draws the braille line or panics.
func mustBrailleLine(bc *braille.Canvas, start, end image.Point, opts ...BrailleLineOption) {
	if err := BrailleLine(bc, start, end, opts...); err != nil {
		panic(err)
	}
}

func TestBrailleCircle(t *testing.T) {
	tests := []struct {
		desc   string
		canvas image.Rectangle
		mid    image.Point
		radius int

		// If not nil, called to prepare the braille canvas before running the test.
		prepare func(*braille.Canvas) error

		opts    []BrailleCircleOption
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc:    "fails when mid isn't in the canvas",
			canvas:  image.Rect(0, 0, 1, 1),
			mid:     image.Point{-1, 0},
			radius:  1,
			wantErr: true,
		},
		{
			desc:    "fails when radius too small",
			canvas:  image.Rect(0, 0, 1, 1),
			mid:     image.Point{0, 0},
			radius:  0,
			wantErr: true,
		},
		{
			desc:    "fails when the circle doesn't fit",
			canvas:  image.Rect(0, 0, 1, 1),
			mid:     image.Point{0, 0},
			radius:  2,
			wantErr: true,
		},
		{
			desc:   "fails when clearing a circle that doesn't fit",
			canvas: image.Rect(0, 0, 1, 1),
			mid:    image.Point{0, 0},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleClearPixels(),
			},
			wantErr: true,
		},
		{
			desc:   "fails when the filled circle doesn't fit",
			canvas: image.Rect(0, 0, 1, 1),
			mid:    image.Point{0, 0},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleFilled(),
			},
			wantErr: true,
		},
		{
			desc:   "fails on arc start too small",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(-1, 0),
			},
			wantErr: true,
		},
		{
			desc:   "fails on arc start too large",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(361, 0),
			},
			wantErr: true,
		},
		{
			desc:   "fails on arc end too small",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(0, -1),
			},
			wantErr: true,
		},
		{
			desc:   "fails on arc end too large",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(0, 361),
			},
			wantErr: true,
		},
		{
			desc:   "fails on arc start and end equal",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(7, 7),
			},
			wantErr: true,
		},
		{
			desc:   "empty circle with radius of two",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{3, 0})

				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{4, 1})

				testbraille.MustSetPixel(bc, image.Point{0, 2})
				testbraille.MustSetPixel(bc, image.Point{4, 2})

				testbraille.MustSetPixel(bc, image.Point{0, 3})
				testbraille.MustSetPixel(bc, image.Point{4, 3})

				testbraille.MustSetPixel(bc, image.Point{1, 4})
				testbraille.MustSetPixel(bc, image.Point{2, 4})
				testbraille.MustSetPixel(bc, image.Point{3, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "sets cell options",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleCellOpts(cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorGreen)),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				opts := []cell.Option{
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorGreen),
				}
				testbraille.MustSetPixel(bc, image.Point{1, 0}, opts...)
				testbraille.MustSetPixel(bc, image.Point{2, 0}, opts...)
				testbraille.MustSetPixel(bc, image.Point{3, 0}, opts...)

				testbraille.MustSetPixel(bc, image.Point{0, 1}, opts...)
				testbraille.MustSetPixel(bc, image.Point{4, 1}, opts...)

				testbraille.MustSetPixel(bc, image.Point{0, 2}, opts...)
				testbraille.MustSetPixel(bc, image.Point{4, 2}, opts...)

				testbraille.MustSetPixel(bc, image.Point{0, 3}, opts...)
				testbraille.MustSetPixel(bc, image.Point{4, 3}, opts...)

				testbraille.MustSetPixel(bc, image.Point{1, 4}, opts...)
				testbraille.MustSetPixel(bc, image.Point{2, 4}, opts...)
				testbraille.MustSetPixel(bc, image.Point{3, 4}, opts...)

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "empty circle with radius of two, specified as arc",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(0, 360),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{3, 0})

				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{4, 1})

				testbraille.MustSetPixel(bc, image.Point{0, 2})
				testbraille.MustSetPixel(bc, image.Point{4, 2})

				testbraille.MustSetPixel(bc, image.Point{0, 3})
				testbraille.MustSetPixel(bc, image.Point{4, 3})

				testbraille.MustSetPixel(bc, image.Point{1, 4})
				testbraille.MustSetPixel(bc, image.Point{2, 4})
				testbraille.MustSetPixel(bc, image.Point{3, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "empty circle with radius of two, specified as inverse arc",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(360, 0),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{3, 0})

				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{4, 1})

				testbraille.MustSetPixel(bc, image.Point{0, 2})
				testbraille.MustSetPixel(bc, image.Point{4, 2})

				testbraille.MustSetPixel(bc, image.Point{0, 3})
				testbraille.MustSetPixel(bc, image.Point{4, 3})

				testbraille.MustSetPixel(bc, image.Point{1, 4})
				testbraille.MustSetPixel(bc, image.Point{2, 4})
				testbraille.MustSetPixel(bc, image.Point{3, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "filled circle with radius of two",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleFilled(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 0})
				mustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1})
				mustBrailleLine(bc, image.Point{0, 2}, image.Point{4, 2})
				mustBrailleLine(bc, image.Point{0, 3}, image.Point{4, 3})
				mustBrailleLine(bc, image.Point{1, 4}, image.Point{3, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "clears pixels on a filled circle",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			prepare: func(bc *braille.Canvas) error {
				// Draw a filled circle so we can erase part of it.
				return BrailleCircle(bc, image.Point{2, 2}, 2, BrailleCircleFilled())
			},
			opts: []BrailleCircleOption{
				BrailleCircleClearPixels(),
				BrailleCircleFilled(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 0})
				mustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1})
				mustBrailleLine(bc, image.Point{0, 2}, image.Point{4, 2})
				mustBrailleLine(bc, image.Point{0, 3}, image.Point{4, 3})
				mustBrailleLine(bc, image.Point{1, 4}, image.Point{3, 4})

				mustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 0}, BrailleLineClearPixels())
				mustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1}, BrailleLineClearPixels())
				mustBrailleLine(bc, image.Point{0, 2}, image.Point{4, 2}, BrailleLineClearPixels())
				mustBrailleLine(bc, image.Point{0, 3}, image.Point{4, 3}, BrailleLineClearPixels())
				mustBrailleLine(bc, image.Point{1, 4}, image.Point{3, 4}, BrailleLineClearPixels())

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "clears pixels by drawing a smaller filled circle",
			canvas: image.Rect(0, 0, 4, 4),
			mid:    image.Point{3, 3},
			radius: 2,
			prepare: func(bc *braille.Canvas) error {
				// Draw a filled circle so we can erase part of it.
				return BrailleCircle(bc, image.Point{3, 3}, 3, BrailleCircleFilled())
			},
			opts: []BrailleCircleOption{
				BrailleCircleClearPixels(),
				BrailleCircleFilled(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{1, 0}, image.Point{5, 0})

				mustBrailleLine(bc, image.Point{0, 1}, image.Point{1, 1})
				mustBrailleLine(bc, image.Point{5, 1}, image.Point{6, 1})

				mustBrailleLine(bc, image.Point{0, 2}, image.Point{0, 2})
				mustBrailleLine(bc, image.Point{6, 2}, image.Point{6, 2})
				mustBrailleLine(bc, image.Point{0, 3}, image.Point{0, 3})
				mustBrailleLine(bc, image.Point{6, 3}, image.Point{6, 3})
				mustBrailleLine(bc, image.Point{0, 4}, image.Point{0, 4})
				mustBrailleLine(bc, image.Point{6, 4}, image.Point{6, 4})

				mustBrailleLine(bc, image.Point{0, 5}, image.Point{1, 5})
				mustBrailleLine(bc, image.Point{5, 5}, image.Point{6, 5})

				mustBrailleLine(bc, image.Point{1, 6}, image.Point{5, 6})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "filled circle with radius of two, specified as arc",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleFilled(),
				BrailleCircleArcOnly(0, 360),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 0})
				mustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1})
				mustBrailleLine(bc, image.Point{0, 2}, image.Point{4, 2})
				mustBrailleLine(bc, image.Point{0, 3}, image.Point{4, 3})
				mustBrailleLine(bc, image.Point{1, 4}, image.Point{3, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "filled partial circle that is almost full",
			canvas: image.Rect(0, 0, 5, 5),
			mid:    image.Point{4, 4},
			radius: 4,
			opts: []BrailleCircleOption{
				BrailleCircleFilled(),
				BrailleCircleArcOnly(150, 90),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{4, 0}, image.Point{6, 0})
				mustBrailleLine(bc, image.Point{4, 1}, image.Point{7, 1})

				mustBrailleLine(bc, image.Point{0, 2}, image.Point{1, 2})
				mustBrailleLine(bc, image.Point{4, 2}, image.Point{8, 2})

				mustBrailleLine(bc, image.Point{0, 3}, image.Point{8, 3})
				mustBrailleLine(bc, image.Point{0, 4}, image.Point{8, 4})
				mustBrailleLine(bc, image.Point{0, 5}, image.Point{8, 5})
				mustBrailleLine(bc, image.Point{0, 6}, image.Point{8, 6})

				mustBrailleLine(bc, image.Point{1, 7}, image.Point{7, 7})
				mustBrailleLine(bc, image.Point{2, 8}, image.Point{6, 8})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "filled partial circle in the first quadrant that is very small",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleFilled(),
				BrailleCircleArcOnly(30, 90),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{2, 0}, image.Point{3, 0})
				mustBrailleLine(bc, image.Point{2, 1}, image.Point{4, 1})
				mustBrailleLine(bc, image.Point{2, 2}, image.Point{3, 2})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "filled partial circle in the third quadrant, 36 degree arc",
			canvas: image.Rect(0, 0, 4, 4),
			mid:    image.Point{4, 4},
			radius: 3,
			opts: []BrailleCircleOption{
				BrailleCircleFilled(),
				BrailleCircleArcOnly(194, 230),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{3, 4}, image.Point{4, 4})
				mustBrailleLine(bc, image.Point{1, 5}, image.Point{3, 5})
				mustBrailleLine(bc, image.Point{1, 6}, image.Point{2, 6})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "filled partial circle in the third quadrant, four degree arc",
			canvas: image.Rect(0, 0, 4, 4),
			mid:    image.Point{4, 4},
			radius: 3,
			opts: []BrailleCircleOption{
				BrailleCircleFilled(),
				BrailleCircleArcOnly(226, 230),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{4, 4})
				testbraille.MustSetPixel(bc, image.Point{3, 5})
				testbraille.MustSetPixel(bc, image.Point{2, 6})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "filled partial circle in the first quadrant, four degree arc",
			canvas: image.Rect(0, 0, 6, 6),
			mid:    image.Point{6, 6},
			radius: 5,
			opts: []BrailleCircleOption{
				BrailleCircleFilled(),
				BrailleCircleArcOnly(66, 70),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{8, 1})
				testbraille.MustSetPixel(bc, image.Point{8, 2})
				testbraille.MustSetPixel(bc, image.Point{7, 3})
				testbraille.MustSetPixel(bc, image.Point{7, 4})
				testbraille.MustSetPixel(bc, image.Point{6, 5})
				testbraille.MustSetPixel(bc, image.Point{6, 6})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "filled partial circle in the third quadrant, 72 degree arc",
			canvas: image.Rect(0, 0, 4, 4),
			mid:    image.Point{4, 4},
			radius: 3,
			opts: []BrailleCircleOption{
				BrailleCircleFilled(),
				BrailleCircleArcOnly(158, 230),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{1, 3}, image.Point{2, 3})
				mustBrailleLine(bc, image.Point{1, 4}, image.Point{4, 4})
				mustBrailleLine(bc, image.Point{1, 5}, image.Point{3, 5})
				mustBrailleLine(bc, image.Point{1, 6}, image.Point{2, 6})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "filled circle with radius of two and cell options",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleFilled(),
				BrailleCircleCellOpts(cell.FgColor(cell.ColorRed)),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				opts := []BrailleLineOption{
					BrailleLineCellOpts(cell.FgColor(cell.ColorRed)),
				}
				mustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 0}, opts...)
				mustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1}, opts...)
				mustBrailleLine(bc, image.Point{0, 2}, image.Point{4, 2}, opts...)
				mustBrailleLine(bc, image.Point{0, 3}, image.Point{4, 3}, opts...)
				mustBrailleLine(bc, image.Point{1, 4}, image.Point{3, 4}, opts...)

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, partial first quadrant",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(20, 70),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{3, 0})
				testbraille.MustSetPixel(bc, image.Point{4, 1})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, full first quadrant",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(0, 90),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{3, 0})
				testbraille.MustSetPixel(bc, image.Point{4, 1})
				testbraille.MustSetPixel(bc, image.Point{4, 2})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw filled arc only, full first quadrant",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(0, 90),
				BrailleCircleFilled(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{2, 0}, image.Point{3, 0})
				mustBrailleLine(bc, image.Point{2, 1}, image.Point{4, 1})
				mustBrailleLine(bc, image.Point{2, 2}, image.Point{4, 2})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, partial second quadrant",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(100, 170),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{0, 1})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, full second quadrant",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(90, 180),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{0, 2})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw filled arc only, full second quadrant",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(90, 180),
				BrailleCircleFilled(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{1, 0}, image.Point{2, 0})
				mustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				mustBrailleLine(bc, image.Point{0, 2}, image.Point{2, 2})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, partial third quadrant",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(190, 260),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 3})
				testbraille.MustSetPixel(bc, image.Point{1, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, full third quadrant",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(180, 270),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 2})
				testbraille.MustSetPixel(bc, image.Point{0, 3})
				testbraille.MustSetPixel(bc, image.Point{1, 4})
				testbraille.MustSetPixel(bc, image.Point{2, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw filled arc only, full third quadrant",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(180, 270),
				BrailleCircleFilled(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{0, 2}, image.Point{2, 2})
				mustBrailleLine(bc, image.Point{0, 3}, image.Point{2, 3})
				mustBrailleLine(bc, image.Point{1, 4}, image.Point{2, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, partial fourth quadrant",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(280, 350),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{4, 3})
				testbraille.MustSetPixel(bc, image.Point{3, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, full fourth quadrant",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(260, 360),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{4, 2})
				testbraille.MustSetPixel(bc, image.Point{4, 3})
				testbraille.MustSetPixel(bc, image.Point{2, 4})
				testbraille.MustSetPixel(bc, image.Point{3, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw filled arc only, full fourth quadrant",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(270, 360),
				BrailleCircleFilled(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{2, 2}, image.Point{4, 2})
				mustBrailleLine(bc, image.Point{2, 3}, image.Point{4, 3})
				mustBrailleLine(bc, image.Point{2, 4}, image.Point{3, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, partial two quadrants first and second",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(10, 170),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{3, 0})
				testbraille.MustSetPixel(bc, image.Point{4, 1})

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{0, 1})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, full two quadrants first and second",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(0, 180),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{3, 0})
				testbraille.MustSetPixel(bc, image.Point{4, 1})
				testbraille.MustSetPixel(bc, image.Point{4, 2})

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{0, 2})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw filled arc only, full two quadrants first and second",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(0, 180),
				BrailleCircleFilled(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 0})
				mustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1})
				mustBrailleLine(bc, image.Point{0, 2}, image.Point{4, 2})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, partial two quadrants second and third",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(100, 260),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{0, 2})

				testbraille.MustSetPixel(bc, image.Point{0, 2})
				testbraille.MustSetPixel(bc, image.Point{0, 3})
				testbraille.MustSetPixel(bc, image.Point{1, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, full two quadrants second and third",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(90, 270),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{0, 2})

				testbraille.MustSetPixel(bc, image.Point{0, 2})
				testbraille.MustSetPixel(bc, image.Point{0, 3})
				testbraille.MustSetPixel(bc, image.Point{1, 4})
				testbraille.MustSetPixel(bc, image.Point{2, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw filled arc only, full two quadrants second and third",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(90, 270),
				BrailleCircleFilled(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{1, 0}, image.Point{2, 0})
				mustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				mustBrailleLine(bc, image.Point{0, 2}, image.Point{2, 2})
				mustBrailleLine(bc, image.Point{0, 2}, image.Point{2, 2})
				mustBrailleLine(bc, image.Point{0, 3}, image.Point{2, 3})
				mustBrailleLine(bc, image.Point{1, 4}, image.Point{2, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, partial two quadrants third and fourth",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(190, 350),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 3})
				testbraille.MustSetPixel(bc, image.Point{1, 4})
				testbraille.MustSetPixel(bc, image.Point{2, 4})

				testbraille.MustSetPixel(bc, image.Point{4, 3})
				testbraille.MustSetPixel(bc, image.Point{2, 4})
				testbraille.MustSetPixel(bc, image.Point{3, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, full two quadrants third and fourth",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(180, 360),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 2})
				testbraille.MustSetPixel(bc, image.Point{0, 3})
				testbraille.MustSetPixel(bc, image.Point{1, 4})
				testbraille.MustSetPixel(bc, image.Point{2, 4})

				testbraille.MustSetPixel(bc, image.Point{4, 2})
				testbraille.MustSetPixel(bc, image.Point{4, 3})
				testbraille.MustSetPixel(bc, image.Point{2, 4})
				testbraille.MustSetPixel(bc, image.Point{3, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw filled arc only, full two quadrants third and fourth",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(180, 360),
				BrailleCircleFilled(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{0, 2}, image.Point{4, 2})
				mustBrailleLine(bc, image.Point{0, 3}, image.Point{4, 3})
				mustBrailleLine(bc, image.Point{1, 4}, image.Point{3, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, partial two quadrants fourth and first",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(280, 80),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{4, 2})
				testbraille.MustSetPixel(bc, image.Point{4, 3})
				testbraille.MustSetPixel(bc, image.Point{3, 4})
				testbraille.MustSetPixel(bc, image.Point{3, 0})
				testbraille.MustSetPixel(bc, image.Point{4, 1})
				testbraille.MustSetPixel(bc, image.Point{4, 2})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, full two quadrants fourth and first",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(270, 90),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{4, 2})
				testbraille.MustSetPixel(bc, image.Point{4, 3})
				testbraille.MustSetPixel(bc, image.Point{2, 4})
				testbraille.MustSetPixel(bc, image.Point{3, 4})
				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{3, 0})
				testbraille.MustSetPixel(bc, image.Point{4, 1})
				testbraille.MustSetPixel(bc, image.Point{4, 2})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw filled arc only, full two quadrants fourth and first",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(270, 90),
				BrailleCircleFilled(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{2, 2}, image.Point{4, 2})
				mustBrailleLine(bc, image.Point{2, 3}, image.Point{4, 3})
				mustBrailleLine(bc, image.Point{2, 4}, image.Point{3, 4})
				mustBrailleLine(bc, image.Point{2, 0}, image.Point{3, 0})
				mustBrailleLine(bc, image.Point{2, 1}, image.Point{4, 1})
				mustBrailleLine(bc, image.Point{2, 2}, image.Point{4, 2})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, full three quarters first, second and third",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(0, 270),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{3, 0})
				testbraille.MustSetPixel(bc, image.Point{4, 1})
				testbraille.MustSetPixel(bc, image.Point{4, 2})

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{0, 2})

				testbraille.MustSetPixel(bc, image.Point{0, 2})
				testbraille.MustSetPixel(bc, image.Point{0, 3})
				testbraille.MustSetPixel(bc, image.Point{1, 4})
				testbraille.MustSetPixel(bc, image.Point{2, 4})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draw empty arc only, full three quarters fourth, first and second",
			canvas: image.Rect(0, 0, 3, 3),
			mid:    image.Point{2, 2},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleArcOnly(270, 180),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{4, 2})
				testbraille.MustSetPixel(bc, image.Point{4, 3})
				testbraille.MustSetPixel(bc, image.Point{2, 4})
				testbraille.MustSetPixel(bc, image.Point{3, 4})

				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{3, 0})
				testbraille.MustSetPixel(bc, image.Point{4, 1})
				testbraille.MustSetPixel(bc, image.Point{4, 2})

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{0, 2})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			bc, err := braille.New(tc.canvas)
			if err != nil {
				t.Fatalf("braille.New => unexpected error: %v", err)
			}

			if tc.prepare != nil {
				if err := tc.prepare(bc); err != nil {
					t.Fatalf("tc.prepare => unexpected error: %v", err)
				}
			}

			err = BrailleCircle(bc, tc.mid, tc.radius, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("BrailleCircle => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			size := area.Size(tc.canvas)
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
				t.Fatalf("BrailleCircle => %v", diff)
			}

		})
	}
}
