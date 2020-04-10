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

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/area"
	"github.com/mum4k/termdash/private/canvas/braille"
	"github.com/mum4k/termdash/private/canvas/braille/testbraille"
	"github.com/mum4k/termdash/private/faketerm"
)

func TestBrailleFill(t *testing.T) {
	tests := []struct {
		desc   string
		canvas image.Rectangle
		start  image.Point
		border []image.Point

		// If not nil, called to prepare the braille canvas before running the test.
		prepare func(*braille.Canvas) error

		opts    []BrailleFillOption
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc:    "fails when start isn't in the canvas",
			canvas:  image.Rect(0, 0, 1, 1),
			start:   image.Point{-1, 0},
			wantErr: true,
		},
		{
			desc:   "fills the full canvas without a border",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{1, 1},
			border: nil,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{1, 1})
				testbraille.MustSetPixel(bc, image.Point{0, 2})
				testbraille.MustSetPixel(bc, image.Point{1, 2})
				testbraille.MustSetPixel(bc, image.Point{0, 3})
				testbraille.MustSetPixel(bc, image.Point{1, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "fills the full canvas and sets cell options",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{1, 1},
			border: nil,
			opts: []BrailleFillOption{
				BrailleFillCellOpts(cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				opts := []cell.Option{
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorBlue),
				}
				testbraille.MustSetPixel(bc, image.Point{0, 0}, opts...)
				testbraille.MustSetPixel(bc, image.Point{1, 0}, opts...)
				testbraille.MustSetPixel(bc, image.Point{0, 1}, opts...)
				testbraille.MustSetPixel(bc, image.Point{1, 1}, opts...)
				testbraille.MustSetPixel(bc, image.Point{0, 2}, opts...)
				testbraille.MustSetPixel(bc, image.Point{1, 2}, opts...)
				testbraille.MustSetPixel(bc, image.Point{0, 3}, opts...)
				testbraille.MustSetPixel(bc, image.Point{1, 3}, opts...)

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "clears pixels instead of setting them",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{1, 1},
			border: nil,
			opts: []BrailleFillOption{
				BrailleFillClearPixels(),
			},
			prepare: func(bc *braille.Canvas) error {
				// Set some pixels, see if they get cleared.
				if err := bc.SetPixel(image.Point{1, 0}); err != nil {
					return err
				}
				return bc.SetPixel(image.Point{1, 0})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustClearPixel(bc, image.Point{0, 0})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "clears pixels and sets cell options",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{1, 1},
			border: nil,
			opts: []BrailleFillOption{
				BrailleFillCellOpts(cell.FgColor(cell.ColorRed)),
				BrailleFillClearPixels(),
			},
			prepare: func(bc *braille.Canvas) error {
				// Set some pixels, see if they get cleared.
				if err := bc.SetPixel(image.Point{1, 0}); err != nil {
					return err
				}
				return bc.SetPixel(image.Point{1, 0})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				opts := []cell.Option{
					cell.FgColor(cell.ColorRed),
				}
				testbraille.MustSetPixel(bc, image.Point{0, 0}, opts...)
				testbraille.MustClearPixel(bc, image.Point{0, 0}, opts...)

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "avoids the border",
			canvas: image.Rect(0, 0, 3, 3),
			start:  image.Point{1, 1},
			border: []image.Point{
				{1, 3},
				{2, 3},
				{3, 3},
				{4, 3},

				{1, 4},
				{1, 5},
				{1, 6},
				{1, 7},
				{1, 8},

				{4, 4},
				{4, 5},
				{4, 6},
				{4, 7},
				{4, 8},

				{1, 9},
				{2, 9},
				{3, 9},
				{4, 9},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 11})
				mustBrailleLine(bc, image.Point{0, 0}, image.Point{5, 0})
				mustBrailleLine(bc, image.Point{0, 1}, image.Point{5, 1})
				mustBrailleLine(bc, image.Point{0, 2}, image.Point{5, 2})
				mustBrailleLine(bc, image.Point{0, 10}, image.Point{5, 10})
				mustBrailleLine(bc, image.Point{0, 11}, image.Point{5, 11})
				mustBrailleLine(bc, image.Point{5, 0}, image.Point{5, 11})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "fills outside of a circle",
			canvas: image.Rect(0, 0, 4, 2),
			start:  image.Point{0, 0},
			border: circlePoints(image.Point{4, 4}, 2),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mustBrailleLine(bc, image.Point{0, 0}, image.Point{7, 0})
				mustBrailleLine(bc, image.Point{0, 1}, image.Point{7, 1})

				mustBrailleLine(bc, image.Point{0, 2}, image.Point{2, 2})
				mustBrailleLine(bc, image.Point{6, 2}, image.Point{7, 2})

				mustBrailleLine(bc, image.Point{0, 3}, image.Point{1, 3})
				mustBrailleLine(bc, image.Point{7, 3}, image.Point{7, 3})
				mustBrailleLine(bc, image.Point{0, 4}, image.Point{1, 4})
				mustBrailleLine(bc, image.Point{7, 4}, image.Point{7, 4})
				mustBrailleLine(bc, image.Point{0, 5}, image.Point{1, 5})
				mustBrailleLine(bc, image.Point{7, 5}, image.Point{7, 5})

				mustBrailleLine(bc, image.Point{0, 6}, image.Point{2, 6})
				mustBrailleLine(bc, image.Point{6, 6}, image.Point{7, 6})

				mustBrailleLine(bc, image.Point{0, 7}, image.Point{7, 7})

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

			err = BrailleFill(bc, tc.start, tc.border, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("BrailleFill => unexpected error: %v, wantErr: %v", err, tc.wantErr)
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
				t.Fatalf("BrailleFill => %v", diff)
			}
		})
	}
}
