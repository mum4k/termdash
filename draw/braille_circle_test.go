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

	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/canvas/braille"
	"github.com/mum4k/termdash/canvas/braille/testbraille"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/faketerm"
)

// mustBrailleLine draws the braille line or panics.
func mustBrailleLine(bc *braille.Canvas, start, end image.Point, opts ...BrailleLineOption) {
	if err := BrailleLine(bc, start, end, opts...); err != nil {
		panic(err)
	}
}

func TestBrailleCircle(t *testing.T) {
	tests := []struct {
		desc    string
		canvas  image.Rectangle
		mid     image.Point
		radius  int
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
			desc:   "draws a single pixel",
			canvas: image.Rect(0, 0, 1, 1),
			mid:    image.Point{0, 0},
			radius: 1,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())
				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "sets cell options",
			canvas: image.Rect(0, 0, 1, 1),
			mid:    image.Point{0, 0},
			radius: 1,
			opts: []BrailleCircleOption{
				BrailleCircleCellOpts(cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorGreen)),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())
				testbraille.MustSetPixel(bc, image.Point{0, 0}, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorGreen))
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "circle with radius of two",
			canvas: image.Rect(0, 0, 3, 2),
			mid:    image.Point{1, 1},
			radius: 2,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{2, 0})

				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{2, 1})

				testbraille.MustSetPixel(bc, image.Point{0, 2})
				testbraille.MustSetPixel(bc, image.Point{1, 2})
				testbraille.MustSetPixel(bc, image.Point{2, 2})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "filled circle with radius of two",
			canvas: image.Rect(0, 0, 3, 2),
			mid:    image.Point{1, 1},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleFilled(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{2, 0})

				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{1, 1})
				testbraille.MustSetPixel(bc, image.Point{2, 1})

				testbraille.MustSetPixel(bc, image.Point{0, 2})
				testbraille.MustSetPixel(bc, image.Point{1, 2})
				testbraille.MustSetPixel(bc, image.Point{2, 2})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "filled circle with radius of two and cell options",
			canvas: image.Rect(0, 0, 3, 2),
			mid:    image.Point{1, 1},
			radius: 2,
			opts: []BrailleCircleOption{
				BrailleCircleFilled(),
				BrailleCircleCellOpts(cell.FgColor(cell.ColorRed)),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0}, cell.FgColor(cell.ColorRed))
				testbraille.MustSetPixel(bc, image.Point{1, 0}, cell.FgColor(cell.ColorRed))
				testbraille.MustSetPixel(bc, image.Point{2, 0}, cell.FgColor(cell.ColorRed))

				testbraille.MustSetPixel(bc, image.Point{0, 1}, cell.FgColor(cell.ColorRed))
				testbraille.MustSetPixel(bc, image.Point{1, 1}, cell.FgColor(cell.ColorRed))
				testbraille.MustSetPixel(bc, image.Point{2, 1}, cell.FgColor(cell.ColorRed))

				testbraille.MustSetPixel(bc, image.Point{0, 2}, cell.FgColor(cell.ColorRed))
				testbraille.MustSetPixel(bc, image.Point{1, 2}, cell.FgColor(cell.ColorRed))
				testbraille.MustSetPixel(bc, image.Point{2, 2}, cell.FgColor(cell.ColorRed))

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "circle with radius of four",
			canvas: image.Rect(0, 0, 4, 2),
			mid:    image.Point{3, 3},
			radius: 4,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{3, 0})
				testbraille.MustSetPixel(bc, image.Point{4, 0})
				testbraille.MustSetPixel(bc, image.Point{5, 0})

				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{6, 1})
				testbraille.MustSetPixel(bc, image.Point{0, 2})
				testbraille.MustSetPixel(bc, image.Point{6, 2})
				testbraille.MustSetPixel(bc, image.Point{0, 3})
				testbraille.MustSetPixel(bc, image.Point{6, 3})
				testbraille.MustSetPixel(bc, image.Point{0, 4})
				testbraille.MustSetPixel(bc, image.Point{6, 4})
				testbraille.MustSetPixel(bc, image.Point{0, 5})
				testbraille.MustSetPixel(bc, image.Point{6, 5})

				testbraille.MustSetPixel(bc, image.Point{1, 6})
				testbraille.MustSetPixel(bc, image.Point{2, 6})
				testbraille.MustSetPixel(bc, image.Point{3, 6})
				testbraille.MustSetPixel(bc, image.Point{4, 6})
				testbraille.MustSetPixel(bc, image.Point{5, 6})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "filled circle with radius of four",
			canvas: image.Rect(0, 0, 4, 2),
			mid:    image.Point{3, 3},
			radius: 4,
			opts: []BrailleCircleOption{
				BrailleCircleFilled(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{2, 0})
				testbraille.MustSetPixel(bc, image.Point{3, 0})
				testbraille.MustSetPixel(bc, image.Point{4, 0})
				testbraille.MustSetPixel(bc, image.Point{5, 0})

				mustBrailleLine(bc, image.Point{0, 1}, image.Point{6, 1})
				mustBrailleLine(bc, image.Point{0, 2}, image.Point{6, 2})
				mustBrailleLine(bc, image.Point{0, 3}, image.Point{6, 3})
				mustBrailleLine(bc, image.Point{0, 4}, image.Point{6, 4})
				mustBrailleLine(bc, image.Point{0, 5}, image.Point{6, 5})

				testbraille.MustSetPixel(bc, image.Point{1, 6})
				testbraille.MustSetPixel(bc, image.Point{2, 6})
				testbraille.MustSetPixel(bc, image.Point{3, 6})
				testbraille.MustSetPixel(bc, image.Point{4, 6})
				testbraille.MustSetPixel(bc, image.Point{5, 6})

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
