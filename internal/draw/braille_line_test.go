// Copyright 2018 Google Inc.
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

	"github.com/mum4k/termdash/internal/area"
	"github.com/mum4k/termdash/internal/canvas/braille"
	"github.com/mum4k/termdash/internal/canvas/braille/testbraille"
	"github.com/mum4k/termdash/internal/cell"
	"github.com/mum4k/termdash/internal/faketerm"
)

func TestBrailleLine(t *testing.T) {
	tests := []struct {
		desc   string
		canvas image.Rectangle
		start  image.Point
		end    image.Point

		// If not nil, called to prepare the braille canvas before running the test.
		prepare func(*braille.Canvas) error

		opts    []BrailleLineOption
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc:    "fails when start has negative X",
			canvas:  image.Rect(0, 0, 1, 1),
			start:   image.Point{-1, 0},
			end:     image.Point{0, 0},
			wantErr: true,
		},
		{
			desc:    "fails when start has negative Y",
			canvas:  image.Rect(0, 0, 1, 1),
			start:   image.Point{0, -1},
			end:     image.Point{0, 0},
			wantErr: true,
		},
		{
			desc:    "fails when end has negative X",
			canvas:  image.Rect(0, 0, 1, 1),
			start:   image.Point{0, 0},
			end:     image.Point{-1, 0},
			wantErr: true,
		},
		{
			desc:    "fails when end has negative Y",
			canvas:  image.Rect(0, 0, 1, 1),
			start:   image.Point{0, 0},
			end:     image.Point{0, -1},
			wantErr: true,
		},
		{
			desc:    "high line, fails on start point outside of the canvas",
			canvas:  image.Rect(0, 0, 1, 1),
			start:   image.Point{2, 2},
			end:     image.Point{2, 2},
			wantErr: true,
		},
		{
			desc:    "low line, fails on end point outside of the canvas",
			canvas:  image.Rect(0, 0, 3, 1),
			start:   image.Point{0, 0},
			end:     image.Point{6, 3},
			wantErr: true,
		},
		{
			desc:   "low line, fails on end point outside of the canvas when clearing pixels",
			canvas: image.Rect(0, 0, 3, 1),
			start:  image.Point{0, 0},
			end:    image.Point{6, 3},
			opts: []BrailleLineOption{
				BrailleLineClearPixels(),
			},
			wantErr: true,
		},
		{
			desc:    "high line, fails on end point outside of the canvas",
			canvas:  image.Rect(0, 0, 1, 1),
			start:   image.Point{0, 0},
			end:     image.Point{2, 2},
			wantErr: true,
		},
		{
			desc:   "draws single point",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{0, 0},
			end:    image.Point{0, 0},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())
				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "clears a single point",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{0, 0},
			end:    image.Point{0, 0},
			prepare: func(bc *braille.Canvas) error {
				return bc.SetPixel(image.Point{0, 0})
			},
			opts: []BrailleLineOption{
				BrailleLineClearPixels(),
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
			desc:   "draws single point with cell options",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{0, 0},
			end:    image.Point{0, 0},
			opts: []BrailleLineOption{
				BrailleLineCellOpts(
					cell.FgColor(cell.ColorRed),
				),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())
				testbraille.MustSetPixel(bc, image.Point{0, 0}, cell.FgColor(cell.ColorRed))
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "clears a single point with cell options",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{0, 0},
			end:    image.Point{0, 0},
			prepare: func(bc *braille.Canvas) error {
				return bc.SetPixel(image.Point{0, 0}, cell.FgColor(cell.ColorBlue))
			},
			opts: []BrailleLineOption{
				BrailleLineClearPixels(),
				BrailleLineCellOpts(
					cell.FgColor(cell.ColorRed),
				),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())
				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustClearPixel(bc, image.Point{0, 0}, cell.FgColor(cell.ColorRed))
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draws high line, octant SE",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{0, 0},
			end:    image.Point{1, 3},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{1, 2})
				testbraille.MustSetPixel(bc, image.Point{1, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "clears a high line, octant SE",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{0, 0},
			end:    image.Point{1, 3},
			prepare: func(bc *braille.Canvas) error {
				return BrailleLine(bc, image.Point{0, 0}, image.Point{1, 3})
			},
			opts: []BrailleLineOption{
				BrailleLineClearPixels(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{1, 2})
				testbraille.MustSetPixel(bc, image.Point{1, 3})
				testbraille.MustClearPixel(bc, image.Point{0, 0})
				testbraille.MustClearPixel(bc, image.Point{0, 1})
				testbraille.MustClearPixel(bc, image.Point{1, 2})
				testbraille.MustClearPixel(bc, image.Point{1, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draws high line, octant NW",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{1, 3},
			end:    image.Point{0, 0},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustSetPixel(bc, image.Point{0, 1})
				testbraille.MustSetPixel(bc, image.Point{1, 2})
				testbraille.MustSetPixel(bc, image.Point{1, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draws high line, octant SW",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{1, 0},
			end:    image.Point{0, 3},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{1, 1})
				testbraille.MustSetPixel(bc, image.Point{0, 2})
				testbraille.MustSetPixel(bc, image.Point{0, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draws high line, octant NE",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{0, 3},
			end:    image.Point{1, 0},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{1, 0})
				testbraille.MustSetPixel(bc, image.Point{1, 1})
				testbraille.MustSetPixel(bc, image.Point{0, 2})
				testbraille.MustSetPixel(bc, image.Point{0, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draws low line, octant SE",
			canvas: image.Rect(0, 0, 3, 1),
			start:  image.Point{0, 0},
			end:    image.Point{4, 3},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustSetPixel(bc, image.Point{1, 1})
				testbraille.MustSetPixel(bc, image.Point{2, 1})
				testbraille.MustSetPixel(bc, image.Point{3, 2})
				testbraille.MustSetPixel(bc, image.Point{4, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draws low line, octant NW",
			canvas: image.Rect(0, 0, 3, 1),
			start:  image.Point{4, 3},
			end:    image.Point{0, 0},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustSetPixel(bc, image.Point{1, 1})
				testbraille.MustSetPixel(bc, image.Point{2, 1})
				testbraille.MustSetPixel(bc, image.Point{3, 2})
				testbraille.MustSetPixel(bc, image.Point{4, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draws high line, octant SW",
			canvas: image.Rect(0, 0, 3, 1),
			start:  image.Point{4, 0},
			end:    image.Point{0, 3},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{4, 0})
				testbraille.MustSetPixel(bc, image.Point{3, 1})
				testbraille.MustSetPixel(bc, image.Point{2, 2})
				testbraille.MustSetPixel(bc, image.Point{1, 2})
				testbraille.MustSetPixel(bc, image.Point{0, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draws high line, octant NE",
			canvas: image.Rect(0, 0, 3, 1),
			start:  image.Point{0, 3},
			end:    image.Point{4, 0},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{4, 0})
				testbraille.MustSetPixel(bc, image.Point{3, 1})
				testbraille.MustSetPixel(bc, image.Point{2, 2})
				testbraille.MustSetPixel(bc, image.Point{1, 2})
				testbraille.MustSetPixel(bc, image.Point{0, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draws horizontal line, octant E",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{0, 0},
			end:    image.Point{1, 0},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustSetPixel(bc, image.Point{1, 0})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draws horizontal line, octant W",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{1, 0},
			end:    image.Point{0, 0},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustSetPixel(bc, image.Point{1, 0})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draws vertical line, octant S",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{0, 0},
			end:    image.Point{0, 1},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustSetPixel(bc, image.Point{0, 1})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draws vertical line, octant N",
			canvas: image.Rect(0, 0, 1, 1),
			start:  image.Point{0, 1},
			end:    image.Point{0, 0},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustSetPixel(bc, image.Point{0, 1})

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

			err = BrailleLine(bc, tc.start, tc.end, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("BrailleLine => unexpected error: %v, wantErr: %v", err, tc.wantErr)
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
				t.Fatalf("BrailleLine => %v", diff)
			}

		})
	}
}
